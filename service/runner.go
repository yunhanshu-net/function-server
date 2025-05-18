package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/repo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 辅助函数，将time.Time转换为model.Time
func timeToModelTime(t time.Time) model.Time {
	return model.Time(t)
}

// Runner Runner服务实现
type Runner struct {
	repo           *repo.RunnerRepo
	runcherService RuncherService
}

// NewRunner 创建Runner服务
func NewRunner(db *gorm.DB) *Runner {
	return &Runner{
		runcherService: GetRuncherService(),
		repo:           repo.NewRunnerRepo(db),
	}
}

// Create 创建Runner
func (s *Runner) Create(ctx context.Context, runner *model.Runner) error {
	logger.Debug(ctx, "开始创建Runner", zap.String("title", runner.Title))

	// 业务逻辑校验
	if runner.Title == "" {
		return errors.New("标题不能为空")
	}
	if runner.Name == "" {
		return errors.New("名称不能为空")
	}

	// 检查名称是否已存在
	exist, err := s.repo.GetByUserAndName(ctx, runner.User, runner.Name)
	if err != nil {
		logger.Error(ctx, "检查Runner名称失败", err, zap.String("name", runner.Name))
		return fmt.Errorf("检查名称失败: %w", err)
	}
	if exist != nil {
		logger.Info(ctx, "Runner名称已存在", zap.String("name", runner.Name))
		return errors.New("名称已存在")
	}

	// 设置默认值
	if runner.Status == 0 {
		runner.Status = 1 // 默认启用
	}

	//todo 这里先忽略错误
	versionString, _ := s.runcherService.CreateProject(ctx, runner)
	//if err != nil {
	//	return err
	//}
	runner.Version = versionString

	// 开始事务
	tx := s.repo.GetDB().Begin()
	if tx.Error != nil {
		logger.Error(ctx, "开始事务失败", tx.Error)
		return fmt.Errorf("开始事务失败: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error(ctx, "事务回滚", nil, zap.Any("recover", r))
		}
	}()

	// 创建一个服务树根节点
	serviceTreeSvc := NewServiceTree(s.repo.GetDB())
	serviceTree := &model.ServiceTree{
		Name:     runner.Name,
		Title:    runner.Title,
		ParentID: 0, // 根节点
		User:     runner.User,
		Type:     model.ServiceTreeTypePackage,
	}
	serviceTree.CreatedBy = runner.CreatedBy
	serviceTree.UpdatedBy = runner.CreatedBy

	// 使用事务创建服务树
	err = serviceTreeSvc.CreateWithTx(ctx, tx, serviceTree)
	if err != nil {
		tx.Rollback()
		logger.Error(ctx, "创建Runner对应的服务树节点失败", err, zap.String("name", runner.Name))
		return fmt.Errorf("创建服务树节点失败: %w", err)
	}

	runner.TreeID = serviceTree.ID
	// 创建Runner
	err = s.repo.CreateWithTx(ctx, tx, runner)
	if err != nil {
		tx.Rollback()
		logger.Error(ctx, "创建Runner失败", err, zap.String("name", runner.Name))
		return fmt.Errorf("创建Runner失败: %w", err)
	}
	err = serviceTreeSvc.UpdateWithTx(ctx, tx, serviceTree.ID, &model.ServiceTree{RunnerID: runner.ID})
	if err != nil {
		tx.Rollback()
		logger.Error(ctx, "更新ServiceTree失败", err, zap.Any("serviceTree.ID", serviceTree.ID))
		return fmt.Errorf("更新ServiceTree失败: %w", err)
	}

	// 创建版本
	version := &model.RunnerVersion{
		RunnerID:  runner.ID,
		Version:   versionString,
		Comment:   "初始版本",
		CreatedBy: runner.CreatedBy}

	err = s.repo.SaveVersionWithTx(ctx, tx, version)
	if err != nil {
		tx.Rollback()
		logger.Error(ctx, "保存Runner版本失败", err, zap.String("name", runner.Name))
		return fmt.Errorf("保存版本失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error(ctx, "提交事务失败", err)
		return fmt.Errorf("提交事务失败: %w", err)
	}

	logger.Info(ctx, "创建Runner成功", zap.Int64("id", runner.ID), zap.String("title", runner.Title))
	return nil
}

// Get 获取Runner详情
func (s *Runner) Get(ctx context.Context, id int64) (*model.Runner, error) {
	logger.Debug(ctx, "开始获取Runner详情", zap.Int64("id", id))
	runner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner详情失败", err, zap.Int64("id", id))
		return nil, fmt.Errorf("获取Runner详情失败: %w", err)
	}

	if runner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
		return nil, nil
	}

	logger.Debug(ctx, "获取Runner详情成功", zap.Int64("id", id))
	return runner, nil
}
func (s *Runner) GetByUserName(ctx context.Context, user, name string) (*model.Runner, error) {
	runner, err := s.repo.GetByUserAndName(ctx, user, name)
	if err != nil {
		return nil, fmt.Errorf("获取Runner详情失败: %w", err)
	}

	if runner == nil {
		return nil, fmt.Errorf("runner不存在")
	}

	return runner, nil
}

// Update 更新Runner
func (s *Runner) Update(ctx context.Context, id int64, runner *model.Runner) error {
	logger.Debug(ctx, "开始更新Runner", zap.Int64("id", id))
	// 先检查是否存在
	existingRunner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取Runner失败: %w", err)
	}
	if existingRunner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
		return errors.New("runner不存在")
	}

	// 业务逻辑校验
	if runner.Title == "" {
		return errors.New("标题不能为空")
	}

	// 如果名称发生变更，检查名称是否已存在
	if runner.Name != "" && runner.Name != existingRunner.Name {
		existing, err := s.repo.GetByName(ctx, runner.Name)
		if err != nil {
			logger.Error(ctx, "检查Runner名称失败", err, zap.String("name", runner.Name))
			return fmt.Errorf("检查名称失败: %w", err)
		}
		if existing != nil && existing.ID != id {
			return errors.New("名称已存在")
		}
	}

	// 设置更新时间
	runner.UpdatedAt = timeToModelTime(time.Now())

	// 更新Runner
	if err := s.repo.Update(ctx, id, runner); err != nil {
		logger.Error(ctx, "更新Runner失败", err, zap.Int64("id", id))
		return fmt.Errorf("更新Runner失败: %w", err)
	}

	logger.Info(ctx, "更新Runner成功", zap.Int64("id", id))
	return nil
}

// Delete 删除Runner
func (s *Runner) Delete(ctx context.Context, id int64, operator string) error {
	logger.Debug(ctx, "开始删除Runner", zap.Int64("id", id))
	// 先检查是否存在
	existingRunner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取Runner失败: %w", err)
	}
	if existingRunner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
		return errors.New("Runner不存在")
	}

	// 设置删除者
	if err := s.repo.SetDeletedBy(ctx, id, operator); err != nil {
		logger.Error(ctx, "设置Runner删除者失败", err, zap.Int64("id", id))
		return fmt.Errorf("设置删除者失败: %w", err)
	}

	// 删除Runner
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error(ctx, "删除Runner失败", err, zap.Int64("id", id))
		return fmt.Errorf("删除Runner失败: %w", err)
	}

	logger.Info(ctx, "删除Runner成功", zap.Int64("id", id))
	return nil
}

// List 获取Runner列表
func (s *Runner) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.Runner, int64, error) {
	logger.Debug(ctx, "开始获取Runner列表", zap.Int("page", page), zap.Int("pageSize", pageSize))
	offset := (page - 1) * pageSize
	runners, total, err := s.repo.List(ctx, offset, pageSize, conditions)
	if err != nil {
		logger.Error(ctx, "获取Runner列表失败", err)
		return nil, 0, fmt.Errorf("获取Runner列表失败: %w", err)
	}

	logger.Debug(ctx, "获取Runner列表成功", zap.Int64("total", total))
	return runners, total, nil
}

// Fork 复制Runner
func (s *Runner) Fork(ctx context.Context, id int64, operator string) (*model.Runner, error) {
	logger.Debug(ctx, "开始复制Runner", zap.Int64("id", id))
	// 先检查是否存在
	sourceRunner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取源Runner失败", err, zap.Int64("id", id))
		return nil, fmt.Errorf("获取源Runner失败: %w", err)
	}
	if sourceRunner == nil {
		logger.Info(ctx, "源Runner不存在", zap.Int64("id", id))
		return nil, errors.New("源Runner不存在")
	}

	// 创建一个新的Runner
	newRunner := *sourceRunner
	newRunner.ID = 0 // 重置ID，让数据库自动生成

	// 修改标题，添加"Copy of"前缀
	newRunner.Title = fmt.Sprintf("Copy of %s", sourceRunner.Title)

	// 生成唯一名称
	newRunner.Name = fmt.Sprintf("%s_copy_%d", sourceRunner.Name, time.Now().UnixNano())

	// 设置创建者和更新者信息
	newRunner.CreatedBy = operator
	newRunner.UpdatedBy = operator

	// 设置fork源信息
	newRunner.ForkFromID = &sourceRunner.ID
	newRunner.ForkFromUser = sourceRunner.User

	// 设置当前用户
	newRunner.User = operator

	// 设置创建时间和更新时间
	now := time.Now()
	newRunner.CreatedAt = timeToModelTime(now)
	newRunner.UpdatedAt = timeToModelTime(now)

	// 保存新Runner
	if err := s.repo.Create(ctx, &newRunner); err != nil {
		logger.Error(ctx, "创建Fork Runner失败", err)
		return nil, fmt.Errorf("创建Fork Runner失败: %w", err)
	}

	// 创建新Runner的版本记录
	version := &model.RunnerVersion{
		RunnerID:  newRunner.ID,
		Version:   "1.0.0",
		CreatedBy: operator,
		CreatedAt: timeToModelTime(now),
		Comment:   fmt.Sprintf("从 %s(%d) Fork", sourceRunner.Name, sourceRunner.ID),
	}

	err = s.repo.SaveVersion(ctx, version)
	if err != nil {
		logger.Error(ctx, "保存Fork Runner版本失败", err, zap.Int64("runner_id", newRunner.ID))
		// 不返回错误，因为Runner已经Fork成功
	}

	logger.Info(ctx, "Fork Runner成功",
		zap.Int64("source_id", sourceRunner.ID),
		zap.Int64("new_id", newRunner.ID),
		zap.String("title", newRunner.Title))

	return &newRunner, nil
}

// GetVersionHistory 获取Runner版本历史
func (s *Runner) GetVersionHistory(ctx context.Context, id int64) ([]model.RunnerVersion, error) {
	logger.Debug(ctx, "开始获取Runner版本历史", zap.Int64("id", id))

	// 先检查Runner是否存在
	runner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", id))
		return nil, fmt.Errorf("获取Runner失败: %w", err)
	}
	if runner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
		return nil, errors.New("Runner不存在")
	}

	// 获取版本历史
	versions, err := s.repo.GetVersions(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner版本历史失败", err, zap.Int64("id", id))
		return nil, fmt.Errorf("获取版本历史失败: %w", err)
	}

	logger.Debug(ctx, "获取Runner版本历史成功", zap.Int64("id", id), zap.Int("count", len(versions)))
	return versions, nil
}

// UpdateStatus 更新Runner状态
func (s *Runner) UpdateStatus(ctx context.Context, id int64, status int) error {
	logger.Debug(ctx, "开始更新Runner状态", zap.Int64("id", id), zap.Int("status", status))

	// 先检查Runner是否存在
	runner, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取Runner失败: %w", err)
	}
	if runner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
		return errors.New("Runner不存在")
	}

	// 更新状态
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		logger.Error(ctx, "更新Runner状态失败", err, zap.Int64("id", id), zap.Int("status", status))
		return fmt.Errorf("更新状态失败: %w", err)
	}

	logger.Info(ctx, "更新Runner状态成功", zap.Int64("id", id), zap.Int("status", status))
	return nil
}

// SaveVersion 保存Runner版本
func (s *Runner) SaveVersion(ctx context.Context, runnerID int64, version string, comment string, operator string) error {
	logger.Debug(ctx, "开始保存Runner版本",
		zap.Int64("runner_id", runnerID),
		zap.String("version", version),
		zap.String("comment", comment))

	// 先检查Runner是否存在
	runner, err := s.repo.Get(ctx, runnerID)
	if err != nil {
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", runnerID))
		return fmt.Errorf("获取Runner失败: %w", err)
	}
	if runner == nil {
		logger.Info(ctx, "Runner不存在", zap.Int64("id", runnerID))
		return errors.New("Runner不存在")
	}

	// 保存版本记录
	versionRecord := &model.RunnerVersion{
		RunnerID:  runnerID,
		Version:   version,
		CreatedBy: operator,
		CreatedAt: timeToModelTime(time.Now()),
		Comment:   comment,
	}

	if err := s.repo.SaveVersion(ctx, versionRecord); err != nil {
		logger.Error(ctx, "保存Runner版本失败", err,
			zap.Int64("runner_id", runnerID),
			zap.String("version", version))
		return fmt.Errorf("保存版本失败: %w", err)
	}

	logger.Info(ctx, "保存Runner版本成功",
		zap.Int64("runner_id", runnerID),
		zap.String("version", version))
	return nil
}

// GetByName 根据名称获取Runner
func (s *Runner) GetByName(ctx context.Context, name string) (*model.Runner, error) {
	logger.Debug(ctx, "根据名称获取Runner", zap.String("name", name))
	return s.repo.GetByName(ctx, name)
}
