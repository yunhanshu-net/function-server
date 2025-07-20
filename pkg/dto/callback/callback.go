package callback

import (
	resp "github.com/yunhanshu-net/function-go/pkg/dto/response"
)

// CallbackContext 回调上下文
type CallbackContext struct {
	User     string
	Runner   string
	FuncID   int64
	Type     string
	Body     interface{}
	Response *resp.RunFunctionResp
}

// UpdateConfigBody OnUpdateConfig回调的body结构
type UpdateConfigBody struct {
	Router     string                 `json:"router"`
	Method     string                 `json:"method"`
	ConfigData map[string]interface{} `json:"config_data"` // 动态配置数据
}

// GetConfigBody OnGetConfig回调的body结构
type GetConfigBody struct {
	Router    string `json:"router"`
	Method    string `json:"method"`
	ConfigKey string `json:"config_key"`
}

// ConfigUpdateReq 配置更新请求
type ConfigUpdateReq struct {
	FuncID     int64
	ConfigData map[string]interface{}
}

// ConfigGetReq 配置获取请求
type ConfigGetReq struct {
	FuncID int64
	ConfigKey string
} 