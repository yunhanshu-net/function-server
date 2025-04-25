// base.go
// 包含所有模型共用的基础结构和接口
package base

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 所有模型的基础结构
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`      // 主键ID
	CreatedAt time.Time      `json:"created_at"`                // 创建时间
	UpdatedAt time.Time      `json:"updated_at"`                // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`            // 软删除时间(不输出到JSON)
	CreatedBy string         `json:"created_by" gorm:"size:50"` // 创建者
	UpdatedBy string         `json:"updated_by" gorm:"size:50"` // 更新者
}

// AuditInfo 审计信息
// 用于跟踪各种实体的使用情况和性能指标
type AuditInfo struct {
	LastAccessTime time.Time `json:"last_access_time"` // 最后访问时间
	AccessCount    int64     `json:"access_count"`     // 访问总次数
	SuccessCount   int64     `json:"success_count"`    // 成功次数
	FailureCount   int64     `json:"failure_count"`    // 失败次数
	AvgResponseMs  int64     `json:"avg_response_ms"`  // 平均响应时间(毫秒)
}
