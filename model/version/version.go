// version.go
// 定义版本相关模型，用于管理Runner的版本
package version

import (
	"time"

	"github.com/yunhanshu-net/api-server/model/base"
)

// Version Runner版本信息
// 记录Runner的版本信息和部署状态
type Version struct {
	base.BaseModel

	// 关联
	RunnerID uint `json:"runner_id" gorm:"index;not null"` // 关联的Runner ID

	// 版本信息
	Version     string `json:"version" gorm:"size:50;index;not null"` // 版本号(如v1)
	Description string `json:"description" gorm:"size:1000"`          // 版本描述

	// 发布信息
	PublishedAt time.Time `json:"published_at"` // 发布时间
	PublishedBy uint      `json:"published_by"` // 发布者ID

	// 部署信息
	IsActive      bool   `json:"is_active" gorm:"default:false"` // 是否为当前活跃版本
	DeployStatus  string `json:"deploy_status" gorm:"size:20"`   // 部署状态
	DeployMessage string `json:"deploy_message" gorm:"size:500"` // 部署消息

	// API变更
	AddedAPIs   int `json:"added_apis" gorm:"default:0"`   // 新增API数量
	UpdatedAPIs int `json:"updated_apis" gorm:"default:0"` // 更新API数量
	DeletedAPIs int `json:"deleted_apis" gorm:"default:0"` // 删除API数量

	// 文件存储
	FilePath     string `json:"file_path" gorm:"size:500"`     // 文件存储路径
	FileChecksum string `json:"file_checksum" gorm:"size:100"` // 文件校验和
	FileSize     int64  `json:"file_size" gorm:"default:0"`    // 文件大小(字节)
}

// TableName 指定表名
func (Version) TableName() string {
	return "runner_version"
}

// VersionSummary 版本简要信息
// 用于API返回简要信息，避免返回过多数据
type VersionSummary struct {
	ID          uint      `json:"id"`
	RunnerID    uint      `json:"runner_id"`
	Version     string    `json:"version"`
	IsActive    bool      `json:"is_active"`
	PublishedAt time.Time `json:"published_at"`
	APIChanges  int       `json:"api_changes"`
}

// VersionHistory 记录版本历史变更
// 包括版本间的差异和变更内容
type VersionHistory struct {
	base.BaseModel

	// 关联
	VersionID   uint   `json:"version_id" gorm:"index;not null"` // 版本ID
	PrevVersion string `json:"prev_version" gorm:"size:50"`      // 前一个版本
	RunnerID    uint   `json:"runner_id" gorm:"index;not null"`  // Runner ID

	// 变更内容
	ChangeType    string `json:"change_type" gorm:"size:20"`     // 变更类型(新增/修改/删除)
	ChangeSummary string `json:"change_summary" gorm:"size:500"` // 变更摘要
	ChangeDetail  string `json:"change_detail" gorm:"type:text"` // 变更详情(JSON格式)

	// 审计
	ChangedBy uint      `json:"changed_by"` // 变更人ID
	ChangedAt time.Time `json:"changed_at"` // 变更时间
}

// TableName 指定表名
func (VersionHistory) TableName() string {
	return "runner_version_history"
}
