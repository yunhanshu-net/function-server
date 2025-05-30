package utils

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/logger"
)

// GetTraceID 从gin上下文中获取跟踪ID
func GetTraceID(c *gin.Context) string {
	// 首先尝试从context的Keys中获取
	if id, exists := c.Get(constants.TraceID); exists {
		if traceID, ok := id.(string); ok {
			return traceID
		}
	}
	// 如果在Keys中没有找到，则从HTTP头中获取
	return c.GetHeader(constants.HttpTraceID)
}

// GetTraceIDFromContext 从context中获取TraceID
func GetTraceIDFromContext(ctx context.Context) string {
	if id := ctx.Value(constants.TraceID); id != nil {
		if traceID, ok := id.(string); ok {
			return traceID
		}
	}
	return ""
}

// GetContextWithTraceID 创建带有跟踪ID的上下文
func GetContextWithTraceID(c *gin.Context) context.Context {
	traceID := GetTraceID(c)
	return context.WithValue(context.Background(), constants.TraceID, traceID)
}

// FromGinContext 从gin.Context创建标准Context
// 这个方法可用于将gin.Context转换为符合context.Context接口的对象
// 同时保留traceID等关键信息
func FromGinContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// GinLog 提供了一组便捷的日志记录函数，直接接受gin.Context
type GinLog struct{}

// Debug 输出调试日志
func (l *GinLog) Debug(c *gin.Context, msg string) {
	logger.Debug(c, msg)
}

// Debugf 格式化输出调试日志
func (l *GinLog) Debugf(c *gin.Context, format string, args ...interface{}) {
	// 使用fmt格式化字符串后传递给Debug方法
	message := fmt.Sprintf(format, args...)
	logger.Debug(c, message)
}

// Info 输出信息日志
func (l *GinLog) Info(c *gin.Context, msg string) {
	logger.Info(c, msg)
}

// Infof 格式化输出信息日志
func (l *GinLog) Infof(c *gin.Context, format string, args ...interface{}) {
	// 使用fmt格式化字符串后传递给Info方法
	message := fmt.Sprintf(format, args...)
	logger.Info(c, message)
}

// Error 输出错误日志
func (l *GinLog) Error(c *gin.Context, msg string) {
	// 保持签名不变，但内部传递nil作为error
	logger.Error(c, msg, nil)
}

// Errorf 格式化输出错误日志
func (l *GinLog) Errorf(c *gin.Context, format string, args ...interface{}) {
	// 使用fmt格式化字符串后传递给Error方法
	message := fmt.Sprintf(format, args...)

	// 尝试从args中提取error
	var err error
	for _, arg := range args {
		if e, ok := arg.(error); ok {
			err = e
			break
		}
	}

	logger.Error(c, message, err)
}

// NewGinLog 创建一个新的GinLog实例
func NewGinLog() *GinLog {
	return &GinLog{}
}
