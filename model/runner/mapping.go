// mapping.go
// 定义Runner与Runcher节点的映射关系
package runner

import (
	"time"

	"github.com/yunhanshu-net/api-server/model/base"
)

// RuncherMapping Runner与Runcher节点的映射关系
// 记录Runner部署在哪个Runcher节点上以及运行状态
type RuncherMapping struct {
	base.BaseModel

	// 关联
	RunnerID  uint `json:"runner_id" gorm:"index;not null"`  // Runner ID
	RuncherID uint `json:"runcher_id" gorm:"index;not null"` // Runcher ID

	// 映射信息
	Status     string    `json:"status" gorm:"size:20;default:'pending'"` // 映射状态
	LastActive time.Time `json:"last_active"`                             // 最后活跃时间
	DeployedAt time.Time `json:"deployed_at"`                             // 部署时间
	IsActive   bool      `json:"is_active" gorm:"default:false"`          // 是否活跃

	// 运行信息
	CPUUsage    float64 `json:"cpu_usage" gorm:"default:0"`    // CPU使用率
	MemoryUsage float64 `json:"memory_usage" gorm:"default:0"` // 内存使用率
	RequestRate float64 `json:"request_rate" gorm:"default:0"` // 每秒请求数
	ErrorRate   float64 `json:"error_rate" gorm:"default:0"`   // 错误率

	// 关联 - 定义关联关系，在查询时可以预加载
	Runner *Runner `json:"runner,omitempty" gorm:"foreignKey:RunnerID"`
	// Runcher类型的关联关系可以通过查询时联结实现
}

// TableName 指定表名
func (RuncherMapping) TableName() string {
	return "runner_runcher_mapping"
}
