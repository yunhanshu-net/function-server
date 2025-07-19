package dto

// GetFuncConfigReq 获取函数配置请求
type GetFuncConfigReq struct {
	FuncID int64 `json:"func_id" form:"func_id" binding:"required"` // 函数ID
}

// GetFuncConfigResp 获取函数配置响应
type GetFuncConfigResp struct {
	FuncID      int64       `json:"func_id"`       // 函数ID
	ConfigKey   string      `json:"config_key"`    // 配置键
	ConfigType  string      `json:"config_type"`   // 配置类型
	ConfigStruct interface{} `json:"config_struct"` // 配置结构体定义
	ConfigData  interface{} `json:"config_data"`   // 配置数据
	Description string      `json:"description"`   // 配置描述
	Version     string      `json:"version"`       // 配置版本
	IsActive    bool        `json:"is_active"`     // 是否激活
}

// UpdateFuncConfigReq 更新函数配置请求
type UpdateFuncConfigReq struct {
	FuncID     int64       `json:"func_id" binding:"required"`     // 函数ID
	ConfigData interface{} `json:"config_data" binding:"required"` // 配置数据
}

// UpdateFuncConfigResp 更新函数配置响应
type UpdateFuncConfigResp struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 响应消息
} 