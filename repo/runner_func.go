package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/logger"
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

// GetByFullPath 创建函数
func (r *RunnerFuncRepo) GetByFullPath(ctx context.Context, method string, fullPath string) (runnerFunc *model.RunnerFunc, err error) {
	err = r.db.WithContext(ctx).Where("method = ? AND path = ?", strings.ToUpper(method), strings.TrimPrefix(fullPath, "/")).First(&runnerFunc).Error
	if err != nil {
		return nil, err
	}
	return runnerFunc, nil
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
		FuncID:  newFunc.ID,
		Version: "1.0.0",
		Comment: "从 " + sourceFunc.Name + " Fork",
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

// GetUserRecentFuncRecords 获取用户最近执行过的函数记录（去重）
func (r *RunnerFuncRepo) GetUserRecentFuncRecords(ctx context.Context, user string, page, pageSize int) ([]model.FuncRunRecord, int64, error) {
	logger.Debug(ctx, "开始获取用户最近执行函数记录", zap.String("user", user), zap.Int("page", page), zap.Int("pageSize", pageSize))

	// 构建子查询，获取每个函数的最新执行记录ID
	subQuery := r.db.WithContext(ctx).
		Model(&model.FuncRunRecord{}).
		Select("MAX(id) as max_id").
		Joins("JOIN runner_func rf ON func_run_record.func_id = rf.id").
		Where("rf.user = ?", user).
		Group("func_run_record.func_id")

	// 获取总数
	var total int64
	countQuery := r.db.WithContext(ctx).
		Model(&model.FuncRunRecord{}).
		Joins("JOIN runner_func rf ON func_run_record.func_id = rf.id").
		Where("rf.user = ? AND func_run_record.id IN (?)", user, subQuery)

	if err := countQuery.Count(&total).Error; err != nil {
		logger.Error(ctx, "获取用户最近执行函数记录总数失败", err, zap.String("user", user))
		return nil, 0, err
	}

	// 获取分页数据
	var records []model.FuncRunRecord
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).
		Model(&model.FuncRunRecord{}).
		Joins("JOIN runner_func rf ON func_run_record.func_id = rf.id").
		Where("rf.user = ? AND func_run_record.id IN (?)", user, subQuery).
		Order("func_run_record.end_ts DESC").
		Offset(offset).
		Limit(pageSize)

	if err := query.Find(&records).Error; err != nil {
		logger.Error(ctx, "获取用户最近执行函数记录失败", err, zap.String("user", user))
		return nil, 0, err
	}

	logger.Info(ctx, "获取用户最近执行函数记录成功", zap.String("user", user), zap.Int("count", len(records)), zap.Int64("total", total))
	return records, total, nil
}

// GetFuncRunRecordWithDetails 获取函数执行记录及其关联的详细信息
func (r *RunnerFuncRepo) GetFuncRunRecordWithDetails(ctx context.Context, recordID int64) (*model.FuncRunRecord, *model.RunnerFunc, *model.Runner, *model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取函数执行记录详细信息", zap.Int64("record_id", recordID))

	var record model.FuncRunRecord
	if err := r.db.WithContext(ctx).First(&record, recordID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil, nil
		}
		logger.Error(ctx, "获取函数执行记录失败", err, zap.Int64("record_id", recordID))
		return nil, nil, nil, nil, err
	}

	// 获取函数信息
	var runnerFunc model.RunnerFunc
	if err := r.db.WithContext(ctx).First(&runnerFunc, record.FuncId).Error; err != nil {
		logger.Error(ctx, "获取函数信息失败", err, zap.Int64("func_id", record.FuncId))
		return &record, nil, nil, nil, err
	}

	// 获取Runner信息
	var runner model.Runner
	if err := r.db.WithContext(ctx).First(&runner, runnerFunc.RunnerID).Error; err != nil {
		logger.Error(ctx, "获取Runner信息失败", err, zap.Int64("runner_id", runnerFunc.RunnerID))
		return &record, &runnerFunc, nil, nil, err
	}

	// 获取ServiceTree信息
	var serviceTree model.ServiceTree
	if err := r.db.WithContext(ctx).First(&serviceTree, runnerFunc.TreeID).Error; err != nil {
		logger.Error(ctx, "获取ServiceTree信息失败", err, zap.Int64("tree_id", runnerFunc.TreeID))
		return &record, &runnerFunc, &runner, nil, err
	}

	return &record, &runnerFunc, &runner, &serviceTree, nil
}

// GetUserFuncRunCount 获取用户函数执行次数统计
func (r *RunnerFuncRepo) GetUserFuncRunCount(ctx context.Context, user string, funcID int64) (int64, error) {
	logger.Debug(ctx, "开始获取用户函数执行次数", zap.String("user", user), zap.Int64("func_id", funcID))

	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.FuncRunRecord{}).
		Joins("JOIN runner_func rf ON func_run_record.func_id = rf.id").
		Where("rf.user = ? AND func_run_record.func_id = ?", user, funcID).
		Count(&count).Error

	if err != nil {
		logger.Error(ctx, "获取用户函数执行次数失败", err, zap.String("user", user), zap.Int64("func_id", funcID))
		return 0, err
	}

	return count, nil
}
