package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/runcher/pkg/dto/coder"
	"strings"
	"time"

	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/repo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceTree 服务树服务
type ServiceTree struct {
	repo       *repo.ServiceTreeRepo
	runnerRepo *repo.RunnerRepo
}

// NewServiceTree 创建ServiceTree服务
func NewServiceTree(db *gorm.DB) *ServiceTree {
	return &ServiceTree{
		repo:       repo.NewServiceTreeRepo(db),
		runnerRepo: repo.NewRunnerRepo(db),
	}
}

// GetRepo 获取仓库实例
func (s *ServiceTree) GetRepo() *repo.ServiceTreeRepo {
	return s.repo
}

// Create 创建服务树
func (s *ServiceTree) Create(ctx context.Context, serviceTree *model.ServiceTree) error {
	logger.Debug(ctx, "开始创建服务树", zap.String("name", serviceTree.Name))

	// 业务逻辑校验
	if serviceTree.Title == "" {
		return errors.New("标题不能为空")
	}
	if serviceTree.Name == "" {
		return errors.New("名称不能为空")
	}

	// 检查同级目录下名称是否已存在
	existing, err := s.repo.GetByName(ctx, serviceTree.ParentID, serviceTree.Name)
	if err != nil {
		logger.Error(ctx, "检查服务树名称失败", err,
			zap.String("name", serviceTree.Name),
			zap.Any("parent_id", serviceTree.ParentID))
		return fmt.Errorf("检查名称失败: %w", err)
	}
	if existing != nil {
		return errors.New("同级目录下名称已存在")
	}

	if serviceTree.ParentID == 0 {
		return fmt.Errorf("ParentID 不能为0")
	}

	parent, err := s.repo.Get(ctx, serviceTree.ParentID)
	if err != nil {
		logger.Error(ctx, "获取父级目录失败", err, zap.Any("parent_id", serviceTree.ParentID))
		return fmt.Errorf("获取父级目录失败: %w", err)
	}
	if parent == nil {
		logger.Info(ctx, "父级目录不存在", zap.Any("parent_id", serviceTree.ParentID))
		return errors.New("父级目录不存在")
	}

	serviceTree.RunnerID = parent.RunnerID
	serviceTree.User = parent.User

	gotRunner, err := s.runnerRepo.Get(ctx, serviceTree.RunnerID)
	if err != nil {
		return err
	}
	// 创建服务树（需要先创建以获得ID）
	if err := s.repo.Create(ctx, serviceTree); err != nil {
		logger.Error(ctx, "创建服务树失败", err)
		return fmt.Errorf("创建服务树失败: %w", err)
	}

	// 使用ID构建FullIDPath
	serviceTree.FullIDPath = parent.FullIDPath + "/" + fmt.Sprintf("%d", serviceTree.ID)

	// 构建FullNamePath
	serviceTree.FullNamePath = parent.FullNamePath + "/" + serviceTree.Name

	// 设置当前目录的级别
	serviceTree.Level = parent.Level + 1

	// 更新路径字段
	if err := s.repo.Update(ctx, serviceTree.ID, &model.ServiceTree{
		FullIDPath:   serviceTree.FullIDPath,
		FullNamePath: serviceTree.FullNamePath,
		Level:        serviceTree.Level,
	}); err != nil {
		logger.Error(ctx, "更新服务树路径失败", err)
		// 不返回错误，因为目录已经创建成功
	}

	if serviceTree.Type == model.ServiceTreeTypePackage {
		runcherServiceIns := GetRuncherService()

		runner, err := runnerproject.NewRunner(gotRunner.User, gotRunner.Name, gotRunner.Version)
		if err != nil {
			return err
		}
		runner.Language = "go"
		pkg := &coder.BizPackage{
			Runner:         runner,
			AbsPackagePath: serviceTree.GetSubFullPath(),
			Language:       gotRunner.Language,
			EnName:         serviceTree.Name,
			CnName:         serviceTree.Title,
			Desc:           serviceTree.Description,
		}

		pkgResp, err := runcherServiceIns.AddBizPackage2(ctx, pkg)
		if err != nil {
			logger.Errorf(ctx, "runcherServiceIns.AddBizPackage err:%s req:%+v  resp:%+v", err.Error(), pkg, pkgResp)
		}
	}

	//_, err = runcherServiceIns.AddBizPackage(ctx, serviceTree.RunnerID, serviceTree.Name, serviceTree.Title, "", serviceTree.ID, true)
	//if err != nil {
	//	logger.Errorf(ctx, "runcherServiceIns.AddBizPackage err:%s", err.Error())
	//	err = nil
	//}

	// 更新父级目录的子目录数量
	if err := s.repo.UpdateChildrenCount(ctx, serviceTree.ParentID, 1); err != nil {
		logger.Error(ctx, "更新父级目录子目录数量失败", err, zap.Any("parent_id", serviceTree.ParentID))
		// 不返回错误，因为目录已经创建成功
	}

	logger.Info(ctx, "创建服务树成功", zap.Any("id", serviceTree.ID), zap.String("name", serviceTree.Name))
	return nil
}

// Get 获取服务树详情
func (s *ServiceTree) Get(ctx context.Context, id int64) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取服务树详情", zap.Int64("id", id))
	serviceTree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树详情失败", err, zap.Int64("id", id))
		return nil, fmt.Errorf("获取服务树详情失败: %w", err)
	}

	if serviceTree == nil {
		logger.Info(ctx, "服务树不存在", zap.Int64("id", id))
		return nil, nil
	}

	logger.Debug(ctx, "获取服务树详情成功", zap.Int64("id", id))
	return serviceTree, nil
}
func (s *ServiceTree) Children(ctx context.Context, parentId int64) ([]*model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取服务树详情", zap.Int64("parentId", parentId))
	serviceTrees, err := s.repo.Children(ctx, parentId)
	if err != nil {
		logger.Error(ctx, "获取服务树详情失败", err, zap.Int64("parentId", parentId))
		return nil, fmt.Errorf("获取服务树详情失败: %w", err)
	}
	return serviceTrees, nil
}

// Update 更新服务树
func (s *ServiceTree) Update(ctx context.Context, id int64, serviceTree *model.ServiceTree) error {
	logger.Debug(ctx, "开始更新服务树", zap.Int64("id", id))
	// 先检查是否存在
	existingTree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取服务树失败: %w", err)
	}
	if existingTree == nil {
		logger.Info(ctx, "目录不存在", zap.Int64("id", id))
		return errors.New("目录不存在")
	}

	// 业务逻辑校验
	if serviceTree.Title == "" {
		return errors.New("标题不能为空")
	}

	// 如果名称发生变更，检查名称是否已存在并更新路径
	if serviceTree.Name != "" && serviceTree.Name != existingTree.Name {
		existing, err := s.repo.GetByName(ctx, int64(existingTree.ParentID), serviceTree.Name)
		if err != nil {
			logger.Error(ctx, "检查服务树名称失败", err,
				zap.String("name", serviceTree.Name),
				zap.Any("parent_id", existingTree.ParentID))
			return fmt.Errorf("检查名称失败: %w", err)
		}
		if existing != nil && existing.ID != id {
			return errors.New("同级目录下名称已存在")
		}

		// 更新完整名称路径
		// ID路径不变，只更新名称路径
		if existingTree.ParentID != 0 {
			parent, err := s.repo.Get(ctx, int64(existingTree.ParentID))
			if err != nil {
				logger.Error(ctx, "获取父级目录失败", err, zap.Any("parent_id", existingTree.ParentID))
				return fmt.Errorf("获取父级目录失败: %w", err)
			}
			if parent != nil {
				serviceTree.FullNamePath = parent.FullNamePath + "/" + serviceTree.Name
			}
		} else {
			serviceTree.FullNamePath = serviceTree.Name
		}

		// 需要更新所有子目录的名称路径
		if err := s.updateChildrenNamePath(ctx, id, existingTree.FullNamePath, serviceTree.FullNamePath); err != nil {
			logger.Error(ctx, "更新子目录名称路径失败", err, zap.Int64("id", id))
			// 继续执行，不中断更新
		}
	}

	// 设置更新时间
	serviceTree.UpdatedAt = model.Time(time.Now())

	// 更新服务树
	if err := s.repo.Update(ctx, id, serviceTree); err != nil {
		logger.Error(ctx, "更新服务树失败", err, zap.Int64("id", id))
		return fmt.Errorf("更新服务树失败: %w", err)
	}

	logger.Info(ctx, "更新服务树成功", zap.Int64("id", id))
	return nil
}

// updateChildrenNamePath 更新所有子目录的名称路径
func (s *ServiceTree) updateChildrenNamePath(ctx context.Context, parentID int64, oldParentPath, newParentPath string) error {
	logger.Debug(ctx, "开始更新子目录名称路径",
		zap.Any("parent_id", parentID),
		zap.Any("old_path", oldParentPath),
		zap.Any("new_path", newParentPath))

	// 获取所有子目录
	children, err := s.repo.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	for _, child := range children {
		// 计算新的名称路径
		newChildPath := strings.Replace(child.FullNamePath, oldParentPath, newParentPath, 1)

		// 更新子目录的名称路径
		if err := s.repo.Update(ctx, int64(child.ID), &model.ServiceTree{
			FullNamePath: newChildPath,
		}); err != nil {
			logger.Error(ctx, "更新子目录名称路径失败", err, zap.Any("child_id", child.ID))
			continue
		}

		// 递归更新子目录的子目录
		if err := s.updateChildrenNamePath(ctx, int64(child.ID), child.FullNamePath, newChildPath); err != nil {
			logger.Error(ctx, "递归更新子目录名称路径失败", err, zap.Any("child_id", child.ID))
			continue
		}
	}

	return nil
}

// Delete 删除服务树
func (s *ServiceTree) Delete(ctx context.Context, id int64, operator string) error {
	logger.Debug(ctx, "开始删除服务树", zap.Any("id", id))
	// 先检查是否存在
	existingTree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树失败", err, zap.Any("id", id))
		return fmt.Errorf("获取服务树失败: %w", err)
	}
	if existingTree == nil {
		logger.Info(ctx, "目录不存在", zap.Any("id", id))
		return errors.New("目录不存在")
	}

	// 检查是否有子目录
	childrenCount, err := s.repo.CountChildren(ctx, id)
	if err != nil {
		logger.Error(ctx, "统计子目录数量失败", err, zap.Any("id", id))
		return fmt.Errorf("检查子目录失败: %w", err)
	}
	if childrenCount > 0 {
		logger.Info(ctx, "该目录下有子目录，无法删除", zap.Any("id", id), zap.Any("children_count", childrenCount))
		return errors.New("该目录下有子目录，无法删除")
	}

	// 检查是否有函数
	funcCount, err := s.repo.CountFunctions(ctx, id)
	if err != nil {
		logger.Error(ctx, "统计目录下的函数数量失败", err, zap.Any("id", id))
		return fmt.Errorf("检查目录下的函数失败: %w", err)
	}
	if funcCount > 0 {
		logger.Info(ctx, "该目录下有函数，无法删除", zap.Any("id", id), zap.Any("func_count", funcCount))
		return errors.New("该目录下有函数，无法删除")
	}

	// 设置删除者
	if err := s.repo.SetDeletedBy(ctx, id, operator); err != nil {
		logger.Error(ctx, "设置服务树删除者失败", err, zap.Any("id", id))
		return fmt.Errorf("设置删除者失败: %w", err)
	}

	// 删除服务树
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error(ctx, "删除服务树失败", err, zap.Any("id", id))
		return fmt.Errorf("删除服务树失败: %w", err)
	}

	// 如果有父级目录，需要更新父级目录的子目录数量
	if existingTree.ParentID != 0 {
		if err := s.repo.UpdateChildrenCount(ctx, int64(existingTree.ParentID), -1); err != nil {
			logger.Error(ctx, "更新父级目录子目录数量失败", err, zap.Any("parent_id", existingTree.ParentID))
			// 不返回错误，因为目录已经删除成功
		}
	}

	logger.Info(ctx, "删除服务树成功", zap.Any("id", id))
	return nil
}

// List 获取服务树列表
func (s *ServiceTree) List(ctx context.Context, page, pageSize int, conditions map[string]interface{}) ([]model.ServiceTree, int64, error) {
	logger.Debug(ctx, "开始获取服务树列表", zap.Any("page", page), zap.Any("pageSize", pageSize))
	offset := (page - 1) * pageSize
	trees, total, err := s.repo.List(ctx, offset, pageSize, conditions)
	if err != nil {
		logger.Error(ctx, "获取服务树列表失败", err)
		return nil, 0, fmt.Errorf("获取服务树列表失败: %w", err)
	}

	logger.Debug(ctx, "获取服务树列表成功", zap.Any("total", total))
	return trees, total, nil
}

// GetChildren 获取子目录列表
func (s *ServiceTree) GetChildren(ctx context.Context, parentID int64) ([]model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取子目录列表", zap.Any("parent_id", parentID))

	// 如果不是根目录，先检查父级目录是否存在
	if parentID != 0 {
		parent, err := s.repo.Get(ctx, parentID)
		if err != nil {
			logger.Error(ctx, "获取父级目录失败", err, zap.Any("parent_id", parentID))
			return nil, fmt.Errorf("获取父级目录失败: %w", err)
		}
		if parent == nil {
			logger.Info(ctx, "父级目录不存在", zap.Any("parent_id", parentID))
			return nil, errors.New("父级目录不存在")
		}
	}

	// 获取子目录列表
	children, err := s.repo.GetChildren(ctx, parentID)
	if err != nil {
		logger.Error(ctx, "获取子目录列表失败", err, zap.Any("parent_id", parentID))
		return nil, fmt.Errorf("获取子目录列表失败: %w", err)
	}

	logger.Debug(ctx, "获取子目录列表成功", zap.Any("parent_id", parentID), zap.Any("count", len(children)))
	return children, nil
}

func (s *ServiceTree) GetChildrenByFullPath(ctx context.Context, user string, fullPath string) ([]model.ServiceTree, error) {
	// 如果不是根目录，先检查父级目录是否存在
	tree, err := s.repo.GetByFullPath(ctx, user, fullPath)
	if err != nil {
		return nil, err
	}
	// 获取子目录列表
	children, err := s.repo.GetChildren(ctx, tree.ID)
	if err != nil {
		return nil, fmt.Errorf("获取子目录列表失败: %w", err)
	}
	return children, nil
}
func (s *ServiceTree) GetByFullPath(ctx context.Context, user string, fullPath string) (*model.ServiceTree, error) {
	// 如果不是根目录，先检查父级目录是否存在
	tree, err := s.repo.GetByFullPath(ctx, user, fullPath)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// GetPath 获取服务树路径
func (s *ServiceTree) GetPath(ctx context.Context, id int64) ([]model.ServiceTree, error) {
	logger.Debug(ctx, "开始获取服务树路径", zap.Any("id", id))

	// 先检查目录是否存在
	tree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树失败", err, zap.Any("id", id))
		return nil, fmt.Errorf("获取服务树失败: %w", err)
	}
	if tree == nil {
		logger.Info(ctx, "目录不存在", zap.Any("id", id))
		return nil, errors.New("目录不存在")
	}

	// 获取路径
	path, err := s.repo.GetPath(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树路径失败", err, zap.Any("id", id))
		return nil, fmt.Errorf("获取服务树路径失败: %w", err)
	}

	logger.Debug(ctx, "获取服务树路径成功", zap.Any("id", id), zap.Any("path_length", len(path)))
	return path, nil
}

// Fork 复制服务树
func (s *ServiceTree) Fork(ctx context.Context, id int64, targetParentID int64, newName string, operator string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始复制服务树",
		zap.Any("id", id),
		zap.Any("target_parent_id", targetParentID),
		zap.Any("new_name", newName))

	// 先检查源目录是否存在
	sourceTree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取源目录失败", err, zap.Any("id", id))
		return nil, fmt.Errorf("获取源目录失败: %w", err)
	}
	if sourceTree == nil {
		logger.Info(ctx, "源目录不存在", zap.Any("id", id))
		return nil, errors.New("源目录不存在")
	}

	// 检查目标父级目录是否存在
	if targetParentID != 0 {
		targetParent, err := s.repo.Get(ctx, targetParentID)
		if err != nil {
			logger.Error(ctx, "获取目标父级目录失败", err, zap.Any("target_parent_id", targetParentID))
			return nil, fmt.Errorf("获取目标父级目录失败: %w", err)
		}
		if targetParent == nil {
			logger.Info(ctx, "目标父级目录不存在", zap.Any("target_parent_id", targetParentID))
			return nil, errors.New("目标父级目录不存在")
		}
	}

	// 确定新名称
	actualNewName := newName
	if actualNewName == "" {
		actualNewName = sourceTree.Name + "_fork"
	}

	// 检查目标目录下是否已存在同名目录
	existing, err := s.repo.GetByName(ctx, targetParentID, actualNewName)
	if err != nil {
		logger.Error(ctx, "检查目标目录名称失败", err,
			zap.Any("name", actualNewName),
			zap.Any("parent_id", targetParentID))
		return nil, fmt.Errorf("检查目标目录名称失败: %w", err)
	}
	if existing != nil {
		logger.Info(ctx, "目标目录下已存在同名目录",
			zap.Any("name", actualNewName),
			zap.Any("parent_id", targetParentID))
		return nil, errors.New("目标目录下已存在同名目录")
	}

	// 创建新目录
	newTree := *sourceTree
	newTree.ID = 0 // 重置ID，让数据库自动生成
	newTree.ParentID = int64(int(targetParentID))

	// 设置名称和标题
	newTree.Name = actualNewName
	if newName != "" {
		newTree.Title = newName
	} else {
		newTree.Title = sourceTree.Title + " (Fork)"
	}

	// 设置创建者和更新者信息
	newTree.CreatedBy = operator
	newTree.UpdatedBy = operator

	// 设置fork来源信息
	newTree.ForkFromID = &sourceTree.ID

	// 设置用户
	newTree.User = operator

	// 设置创建时间和更新时间
	now := time.Now()
	newTree.CreatedAt = model.Time(now)
	newTree.UpdatedAt = model.Time(now)

	// 重置子目录数量
	newTree.ChildrenCount = 0

	// 先创建获取ID
	if err := s.repo.Create(ctx, &newTree); err != nil {
		logger.Error(ctx, "创建Fork目录失败", err)
		return nil, fmt.Errorf("创建Fork目录失败: %w", err)
	}

	// 更新FullIDPath、FullNamePath和Level
	if targetParentID != 0 {
		targetParent, _ := s.repo.Get(ctx, targetParentID)
		// 使用ID构建FullIDPath
		newTree.FullIDPath = targetParent.FullIDPath + "/" + fmt.Sprintf("%d", newTree.ID)
		// 构建FullNamePath
		newTree.FullNamePath = targetParent.FullNamePath + "/" + newTree.Name
		newTree.Level = targetParent.Level + 1
	} else {
		// 使用ID构建FullIDPath
		newTree.FullIDPath = fmt.Sprintf("%d", newTree.ID)
		// 构建FullNamePath
		newTree.FullNamePath = newTree.Name
		newTree.Level = 0
	}

	// 更新路径字段
	if err := s.repo.Update(ctx, newTree.ID, &model.ServiceTree{
		FullIDPath:   newTree.FullIDPath,
		FullNamePath: newTree.FullNamePath,
		Level:        newTree.Level,
	}); err != nil {
		logger.Error(ctx, "更新Fork目录路径失败", err)
		// 不返回错误，因为新目录已创建成功
	}

	// 如果有父级目录，更新父级目录的子目录数量
	if targetParentID != 0 {
		if err := s.repo.UpdateChildrenCount(ctx, targetParentID, 1); err != nil {
			logger.Error(ctx, "更新父级目录子目录数量失败", err, zap.Any("parent_id", targetParentID))
			// 不返回错误，因为新目录已创建成功
		}
	}

	logger.Info(ctx, "Fork目录成功",
		zap.Any("source_id", sourceTree.ID),
		zap.Any("new_id", newTree.ID),
		zap.Any("name", newTree.Name))

	return &newTree, nil
}

// UpdateSort 更新服务树排序
func (s *ServiceTree) UpdateSort(ctx context.Context, id int64, sort int) error {
	logger.Debug(ctx, "开始更新服务树排序", zap.Int64("id", id), zap.Int("sort", sort))

	// 先检查目录是否存在
	tree, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.Error(ctx, "获取服务树失败", err, zap.Int64("id", id))
		return fmt.Errorf("获取服务树失败: %w", err)
	}
	if tree == nil {
		logger.Info(ctx, "目录不存在", zap.Int64("id", id))
		return errors.New("目录不存在")
	}

	// 更新排序
	if err := s.repo.UpdateSort(ctx, id, sort); err != nil {
		logger.Error(ctx, "更新服务树排序失败", err, zap.Int64("id", id), zap.Int("sort", sort))
		return fmt.Errorf("更新排序失败: %w", err)
	}

	logger.Info(ctx, "更新服务树排序成功", zap.Int64("id", id), zap.Int("sort", sort))
	return nil
}

// GetByPath 根据路径获取服务树
func (s *ServiceTree) GetByPath(ctx context.Context, pathStr string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始根据路径获取服务树", zap.String("path", pathStr))

	// 将路径字符串分割为路径数组
	var path []string
	if pathStr != "" {
		// 去除开头和结尾的斜杠，然后按斜杠分割
		cleanPath := pathStr
		if cleanPath[0] == '/' {
			cleanPath = cleanPath[1:]
		}
		if len(cleanPath) > 0 && cleanPath[len(cleanPath)-1] == '/' {
			cleanPath = cleanPath[:len(cleanPath)-1]
		}

		if cleanPath != "" {
			path = strings.Split(cleanPath, "/")
		}
	}

	if len(path) == 0 {
		logger.Info(ctx, "路径为空")
		return nil, errors.New("路径为空")
	}

	// 获取目录
	tree, err := s.repo.GetByPath(ctx, path)
	if err != nil {
		logger.Error(ctx, "根据路径获取服务树失败", err, zap.String("path", pathStr))
		return nil, fmt.Errorf("根据路径获取服务树失败: %w", err)
	}

	if tree == nil {
		logger.Info(ctx, "路径对应的目录不存在", zap.String("path", pathStr))
		return nil, nil
	}

	logger.Debug(ctx, "根据路径获取服务树成功", zap.String("path", pathStr), zap.Int64("id", tree.ID))
	return tree, nil
}

// GetByNamePath 根据名称路径获取服务树
func (s *ServiceTree) GetByNamePath(ctx context.Context, namePath string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始根据名称路径获取服务树", zap.String("name_path", namePath))

	if namePath == "" {
		logger.Info(ctx, "路径为空")
		return nil, errors.New("路径为空")
	}

	// 获取目录
	tree, err := s.repo.GetByNamePath(ctx, namePath)
	if err != nil {
		logger.Error(ctx, "根据名称路径获取服务树失败", err, zap.String("name_path", namePath))
		return nil, fmt.Errorf("根据名称路径获取服务树失败: %w", err)
	}

	if tree == nil {
		logger.Info(ctx, "名称路径对应的服务树不存在", zap.String("name_path", namePath))
		return nil, nil
	}

	logger.Info(ctx, "根据名称路径获取服务树成功", zap.String("name_path", namePath), zap.Int64("id", tree.ID))
	return tree, nil
}

// GetByIDPath 根据ID路径获取服务树
func (s *ServiceTree) GetByIDPath(ctx context.Context, idPath string) (*model.ServiceTree, error) {
	logger.Debug(ctx, "开始根据ID路径获取服务树", zap.String("id_path", idPath))

	if idPath == "" {
		logger.Info(ctx, "路径为空")
		return nil, errors.New("路径为空")
	}

	// 获取目录
	tree, err := s.repo.GetByIDPath(ctx, idPath)
	if err != nil {
		logger.Error(ctx, "根据ID路径获取服务树失败", err, zap.String("id_path", idPath))
		return nil, fmt.Errorf("根据ID路径获取服务树失败: %w", err)
	}

	if tree == nil {
		logger.Info(ctx, "ID路径对应的服务树不存在", zap.String("id_path", idPath))
		return nil, nil
	}

	logger.Info(ctx, "根据ID路径获取服务树成功", zap.String("id_path", idPath), zap.Int64("id", tree.ID))
	return tree, nil
}

// CreateWithTx 使用事务创建服务树
func (s *ServiceTree) CreateWithTx(ctx context.Context, tx *gorm.DB, serviceTree *model.ServiceTree) error {
	logger.Debug(ctx, "使用事务创建服务树", zap.String("name", serviceTree.Name))

	// 业务逻辑校验
	if serviceTree.Title == "" {
		return errors.New("标题不能为空")
	}
	if serviceTree.Name == "" {
		return errors.New("名称不能为空")
	}

	// 检查同级目录下名称是否已存在
	existing, err := s.repo.GetByName(ctx, serviceTree.ParentID, serviceTree.Name)
	if err != nil {
		logger.Error(ctx, "检查服务树名称失败", err,
			zap.String("name", serviceTree.Name),
			zap.Any("parent_id", serviceTree.ParentID))
		return fmt.Errorf("检查名称失败: %w", err)
	}
	if existing != nil {
		return errors.New("同级目录下名称已存在")
	}

	// 如果有父级目录，需要更新父级目录的子目录数量并验证其存在性
	if serviceTree.ParentID != 0 {
		parent, err := s.repo.Get(ctx, serviceTree.ParentID)
		if err != nil {
			logger.Error(ctx, "获取父级目录失败", err, zap.Any("parent_id", serviceTree.ParentID))
			return fmt.Errorf("获取父级目录失败: %w", err)
		}
		if parent == nil {
			logger.Info(ctx, "父级目录不存在", zap.Any("parent_id", serviceTree.ParentID))
			return errors.New("父级目录不存在")
		}

		// 使用事务创建服务树
		if err := s.repo.CreateWithTx(ctx, tx, serviceTree); err != nil {
			logger.Error(ctx, "使用事务创建服务树失败", err)
			return fmt.Errorf("创建服务树失败: %w", err)
		}

		// 使用ID构建FullIDPath
		serviceTree.FullIDPath = parent.FullIDPath + "/" + fmt.Sprintf("%d", serviceTree.ID)

		// 构建FullNamePath
		serviceTree.FullNamePath = parent.FullNamePath + "/" + serviceTree.Name

		// 设置当前目录的级别
		serviceTree.Level = parent.Level + 1

		// 更新路径字段
		if err := s.repo.UpdateWithTx(ctx, tx, serviceTree.ID, &model.ServiceTree{
			FullIDPath:   serviceTree.FullIDPath,
			FullNamePath: serviceTree.FullNamePath,
			Level:        serviceTree.Level,
		}); err != nil {
			logger.Error(ctx, "更新服务树路径失败", err)
			return fmt.Errorf("更新服务树路径失败: %w", err)
		}

		// 更新父级目录的子目录数量
		if err := s.repo.UpdateChildrenCountWithTx(ctx, tx, serviceTree.ParentID, 1); err != nil {
			logger.Error(ctx, "更新父级目录子目录数量失败", err, zap.Any("parent_id", serviceTree.ParentID))
			return fmt.Errorf("更新父级目录子目录数量失败: %w", err)
		}
	} else {
		// 根目录
		// 使用事务创建服务树
		if err := s.repo.CreateWithTx(ctx, tx, serviceTree); err != nil {
			logger.Error(ctx, "使用事务创建服务树失败", err)
			return fmt.Errorf("创建服务树失败: %w", err)
		}

		// 使用ID构建FullIDPath
		serviceTree.FullIDPath = fmt.Sprintf("%d", serviceTree.ID)

		// 构建FullNamePath
		serviceTree.FullNamePath = serviceTree.Name

		// 根目录级别为0
		serviceTree.Level = 0

		// 更新路径字段
		if err := s.repo.UpdateWithTx(ctx, tx, serviceTree.ID, &model.ServiceTree{
			FullIDPath:   serviceTree.FullIDPath,
			FullNamePath: serviceTree.FullNamePath,
			Level:        serviceTree.Level,
		}); err != nil {
			logger.Error(ctx, "更新服务树路径失败", err)
			return fmt.Errorf("更新服务树路径失败: %w", err)
		}
	}

	logger.Info(ctx, "使用事务创建服务树成功", zap.Any("id", serviceTree.ID), zap.String("name", serviceTree.Name))
	return nil
}

func (s *ServiceTree) UpdateWithTx(ctx context.Context, tx *gorm.DB, treeId int64, serviceTree *model.ServiceTree) error {
	return tx.WithContext(ctx).Where("id =?", treeId).Updates(serviceTree).Error
}
