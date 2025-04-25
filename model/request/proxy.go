// request/proxy.go
package request

import (
	"fmt"
	"time"
)

// ProxyRequest 代理请求模型
type ProxyRequest struct {
	User      string                 `json:"user"`      // 用户
	Runner    string                 `json:"runner"`    // Runner名称
	Route     string                 `json:"route"`     // 路由
	Method    string                 `json:"method"`    // HTTP方法
	Headers   map[string]string      `json:"headers"`   // 请求头
	Body      interface{}            `json:"body"`      // 请求体
	TraceID   string                 `json:"trace_id"`  // 追踪ID
	Timestamp int64                  `json:"timestamp"` // 时间戳
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// NewProxyRequest 创建代理请求
func NewProxyRequest(user, runner, route, method string, headers map[string]string, body interface{}) *ProxyRequest {
	return &ProxyRequest{
		User:      user,
		Runner:    runner,
		Route:     route,
		Method:    method,
		Headers:   headers,
		Body:      body,
		TraceID:   generateTraceID(),
		Timestamp: time.Now().Unix(),
		Metadata:  make(map[string]interface{}),
	}
}

// generateTraceID 生成追踪ID
func generateTraceID() string {
	return fmt.Sprintf("%d-%x", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// ProxyResponse 代理响应模型
type ProxyResponse struct {
	TraceID    string                 `json:"trace_id"`        // 追踪ID
	StatusCode int                    `json:"status_code"`     // 状态码
	Headers    map[string]string      `json:"headers"`         // 响应头
	Body       interface{}            `json:"body"`            // 响应体
	Metadata   map[string]interface{} `json:"metadata"`        // 元数据
	Error      string                 `json:"error,omitempty"` // 错误信息
	Timestamp  int64                  `json:"timestamp"`       // 时间戳
	Duration   int64                  `json:"duration"`        // 处理时间(毫秒)
}
