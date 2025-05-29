package v1

import (
	"fmt"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/logger"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceTreeAPI ServiceTree API控制器
type ServiceTreeAPI struct {
	service *service.ServiceTree
}

// NewServiceTreeAPI 创建ServiceTree API控制器
func NewServiceTreeAPI(db *gorm.DB) *ServiceTreeAPI {
	return &ServiceTreeAPI{
		service: service.NewServiceTree(db),
	}
}

// Create 创建目录
func (api *ServiceTreeAPI) Create(c *gin.Context) {
	logger.Debug(c, "开始处理ServiceTree创建请求")

	var serviceTree model.ServiceTree
	if err := c.ShouldBindJSON(&serviceTree); err != nil {
		logger.Error(c, "解析ServiceTree参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	serviceTree.User = c.GetString("user")
	// 设置创建者信息，实际项目中应从JWT或Session获取
	serviceTree.CreatedBy = c.GetString("user")
	serviceTree.UpdatedBy = c.GetString("user")
	if serviceTree.Type != model.ServiceTreeTypePackage {
		logger.Error(c, "接口只能创建package类型的服务目录", fmt.Errorf("接口只能创建package类型的服务目录"))
		response.ParamError(c, "接口只能创建package类型的服务目录")
		return
	}

	// 调用服务层创建目录
	if err := api.service.CreateNode(c, &serviceTree); err != nil {
		logger.Error(c, "创建ServiceTree失败", err)
		response.ServerError(c, "创建目录失败: "+err.Error())
		return
	}

	logger.Info(c, "创建ServiceTree成功", zap.Int64("id", serviceTree.ID), zap.String("name", serviceTree.Name))
	response.Success(c, serviceTree)
}

// List 获取目录列表
func (api *ServiceTreeAPI) List(c *gin.Context) {
	logger.Debug(c, "开始处理ServiceTree列表请求")

	// 从查询参数获取分页信息
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logger.Debug(c, "分页参数", zap.Int("page", page), zap.Int("page_size", pageSize))

	// 查询参数
	conditions := make(map[string]interface{})

	// 添加其他过滤条件
	if user := c.Query("user"); user != "" {
		conditions["user"] = user
		logger.Debug(c, "添加user过滤条件", zap.String("user", user))
	}

	if parentID := c.Query("parent_id"); parentID != "" {
		parentIDInt, _ := strconv.ParseInt(parentID, 10, 64)
		conditions["parent_id"] = parentIDInt
		logger.Debug(c, "添加parent_id过滤条件", zap.Int64("parent_id", parentIDInt))
	}

	// 调用服务层获取目录列表
	serviceTreeList, total, err := api.service.List(c, page, pageSize, conditions)
	if err != nil {
		logger.Error(c, "获取ServiceTree列表失败", err)
		response.ServerError(c, "获取目录列表失败")
		return
	}

	logger.Info(c, "获取ServiceTree列表成功", zap.Int64("total", total), zap.Int("count", len(serviceTreeList)))
	response.Success(c, gin.H{
		"total": total,
		"items": serviceTreeList,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取目录详情
func (api *ServiceTreeAPI) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree详情请求", zap.Int64("id", id))

	// 调用服务层获取目录详情
	serviceTree, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取ServiceTree详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取目录详情失败")
		return
	}

	if serviceTree == nil {
		logger.Info(c, "目录不存在", zap.Int64("id", id))
		response.NotFound(c, "目录不存在")
		return
	}
	serviceTree.Runner = nil

	logger.Info(c, "获取ServiceTree详情成功", zap.Int64("id", id))
	response.Success(c, serviceTree)
}

// Get 获取目录详情
func (api *ServiceTreeAPI) Children(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree详情请求", zap.Int64("id", id))

	// 调用服务层获取目录详情
	serviceTree, err := api.service.Children(c, id)
	if err != nil {
		logger.Error(c, "获取ServiceTree详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取目录详情失败")
		return
	}

	logger.Info(c, "获取ServiceTree详情成功", zap.Int64("id", id))
	response.Success(c, serviceTree)
}

// Get 获取目录详情
func (api *ServiceTreeAPI) GetByName(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree详情请求", zap.Int64("id", id))

	// 调用服务层获取目录详情
	serviceTree, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取ServiceTree详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取目录详情失败")
		return
	}

	if serviceTree == nil {
		logger.Info(c, "目录不存在", zap.Int64("id", id))
		response.NotFound(c, "目录不存在")
		return
	}

	logger.Info(c, "获取ServiceTree详情成功", zap.Int64("id", id))
	response.Success(c, serviceTree)
}

// Update 更新目录
func (api *ServiceTreeAPI) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree更新请求", zap.Int64("id", id))

	var updateData model.ServiceTree
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Error(c, "解析ServiceTree参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 设置更新者信息，实际项目中应从JWT或Session获取
	updateData.UpdatedBy = "admin"

	// 调用服务层更新目录
	if err := api.service.Update(c, id, &updateData); err != nil {
		logger.Error(c, "更新ServiceTree失败", err, zap.Int64("id", id))
		response.ServerError(c, "更新目录失败: "+err.Error())
		return
	}

	// 重新获取最新数据
	serviceTree, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取更新后的ServiceTree详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取更新后的目录详情失败")
		return
	}

	logger.Info(c, "更新ServiceTree成功", zap.Int64("id", id))
	response.Success(c, serviceTree)
}

// Delete 删除目录
func (api *ServiceTreeAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree删除请求", zap.Int64("id", id))

	// 设置删除者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	// 调用服务层删除目录
	if err := api.service.Delete(c, id, operator); err != nil {
		logger.Error(c, "删除ServiceTree失败", err, zap.Int64("id", id))
		response.ServerError(c, "删除目录失败: "+err.Error())
		return
	}

	logger.Info(c, "删除ServiceTree成功", zap.Int64("id", id))
	response.Success(c, nil)
}

// Fork 复制目录
func (api *ServiceTreeAPI) Fork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析ServiceTree ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理ServiceTree复制请求", zap.Int64("id", id))

	// 解析请求参数，获取目标父目录
	var req struct {
		TargetParentID int64  `json:"target_parent_id"`
		NewName        string `json:"new_name,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析Fork参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 设置操作者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	// 调用服务层复制目录
	newTree, err := api.service.Fork(c, id, req.TargetParentID, req.NewName, operator)
	if err != nil {
		logger.Error(c, "Fork ServiceTree失败", err, zap.Int64("id", id))
		response.ServerError(c, "Fork目录失败: "+err.Error())
		return
	}

	logger.Info(c, "Fork ServiceTree成功", zap.Int64("source_id", id), zap.Int64("new_id", newTree.ID), zap.String("name", newTree.Name))
	response.Success(c, newTree)
}

// GetChildren 获取子目录列表
func (api *ServiceTreeAPI) GetChildren(c *gin.Context) {
	parentID, err := strconv.ParseInt(c.Param("parent_id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Parent ID失败", err, zap.String("parent_id_param", c.Param("parent_id")))
		response.ParamError(c, "无效的父级ID")
		return
	}

	logger.Debug(c, "开始获取子目录列表", zap.Int64("parent_id", parentID))

	// 调用服务层获取子目录列表
	children, err := api.service.GetChildren(c, parentID)
	if err != nil {
		logger.Error(c, "获取子目录列表失败", err, zap.Int64("parent_id", parentID))
		response.ServerError(c, "获取子目录列表失败")
		return
	}

	logger.Info(c, "获取子目录列表成功", zap.Int64("parent_id", parentID), zap.Int("children_count", len(children)))
	response.Success(c, children)
}
func (api *ServiceTreeAPI) GetChildrenByFullPath(c *gin.Context) {
	var req dto.GetChildrenByFullPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求参数传递错误")
		return
	}
	// 调用服务层获取子目录列表
	children, err := api.service.GetChildrenByFullPath(c, req.User, req.FullNamePath)
	if err != nil {
		response.ServerError(c, "获取子目录列表失败")
		return
	}

	response.Success(c, children)
}

func (api *ServiceTreeAPI) GetChildrenByPath(c *gin.Context) {
	path := c.Param("full_path")
	var s model.ServiceTree
	err := db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path = ?", strings.TrimPrefix(path, "/")).First(&s).Error
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	var list []model.ServiceTree
	err = db.GetDB().Model(&model.ServiceTree{}).Where("parent_id = ?", s.ID).Find(&list).Error
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, list)
}
func (api *ServiceTreeAPI) GetByFullPath(c *gin.Context) {
	var req dto.GetByFullPathReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求参数传递错误")
		return
	}
	// 调用服务层获取子目录列表
	tree, err := api.service.GetByFullPath(c, req.User, req.FullNamePath)
	if err != nil {
		response.ServerError(c, "获取子目录列表失败")
		return
	}
	response.Success(c, tree)
}
func (api *ServiceTreeAPI) Tree(c *gin.Context) {

	var list []*model.ServiceTree
	fullNamePath := strings.Trim(c.Param("full_name_path"), "/")
	fullNamePath = "/" + fullNamePath + "/"
	db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path like ?", fullNamePath+"%").Find(&list)
	tree := model.BuildServiceTree(list)
	response.Success(c, tree)
}

// GetByNamePath 根据名称路径获取服务树
func (api *ServiceTreeAPI) GetByNamePath(c *gin.Context) {
	namePath := c.Query("path")
	if namePath == "" {
		logger.Error(c, "未提供名称路径参数", fmt.Errorf("缺少路径参数"))
		response.ParamError(c, "未提供路径参数")
		return
	}

	logger.Debug(c, "开始处理根据名称路径获取服务树请求", zap.String("name_path", namePath))

	// 直接通过repo层获取服务树
	serviceTree, err := api.service.GetRepo().GetByNamePath(c, namePath)
	if err != nil {
		logger.Error(c, "根据名称路径获取服务树失败", err, zap.String("name_path", namePath))
		response.ServerError(c, "获取服务树失败")
		return
	}

	if serviceTree == nil {
		logger.Info(c, "名称路径对应的服务树不存在", zap.String("name_path", namePath))
		response.NotFound(c, "服务树不存在")
		return
	}

	logger.Info(c, "根据名称路径获取服务树成功", zap.String("name_path", namePath), zap.Int64("id", serviceTree.ID))
	response.Success(c, serviceTree)
}

// GetByIDPath 根据ID路径获取服务树
func (api *ServiceTreeAPI) GetByIDPath(c *gin.Context) {
	idPath := c.Query("path")
	if idPath == "" {
		logger.Error(c, "未提供ID路径参数", fmt.Errorf("缺少路径参数"))
		response.ParamError(c, "未提供路径参数")
		return
	}

	logger.Debug(c, "开始处理根据ID路径获取服务树请求", zap.String("id_path", idPath))

	// 直接通过repo层获取服务树
	serviceTree, err := api.service.GetRepo().GetByIDPath(c, idPath)
	if err != nil {
		logger.Error(c, "根据ID路径获取服务树失败", err, zap.String("id_path", idPath))
		response.ServerError(c, "获取服务树失败")
		return
	}

	if serviceTree == nil {
		logger.Info(c, "ID路径对应的服务树不存在", zap.String("id_path", idPath))
		response.NotFound(c, "服务树不存在")
		return
	}

	logger.Info(c, "根据ID路径获取服务树成功", zap.String("id_path", idPath), zap.Int64("id", serviceTree.ID))
	response.Success(c, serviceTree)
}
