package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/dto/coder"
	"sync"
	"time"

	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/repo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunnerFuncService 函数服务接口
type RunnerFuncService interface {
	Create(ctx context.Context, runnerFunc *model.RunnerFunc) error
	Get(ctx context.Context, id int64) (*model.RunnerFunc, error)
	Update(ctx context.Context, id int64, updateData *model.RunnerFunc) error
	Delete(ctx context.Context, id int64, operator string) error
	List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.RunnerFunc, int64, error)
	GetByRunner(ctx context.Context, runnerID int64) ([]model.RunnerFunc, error)
	Fork(ctx context.Context, sourceID int64, targetTreeID int64, targetRunnerID int64, newName string, operator string) (*model.RunnerFunc, error)
	GetVersionHistory(ctx context.Context, funcID int64) ([]model.FuncVersion, error)
	SaveVersion(ctx context.Context, funcID int64, version string, comment string, operator string) error
	UpdateStatus(ctx context.Context, id int64, status int) error
	GetRunnerFuncByID(ctx context.Context, id int64) (*model.RunnerFunc, error)
}

// RunnerFunc 函数服务实现
type RunnerFunc struct {
	repo            repo.RunnerFuncRepository
	runnerRepo      *repo.RunnerRepo
	serviceTreeRepo *repo.ServiceTreeRepo
	serviceTree     *ServiceTree
}

// RunnerFuncParam 全局参数
var (
	runcherFuncService RunnerFuncService
	runcherFuncMutex   = &sync.RWMutex{}
)

// GetRunnerFuncService 获取RunnerFunc服务实例
func GetRunnerFuncService() RunnerFuncService {
	runcherFuncMutex.RLock()
	defer runcherFuncMutex.RUnlock()
	return runcherFuncService
}

// NewRunnerFunc 创建函数服务
func NewRunnerFunc(db *gorm.DB) *RunnerFunc {
	svc := &RunnerFunc{
		repo:            repo.NewRunnerFuncRepo(db),
		serviceTreeRepo: repo.NewServiceTreeRepo(db),
		serviceTree:     NewServiceTree(db),
		runnerRepo:      repo.NewRunnerRepo(db),
	}

	// 设置全局实例
	runcherFuncMutex.Lock()
	runcherFuncService = svc
	runcherFuncMutex.Unlock()

	return svc
}

// Create 创建函数
func (s *RunnerFunc) Create(ctx context.Context, runnerFunc *model.RunnerFunc) error {
	logger.Debug(ctx, "开始创建函数", zap.String("name", runnerFunc.Name))

	// 业务逻辑校验
	if runnerFunc.Name == "" {
		return errors.New("函数名称不能为空")
	}
	if runnerFunc.Title == "" {
		return errors.New("函数标题不能为空")
	}

	t, err := repo.NewServiceTreeRepo(db.GetDB()).Get(ctx, runnerFunc.TreeID)
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
	//// 检查服务树是否存在
	//if runnerFunc.TreeID > 0 {
	//
	if packageTree == nil {
		return errors.New("关联的服务树不存在")
	}
	//}

	// 检查名称是否已存在
	existingFunc, err := s.repo.GetByName(ctx, runnerFunc.RunnerID, runnerFunc.Name)
	if err != nil {
		logger.Error(ctx, "检查函数名称失败", err, zap.String("name", runnerFunc.Name))
		return fmt.Errorf("检查函数名称失败: %w", err)
	}
	if existingFunc != nil {
		return errors.New("函数名称已存在")
	}

	// 设置默认值
	if runnerFunc.User == "" {
		runnerFunc.User = "admin"
	}

	// 创建函数
	err = s.repo.Create(ctx, runnerFunc)
	if err != nil {
		logger.Error(ctx, "创建函数失败", err)
		return fmt.Errorf("创建函数失败: %w", err)
	}

	tree := &model.ServiceTree{
		Type:     model.ServiceTreeTypeFunction,
		ParentID: runnerFunc.TreeID,
		Name:     runnerFunc.Name,
		Title:    runnerFunc.Title,
		User:     t.User,
		Base: model.Base{
			CreatedBy: runnerFunc.CreatedBy,
			UpdatedBy: runnerFunc.UpdatedBy,
		},
	}

	err = s.serviceTree.Create(ctx, tree)
	if err != nil {
		return err
	}

	// 创建版本记录
	version := &model.FuncVersion{
		FuncID:    runnerFunc.ID,
		Version:   "1.0.0",
		Comment:   "初始版本",
		CreatedBy: runnerFunc.CreatedBy,
		CreatedAt: runnerFunc.CreatedAt,
	}
	if runnerFunc.Code != "" {
		service := GetRuncherService()
		r := &coder.AddApiReq{
			Runner: &coder.Runner{
				Language: gotRunner.Language,
				Name:     gotRunner.Name,
				Version:  gotRunner.Version,
				User:     gotRunner.User,
			},
			CodeApi: &coder.CodeApi{
				EnName:         runnerFunc.Name,
				CnName:         runnerFunc.Title,
				Desc:           runnerFunc.Description,
				Language:       gotRunner.Language,
				Code:           runnerFunc.Code,
				Package:        packageTree.Name,
				AbsPackagePath: packageTree.GetSubFullPath(),
			},
		}
		rsp, err := service.AddAPI2(ctx, r)
		if err != nil {
			logger.Error(ctx, "添加api失败", err, zap.Int64("func_id", runnerFunc.ID))
			return err
		}
		err = s.runnerRepo.Update(ctx, runnerFunc.RunnerID, &model.Runner{Version: rsp.Version})
		if err != nil {
			logger.Error(ctx, "更新版本失败", err, zap.Int64("func_id", runnerFunc.ID))
			return err
		}
	}

	if err := s.repo.SaveVersion(ctx, version); err != nil {
		logger.Error(ctx, "保存函数版本失败", err, zap.Int64("func_id", runnerFunc.ID))
		// 不返回错误，因为函数已创建成功
	}

	logger.Info(ctx, "创建函数成功", zap.Int64("id", runnerFunc.ID), zap.String("name", runnerFunc.Name))
	return nil
}

// Get 获取函数详情
func (s *RunnerFunc) Get(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取函数详情", zap.Int64("id", id))
	return s.repo.Get(ctx, id)
}

// GetRunnerFuncByID 通过ID获取运行函数（适配接口）
func (s *RunnerFunc) GetRunnerFuncByID(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	return s.Get(ctx, id)
}

// Update 更新函数
func (s *RunnerFunc) Update(ctx context.Context, id int64, updateData *model.RunnerFunc) error {
	logger.Debug(ctx, "开始更新函数", zap.Int64("id", id))

	// 检查函数是否存在
	existingFunc, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 如果要更新名称，需要检查名称是否已存在
	if updateData.Name != "" && updateData.Name != existingFunc.Name {
		existing, err := s.repo.GetByName(ctx, existingFunc.RunnerID, updateData.Name)
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
		treeExists, err := s.repo.CheckServiceTreeExists(ctx, updateData.TreeID)
		if err != nil {
			logger.Error(ctx, "检查服务树存在性失败", err, zap.Int64("tree_id", updateData.TreeID))
			return fmt.Errorf("检查服务树存在性失败: %w", err)
		}
		if !treeExists {
			return errors.New("关联的服务树不存在")
		}
	}

	// 更新函数
	if err := s.repo.Update(ctx, id, updateData); err != nil {
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
	existingFunc, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 设置删除者
	if err := s.repo.SetDeletedBy(ctx, id, operator); err != nil {
		logger.Error(ctx, "设置函数删除者失败", err, zap.Int64("id", id))
		return fmt.Errorf("设置删除者失败: %w", err)
	}

	// 删除函数
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error(ctx, "删除函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("删除函数失败: %w", err)
	}

	logger.Info(ctx, "删除函数成功", zap.Int64("id", id))
	return nil
}

// List 获取函数列表
func (s *RunnerFunc) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.RunnerFunc, int64, error) {
	logger.Debug(ctx, "开始获取函数列表", zap.Int("page", page), zap.Int("pageSize", pageSize))
	return s.repo.List(ctx, page, pageSize, conditions)
}

// GetByRunner 获取Runner下的所有函数
func (s *RunnerFunc) GetByRunner(ctx context.Context, runnerID int64) ([]model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取Runner下的函数", zap.Int64("runner_id", runnerID))
	return s.repo.GetByRunner(ctx, runnerID)
}

// Fork 复制函数
func (s *RunnerFunc) Fork(ctx context.Context, sourceID int64, targetTreeID int64, targetRunnerID int64, newName string, operator string) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始复制函数",
		zap.Int64("source_id", sourceID),
		zap.Int64("target_tree_id", targetTreeID),
		zap.Int64("target_runner_id", targetRunnerID),
		zap.String("new_name", newName))

	// 调用仓库层实现Fork
	return s.repo.Fork(ctx, sourceID, targetTreeID, targetRunnerID, newName, operator)
}

// GetVersionHistory 获取函数版本历史
func (s *RunnerFunc) GetVersionHistory(ctx context.Context, funcID int64) ([]model.FuncVersion, error) {
	logger.Debug(ctx, "开始获取函数版本历史", zap.Int64("func_id", funcID))
	return s.repo.GetVersions(ctx, funcID)
}

// SaveVersion 保存函数版本
func (s *RunnerFunc) SaveVersion(ctx context.Context, funcID int64, version string, comment string, operator string) error {
	logger.Debug(ctx, "开始保存函数版本", zap.Int64("func_id", funcID), zap.String("version", version))

	// 检查函数是否存在
	existingFunc, err := s.repo.Get(ctx, funcID)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", funcID))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 创建版本记录
	funcVersion := &model.FuncVersion{
		FuncID:    funcID,
		Version:   version,
		Comment:   comment,
		CreatedBy: operator,
		CreatedAt: model.Time(time.Now()),
	}

	// 保存版本记录
	if err := s.repo.SaveVersion(ctx, funcVersion); err != nil {
		logger.Error(ctx, "保存函数版本失败", err, zap.Int64("func_id", funcID))
		return fmt.Errorf("保存函数版本失败: %w", err)
	}

	// 更新函数版本
	//updateData := &model.RunnerFunc{
	//	Version: version,
	//}
	//if err := s.repo.Update(ctx, funcID, updateData); err != nil {
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
	existingFunc, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取函数失败: %w", err)
	}
	if existingFunc == nil {
		return errors.New("函数不存在")
	}

	// 更新状态
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		logger.Error(ctx, "更新函数状态失败", err, zap.Int64("id", id))
		return fmt.Errorf("更新函数状态失败: %w", err)
	}

	logger.Info(ctx, "更新函数状态成功", zap.Int64("id", id), zap.Int("status", status))
	return nil
}
