package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/pkg/response"
	"github.com/yunhanshu-net/api-server/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceTreePathAPI ServiceTree路径API控制器
type ServiceTreePathAPI struct {
	service *service.ServiceTree
}

// NewServiceTreePathAPI 创建ServiceTreePath API控制器
func NewServiceTreePathAPI(db *gorm.DB) *ServiceTreePathAPI {
	return &ServiceTreePathAPI{
		service: service.NewServiceTree(db),
	}
}

// GetByName 根据名称路径获取服务树
func (api *ServiceTreePathAPI) GetByName(c *gin.Context) {
	namePath := c.Query("path")
	if namePath == "" {
		logger.Error(c, "未提供名称路径参数", fmt.Errorf("缺少路径参数"))
		response.ParamError(c, "未提供路径参数")
		return
	}

	logger.Debug(c, "开始处理根据名称路径获取服务树请求", zap.String("name_path", namePath))

	// 通过repo层获取服务树
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

// GetByID 根据ID路径获取服务树
func (api *ServiceTreePathAPI) GetByID(c *gin.Context) {
	idPath := c.Query("path")
	if idPath == "" {
		logger.Error(c, "未提供ID路径参数", fmt.Errorf("缺少路径参数"))
		response.ParamError(c, "未提供路径参数")
		return
	}

	logger.Debug(c, "开始处理根据ID路径获取服务树请求", zap.String("id_path", idPath))

	// 通过repo层获取服务树 - 确保这里调用的是正确的方法
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
