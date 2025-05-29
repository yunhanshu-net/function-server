package repo

import (
	"context"
	"errors"
	"go.uber.org/zap"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/logger"
	"gorm.io/gorm"
)

// RunnerRepo Runner仓库实现
type RunnerRepo struct {
	db *gorm.DB
}

func (r *RunnerRepo) GetDB() *gorm.DB {
	return r.db
}

// NewRunnerRepo 创建Runner仓库
func NewRunnerRepo(db *gorm.DB) *RunnerRepo {
	return &RunnerRepo{db: db}
}

// Create 创建Runner
func (r *RunnerRepo) Create(ctx context.Context, runner *model.Runner) error {
	logger.Debug(ctx, "开始创建Runner", zap.String("name", runner.Name))
	return r.db.WithContext(ctx).Create(runner).Error
}

//func (r *RunnerRepo) AddFunctions(ctx context.Context, functions []*api.Info) error {
//	var fns []*model.RunnerFunc
//	for i, function := range functions {
//		f := &model.RunnerFunc{
//
//		}
//	}
//}

func (r *RunnerRepo) CreateRunnerVersion(ctx context.Context, v *model.RunnerVersion) error {
	return r.db.WithContext(ctx).Create(v).Error
}

// Get 获取Runner详情
func (r *RunnerRepo) Get(ctx context.Context, id int64) (*model.Runner, error) {
	logger.Debug(ctx, "开始获取Runner", zap.Int64("id", id))
	var runner model.Runner
	err := r.db.WithContext(ctx).First(&runner, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info(ctx, "Runner不存在", zap.Int64("id", id))
			return nil, nil
		}
		logger.Error(ctx, "获取Runner失败", err, zap.Int64("id", id))
		return nil, err
	}
	return &runner, nil
}

// Update 更新Runner
func (r *RunnerRepo) Update(ctx context.Context, id int64, runner *model.Runner) error {
	logger.Debug(ctx, "开始更新Runner", zap.Int64("id", id))
	return r.db.WithContext(ctx).Model(&model.Runner{}).Where("id = ?", id).Updates(runner).Error
}

// Delete 删除Runner
func (r *RunnerRepo) Delete(ctx context.Context, id int64) error {
	logger.Debug(ctx, "开始删除Runner", zap.Int64("id", id))
	return r.db.WithContext(ctx).Delete(&model.Runner{}, id).Error
}

// List 获取Runner列表
func (r *RunnerRepo) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.Runner, int64, error) {
	logger.Debug(ctx, "开始获取Runner列表", zap.Int("page", page), zap.Int("pageSize", pageSize))
	var total int64
	var runnerList []model.Runner

	query := r.db.WithContext(ctx).Model(&model.Runner{})

	// 应用查询条件
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error(ctx, "获取Runner总数失败", err)
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&runnerList).Error; err != nil {
		logger.Error(ctx, "获取Runner列表失败", err)
		return nil, 0, err
	}

	return runnerList, total, nil
}

// GetByName 根据名称获取Runner
func (r *RunnerRepo) GetByName(ctx context.Context, name string) (*model.Runner, error) {
	logger.Debug(ctx, "根据名称获取Runner", zap.String("name", name))
	var runner model.Runner
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&runner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称获取Runner失败", err, zap.String("name", name))
		return nil, err
	}
	return &runner, nil
}

// GetByUserAndName 根据用户和名称获取Runner
func (r *RunnerRepo) GetByUserAndName(ctx context.Context, user, name string) (*model.Runner, error) {
	logger.Debug(ctx, "根据名称获取Runner", zap.String("name", name), zap.String("user", user))
	var runner model.Runner
	err := r.db.WithContext(ctx).Where("user = ? and name = ?", user, name).First(&runner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称获取Runner失败", err, zap.String("name", name))
		return nil, err
	}
	return &runner, nil
}

// GetByTreeId 根据treeId获取Runner
func (r *RunnerRepo) GetByTreeId(ctx context.Context, treeId int64) (*model.Runner, error) {
	logger.Debug(ctx, "根据名称获取Runner", zap.Any("treeId", treeId))
	var runner model.Runner
	err := r.db.WithContext(ctx).Where("tree_id = ?", treeId).First(&runner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称获取Runner失败", err, zap.Any("tree_id", treeId))
		return nil, err
	}
	return &runner, nil
}

// SetDeletedBy 设置删除者
func (r *RunnerRepo) SetDeletedBy(ctx context.Context, id int64, deletedBy string) error {
	logger.Debug(ctx, "设置Runner删除者", zap.Int64("id", id), zap.String("deleted_by", deletedBy))
	return r.db.WithContext(ctx).Model(&model.Runner{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error
}

// UpdateStatus 更新Runner状态
func (r *RunnerRepo) UpdateStatus(ctx context.Context, id int64, status int) error {
	logger.Debug(ctx, "更新Runner状态", zap.Int64("id", id), zap.Int("status", status))
	return r.db.WithContext(ctx).Model(&model.Runner{}).Where("id = ?", id).Update("status", status).Error
}

// SaveVersion 保存Runner版本
func (r *RunnerRepo) SaveVersion(ctx context.Context, version *model.RunnerVersion) error {
	logger.Debug(ctx, "保存Runner版本", zap.Int64("runner_id", version.RunnerID), zap.String("version", version.Version))
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersions 获取Runner版本列表
func (r *RunnerRepo) GetVersions(ctx context.Context, runnerID int64) ([]model.RunnerVersion, error) {
	logger.Debug(ctx, "获取Runner版本列表", zap.Int64("runner_id", runnerID))
	var versions []model.RunnerVersion
	err := r.db.WithContext(ctx).Where("runner_id = ?", runnerID).Order("created_at DESC").Find(&versions).Error
	if err != nil {
		logger.Error(ctx, "获取Runner版本列表失败", err, zap.Int64("runner_id", runnerID))
		return nil, err
	}
	return versions, nil
}

// BatchCreate 批量创建Runner
func (r *RunnerRepo) BatchCreate(ctx context.Context, runners []model.Runner) error {
	logger.Debug(ctx, "开始批量创建Runner", zap.Int("count", len(runners)))
	return r.db.WithContext(ctx).Create(&runners).Error
}

// CreateWithTx 使用事务创建Runner
func (r *RunnerRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, runner *model.Runner) error {
	logger.Debug(ctx, "使用事务创建Runner", zap.String("name", runner.Name))
	return tx.Create(runner).Error
}

// SaveVersionWithTx 使用事务保存Runner版本
func (r *RunnerRepo) SaveVersionWithTx(ctx context.Context, tx *gorm.DB, version *model.RunnerVersion) error {
	logger.Debug(ctx, "使用事务保存Runner版本", zap.Int64("runner_id", version.RunnerID), zap.String("version", version.Version))
	return tx.Create(version).Error
}
