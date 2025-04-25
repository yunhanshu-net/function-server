// param.go
// 参数处理工具
package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseUint 将字符串转换为uint
func ParseUint(str string) (uint, error) {
	if str == "" {
		return 0, errors.New("空字符串不能转换为uint")
	}

	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(val), nil
}

// ParseUintWithDefault 将字符串转换为uint，失败时返回默认值
func ParseUintWithDefault(str string, defaultVal uint) (uint, error) {
	if str == "" {
		return defaultVal, nil
	}

	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return defaultVal, err
	}

	return uint(val), nil
}

// ParseInt 将字符串转换为int
func ParseInt(str string) (int, error) {
	if str == "" {
		return 0, errors.New("空字符串不能转换为int")
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// ParseIntWithDefault 将字符串转换为int，失败时返回默认值
func ParseIntWithDefault(str string, defaultVal int) (int, error) {
	if str == "" {
		return defaultVal, nil
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return defaultVal, err
	}

	return val, nil
}

// GetUserFromContext 从上下文中获取当前用户
// 在实际项目中，这通常从JWT令牌或会话中提取
func GetUserFromContext(ctx *gin.Context) string {
	// 从上下文中获取用户信息
	user, exists := ctx.Get("user")
	if !exists {
		// 在实际项目中，可能需要处理未认证的情况
		// 这里为了简化，返回空字符串或默认用户
		return "anonymous"
	}

	// 类型断言
	userStr, ok := user.(string)
	if !ok {
		return "anonymous"
	}

	return userStr
}
