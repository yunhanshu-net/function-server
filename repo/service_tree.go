package repo

import (
	"context"
	"errors"
	"github.com/yunhanshu-net/api-server/pkg/db"

	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceTreeRepository ServiceTree仓库接口
type ServiceTreeRepository interface {
	Create(ctx context.Context, tree *model.ServiceTree) error
	Get(ctx context.Context, id int64) (*model.ServiceTree, error)
	Update(ctx context.Context, id int64, tree *model.ServiceTree) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.ServiceTree, int64, error)
	GetByName(ctx context.Context, parentID int64, name string) (*model.ServiceTree, error)
	GetChildren(ctx context.Context, parentID int64) ([]model.ServiceTree, error)
}

// ServiceTreeRepo ServiceTree仓库实现
type ServiceTreeRepo struct {
	db *gorm.DB
}

// NewServiceTreeRepo 创建ServiceTree仓库
func NewServiceTreeRepo(db *gorm.DB) *ServiceTreeRepo {
	return &ServiceTreeRepo{db: db}
}

// Create 创建ServiceTree
func (r *ServiceTreeRepo) Create(ctx context.Context, tree *model.ServiceTree) error {
	//这里先从ctx获取上层的db，假如存在事务的db，那就用事务来操作数据库，不存在再用默认的db对象
	return db.GetContextDB(ctx, r.db).Create(tree).Error
}

// Get 获取ServiceTree详情
func (r *ServiceTreeRepo) Get(ctx context.Context, id int64) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取ServiceTree", zap.Any("id", id))
	var tree model.ServiceTree
	err := r.db.WithContext(ctx).First(&tree, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info(ctx, "ServiceTree不存在", zap.Any("id", id))
			return nil, nil
		}
		logger.Error(ctx, "获取ServiceTree失败", err, zap.Any("id", id))
		return nil, err
	}

	if tree.ParentID == 0 {
		tree.Runner = new(model.Runner)
		err = r.db.WithContext(ctx).Model(&model.Runner{}).Where("tree_id = ?", tree.ID).First(tree.Runner).Error
		if err != nil {
			return nil, err
		}
	}
	return &tree, nil
}
func (r *ServiceTreeRepo) Children(ctx context.Context, parentId int64) ([]*model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取ServiceTree", zap.Any("parentId", parentId))
	var trees []*model.ServiceTree
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentId).Find(&trees).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info(ctx, "ServiceTree不存在", zap.Any("parentId", parentId))
			return nil, nil
		}
		logger.Error(ctx, "获取ServiceTree失败", err, zap.Any("parentId", parentId))
		return nil, err
	}
	return trees, nil
}

// Update 更新ServiceTree
func (r *ServiceTreeRepo) Update(ctx context.Context, id int64, tree *model.ServiceTree) error {
	logger.Debug(ctx, "开始更新ServiceTree", zap.Any("id", id))
	return r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).Updates(tree).Error
}

// Delete 删除ServiceTree
func (r *ServiceTreeRepo) Delete(ctx context.Context, id int64) error {
	logger.Debug(ctx, "开始删除ServiceTree", zap.Any("id", id))
	return r.db.WithContext(ctx).Delete(&model.ServiceTree{}, id).Error
}

// List 获取ServiceTree列表
func (r *ServiceTreeRepo) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.ServiceTree, int64, error) {
	logger.Debug(ctx, "开始获取ServiceTree列表", zap.Any("page", page), zap.Any("pageSize", pageSize))
	var total int64
	var treeList []model.ServiceTree

	query := r.db.WithContext(ctx).Model(&model.ServiceTree{})

	// 应用查询条件
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		logger.Error(ctx, "获取ServiceTree总数失败", err)
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&treeList).Error; err != nil {
		logger.Error(ctx, "获取ServiceTree列表失败", err)
		return nil, 0, err
	}

	return treeList, total, nil
}

// GetByName 根据名称和父ID获取ServiceTree
func (r *ServiceTreeRepo) GetByName(ctx context.Context, parentID int64, name string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "根据名称获取ServiceTree", zap.Any("parent_id", parentID), zap.Any("name", name))
	var tree model.ServiceTree

	query := r.db.WithContext(ctx).Where("name = ?", name)
	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.First(&tree).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称获取ServiceTree失败", err, zap.Any("parent_id", parentID), zap.Any("name", name))
		return nil, err
	}
	return &tree, nil
}

// GetChildren 获取子节点
func (r *ServiceTreeRepo) GetChildren(ctx context.Context, parentID int64) ([]model.ServiceTree, error) {
	logger.Debug(ctx, "获取ServiceTree子节点", zap.Any("parent_id", parentID))
	var children []model.ServiceTree

	query := r.db.WithContext(ctx)
	if parentID > 0 {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Find(&children).Error
	if err != nil {
		logger.Error(ctx, "获取ServiceTree子节点失败", err, zap.Any("parent_id", parentID))
		return nil, err
	}

	return children, nil
}

// UpdateChildrenCount 更新子目录数量
func (r *ServiceTreeRepo) UpdateChildrenCount(ctx context.Context, id int64, increment int) error {
	logger.Debug(ctx, "更新子目录数量", zap.Any("id", id), zap.Any("increment", increment))
	if increment > 0 {
		return r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).
			Update("children_count", gorm.Expr("children_count + ?", increment)).Error
	} else {
		return r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).
			Update("children_count", gorm.Expr("children_count - ?", -increment)).Error
	}
}

// SetDeletedBy 设置删除者
func (r *ServiceTreeRepo) SetDeletedBy(ctx context.Context, id int64, deletedBy string) error {
	logger.Debug(ctx, "设置服务树删除者", zap.Any("id", id), zap.Any("deleted_by", deletedBy))
	return r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).Update("deleted_by", deletedBy).Error
}

// CountChildren 统计子目录数量
func (r *ServiceTreeRepo) CountChildren(ctx context.Context, parentID int64) (int64, error) {
	logger.Debug(ctx, "统计子目录数量", zap.Any("parent_id", parentID))
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("parent_id = ?", parentID).Count(&count).Error
	if err != nil {
		logger.Error(ctx, "统计子目录数量失败", err, zap.Any("parent_id", parentID))
	}
	return count, err
}

// CountFunctions 统计目录下的函数数量
func (r *ServiceTreeRepo) CountFunctions(ctx context.Context, treeID int64) (int64, error) {
	logger.Debug(ctx, "统计目录下的函数数量", zap.Any("tree_id", treeID))
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RunnerFunc{}).Where("tree_id = ?", treeID).Count(&count).Error
	if err != nil {
		logger.Error(ctx, "统计目录下的函数数量失败", err, zap.Any("tree_id", treeID))
	}
	return count, err
}

// GetPath 获取服务树路径
func (r *ServiceTreeRepo) GetPath(ctx context.Context, id int64) ([]model.ServiceTree, error) {
	logger.Debug(ctx, "获取服务树路径", zap.Any("id", id))
	var path []model.ServiceTree
	current, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if current == nil {
		return path, nil
	}

	path = append(path, *current)

	// 递归获取父级目录，直到根目录
	for current.ParentID > 0 {
		parentID := int64(current.ParentID)
		parent, err := r.Get(ctx, parentID)
		if err != nil {
			logger.Error(ctx, "获取父级目录失败", err, zap.Any("parent_id", current.ParentID))
			return path, err
		}

		if parent == nil {
			break
		}

		path = append([]model.ServiceTree{*parent}, path...)
		current = parent
	}

	return path, nil
}

// UpdateSort 更新服务树排序
func (r *ServiceTreeRepo) UpdateSort(ctx context.Context, id int64, sort int) error {
	logger.Debug(ctx, "更新服务树排序", zap.Any("id", id), zap.Any("sort", sort))
	return r.db.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).Update("sort", sort).Error
}

// GetByPath 根据路径获取服务树
func (r *ServiceTreeRepo) GetByPath(ctx context.Context, path []string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "根据路径获取服务树", zap.Any("path", path))
	if len(path) == 0 {
		return nil, errors.New("路径为空")
	}

	var current *model.ServiceTree
	var err error
	parentID := int64(0) // 根目录

	for _, name := range path {
		current, err = r.GetByName(ctx, parentID, name)
		if err != nil {
			return nil, err
		}

		if current == nil {
			return nil, nil
		}

		// 显式转换uint到int64
		parentID = int64(current.ID)
	}

	return current, nil
}

// GetByIDPath 根据ID路径获取服务树
func (r *ServiceTreeRepo) GetByIDPath(ctx context.Context, idPath string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "根据ID路径获取服务树", zap.Any("id_path", idPath))
	if idPath == "" {
		return nil, errors.New("路径为空")
	}

	var tree model.ServiceTree
	// 尝试使用full_id_path查询
	err := r.db.WithContext(ctx).Where("full_id_path = ?", idPath).First(&tree).Error
	if err != nil {
		// 如果找不到记录，尝试使用full_path查询
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 尝试使用full_id_path字段查询
			err = r.db.WithContext(ctx).Where("full_id_path = ?", idPath).First(&tree).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, nil
				}
				logger.Error(ctx, "根据ID路径获取ServiceTree失败(full_id_path)", err, zap.Any("id_path", idPath))
				return nil, err
			}
			return &tree, nil
		}
		logger.Error(ctx, "根据ID路径获取ServiceTree失败(full_id_path)", err, zap.Any("id_path", idPath))
		return nil, err
	}

	return &tree, nil
}

// GetByNamePath 根据名称路径获取服务树
func (r *ServiceTreeRepo) GetByNamePath(ctx context.Context, namePath string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "根据名称路径获取服务树", zap.Any("name_path", namePath))
	if namePath == "" {
		return nil, errors.New("路径为空")
	}

	var tree model.ServiceTree
	err := r.db.WithContext(ctx).Where("full_name_path = ?", namePath).First(&tree).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logger.Error(ctx, "根据名称路径获取ServiceTree失败", err, zap.Any("name_path", namePath))
		return nil, err
	}

	return &tree, nil
}

// CreateWithTx 使用事务创建服务树
func (r *ServiceTreeRepo) CreateWithTx(ctx context.Context, tx *gorm.DB, tree *model.ServiceTree) error {
	logger.Debug(ctx, "使用事务创建服务树", zap.String("name", tree.Name))
	return tx.WithContext(ctx).Create(tree).Error
}

// UpdateWithTx 使用事务更新服务树
func (r *ServiceTreeRepo) UpdateWithTx(ctx context.Context, tx *gorm.DB, id int64, updates *model.ServiceTree) error {
	logger.Debug(ctx, "使用事务更新服务树", zap.Int64("id", id))
	return tx.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateChildrenCountWithTx 使用事务更新服务树子目录数量
func (r *ServiceTreeRepo) UpdateChildrenCountWithTx(ctx context.Context, tx *gorm.DB, id int64, increment int) error {
	logger.Debug(ctx, "使用事务更新服务树子目录数量", zap.Int64("id", id), zap.Int("increment", increment))

	// 先获取当前子目录数
	var tree model.ServiceTree
	if err := tx.WithContext(ctx).Select("children_count").Where("id = ?", id).First(&tree).Error; err != nil {
		return err
	}

	// 更新子目录数
	newCount := tree.ChildrenCount + increment
	if newCount < 0 {
		newCount = 0
	}

	return tx.WithContext(ctx).Model(&model.ServiceTree{}).Where("id = ?", id).Update("children_count", newCount).Error
}
