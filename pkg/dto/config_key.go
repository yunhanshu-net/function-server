package dto

import (
	"fmt"
	"strings"
)

// GenerateConfigKey 统一的配置key生成函数
func GenerateConfigKey(router, method string) string {
	// 将路由中的路径分隔符替换为点号
	routerKey := strings.ReplaceAll(strings.Trim(router, "/"), "/", ".")
	// 去除前后多余的点号
	routerKey = strings.Trim(routerKey, ".")
	// 使用大写 method
	return fmt.Sprintf("function.%s.%s", routerKey, strings.ToUpper(method))
} 