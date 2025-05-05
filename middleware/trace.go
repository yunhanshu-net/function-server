package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yunhanshu-net/api-server/pkg/constants"
	"time"
)

// WithTraceID 为请求添加跟踪ID的中间件
func WithTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头中是否已有跟踪ID
		traceID := c.GetHeader(constants.HttpTraceID)
		if traceID == "" {
			// 如果没有，则生成一个新的跟踪ID，格式为: 时间戳-UUID
			traceID = time.Now().Format("20060102150405") + "-" + uuid.New().String()
			c.Request.Header.Set(constants.HttpTraceID, traceID)
		}

		// 在响应头中也返回跟踪ID
		c.Header(constants.HttpTraceID, traceID)

		// 将跟踪ID存储在gin.Context中，这样可以直接通过context获取
		c.Set(constants.TraceID, traceID)

		c.Next()
	}
}
