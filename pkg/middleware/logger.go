package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/pkg/logger"
	"go.uber.org/zap"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器，以便记录响应体
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 从上下文中获取错误（如果有）
		var errs []string
		for _, err := range c.Errors.Errors() {
			errs = append(errs, err)
		}

		// 记录访问日志
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("latency", latency),
		}

		// 只在调试模式下记录请求体和响应体，避免记录敏感信息
		if gin.Mode() == gin.DebugMode {
			fields = append(fields, zap.String("request", string(requestBody)))
			fields = append(fields, zap.String("response", w.body.String()))
		}

		if len(errs) > 0 {
			fields = append(fields, zap.Strings("errors", errs))
			logger.Warn(c, "HTTP请求", fields...)
		} else {
			logger.Info(c, "HTTP请求", fields...)
		}
	}
}
