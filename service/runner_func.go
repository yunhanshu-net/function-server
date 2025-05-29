package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/x/jsonx"
	"strings"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/logger"
	"github.com/yunhanshu-net/function-server/repo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunnerFunc 函数服务实现
type RunnerFunc struct {
	runnerFuncRepo  *repo.RunnerFuncRepo
	runnerRepo      *repo.RunnerRepo
	serviceTreeRepo *repo.ServiceTreeRepo
	serviceTree     *ServiceTree
}

// NewRunnerFunc 创建函数服务
func NewRunnerFunc(db *gorm.DB) *RunnerFunc {
	svc := &RunnerFunc{
		runnerFuncRepo:  repo.NewRunnerFuncRepo(db),
		serviceTreeRepo: repo.NewServiceTreeRepo(db),
		serviceTree:     NewServiceTree(db),
		runnerRepo:      repo.NewRunnerRepo(db),
	}
	return svc
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

		fc := *runnerFunc
		fc.Name = addAPI.EnglishName
		fc.Title = addAPI.ChineseName
		fc.Tags = strings.Join(addAPI.Tags, ",")
		fc.Request = json.RawMessage(jsonx.String(addAPI.ParamsIn))
		path := "/" + gotRunner.User + "/" + gotRunner.Name + "/" + strings.Trim(addAPI.Router, "/") + "/"
		fc.Response = json.RawMessage(jsonx.String(addAPI.ParamsOut))
		fc.Path = path
		fc.Method = addAPI.Method
		fc.Callbacks = strings.Join(addAPI.Callbacks, ",")
		fc.UseTables = strings.Join(addAPI.UseTables, ",")
		// 设置默认值
		if fc.User == "" {
			fc.User = "admin"
		}
		// 创建函数
		err = s.runnerFuncRepo.Create(ctx, &fc)
		if err != nil {
			logger.Error(ctx, "创建函数失败", err)
			return fmt.Errorf("创建函数失败: %w", err)
		}
		tree := &model.ServiceTree{
			Type:     model.ServiceTreeTypeFunction,
			ParentID: fc.TreeID,
			Name:     fc.Name,
			Title:    fc.Title,
			User:     t.User,
			RefID:    fc.ID,
			Method:   fc.Method,
			Base:     model.Base{CreatedBy: fc.CreatedBy, UpdatedBy: fc.UpdatedBy},
		}
		err = s.serviceTree.CreateNode(ctx, tree)
		if err != nil {
			return err
		}
		go func() {
			s.runnerFuncRepo.SaveVersion(ctx, &model.FuncVersion{
				Base:     model.Base{CreatedBy: fc.CreatedBy, UpdatedBy: fc.UpdatedBy},
				FuncID:   fc.ID,
				Version:  rsp.Version,
				MetaData: json.RawMessage(jsonx.String(addAPI)),
				Hash:     rsp.Hash,
			})
		}()

	}

	go func() {
		err = s.runnerRepo.Update(ctx, runnerFunc.RunnerID, &model.Runner{Version: rsp.Version})
		if err != nil {
			logger.Error(ctx, "更新版本失败", err, zap.Int64("func_id", runnerFunc.ID))
		}
		s.runnerRepo.CreateRunnerVersion(ctx, &model.RunnerVersion{
			Base:     model.Base{CreatedBy: runnerFunc.CreatedBy, UpdatedBy: runnerFunc.UpdatedBy},
			Desc:     runnerFunc.Description,
			Log:      rsp.ApiChangeInfo.GetChangeLog(),
			Version:  rsp.Version,
			RunnerID: runnerFunc.RunnerID,
			MetaData: json.RawMessage(jsonx.String(rsp)),
			Hash:     rsp.Hash,
		})

	}()

	logger.Info(ctx, "创建函数成功", zap.Int64("id", runnerFunc.ID), zap.String("name", runnerFunc.Name))
	return nil
}

// Get 获取函数详情
func (s *RunnerFunc) Get(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取函数详情", zap.Int64("id", id))
	return s.runnerFuncRepo.Get(ctx, id)
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
	existingFunc, err := s.runnerFuncRepo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

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
