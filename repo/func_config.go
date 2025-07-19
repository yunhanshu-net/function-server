package repo

import (
	"context"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FuncConfigRepo 配置仓库
type FuncConfigRepo struct {
	db *gorm.DB
}

// NewFuncConfigRepo 创建配置仓库
func NewFuncConfigRepo(db *gorm.DB) *FuncConfigRepo {
	return &FuncConfigRepo{db: db}
}

// GetDB 获取数据库连接
func (r *FuncConfigRepo) GetDB() *gorm.DB {
	return r.db
}

// Create 创建配置
func (r *FuncConfigRepo) Create(ctx context.Context, config *model.FuncConfig) error {
	logger.Debug(ctx, "开始创建配置", zap.Int64("func_id", config.FuncID))
	return r.db.Create(config).Error
}

// GetByFuncID 根据函数ID获取配置
func (r *FuncConfigRepo) GetByFuncID(ctx context.Context, funcID int64) (*model.FuncConfig, error) {
	var config model.FuncConfig
	err := r.db.Where("func_id = ? AND is_active = ?", funcID, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetByConfigKey 根据配置键获取配置
func (r *FuncConfigRepo) GetByConfigKey(ctx context.Context, configKey string) (*model.FuncConfig, error) {
	var config model.FuncConfig
	err := r.db.Where("config_key = ? AND is_active = ?", configKey, true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// Update 更新配置
func (r *FuncConfigRepo) Update(ctx context.Context, id int64, config *model.FuncConfig) error {
	logger.Debug(ctx, "开始更新配置", zap.Int64("id", id))
	return r.db.Model(&model.FuncConfig{}).Where("id = ?", id).Updates(config).Error
}

// Delete 删除配置
func (r *FuncConfigRepo) Delete(ctx context.Context, id int64) error {
	logger.Debug(ctx, "开始删除配置", zap.Int64("id", id))
	return r.db.Delete(&model.FuncConfig{}, id).Error
}

// ListByFuncID 获取函数的所有配置版本
func (r *FuncConfigRepo) ListByFuncID(ctx context.Context, funcID int64) ([]model.FuncConfig, error) {
	var configs []model.FuncConfig
	err := r.db.Where("func_id = ?", funcID).Order("created_at DESC").Find(&configs).Error
	return configs, err
}

// SetInactive 设置配置为非激活状态
func (r *FuncConfigRepo) SetInactive(ctx context.Context, funcID int64) error {
	return r.db.Model(&model.FuncConfig{}).Where("func_id = ?", funcID).Update("is_active", false).Error
} 