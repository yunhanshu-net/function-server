package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/yunhanshu-net/api-server/api/v1"
)

// RegisterCoderRouter 注册coder相关路由
func RegisterCoderRouter(r *gin.RouterGroup) {
	coderRouter := r.Group("/coder")
	{
		// API管理相关接口
		coderRouter.POST("/api", v1.AddApi) // 添加单个API

		// 其他待实现的API接口
		// coderRouter.POST("/apis", v1.AddApis)     // 批量添加API
		// coderRouter.POST("/package", v1.AddBizPackage) // 添加业务包
		// coderRouter.POST("/project", v1.CreateProject) // 创建项目
	}
}
