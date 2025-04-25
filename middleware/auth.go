// auth.go
// 认证相关中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yunhanshu-net/api-server/api/response"
)

// Authentication 认证中间件
// 验证用户身份并将用户信息存储到上下文中
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.FailWithCode(c, http.StatusUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		// 提取令牌值
		// 通常格式为: "Bearer xxx.yyy.zzz"
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.FailWithCode(c, http.StatusUnauthorized, "认证格式错误，应为 Bearer XXX")
			c.Abort()
			return
		}

		token := parts[1]
		// 在实际项目中，这里应该验证JWT令牌
		// 为了简化示例，这里直接使用令牌值作为用户名
		// 假设令牌的格式是 user_xxx

		if !strings.HasPrefix(token, "user_") {
			response.FailWithCode(c, http.StatusUnauthorized, "无效的令牌")
			c.Abort()
			return
		}

		// 提取用户名并存储到上下文
		username := strings.TrimPrefix(token, "user_")
		c.Set("user", username)

		c.Next()
	}
}
