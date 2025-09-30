package service

import (
	"context"
	"encoding/json"
	"fmt"
	resp "github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/pkg/x/urlx"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-go/pkg/dto/api"
	"github.com/yunhanshu-net/function-go/pkg/dto/usercall"
	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/x/jsonx"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/repo"
	"github.com/yunhanshu-net/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunnerFunc 函数服务实现
type RunnerFunc struct {
	runnerFuncRepo  *repo.RunnerFuncRepo
	runnerRepo      *repo.RunnerRepo
	serviceTreeRepo *repo.ServiceTreeRepo
	serviceTree     *ServiceTree
	funcConfigRepo  *repo.FuncConfigRepo
}

// NewRunnerFunc 创建函数服务
func NewRunnerFunc(db *gorm.DB) *RunnerFunc {
	svc := &RunnerFunc{
		runnerFuncRepo:  repo.NewRunnerFuncRepo(db),
		serviceTreeRepo: repo.NewServiceTreeRepo(db),
		serviceTree:     NewServiceTree(db),
		runnerRepo:      repo.NewRunnerRepo(db),
		funcConfigRepo:  repo.NewFuncConfigRepo(db),
	}
	return svc
}

func (s *RunnerFunc) Run(ctx context.Context, req *runcher.RunFunctionReq) (*resp.RunFunctionResp, error) {

	//req := &runcher.RunFunctionReq{
	//	User:   c.Param("user"),
	//	Method: c.Request.Method,
	//	Runner: c.Param("runner"),
	//	Router: c.Param("router"),
	//}

	traceId := contextx.GetTraceID(ctx)
	if traceId == "" {
		return nil, fmt.Errorf("traceId is empty")
	}

	log := model.FuncRunRecord{
		Base: model.Base{
			CreatedBy: contextx.GetRequestUserName(ctx),
			UpdatedBy: contextx.GetRequestUserName(ctx),
		},
		FuncId:  1,
		Status:  "running",
		TraceId: traceId,
		Request: json.RawMessage(req.Body),
		StartTs: time.Now().UnixMilli(),
	}
	funcId := contextx.GetFunctionID(ctx)

	log.FuncId = int64(funcId)
	if req.Method == http.MethodGet {
		//req.RawQuery = c.Request.URL.RawQuery
		toMap := urlx.QueryToMap(req.RawQuery)
		marshal, err := json.Marshal(toMap)
		if err != nil {
			//response.ParamError(c, fmt.Sprintf("json.Marshal 失败：%s", err.Error()))
			return nil, err
		}
		log.Request = marshal
	} else {

		//b, err := io.ReadAll(c.Request.Body)
		//if err != nil {
		//	panic(err)
		//}
		//defer c.Request.Body.Close()
		//req.Body = string(b)
		log.Request = json.RawMessage(req.Body)
	}
	rn, err := s.runnerRepo.GetByUserAndName(ctx, req.User, req.Runner)
	if err != nil {
		//response.ParamError(c, fmt.Sprintf("获取runner失败：%s", err.Error()))
		return nil, err
	}
	req.Version = rn.Version
	db.GetDB().Model(&model.FuncRunRecord{}).Create(&log)

	function2, err := GetRuncherService().RunFunction3(ctx, req)
	if err != nil {
		//response.ServerError(c, err.Error())
		return nil, err
	}

	update := model.FuncRunRecord{}
	update.EndTs = time.Now().UnixMilli()
	update.Cost = log.EndTs - log.StartTs

	update.Response = function2.Data
	var res resp.RunFunctionResp
	err = json.Unmarshal(function2.Data, &res)
	if err != nil {
		update.Status = "fail"
		update.Message = err.Error()
		//response.ServerError(c, err.Error())
		return &res, err
	}
	if res.MetaData == nil {
		res.MetaData = make(map[string]interface{})
	}
	for k, v := range function2.Header {
		if k != "code" {
			if len(v) > 0 {
				res.MetaData[k] = v[0]
			}
		}
	}
	res.MetaData["version"] = rn.Version
	update.Status = "success"
	go func() {
		marshal, err2 := json.Marshal(res)
		if err2 != nil {
			logger.Errorf(ctx, err2.Error())
		} else {
			update.Response = marshal
		}
		db.GetDB().Model(&model.FuncRunRecord{}).Where("trace_id = ?", traceId).Updates(update)
	}()
	return &res, nil
}

// Create 创建函数
func (s *RunnerFunc) Create(ctx context.Context, runnerFunc *model.RunnerFunc) error {
	// 业务逻辑校验
	if runnerFunc.Name == "" {
		return errors.New("函数名称不能为空")
	}
	if runnerFunc.Title == "" {
		return errors.New("函数标题不能为空")
	}

	if runnerFunc.Code == "" {
		return errors.New("code 不能为空")
	}

	t, err := s.serviceTreeRepo.Get(ctx, runnerFunc.TreeID)
	if err != nil {
		return err
	}
	runnerFunc.RunnerID = t.RunnerID

	// 检查Runner是否存在

	gotRunner, err := s.runnerRepo.Get(ctx, runnerFunc.RunnerID)
	if err != nil {
		return err
	}
	if gotRunner == nil {
		return errors.New("关联的Runner不存在")
	}

	packageTree, err := s.serviceTreeRepo.Get(ctx, runnerFunc.TreeID)
	if err != nil {
		logger.Error(ctx, "检查服务树存在性失败", err, zap.Int64("tree_id", runnerFunc.TreeID))
		return fmt.Errorf("检查服务树存在性失败: %w", err)
	}
	if packageTree == nil {
		return errors.New("关联的服务树不存在")
	}

	// 检查名称是否已存在
	existingFunc, err := s.runnerFuncRepo.GetByName(ctx, runnerFunc.RunnerID, runnerFunc.Name)
	if err != nil {
		logger.Error(ctx, "检查函数名称失败", err, zap.String("name", runnerFunc.Name))
		return fmt.Errorf("检查函数名称失败: %w", err)
	}
	if existingFunc != nil {
		return errors.New("函数名称已存在")
	}
	//}

	runner, err := runnerproject.NewRunner(gotRunner.User, gotRunner.Name, gotRunner.Version)
	if err != nil {
		return err
	}
	runner.Language = "go"
	service := GetRuncherService()
	r := &coder.AddApisReq{
		Runner: runner,
		Msg:    runnerFunc.Description,
		CodeApis: []*coder.CodeApi{
			{
				EnName:         runnerFunc.Name,
				CnName:         runnerFunc.Title,
				Desc:           runnerFunc.Description,
				Language:       "go",
				Code:           runnerFunc.Code,
				Package:        packageTree.Name,
				AbsPackagePath: packageTree.GetPackagePath(),
			},
		},
	}
	rsp, err := service.AddAPI2(ctx, r)
	if err != nil {
		logger.Error(ctx, "添加api失败", err, zap.Int64("func_id", runnerFunc.ID))
		return err
	}
	logger.Infof(ctx, "rsp:%+v", rsp)

	addAPIs := rsp.ApiChangeInfo.AddApi
	for _, addAPI := range addAPIs {
		// 使用新的通用方法创建函数及其所有依赖项
		err = s.createFunctionWithDependencies(ctx, runnerFunc, addAPI, rsp, false)
		if err != nil {
			return err
		}
	}

	logger.Info(ctx, "创建函数成功", zap.Int64("id", runnerFunc.ID), zap.String("name", runnerFunc.Name))
	return nil
}

// Get 获取函数详情
func (s *RunnerFunc) Get(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	get, err := s.runnerFuncRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	trim := strings.Trim(get.Path, "/")
	split := strings.Split(trim, "/") // a/b/c/d
	router := split[2:]
	runner := split[1:2]
	get.Router = strings.Join(router, "/")
	get.RunnerName = runner[0]
	return get, nil

}
func (s *RunnerFunc) Versions(ctx context.Context, id int64) ([]model.FuncVersion, error) {
	versions, err := s.runnerFuncRepo.GetVersions(ctx, id)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// GetByTreeId GetByTreeId
func (s *RunnerFunc) GetByTreeId(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	get, err := s.serviceTreeRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	logger.Debug(ctx, "开始获取函数详情", zap.Int64("id", id))
	return s.runnerFuncRepo.Get(ctx, get.RefID)
}

func (s *RunnerFunc) GetByFullPath(ctx context.Context, method string, fullPath string) (*model.RunnerFunc, error) {
	runnerFunc, err := s.runnerFuncRepo.GetByFullPath(ctx, method, fullPath)
	if err != nil {
		return nil, err
	}
	return runnerFunc, nil
}

// GetRunnerFuncByID 通过ID获取运行函数（适配接口）
func (s *RunnerFunc) GetRunnerFuncByID(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	return s.Get(ctx, id)
}

// Update 更新函数
func (s *RunnerFunc) Update(ctx context.Context, id int64, updateData *model.RunnerFunc) error {
	logger.Debug(ctx, "开始更新函数", zap.Int64("id", id))

	// 检查函数是否存在
	existingFunc, err := s.runnerFuncRepo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 如果要更新名称，需要检查名称是否已存在
	if updateData.Name != "" && updateData.Name != existingFunc.Name {
		existing, err := s.runnerFuncRepo.GetByName(ctx, existingFunc.RunnerID, updateData.Name)
		if err != nil {
			logger.Error(ctx, "检查函数名称失败", err, zap.String("name", updateData.Name))
			return fmt.Errorf("检查函数名称失败: %w", err)
		}
		if existing != nil && existing.ID != id {
			return errors.New("函数名称已存在")
		}
	}

	// 如果要更新服务树，需要检查服务树是否存在
	if updateData.TreeID > 0 && updateData.TreeID != existingFunc.TreeID {
		treeExists, err := s.runnerFuncRepo.CheckServiceTreeExists(ctx, updateData.TreeID)
		if err != nil {
			logger.Error(ctx, "检查服务树存在性失败", err, zap.Int64("tree_id", updateData.TreeID))
			return fmt.Errorf("检查服务树存在性失败: %w", err)
		}
		if !treeExists {
			return errors.New("关联的服务树不存在")
		}
	}

	// 更新函数
	if err := s.runnerFuncRepo.Update(ctx, id, updateData); err != nil {
		logger.Error(ctx, "更新函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("更新函数失败: %w", err)
	}

	logger.Info(ctx, "更新函数成功", zap.Int64("id", id))
	return nil
}

// Delete 删除函数
func (s *RunnerFunc) Delete(ctx context.Context, id int64, operator string) error {
	logger.Debug(ctx, "开始删除函数", zap.Int64("id", id))

	// 检查函数是否存在
	runnerFunc, err := s.runnerFuncRepo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if runnerFunc == nil {
		return errors.New("函数不存在")
	}
	gotRunner, err := s.runnerRepo.Get(ctx, runnerFunc.RunnerID)
	if err != nil {
		return err
	}
	packageTree, err := s.serviceTreeRepo.Get(ctx, runnerFunc.TreeID)
	if err != nil {
		logger.Error(ctx, "检查服务树存在性失败", err, zap.Int64("tree_id", runnerFunc.TreeID))
		return fmt.Errorf("检查服务树存在性失败: %w", err)
	}

	runner, err := runnerproject.NewRunner(gotRunner.User, gotRunner.Name, gotRunner.Version)
	if err != nil {
		return err
	}
	runner.Language = "go"
	service := GetRuncherService()
	r := &coder.DeleteAPIsReq{
		Runner: runner,
		Msg:    runnerFunc.Description,
		CodeApis: []*coder.CodeApi{
			{
				EnName:         runnerFunc.Name,
				CnName:         runnerFunc.Title,
				Desc:           runnerFunc.Description,
				Language:       "go",
				Code:           runnerFunc.Code,
				Package:        packageTree.Name,
				AbsPackagePath: packageTree.GetPackagePath(),
			},
		},
	}
	_, err = service.DeleteAPIs(ctx, r)
	if err != nil {
		return err
	}
	//删除对应tree和对应函数

	// 设置删除者
	if err := s.runnerFuncRepo.SetDeletedBy(ctx, id, operator); err != nil {
		logger.Error(ctx, "设置函数删除者失败", err, zap.Int64("id", id))
		return fmt.Errorf("设置删除者失败: %w", err)
	}

	// 删除函数
	if err := s.runnerFuncRepo.Delete(ctx, id); err != nil {
		logger.Error(ctx, "删除函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("删除函数失败: %w", err)
	}

	logger.Info(ctx, "删除函数成功", zap.Int64("id", id))
	return nil
}

// DeleteByIds 删除函数
func (s *RunnerFunc) DeleteByIds(ctx context.Context, ids []int64, operator string) error {

	if ids == nil || len(ids) == 0 {
		return fmt.Errorf("ids is empty")
	}
	service := GetRuncherService()

	var gotRunner *model.Runner
	del := &coder.DeleteAPIsReq{}
	var rp *runnerproject.Runner
	var delPaths []string
	for _, id := range ids {
		runnerFunc, err := s.runnerFuncRepo.Get(ctx, id)
		if err != nil {
			logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
			return fmt.Errorf("获取函数失败: %w", err)
		}
		if runnerFunc == nil {
			return errors.New("函数不存在")
		}
		if gotRunner == nil {
			gotRunner, err = s.runnerRepo.Get(ctx, runnerFunc.RunnerID)
			if err != nil {
				return err
			}
		}

		packageTree, err := s.serviceTreeRepo.Get(ctx, runnerFunc.TreeID)
		if err != nil {
			logger.Error(ctx, "检查服务树存在性失败", err, zap.Int64("tree_id", runnerFunc.TreeID))
			return fmt.Errorf("检查服务树存在性失败: %w", err)
		}

		if rp == nil {
			rp, err = runnerproject.NewRunner(gotRunner.User, gotRunner.Name, gotRunner.Version)
			if err != nil {
				return err
			}
			rp.Language = "go"
			del.Runner = rp
		}
		delPaths = append(delPaths, packageTree.FullNamePath+runnerFunc.Name+"/")
		del.CodeApis = append(del.CodeApis, &coder.CodeApi{
			EnName:         runnerFunc.Name,
			CnName:         runnerFunc.Title,
			Desc:           runnerFunc.Description,
			Language:       "go",
			Code:           runnerFunc.Code,
			Package:        packageTree.Name,
			AbsPackagePath: packageTree.GetPackagePath(),
		})

	}
	//// 检查函数是否存在
	//if gotRunner == nil {
	//	return errors.New("runner is nil")
	//}
	//runner, err := runnerproject.NewRunner(gotRunner.User, gotRunner.CnName, gotRunner.Version)
	//if err != nil {
	//	return err
	//}
	if gotRunner == nil {
		return errors.New("gotRunner is nil")
	}

	rsp, err := service.DeleteAPIs(ctx, del)
	if err != nil {
		logger.Errorf(ctx, "DeleteAPIs err:%s", err.Error())
	} else {
		fmt.Println(rsp)
		if len(rsp.DelApis) != len(ids) {
			logger.Warnf(ctx, "删除的api和实际删除数量不符：user：%v rel：%v", len(ids), len(rsp.DelApis))
		}
		db.GetDB().Model(&model.Runner{}).
			Where("id = ?", gotRunner.ID).
			Updates(map[string]interface{}{
				"version": rsp.Version,
			})
		go func() {
			s.runnerRepo.CreateRunnerVersion(ctx, &model.RunnerVersion{
				Base:     model.Base{CreatedBy: gotRunner.CreatedBy, UpdatedBy: gotRunner.UpdatedBy},
				Desc:     gotRunner.Description,
				Log:      rsp.GetDelApisDesc(),
				Version:  rsp.Version,
				RunnerID: gotRunner.ID,
				MetaData: json.RawMessage(jsonx.String(coder.AddApisResp{
					Hash:    rsp.Hash,
					Version: rsp.Version,
					ApiChangeInfo: &coder.ApiChangeInfo{
						CurrentVersion: rsp.Version,
						DelApi:         rsp.DelApis,
					},
				})),
				Hash: rsp.Hash,
			})

		}()
	}

	//删除对应tree和对应函数

	s.deleteFunctions(ctx, ids, delPaths, operator)
	return nil
}

func (s *RunnerFunc) deleteFunctions(ctx context.Context, ids []int64, names []string, operator string) error {
	db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path in ?", names).Updates(map[string]interface{}{
		"deleted_by": operator,
	})
	db.GetDB().Delete(&model.ServiceTree{}, "full_name_path in ?", names)
	db.GetDB().Delete(&model.RunnerFunc{}, "id in ?", ids)
	return nil
}
func (s *RunnerFunc) deleteFunctionsByFullPath(ctx context.Context, names []string, operator string) error {

	var ids []int64

	db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path in ?", names).Pluck("ref_id", &ids)
	db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path in ?", names).Updates(map[string]interface{}{
		"deleted_by": operator,
	})
	db.GetDB().Delete(&model.ServiceTree{}, "full_name_path in ?", names)
	db.GetDB().Delete(&model.RunnerFunc{}, "id in ?", ids)
	return nil
}

// List 获取函数列表
func (s *RunnerFunc) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.RunnerFunc, int64, error) {
	logger.Debug(ctx, "开始获取函数列表", zap.Int("page", page), zap.Int("pageSize", pageSize))
	return s.runnerFuncRepo.List(ctx, page, pageSize, conditions)
}

// GetByRunner 获取Runner下的所有函数
func (s *RunnerFunc) GetByRunner(ctx context.Context, runnerID int64) ([]model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取Runner下的函数", zap.Int64("runner_id", runnerID))
	return s.runnerFuncRepo.GetByRunner(ctx, runnerID)
}

// Fork 复制函数
func (s *RunnerFunc) Fork(ctx context.Context, sourceID int64, targetTreeID int64, targetRunnerID int64, newName string, operator string) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始复制函数",
		zap.Int64("source_id", sourceID),
		zap.Int64("target_tree_id", targetTreeID),
		zap.Int64("target_runner_id", targetRunnerID),
		zap.String("new_name", newName))

	// 调用仓库层实现Fork
	return s.runnerFuncRepo.Fork(ctx, sourceID, targetTreeID, targetRunnerID, newName, operator)
}

// GetVersionHistory 获取函数版本历史
func (s *RunnerFunc) GetVersionHistory(ctx context.Context, funcID int64) ([]model.FuncVersion, error) {
	logger.Debug(ctx, "开始获取函数版本历史", zap.Int64("func_id", funcID))
	return s.runnerFuncRepo.GetVersions(ctx, funcID)
}

// SaveVersion 保存函数版本
func (s *RunnerFunc) SaveVersion(ctx context.Context, funcID int64, version string, comment string, operator string) error {
	logger.Debug(ctx, "开始保存函数版本", zap.Int64("func_id", funcID), zap.String("version", version))

	// 检查函数是否存在
	existingFunc, err := s.runnerFuncRepo.Get(ctx, funcID)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", funcID))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 创建版本记录
	funcVersion := &model.FuncVersion{
		FuncID:  funcID,
		Version: version,
		Comment: comment,
	}

	// 保存版本记录
	if err := s.runnerFuncRepo.SaveVersion(ctx, funcVersion); err != nil {
		logger.Error(ctx, "保存函数版本失败", err, zap.Int64("func_id", funcID))
		return fmt.Errorf("保存函数版本失败: %w", err)
	}

	// 更新函数版本
	//updateData := &model.RunnerFunc{
	//	Version: version,
	//}
	//if err := s.runnerFuncRepo.Update(ctx, funcID, updateData); err != nil {
	//	logger.Error(ctx, "更新函数版本信息失败", err, zap.Int64("func_id", funcID))
	//	// 不返回错误，因为版本记录已保存成功
	//}

	logger.Info(ctx, "保存函数版本成功", zap.Int64("func_id", funcID), zap.String("version", version))
	return nil
}

// UpdateStatus 更新函数状态
func (s *RunnerFunc) UpdateStatus(ctx context.Context, id int64, status int) error {
	logger.Debug(ctx, "开始更新函数状态", zap.Int64("id", id), zap.Int("status", status))

	// 检查函数是否存在
	existingFunc, err := s.runnerFuncRepo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 更新状态
	if err := s.runnerFuncRepo.UpdateStatus(ctx, id, status); err != nil {
		logger.Error(ctx, "更新函数状态失败", err, zap.Int64("id", id))
		return fmt.Errorf("更新函数状态失败: %w", err)
	}

	logger.Info(ctx, "更新函数状态成功", zap.Int64("id", id), zap.Int("status", status))
	return nil
}

// GetUserRecentFuncRecords 获取用户最近执行过的函数记录（去重）
func (s *RunnerFunc) GetUserRecentFuncRecords(ctx context.Context, user string, page, pageSize int) ([]model.FuncRunRecord, int64, error) {
	logger.Debug(ctx, "开始获取用户最近执行函数记录", zap.String("user", user), zap.Int("page", page), zap.Int("pageSize", pageSize))

	// 参数校验
	if user == "" {
		return nil, 0, errors.New("用户名不能为空")
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20 // 默认每页20条
	}

	// 调用仓库层获取数据
	records, total, err := s.runnerFuncRepo.GetUserRecentFuncRecords(ctx, user, page, pageSize)
	if err != nil {
		logger.Error(ctx, "获取用户最近执行函数记录失败", err, zap.String("user", user))
		return nil, 0, fmt.Errorf("获取用户最近执行函数记录失败: %w", err)
	}

	logger.Info(ctx, "获取用户最近执行函数记录成功",
		zap.String("user", user),
		zap.Int("count", len(records)),
		zap.Int64("total", total))

	return records, total, nil
}

// GetUserRecentFuncRecordsWithDetails 获取用户最近执行过的函数记录详细信息（去重）
func (s *RunnerFunc) GetUserRecentFuncRecordsWithDetails(ctx context.Context, user string, page, pageSize int) ([]*dto.GetUserRecentFuncRecordsResp, int64, error) {
	logger.Debug(ctx, "开始获取用户最近执行函数记录详细信息", zap.String("user", user), zap.Int("page", page), zap.Int("pageSize", pageSize))

	// 获取基础记录
	records, total, err := s.GetUserRecentFuncRecords(ctx, user, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 构建详细信息响应
	var respList []*dto.GetUserRecentFuncRecordsResp
	for _, record := range records {
		// 获取关联的详细信息
		_, runnerFunc, runner, serviceTree, err := s.runnerFuncRepo.GetFuncRunRecordWithDetails(ctx, record.ID)
		if err != nil {
			logger.Warn(ctx, "获取函数执行记录详细信息失败", zap.Error(err), zap.Int64("record_id", record.ID))
			// 继续处理其他记录，不中断整个流程
			continue
		}

		// 获取执行次数
		runCount, err := s.runnerFuncRepo.GetUserFuncRunCount(ctx, user, record.FuncId)
		if err != nil {
			logger.Warn(ctx, "获取函数执行次数失败", zap.Error(err), zap.Int64("func_id", record.FuncId))
			runCount = 0 // 设置默认值
		}

		// 构建响应对象
		resp := &dto.GetUserRecentFuncRecordsResp{}
		resp.FromFuncRunRecord(&record, runnerFunc, runner, serviceTree)
		resp.RunCount = runCount

		respList = append(respList, resp)
	}

	logger.Info(ctx, "获取用户最近执行函数记录详细信息成功",
		zap.String("user", user),
		zap.Int("count", len(respList)),
		zap.Int64("total", total))

	return respList, total, nil
}

// createFuncConfig 创建函数配置记录
func (s *RunnerFunc) createFuncConfig(ctx context.Context, runnerFunc *model.RunnerFunc, configDefine interface{}, configData interface{}, version string) error {
	// 生成配置键
	configKey := s.generateConfigKey(runnerFunc.Path, runnerFunc.Method)

	// 序列化配置结构体
	configStructData, err := json.Marshal(configDefine)
	if err != nil {
		return fmt.Errorf("序列化配置结构体失败: %w", err)
	}

	// 序列化配置初始值
	configDataJson, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("序列化配置数据失败: %w", err)
	}

	db := s.funcConfigRepo.GetDB()
	var fg model.FuncConfig
	db.Where("config_key = ?", configKey).First(&fg)

	if fg.ID != 0 { //todo 说明配置结构变更了，更新配置，这里需要更新配置，但是value不更新
		eq := fg.DiffStruct(configStructData)
		if !eq {
			logger.Infof(ctx, "配置发生了变更：%s", configStructData)
			db.Where("config_key = ?", configKey).Updates(&model.FuncConfig{
				FuncID:       runnerFunc.ID,
				ConfigStruct: configStructData,
				Version:      version,
				IsActive:     true,
			})
			return nil
		}
	}

	// 创建配置记录
	funcConfig := &model.FuncConfig{
		FuncID:       runnerFunc.ID,
		ConfigKey:    configKey,
		ConfigType:   "json",
		ConfigStruct: json.RawMessage(configStructData),
		ConfigData:   json.RawMessage(configDataJson),
		Description:  runnerFunc.Description,
		Version:      version,
		IsActive:     true,
		Base:         model.Base{CreatedBy: runnerFunc.CreatedBy, UpdatedBy: runnerFunc.UpdatedBy},
	}

	return s.funcConfigRepo.Create(ctx, funcConfig)
}

// generateConfigKey 生成配置键
func (s *RunnerFunc) generateConfigKey(path, method string) string {
	return usercall.GenerateConfigKey(path, method)
}

// createServiceTree 创建服务树节点
func (s *RunnerFunc) createServiceTree(ctx context.Context, runnerFunc *model.RunnerFunc) error {
	tree := &model.ServiceTree{
		Type:     model.ServiceTreeTypeFunction,
		ParentID: runnerFunc.TreeID,
		Name:     runnerFunc.Name,
		Title:    runnerFunc.Title,
		User:     runnerFunc.User,
		RefID:    runnerFunc.ID,
		Method:   runnerFunc.Method,
		Base:     model.Base{CreatedBy: runnerFunc.CreatedBy, UpdatedBy: runnerFunc.UpdatedBy},
	}

	if runnerFunc.Group != nil {
		tree.Group = runnerFunc.Group
	}

	return s.serviceTree.CreateNode(ctx, tree)
}

// saveFuncVersion 保存函数版本
func (s *RunnerFunc) saveFuncVersion(ctx context.Context, runnerFunc *model.RunnerFunc, version string, hash string, metadata interface{}) {
	go func() {
		s.runnerFuncRepo.SaveVersion(ctx, &model.FuncVersion{
			Base:     model.Base{CreatedBy: runnerFunc.CreatedBy, UpdatedBy: runnerFunc.UpdatedBy},
			FuncID:   runnerFunc.ID,
			Version:  version,
			MetaData: json.RawMessage(jsonx.String(metadata)),
			Hash:     hash,
		})
	}()
}

// updateRunnerVersion 更新Runner版本
func (s *RunnerFunc) updateRunnerVersion(ctx context.Context, runnerID int64, version string, description string, changeLog string, metadata interface{}, hash string, createdBy, updatedBy string) {
	go func() {
		err := s.runnerRepo.Update(ctx, runnerID, &model.Runner{Version: version})
		if err != nil {
			logger.Error(ctx, "更新版本失败", err, zap.Int64("runner_id", runnerID))
		}

		s.runnerRepo.CreateRunnerVersion(ctx, &model.RunnerVersion{
			Base:     model.Base{CreatedBy: createdBy, UpdatedBy: updatedBy},
			Desc:     description,
			Log:      changeLog,
			Version:  version,
			RunnerID: runnerID,
			MetaData: json.RawMessage(jsonx.String(metadata)),
			Hash:     hash,
		})
	}()
}

// processFuncConfig 处理函数配置
func (s *RunnerFunc) processFuncConfig(ctx context.Context, runnerFunc *model.RunnerFunc, paramsConfig, paramsData interface{}, version string) {
	if runnerFunc.HasConfig {
		if err := s.createFuncConfig(ctx, runnerFunc, paramsConfig, paramsData, version); err != nil {
			logger.Error(ctx, "创建配置记录失败", err, zap.Int64("func_id", runnerFunc.ID))
			// 不返回错误，因为函数创建已经成功
		}
	}
}

// createFunctionWithDependencies 创建函数及其所有依赖项
func (s *RunnerFunc) createFunctionWithDependencies(ctx context.Context, runnerFunc *model.RunnerFunc, addAPI *api.Info, rsp *coder.AddApisResp, autoTree bool) error {
	// 1. 使用FromAPIInfo方法进行数据赋值
	fc := *runnerFunc
	fc.ID = 0
	fc.FromAPIInfo(addAPI)

	if autoTree {
		path := addAPI.GetParentTreePath()
		tree, err := s.serviceTree.GetByFullPath(ctx, runnerFunc.User, path)
		if err != nil {
			return err
		}
		if tree == nil {
			fmt.Println(fc, tree)
		}
		fc.TreeID = tree.ID
	}
	//TreeID需要填写TreeID来分辨需要挂在哪个树下

	// 2. 创建函数
	err := s.runnerFuncRepo.Create(ctx, &fc)
	if err != nil {
		logger.Error(ctx, "创建函数失败", err)
		return fmt.Errorf("创建函数失败: %w", err)
	}
	runnerFunc.ID = fc.ID

	// 3. 处理函数配置
	s.processFuncConfig(ctx, &fc, addAPI.ParamsConfig, addAPI.ParamsData, rsp.Version)

	// 4. 创建服务树节点
	err = s.createServiceTree(ctx, &fc)
	if err != nil {
		return err
	}

	// 5. 保存函数版本
	s.saveFuncVersion(ctx, &fc, rsp.Version, rsp.Hash, addAPI)

	// 6. 更新Runner版本
	s.updateRunnerVersion(ctx, runnerFunc.RunnerID, rsp.Version, runnerFunc.Description, rsp.ApiChangeInfo.GetChangeLog(), rsp, rsp.Hash, runnerFunc.CreatedBy, runnerFunc.UpdatedBy)

	logger.Info(ctx, "创建函数成功", zap.Int64("id", runnerFunc.ID), zap.String("name", runnerFunc.Name))
	return nil
}

// GetFuncConfig 获取函数配置
func (s *RunnerFunc) GetFuncConfig(ctx context.Context, funcID int64) (*model.FuncConfig, error) {
	logger.Debug(ctx, "开始获取函数配置", zap.Int64("func_id", funcID))
	return s.funcConfigRepo.GetByFuncID(ctx, funcID)
}

// GetConfigByKey 根据配置键获取配置
func (s *RunnerFunc) GetConfigByKey(ctx context.Context, configKey string) (*model.FuncConfig, error) {
	logger.Debug(ctx, "开始根据配置键获取配置", zap.String("config_key", configKey))
	return s.funcConfigRepo.GetByConfigKey(ctx, configKey)
}

// UpdateFuncConfig 更新函数配置
func (s *RunnerFunc) UpdateFuncConfig(ctx context.Context, funcID int64, configData interface{}) error {
	logger.Debug(ctx, "开始更新函数配置", zap.Int64("func_id", funcID))

	// 获取现有配置
	existingConfig, err := s.funcConfigRepo.GetByFuncID(ctx, funcID)
	if err != nil {
		return fmt.Errorf("获取现有配置失败: %w", err)
	}
	if existingConfig == nil {
		return fmt.Errorf("函数配置不存在")
	}

	// 序列化新配置数据
	newConfigData, err := json.Marshal(configData)
	if err != nil {
		return fmt.Errorf("序列化配置数据失败: %w", err)
	}

	// 更新配置数据
	existingConfig.ConfigData = newConfigData
	existingConfig.IsActive = true

	return s.funcConfigRepo.Update(ctx, existingConfig.ID, existingConfig)
}

// ListFuncConfigs 获取函数的所有配置版本
func (s *RunnerFunc) ListFuncConfigs(ctx context.Context, funcID int64) ([]model.FuncConfig, error) {
	logger.Debug(ctx, "开始获取函数配置版本", zap.Int64("func_id", funcID))
	return s.funcConfigRepo.ListByFuncID(ctx, funcID)
}

// ListFunctionsByRunnerID 获取工作空间下所有函数
func (s *RunnerFunc) ListFunctionsByRunnerID(ctx context.Context, rid int64) ([]model.RunnerFunc, error) {
	var list []model.RunnerFunc
	err := s.runnerRepo.GetDB().Model(&model.RunnerFunc{}).Where("runner_id = ?", rid).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
