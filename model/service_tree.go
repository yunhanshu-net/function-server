package model

import "strings"

const (
	ServiceTreeTypePackage  = "package"
	ServiceTreeTypeFunction = "function"
)

// ServiceTree 表示服务树模型
type ServiceTree struct {
	Base
	Title       string `json:"title"`
	Name        string `json:"name"`
	ParentID    int64  `json:"parent_id" gorm:"default:0"`
	Type        string `json:"type"` //package or function
	Description string `json:"description,omitempty"`
	Tags        string `json:"tags"`
	RunnerID    int64  `json:"runner_id"`
	//下面字段是数据库
	Level         int    `json:"level" gorm:"default:1"`
	Sort          int    `json:"sort" gorm:"default:0"`
	FullIDPath    string `json:"full_id_path"`
	FullNamePath  string `json:"full_name_path"`
	User          string `json:"user"`
	ChildrenCount int    `json:"children_count" gorm:"default:0"`
	ForkFromID    *int64 `json:"fork_from_id"`

	RefID  int64   `json:"ref_id"`
	Runner *Runner `json:"runner,omitempty" gorm:"foreignKey:ID;references:ID"`
}

// TableName 指定表名
func (*ServiceTree) TableName() string {
	return "service_tree"
}

// GetSubFullPath a/b/c -> b/c
func (s *ServiceTree) GetSubFullPath() string {
	split := strings.Split(s.FullNamePath, "/")
	if len(split) == 1 {
		return s.FullNamePath
	}
	return strings.Join(split[1:], "/")
}
