package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/pkg/constants"
)

// WithUserInfo 为请求添加跟踪ID的中间件
func WithUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		get := c.Request.Header.Get(constants.RequestUserInfo)
		if get == "" {
			//todo 其实应该加上token的逻辑，这里为了能快速实现功能先省略
			c.Set("user", "admin")
		} else {
			c.Set("user", get)
		}
	}
}
