package repo

import (
	"context"
	"errors"

	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunnerFuncRepo 函数仓库实现
type RunnerFuncRepo struct {
	db *gorm.DB
}

// NewRunnerFuncRepo 创建函数仓库
func NewRunnerFuncRepo(db *gorm.DB) *RunnerFuncRepo {
	return &RunnerFuncRepo{db: db}
}

// Create 创建函数
func (r *RunnerFuncRepo) Create(ctx context.Context, runnerFunc *model.RunnerFunc) error {
	logger.Debug(ctx, "开始创建函数", zap.String("name", runnerFunc.Name))
	return r.db.WithContext(ctx).Create(runnerFunc).Error
}

// Get 获取函数详情
func (r *RunnerFuncRepo) Get(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取函数", zap.Int64("id", id))
	var runnerFunc model.RunnerFunc
	err := r.db.WithContext(ctx).First(&runnerFunc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info(ctx, "函数不存在", zap.Int64("id", id))
			return nil, nil
		}
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return nil, err
	}
	return &runnerFunc, nil
}

// GetByTreeId 获取函数详情
func (r *RunnerFuncRepo) GetByTreeId(ctx context.Context, id int64) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "开始获取函数", zap.Int64("id", id))
	var runnerFunc model.RunnerFunc
	err := r.db.WithContext(ctx).First(&runnerFunc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info(ctx, "函数不存在", zap.Int64("id", id))
			return nil, nil
		}
		logger.Error(ctx, "获取函数失败", err, zap.Int64("id", id))
		return nil, err
	}
	return &runnerFunc, nil
}

// Update 更新函数
func (r *RunnerFuncRepo) Update(ctx context.Context, id int64, updateData *model.RunnerFunc) error {
	logger.Debug(ctx, "开始更新函数", zap.Int64("id", id))
	return r.db.WithContext(ctx).Model(&model.RunnerFunc{}).Where("id = ?", id).Updates(updateData).Error
}

// Delete 删除函数
func (r *RunnerFuncRepo) Delete(ctx context.Context, id int64) error {
	logger.Debug(ctx, "开始删除函数", zap.Int64("id", id))
	return r.db.WithContext(ctx).Delete(&model.RunnerFunc{}, id).Error
}

// SetDeletedBy 设置删除者
func (r *RunnerFuncRepo) SetDeletedBy(ctx context.Context, id int64, deletedBy string) error {
	logger.Debug(ctx, "设置函数删除者", zap.Int64("id", id), zap.String("deleted_by", deletedBy))
	return r.db.WithContext(ctx).Model(&model.RunnerFunc{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error
}

// List 获取函数列表
func (r *RunnerFuncRepo) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.RunnerFunc, int64, error) {
	logger.Debug(ctx, "开始获取函数列表", zap.Int("page", page), zap.Int("pageSize", pageSize))
	var total int64
	var runnerFuncList []model.RunnerFunc

	query := r.db.WithContext(ctx).Model(&model.RunnerFunc{})

	// 应用查询条件
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error(ctx, "获取函数总数失败", err)
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&runnerFuncList).Error; err != nil {
		logger.Error(ctx, "获取函数列表失败", err)
		return nil, 0, err
	}

	return runnerFuncList, total, nil
}

// GetByRunner 获取Runner下的所有函数
func (r *RunnerFuncRepo) GetByRunner(ctx context.Context, runnerID int64) ([]model.RunnerFunc, error) {
	logger.Debug(ctx, "获取Runner下的所有函数", zap.Int64("runner_id", runnerID))
	var funcs []model.RunnerFunc
	err := r.db.WithContext(ctx).Where("runner_id = ?", runnerID).Find(&funcs).Error
	if err != nil {
		logger.Error(ctx, "获取Runner下的函数失败", err, zap.Int64("runner_id", runnerID))
	}
	return funcs, err
}

// CheckRunnerExists 检查Runner是否存在
func (r *RunnerFuncRepo) CheckRunnerExists(ctx context.Context, runnerID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Runner{}).Where("id = ?", runnerID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CheckServiceTreeExists 检查ServiceTree是否存在
func (r *RunnerFuncRepo) CheckServiceTreeExists(ctx context.Context, treeID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&model.ServiceTree{}).Where("id = ?", treeID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Fork 复制函数
func (r *RunnerFuncRepo) Fork(ctx context.Context, sourceID int64, targetTreeID int64, targetRunnerID int64, newName string, operator string) (*model.RunnerFunc, error) {
	// 获取源函数
	sourceFunc, err := r.Get(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// 创建新函数
	newFunc := *sourceFunc
	newFunc.ID = 0 // 重置ID，让数据库自动生成
	newFunc.TreeID = targetTreeID
	newFunc.RunnerID = targetRunnerID

	// 处理函数名称
	if newName != "" {
		newFunc.Name = newName
		newFunc.Title = newName
	} else {
		// 添加Fork标记，避免重名
		newFunc.Name = newFunc.Name + "_fork"
		newFunc.Title = newFunc.Title + " (Fork)"
	}

	// 设置创建者信息
	newFunc.CreatedBy = operator
	newFunc.UpdatedBy = operator
	newFunc.ForkFromID = &sourceFunc.ID
	newFunc.ForkFromUser = sourceFunc.User
	newFunc.User = operator

	// 创建新函数
	if err := r.db.Create(&newFunc).Error; err != nil {
		return nil, err
	}

	// 创建版本记录
	version := &model.FuncVersion{
		FuncID:    newFunc.ID,
		Version:   "1.0.0",
		CreatedBy: operator,
		CreatedAt: newFunc.CreatedAt,
		Comment:   "从 " + sourceFunc.Name + " Fork",
	}

	// 保存版本记录
	if err := r.SaveVersion(ctx, version); err != nil {
		logger.Warn(ctx, "保存Fork函数版本失败", zap.Error(err), zap.Int64("func_id", newFunc.ID))
		// 不返回错误，因为函数已创建成功
	}

	return &newFunc, nil
}

// GetByTree 获取指定目录下的所有函数
func (r *RunnerFuncRepo) GetByTree(ctx context.Context, treeID int64) ([]model.RunnerFunc, error) {
	logger.Debug(ctx, "获取目录下的所有函数", zap.Int64("tree_id", treeID))
	var runnerFuncs []model.RunnerFunc
	err := r.db.WithContext(ctx).Where("tree_id = ?", treeID).Find(&runnerFuncs).Error
	if err != nil {
		logger.Error(ctx, "获取目录下的函数失败", err, zap.Int64("tree_id", treeID))
	}
	return runnerFuncs, err
}

// GetByName 根据名称获取函数
func (r *RunnerFuncRepo) GetByName(ctx context.Context, runnerID int64, name string) (*model.RunnerFunc, error) {
	logger.Debug(ctx, "根据名称获取函数", zap.Int64("runner_id", runnerID), zap.String("name", name))
	var runnerFunc model.RunnerFunc
	err := r.db.WithContext(ctx).Where("runner_id = ? AND name = ?", runnerID, name).First(&runnerFunc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称获取函数失败", err, zap.Int64("runner_id", runnerID), zap.String("name", name))
		return nil, err
	}
	return &runnerFunc, nil
}

// BatchCreate 批量创建函数
func (r *RunnerFuncRepo) BatchCreate(ctx context.Context, runnerFuncs []model.RunnerFunc) error {
	logger.Debug(ctx, "开始批量创建函数", zap.Int("count", len(runnerFuncs)))
	return r.db.WithContext(ctx).Create(&runnerFuncs).Error
}

// UpdateStatus 更新函数状态
func (r *RunnerFuncRepo) UpdateStatus(ctx context.Context, id int64, status int) error {
	logger.Debug(ctx, "更新函数状态", zap.Int64("id", id), zap.Int("status", status))
	return r.db.WithContext(ctx).Model(&model.RunnerFunc{}).Where("id = ?", id).Update("status", status).Error
}

// SaveVersion 保存函数版本
func (r *RunnerFuncRepo) SaveVersion(ctx context.Context, version *model.FuncVersion) error {
	logger.Debug(ctx, "保存函数版本", zap.Int64("func_id", version.FuncID), zap.String("version", version.Version))
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersions 获取函数版本列表
func (r *RunnerFuncRepo) GetVersions(ctx context.Context, funcID int64) ([]model.FuncVersion, error) {
	logger.Debug(ctx, "获取函数版本列表", zap.Int64("func_id", funcID))
	var versions []model.FuncVersion
	err := r.db.WithContext(ctx).Where("func_id = ?", funcID).Order("created_at DESC").Find(&versions).Error
	if err != nil {
		logger.Error(ctx, "获取函数版本列表失败", err, zap.Int64("func_id", funcID))
		return nil, err
	}
	return versions, nil
}
