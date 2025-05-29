package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/api/v1"
	"github.com/yunhanshu-net/api-server/middleware"
	"github.com/yunhanshu-net/api-server/pkg/config"
	"github.com/yunhanshu-net/api-server/pkg/db"
	pkgmiddleware "github.com/yunhanshu-net/api-server/pkg/middleware"
)

// Init 初始化路由
func Init() *gin.Engine {
	// 设置gin模式
	gin.SetMode(config.Get().ServerConfig.Mode)

	// 创建gin引擎
	r := gin.New()

	// 使用中间件
	r.Use(gin.Recovery())
	//r.Use(pkgmiddleware.Logger())
	r.Use(pkgmiddleware.Cors())
	r.Use(middleware.WithTraceID()) // 添加跟踪ID中间件

	// API版本v1
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.WithUserInfo())

	functionV1 := r.Group("/function")
	functionV1.Use(middleware.WithUserInfo())
	{
		functionApi := v1.NewFunctions(db.GetDB())
		functionV1.Any("/run/:user/:runner/*router", functionApi.Run)
		functionV1.POST("/callback/:user/:runner/*router", functionApi.Callback)
	}
	{
		// Runner 相关路由
		runnerAPI := v1.NewRunnerAPI(db.GetDB())
		runner := apiV1.Group("/runner")
		{
			runner.POST("", runnerAPI.Create)                       // 创建Runner
			runner.GET("", runnerAPI.List)                          // 获取Runner列表
			runner.GET("/:id", runnerAPI.Get)                       // 获取Runner详情
			runner.PUT("/:id", runnerAPI.Update)                    // 更新Runner
			runner.DELETE("/:id", runnerAPI.Delete)                 // 删除Runner
			runner.POST("/:id/fork", runnerAPI.Fork)                // Fork Runner
			runner.GET("/:id/version", runnerAPI.Version)           // 获取Runner版本历史
			runner.GET("/by-name/:user/:name", runnerAPI.GetByName) // 通过用户名和名称获取Runner
		}

		// ServiceTree 相关路由
		serviceTreeAPI := v1.NewServiceTreeAPI(db.GetDB())
		serviceTree := apiV1.Group("/service-tree")
		{
			serviceTree.POST("", serviceTreeAPI.Create) // 创建目录
			serviceTree.GET("", serviceTreeAPI.List)    // 获取目录列表
			serviceTree.GET("/:id", serviceTreeAPI.Get) // 获取目录详情
			//serviceTree.GET("/children/:id", serviceTreeAPI.Children)           // 获取目录详情
			serviceTree.PUT("/:id", serviceTreeAPI.Update)                                // 更新目录
			serviceTree.DELETE("/:id", serviceTreeAPI.Delete)                             // 删除目录
			serviceTree.POST("/:id/fork", serviceTreeAPI.Fork)                            // Fork目录
			serviceTree.GET("/children/:parent_id", serviceTreeAPI.GetChildren)           // 获取子目录列表
			serviceTree.GET("/get_children/*full_path", serviceTreeAPI.GetChildrenByPath) // 获取子目录列表
			serviceTree.GET("/full_path", serviceTreeAPI.GetByFullPath)                   // 获取子目录列表
			serviceTree.GET("/tree/*full_name_path", serviceTreeAPI.Tree)                 // 获取子目录列表
			serviceTree.GET("/search", serviceTreeAPI.Search)                             // 获取子目录列表
			serviceTree.GET("/user_count_list", serviceTreeAPI.UserWorkCountList)         // 获取子目录列表

		}

		// ServiceTreePath 相关路由
		serviceTreePathAPI := v1.NewServiceTreePathAPI(db.GetDB())
		serviceTreePath := apiV1.Group("/service-tree-path")
		{
			serviceTreePath.GET("/by-name", serviceTreePathAPI.GetByName) // 根据名称路径获取服务树
			serviceTreePath.GET("/by-id", serviceTreePathAPI.GetByID)     // 根据ID路径获取服务树
		}

		// RunnerFunc 相关路由
		runnerFuncAPI := v1.NewRunnerFuncAPI(db.GetDB())
		runnerFunc := apiV1.Group("/runner-func")
		{
			runnerFunc.POST("", runnerFuncAPI.Create)                            // 创建函数
			runnerFunc.GET("", runnerFuncAPI.List)                               // 获取函数列表
			runnerFunc.GET("/:id", runnerFuncAPI.Get)                            // 获取函数详情
			runnerFunc.GET("/:id/versions", runnerFuncAPI.Versions)              // 获取函数详情
			runnerFunc.GET("/tree/:tree_id", runnerFuncAPI.GetByTreeId)          // 获取函数详情
			runnerFunc.GET("/full-path/*full_path", runnerFuncAPI.GetByFullPath) // 获取函数详情

			runnerFunc.PUT("/:id", runnerFuncAPI.Update)                    // 更新函数
			runnerFunc.DELETE("/:id", runnerFuncAPI.Delete)                 // 删除函数
			runnerFunc.POST("/:id/fork", runnerFuncAPI.Fork)                // Fork函数
			runnerFunc.GET("/runner/:runner_id", runnerFuncAPI.GetByRunner) // 获取Runner下的函数列表

			runnerFunc.GET("/record/:func_id", runnerFuncAPI.GetFuncRecord)
			runnerFunc.GET("/recent-records", runnerFuncAPI.GetUserRecentFuncRecords) // 获取用户最近执行函数记录（去重）
		}
	}

	return r
}
