package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yunhanshu-net/function-server/pkg/config"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/logger"
	"github.com/yunhanshu-net/function-server/router"
	"github.com/yunhanshu-net/function-server/service"
)

func main() {
	ctx := context.Background()
	// 加载配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(config.Get().LogConfig); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化数据库连接
	if err := db.Init(config.Get().DBConfig); err != nil {
		logger.Fatal(ctx, "初始化数据库连接失败", err)
	}

	// 初始化RuncherService
	runcherOptions := service.RuncherOptions{
		NatsURL: config.Get().RuncherConfig.NatsURL,
		Timeout: time.Duration(config.Get().RuncherConfig.Timeout) * time.Second,
	}

	var err error
	runcherService, err := service.NewRuncherService(runcherOptions)
	if err != nil {
		logger.Error(ctx, "初始化Runcher服务失败", err)
		// 继续执行，不要因为Runcher服务初始化失败而中断启动
		// 这里仅记录错误，让应用可以启动，但函数执行功能将不可用
	} else {
		logger.Info(ctx, "Runcher服务初始化成功")
		defer runcherService.Close()

		// 设置全局RuncherService实例
		service.SetGlobalRuncherService(runcherService)
	}

	// 初始化路由
	r := router.Init()

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Get().ServerConfig.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 在一个单独的goroutine中启动服务器
	go func() {
		logger.Info(ctx, fmt.Sprintf("服务器运行在 %s 端口", addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, "启动服务器出错", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(ctx, "服务器强制关闭", err)
	}

	logger.Info(ctx, "服务器优雅退出")
}
