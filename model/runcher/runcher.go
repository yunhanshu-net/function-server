// runcher.go
// Runcher模型定义
package runcher

import (
	"time"

	"gorm.io/gorm"
)

// Runcher 模型定义
type Runcher struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	RunnerID  uint           `gorm:"index:idx_runner_name,unique;not null" json:"runner_id"`
	Name      string         `gorm:"index:idx_runner_name,unique;size:255;not null" json:"name"`
	Version   string         `gorm:"size:50" json:"version"`
	Command   string         `gorm:"size:255" json:"command"`
	Args      string         `gorm:"size:1024" json:"args"`
	Status    string         `gorm:"size:20;default:'created'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Runcher) TableName() string {
	return "runchers"
}
