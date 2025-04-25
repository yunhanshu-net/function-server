// package.go
// 定义Package模型，表示目录结构中的包/目录
package directory

import (
	"time"

	"gorm.io/gorm"

	"github.com/yunhanshu-net/api-server/model/base"
)

// Package 包/目录模型
// 代表一个功能集合，可以包含多个函数
// 是TreeNode中RefID引用的实际内容之一
type Package struct {
	base.BaseModel
	// 基本信息
	Name        string `json:"name" gorm:"size:100;not null;index"` // 包名称(英文标识)
	DisplayName string `json:"display_name" gorm:"size:100"`        // 显示名称(可中文)
	Description string `json:"description" gorm:"size:500"`         // 描述信息
	User        string `json:"user" gorm:"index;not null"`          // 所属用户/租户

	// 分类与标签
	Category string `json:"category" gorm:"size:50"` // 分类
	Tags     string `json:"tags" gorm:"size:200"`    // 标签(逗号分隔)

	// 权限与状态
	Visibility string `json:"visibility" gorm:"size:20;default:'private'"` // 可见性(private/public)
	Status     string `json:"status" gorm:"size:20;default:'active'"`      // 状态

	// 版本管理
	CurrentVersion string `json:"current_version" gorm:"size:20;default:'v1'"` // 当前版本号

	// 统计信息
	FunctionCount int       `json:"function_count" gorm:"default:0"` // 包含的函数数量
	ForkCount     int       `json:"fork_count" gorm:"default:0"`     // 被fork次数
	ViewCount     int64     `json:"view_count" gorm:"default:0"`     // 查看次数
	LastViewTime  time.Time `json:"last_view_time"`                  // 最后查看时间

	// Fork信息
	SourceID   uint   `json:"source_id" gorm:"default:0"` // 源包ID(如果是fork的)
	SourceUser string `json:"source_user" gorm:"size:50"` // 源用户(fork来源)

	// 元数据
	Metadata string `json:"metadata" gorm:"type:text"` // 元数据(JSON格式)
}

// TableName 指定数据库表名
func (p *Package) TableName() string {
	return "directory_package"
}

// BeforeCreate 创建前钩子
// 设置默认值
func (p *Package) BeforeCreate(tx *gorm.DB) error {
	// 确保有默认显示名称
	if p.DisplayName == "" {
		p.DisplayName = p.Name // 默认使用英文名作为显示名
	}

	return nil
}

// IsFork 判断包是否为fork来源
func (p *Package) IsFork() bool {
	return p.SourceUser != "" && p.SourceID > 0
}

// GetNodeType 获取节点类型
// 包始终返回NodeTypePackage类型
func (p *Package) GetNodeType() NodeType {
	return NodeTypePackage
}
