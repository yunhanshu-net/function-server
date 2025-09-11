package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yunhanshu-net/pkg/query"

	resp "github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/dto/base"
	"github.com/yunhanshu-net/function-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/x/jsonx"
	"gorm.io/gorm"
)

type FunctionExecuteCase struct {
	db         *gorm.DB
	runner     *Runner
	runnerFunc *RunnerFunc
}

func NewFunctionExecuteCase(db *gorm.DB) *FunctionExecuteCase {
	return &FunctionExecuteCase{
		db:         db,
		runner:     NewRunner(db),
		runnerFunc: NewRunnerFunc(db),
	}
}

func (r *FunctionExecuteCase) Create(ctx context.Context, executeCase *model.FunctionExecuteCase) error {
	function, err := r.runnerFunc.Get(ctx, int64(executeCase.FunctionId))
	if err != nil {
		return err
	}
	executeCase.DefineRequest = function.Request

	// 使用事务确保数据一致性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建执行用例
		if err := tx.Create(executeCase).Error; err != nil {
			return err
		}

		// 重新统计并更新RunnerFunc的exec_case_count字段
		var count int64
		if err := tx.Model(&model.FunctionExecuteCase{}).Where("function_id = ?", executeCase.FunctionId).Count(&count).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.RunnerFunc{}).Where("id = ?", executeCase.FunctionId).
			Update("exec_case_count", count).Error; err != nil {
			return err
		}

		logger.Infof(ctx, "创建执行用例成功，用例ID: %d，函数ID: %d，当前用例数量: %d", executeCase.ID, executeCase.FunctionId, count)
		return nil
	})
}

// 执行用例
func (r *FunctionExecuteCase) Exec(ctx context.Context, req dto.FunctionExecuteCaseReq) (record *model.FunctionExecuteCaseRecord, rsp *resp.RunFunctionResp, err error) {
	var execCase model.FunctionExecuteCase
	var runner model.Runner
	err = r.db.Where("id = ?", req.CaseId).First(&execCase).Error
	if err != nil {
		return nil, nil, err
	}
	err = r.db.Where("id = ?", execCase.RunnerId).First(&runner).Error
	if err != nil {
		return nil, nil, err
	}
	execCase.CreatedBy = contextx.GetRequestUserName(ctx)
	execCase.UpdatedBy = contextx.GetRequestUserName(ctx)
	funcInfo, err := r.runnerFunc.Get(ctx, int64(execCase.FunctionId))
	if err != nil {
		return nil, nil, err
	}

	eq := jsonx.EQRawMessage(funcInfo.Request, execCase.DefineRequest)
	if !eq {
		return nil, nil, fmt.Errorf("invalid request，发现函数参数已经变更，请及时更新用例")
	}

	doReq := &runcher.RunFunctionReq{
		//TraceID: contextx.GetTraceID(ctx),
		User:    funcInfo.User,
		Method:  funcInfo.Method,
		Router:  funcInfo.Router,
		Runner:  runner.Name,
		Body:    string(execCase.Request),
		Version: runner.Version,
	}
	if execCase.CanBackground && req.Background {
		go func() {
			_, _, err := r.exec(ctx, &execCase, req, doReq)
			if err != nil {
				logger.Errorf(ctx, "execute background request failed: %v req:%+v doReq：%+v", err, req, doReq)
				return
			}
		}()
		return nil, nil, nil
	} else {
		record, rsp, err = r.exec(ctx, &execCase, req, doReq)
		if err != nil {
			logger.Errorf(ctx, "execute request failed: %v req:%+v doReq：%+v", err, req, doReq)
			return nil, nil, err
		}
		return record, rsp, nil
	}
}

func (r *FunctionExecuteCase) exec(ctx context.Context, execCase *model.FunctionExecuteCase, caseReq dto.FunctionExecuteCaseReq, req *runcher.RunFunctionReq) (record *model.FunctionExecuteCaseRecord, rsp *resp.RunFunctionResp, err error) {

	now := time.Now()
	record = &model.FunctionExecuteCaseRecord{
		Base: model.Base{
			CreatedBy: contextx.GetRequestUserName(ctx),
			UpdatedBy: contextx.GetRequestUserName(ctx),
		},
		CaseId:     int(execCase.ID),
		RunnerId:   execCase.RunnerId,
		FunctionId: execCase.FunctionId,
		Status:     "running",
		Message:    "",
		Request:    execCase.Request,
		Background: caseReq.Background,
		Remark:     caseReq.Remark,
		StartTime:  timeToModelTime(time.Now()),
		TraceId:    contextx.GetTraceID(ctx),
	}
	err = r.db.Create(record).Error
	if err != nil {
		logger.Errorf(ctx, "failed to create record: %+v", err)
	}

	run, err := r.runnerFunc.Run(ctx, req)
	record.Status = "success"
	if err != nil {
		record.ErrorDetails = err.Error()
		record.Status = "sys_failed"
	} else {
		if run.Code != 0 || run.Msg != "ok" {
			record.ErrorDetails = run.Msg
			record.Status = "biz_failed"
		}
	}
	end := timeToModelTime(time.Now())
	record.EndTime = &end
	record.CostMillis = time.Since(now).Milliseconds()
	record.Response = json.RawMessage(jsonx.String(run))
	go func() {
		r.db.Where("id = ?", record.ID).Updates(record)
		r.db.Model(&model.FunctionExecuteCase{}).Where("id = ?", caseReq.CaseId).Updates(map[string]interface{}{
			"last_used_at": timeToModelTime(time.Now()),
			"exec_count":   gorm.Expr("exec_count + 1"),
		})
	}()
	return record, run, nil
}

// BatchDelete 批量删除执行用例
func (r *FunctionExecuteCase) BatchDelete(ctx context.Context, req dto.BatchDeleteFunctionExecuteCaseReq) error {
	if len(req.CaseIds) == 0 {
		return fmt.Errorf("请选择要删除的用例")
	}

	// 检查用例是否存在
	var count int64
	err := r.db.Model(&model.FunctionExecuteCase{}).Where("id IN ?", req.CaseIds).Count(&count).Error
	if err != nil {
		return err
	}
	if count != int64(len(req.CaseIds)) {
		return fmt.Errorf("部分用例不存在，请检查后重试")
	}

	// 使用事务确保数据一致性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先查询要删除的用例，获取对应的FunctionId
		var cases []model.FunctionExecuteCase
		if err := tx.Where("id IN ?", req.CaseIds).Find(&cases).Error; err != nil {
			return err
		}

		// 收集所有涉及的FunctionId
		functionIds := make(map[int]bool)
		for _, caseItem := range cases {
			functionIds[caseItem.FunctionId] = true
		}

		// 删除执行记录
		err := tx.Where("case_id IN ?", req.CaseIds).Delete(&model.FunctionExecuteCaseRecord{}).Error
		if err != nil {
			return err
		}

		// 删除执行用例
		err = tx.Where("id IN ?", req.CaseIds).Delete(&model.FunctionExecuteCase{}).Error
		if err != nil {
			return err
		}

		// 重新统计并更新每个Function的exec_case_count字段
		for functionId := range functionIds {
			var count int64
			if err := tx.Model(&model.FunctionExecuteCase{}).Where("function_id = ?", functionId).Count(&count).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.RunnerFunc{}).Where("id = ?", functionId).
				Update("exec_case_count", count).Error; err != nil {
				return err
			}
		}

		logger.Infof(ctx, "批量删除执行用例成功，删除数量: %d", len(req.CaseIds))
		return nil
	})
}

// Update 更新执行用例
func (r *FunctionExecuteCase) Update(ctx context.Context, req dto.UpdateFunctionExecuteCaseReq) error {
	// 检查用例是否存在
	var execCase model.FunctionExecuteCase
	err := r.db.Where("id = ?", req.CaseId).First(&execCase).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("执行用例不存在")
		}
		return err
	}

	// 如果更新了请求参数，需要重新获取函数信息并验证
	if req.Request != nil {
		funcInfo, err := r.runnerFunc.Get(ctx, int64(execCase.FunctionId))
		if err != nil {
			return err
		}

		// 验证请求参数格式是否与函数定义一致
		eq := jsonx.EQRawMessage(funcInfo.Request, execCase.DefineRequest)
		if !eq {
			return fmt.Errorf("函数参数已变更，请重新创建用例")
		}
	}

	// 构建更新数据
	updateData := map[string]interface{}{
		"updated_by": contextx.GetRequestUserName(ctx),
	}

	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Request != nil {
		updateData["request"] = req.Request
	}
	updateData["can_background"] = req.CanBackground
	updateData["auto_run"] = req.AutoRun
	updateData["use_latest_version"] = req.UseLatestVersion

	// 执行更新
	err = r.db.Model(&execCase).Updates(updateData).Error
	if err != nil {
		return err
	}

	logger.Infof(ctx, "更新执行用例成功，用例ID: %d", req.CaseId)
	return nil
}

// Query 查询执行用例
func (r *FunctionExecuteCase) Query(ctx context.Context, req dto.QueryFunctionExecuteCaseReq) (*base.Paginated[[]*dto.FunctionExecuteCaseWithDetails], error) {

	// 构建基础查询
	db := r.db.Model(&model.FunctionExecuteCase{}).Omit("define_request")

	// 使用AutoPaginate处理分页和查询条件
	var list []*model.FunctionExecuteCase
	result, err := query.AutoPaginateTable(ctx, db, &model.FunctionExecuteCase{}, &list, &req.PageInfoReq)
	if err != nil {
		return nil, err
	}

	// 预加载Runner和Function信息（批量查询优化性能）
	var resultList []*dto.FunctionExecuteCaseWithDetails

	// 收集需要查询的ID
	runnerIds := make(map[int]bool)
	functionIds := make(map[int]bool)
	for _, caseItem := range list {
		if caseItem.RunnerId > 0 {
			runnerIds[caseItem.RunnerId] = true
		}
		if caseItem.FunctionId > 0 {
			functionIds[caseItem.FunctionId] = true
		}
	}

	// 批量查询Runner信息
	runnerMap := make(map[int]*dto.RunnerInfo)
	if len(runnerIds) > 0 {
		var runnerIdList []int
		for id := range runnerIds {
			runnerIdList = append(runnerIdList, id)
		}

		var runners []model.Runner
		err := r.db.Where("id IN ?", runnerIdList).Find(&runners).Error
		if err == nil {
			for _, runner := range runners {
				runnerMap[int(runner.ID)] = &dto.RunnerInfo{
					ID:           uint(runner.ID),
					Name:         runner.Name,
					Title:        runner.Title,
					Version:      runner.Version,
					Language:     runner.Language,
					Status:       runner.Status,
					User:         runner.User,
					FullNamePath: runner.FullNamePath,
				}
			}
		}
	}

	// 批量查询Function信息
	functionMap := make(map[int]*dto.FunctionInfo)
	if len(functionIds) > 0 {
		var functionIdList []int
		for id := range functionIds {
			functionIdList = append(functionIdList, id)
		}

		var functions []model.RunnerFunc
		err := r.db.Where("id IN ?", functionIdList).Find(&functions).Error
		if err == nil {
			for _, function := range functions {
				functionMap[int(function.ID)] = &dto.FunctionInfo{
					ID:          uint(function.ID),
					Name:        function.Name,
					Title:       function.Title,
					Description: function.Description,
					Method:      function.Method,
					Router:      function.GetRouter(),
					Tags:        function.Tags,
					User:        function.User,
					HasConfig:   function.HasConfig,
				}
			}
		}
	}

	// 组装结果
	for _, caseItem := range list {
		item := &dto.FunctionExecuteCaseWithDetails{
			FunctionExecuteCase: caseItem,
		}

		// 设置预加载的Runner信息
		if caseItem.RunnerId > 0 {
			if runner, exists := runnerMap[caseItem.RunnerId]; exists {
				item.Runner = runner
			}
		}

		// 设置预加载的Function信息
		if caseItem.FunctionId > 0 {
			if function, exists := functionMap[caseItem.FunctionId]; exists {
				item.Function = function
			}
		}

		resultList = append(resultList, item)
	}
	for _, resultItem := range resultList {
		resultItem.UrlPath = resultItem.Runner.User + "/" + resultItem.Runner.Name + "/" + strings.Trim(resultItem.Function.Router, "/")
	}

	return &base.Paginated[[]*dto.FunctionExecuteCaseWithDetails]{
		Items:       resultList,
		CurrentPage: result.CurrentPage,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		PageSize:    result.PageSize,
	}, nil
}

// QueryRecords 查询执行记录
func (r *FunctionExecuteCase) QueryRecords(ctx context.Context, req dto.QueryFunctionExecuteCaseRecordReq) (*base.Paginated[[]dto.FunctionExecuteCaseRecordWithDetails], error) {
	// 创建查询配置，限制可查询的字段
	//config := base.NewQueryConfig()
	//config.AllowField("case_id", "eq", "in")
	//config.AllowField("runner_id", "eq", "in")
	//config.AllowField("function_id", "eq", "in")
	//config.AllowField("status", "eq", "in")
	//config.AllowField("background", "eq")
	//config.AllowField("trace_id", "eq", "like")
	//config.AllowField("start_time", "gt", "gte", "lt", "lte")
	//config.AllowField("end_time", "gt", "gte", "lt", "lte")
	//config.AllowField("created_at", "gt", "gte", "lt", "lte")

	// 构建基础查询
	db := r.db.Model(&model.FunctionExecuteCaseRecord{})

	// 使用AutoPaginate处理分页和查询条件
	var list []model.FunctionExecuteCaseRecord
	result, err := query.AutoPaginateTable(ctx, db, &model.FunctionExecuteCaseRecord{}, &list, &req.PageInfoReq)
	if err != nil {
		return nil, err
	}

	// 预加载Runner和Function信息（批量查询优化性能）
	var resultList []dto.FunctionExecuteCaseRecordWithDetails

	// 收集需要查询的ID
	runnerIds := make(map[int]bool)
	functionIds := make(map[int]bool)
	for _, recordItem := range list {
		if recordItem.RunnerId > 0 {
			runnerIds[recordItem.RunnerId] = true
		}
		if recordItem.FunctionId > 0 {
			functionIds[recordItem.FunctionId] = true
		}
	}

	// 批量查询Runner信息
	runnerMap := make(map[int]*dto.RunnerInfo)
	if len(runnerIds) > 0 {
		var runnerIdList []int
		for id := range runnerIds {
			runnerIdList = append(runnerIdList, id)
		}

		var runners []model.Runner
		err := r.db.Where("id IN ?", runnerIdList).Find(&runners).Error
		if err == nil {
			for _, runner := range runners {
				runnerMap[int(runner.ID)] = &dto.RunnerInfo{
					ID:           uint(runner.ID),
					Name:         runner.Name,
					Title:        runner.Title,
					Version:      runner.Version,
					Language:     runner.Language,
					Status:       runner.Status,
					User:         runner.User,
					FullNamePath: runner.FullNamePath,
				}
			}
		}
	}

	// 批量查询Function信息
	functionMap := make(map[int]*dto.FunctionInfo)
	if len(functionIds) > 0 {
		var functionIdList []int
		for id := range functionIds {
			functionIdList = append(functionIdList, id)
		}

		var functions []model.RunnerFunc
		err := r.db.Where("id IN ?", functionIdList).Find(&functions).Error
		if err == nil {
			for _, function := range functions {
				functionMap[int(function.ID)] = &dto.FunctionInfo{
					ID:          uint(function.ID),
					Name:        function.Name,
					Title:       function.Title,
					Description: function.Description,
					Method:      function.Method,
					Router:      function.GetRouter(),
					Tags:        function.Tags,
					User:        function.User,
					HasConfig:   function.HasConfig,
				}
			}
		}
	}

	// 组装结果
	for _, recordItem := range list {
		item := dto.FunctionExecuteCaseRecordWithDetails{
			FunctionExecuteCaseRecord: recordItem,
		}

		// 设置预加载的Runner信息
		if recordItem.RunnerId > 0 {
			if runner, exists := runnerMap[recordItem.RunnerId]; exists {
				item.Runner = runner
			}
		}

		// 设置预加载的Function信息
		if recordItem.FunctionId > 0 {
			if function, exists := functionMap[recordItem.FunctionId]; exists {
				item.Function = function
			}
		}

		resultList = append(resultList, item)
	}

	return &base.Paginated[[]dto.FunctionExecuteCaseRecordWithDetails]{
		Items:       resultList,
		CurrentPage: result.CurrentPage,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		PageSize:    result.PageSize,
	}, nil
}

// Get 根据ID获取执行用例详情
func (r *FunctionExecuteCase) Get(ctx context.Context, caseId int) (*dto.FunctionExecuteCaseWithDetails, error) {
	// 查询执行用例
	var execCase model.FunctionExecuteCase
	err := r.db.Omit("define_request").Where("id = ?", caseId).First(&execCase).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("执行用例不存在")
		}
		return nil, err
	}

	// 查询Runner信息
	var runner model.Runner
	if execCase.RunnerId > 0 {
		err = r.db.Where("id = ?", execCase.RunnerId).First(&runner).Error
		if err != nil {
			logger.Warnf(ctx, "查询Runner信息失败，RunnerId: %d, 错误: %v", execCase.RunnerId, err)
		}
	}

	// 查询Function信息
	var function model.RunnerFunc
	if execCase.FunctionId > 0 {
		err = r.db.Where("id = ?", execCase.FunctionId).First(&function).Error
		if err != nil {
			logger.Warnf(ctx, "查询Function信息失败，FunctionId: %d, 错误: %v", execCase.FunctionId, err)
		}
	}

	// 组装结果
	result := &dto.FunctionExecuteCaseWithDetails{
		FunctionExecuteCase: &execCase,
	}

	// 设置Runner信息
	if runner.ID > 0 {
		result.Runner = &dto.RunnerInfo{
			ID:           uint(runner.ID),
			Name:         runner.Name,
			Title:        runner.Title,
			Version:      runner.Version,
			Language:     runner.Language,
			Status:       runner.Status,
			User:         runner.User,
			FullNamePath: runner.FullNamePath,
		}
	}

	// 设置Function信息
	if function.ID > 0 {
		result.Function = &dto.FunctionInfo{
			ID:          uint(function.ID),
			Name:        function.Name,
			Title:       function.Title,
			Description: function.Description,
			Method:      function.Method,
			Router:      function.GetRouter(),
			Tags:        function.Tags,
			User:        function.User,
			HasConfig:   function.HasConfig,
		}
	}

	// 生成URL路径
	if result.Runner != nil && result.Function != nil {
		result.UrlPath = result.Runner.User + "/" + result.Runner.Name + "/" + strings.Trim(result.Function.Router, "/")
	}

	return result, nil
}

// InitExecCaseCount 初始化所有函数的exec_case_count字段
func (r *FunctionExecuteCase) InitExecCaseCount(ctx context.Context) error {
	logger.Infof(ctx, "开始初始化所有函数的exec_case_count字段")

	// 获取所有函数
	var functions []model.RunnerFunc
	if err := r.db.Find(&functions).Error; err != nil {
		return err
	}

	// 为每个函数统计并更新exec_case_count
	for _, function := range functions {
		var count int64
		if err := r.db.Model(&model.FunctionExecuteCase{}).Where("function_id = ?", function.ID).Count(&count).Error; err != nil {
			logger.Errorf(ctx, "统计函数%d的用例数量失败: %v", function.ID, err)
			continue
		}

		if err := r.db.Model(&model.RunnerFunc{}).Where("id = ?", function.ID).Update("exec_case_count", count).Error; err != nil {
			logger.Errorf(ctx, "更新函数%d的exec_case_count失败: %v", function.ID, err)
			continue
		}

		logger.Debugf(ctx, "函数%d的exec_case_count已更新为: %d", function.ID, count)
	}

	logger.Infof(ctx, "完成初始化所有函数的exec_case_count字段，共处理%d个函数", len(functions))
	return nil
}
