package model

import (
	"github.com/yunhanshu-net/pkg/x/slicesx"
	"strings"
)

const (
	ServiceTreeTypePackage  = "package"
	ServiceTreeTypeFunction = "function"
)

// BuildServiceTree 将ServiceTree切片组装成树形结构并返回根节点
func BuildServiceTree(nodes []*ServiceTree) *ServiceTree {
	if len(nodes) == 0 {
		return nil
	}

	// 创建一个map用于快速查找节点
	nodeMap := make(map[int64]*ServiceTree)
	for _, node := range nodes {
		nodeMap[node.ID] = node
		// 初始化children切片，避免nil
		if node.Children == nil {
			node.Children = []*ServiceTree{}
		}
	}

	var root *ServiceTree

	// 建立父子关系
	for _, node := range nodes {
		if node.ParentID == 0 {
			// ParentID为0表示根节点
			root = node
		} else {
			// 找到父节点并添加到其children中
			if parent, exists := nodeMap[node.ParentID]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return root
}

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
	Level         int            `json:"level" gorm:"default:1"`
	Sort          int            `json:"sort" gorm:"default:0"`
	FullIDPath    string         `json:"full_id_path"`
	FullNamePath  string         `json:"full_name_path"`
	User          string         `json:"user"`
	ChildrenCount int            `json:"children_count" gorm:"default:0"`
	ForkFromID    *int64         `json:"fork_from_id"`
	Method        string         `json:"method" gorm:"column:method"`
	RefID         int64          `json:"ref_id"`
	Runner        *Runner        `json:"runner,omitempty" gorm:"foreignKey:ID;references:ID"`
	Children      []*ServiceTree `json:"children" gorm:"-"`
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
	split = slicesx.RemoveBy(split, func(s string) bool {
		return s == ""
	})
	return strings.Join(split[1:], "/")
}
