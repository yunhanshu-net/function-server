// api.go
// 定义Runner API模型，表示Runner提供的API
package runner

import (
	"github.com/yunhanshu-net/api-server/model/base"
)

// API Runner提供的API
type API struct {
	base.BaseModel

	// 关联
	RunnerID uint `json:"runner_id" gorm:"index;not null"` // 关联的Runner ID

	// API信息
	Route       string `json:"route" gorm:"size:200;index;not null"`         // API路由
	Method      string `json:"method" gorm:"size:10;not null;default:'GET'"` // HTTP方法
	Name        string `json:"name" gorm:"size:100"`                         // API名称
	Description string `json:"description" gorm:"size:1000"`                 // 描述信息

	// 参数定义 (存储为JSON格式)
	RequestSchema  string `json:"request_schema" gorm:"type:text"`  // 请求参数Schema
	ResponseSchema string `json:"response_schema" gorm:"type:text"` // 响应参数Schema

	// 额外信息
	Category     string `json:"category" gorm:"size:50"`            // API分类
	Tags         string `json:"tags" gorm:"size:200"`               // 标签(逗号分隔)
	IsDeprecated bool   `json:"is_deprecated" gorm:"default:false"` // 是否已废弃

	// 审计和统计
	base.AuditInfo
}

// TableName 指定表名
func (API) TableName() string {
	return "runner_api"
}

// Parameter API参数定义
type Parameter struct {
	Name        string `json:"name"`        // 参数名
	Type        string `json:"type"`        // 参数类型
	Required    bool   `json:"required"`    // 是否必须
	Description string `json:"description"` // 参数描述
	Default     string `json:"default"`     // 默认值
}

// APISummary API简要信息
// 用于API返回简要信息，避免返回过多数据
type APISummary struct {
	ID          uint   `json:"id"`
	RunnerID    uint   `json:"runner_id"`
	Route       string `json:"route"`
	Method      string `json:"method"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
