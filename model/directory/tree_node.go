// tree_node.go
// 实现树形目录结构中的节点模型
package directory

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/yunhanshu-net/api-server/model/base"
)

// TreeNode 通用树节点结构
// 这是整个目录系统的核心模型，用于构建树形目录结构
// 通过RefID字段关联到实际内容(Package或Function)
type TreeNode struct {
	base.BaseModel
	// 基本信息部分
	CnName      string `json:"cn_name" gorm:"size:100;not null"`                // 中文名称，用于显示
	EnName      string `json:"en_name" gorm:"size:100;not null;index:idx_name"` // 英文名称，作为唯一标识
	Description string `json:"description" gorm:"size:500"`                     // 描述信息
	Icon        string `json:"icon" gorm:"size:50"`                             // 图标标识符

	// 节点类型与引用部分 - 关联到实际内容
	Type       NodeType `json:"type" gorm:"column:type;not null;comment:1 pkg or 2 func"` // 节点类型(包或函数)
	RefID      uint     `json:"ref_id" gorm:"column:ref_id"`                              // 指向实体(Package/Function)的ID
	RefVersion string   `json:"ref_version" gorm:"size:20"`                               // 引用的实体版本

	// 树形结构部分 - 维护节点间关系
	ParentID *uint  `json:"parent_id" gorm:"default:null;index:idx_parent_id"` // 父节点ID，可为空(根节点)
	Level    int    `json:"level" gorm:"not null;default:0"`                   // 层级深度，根节点为0
	FullPath string `json:"full_path" gorm:"index:idx_full_path;not null"`     // 完整路径，如/1/3/7
	Sort     int    `json:"sort" gorm:"default:0"`                             // 同级节点间的排序值

	// 所有权与权限部分
	User           string              `json:"user" gorm:"index:idx_user;not null"`      // 所属用户/租户
	SourceUser     string              `json:"source_user" gorm:"index:idx_source_user"` // 原始用户(fork来源)
	SourceNodeID   uint                `json:"source_node_id" gorm:"default:0"`          // 原始节点ID(fork来源)
	PermissionType base.PermissionType `json:"permission_type" gorm:"default:2"`         // 权限类型

	// UI相关部分 - 前端展示控制
	Expand bool   `json:"expand" gorm:"default:false"` // 是否默认展开子节点
	Label  string `json:"label" gorm:"-"`              // 显示标签(运行时生成)

	// 状态与统计部分
	IsActive       bool      `json:"is_active" gorm:"default:true"` // 是否处于激活状态
	AccessCount    int64     `json:"access_count" gorm:"default:0"` // 访问计数
	LastAccessTime time.Time `json:"last_access_time"`              // 最后访问时间

	// 树形关系(非持久化字段)
	IsLeaf   bool        `json:"is_leaf" gorm:"-"`  // 是否为叶子节点(无子节点)
	Leaf     bool        `json:"leaf" gorm:"-"`     // 兼容某些前端框架的叶子标记
	Position int         `json:"position" gorm:"-"` // 展开时的位置
	Children []*TreeNode `json:"children" gorm:"-"` // 子节点集合
}

// TableName 指定数据库表名
func (t *TreeNode) TableName() string {
	return "directory_tree_node"
}

// BeforeCreate 创建记录前的处理
// 确保节点类型正确、设置默认值，并标记函数类型为叶子节点
func (t *TreeNode) BeforeCreate(tx *gorm.DB) error {
	// 验证节点类型有效性
	if t.Type != NodeTypePackage && t.Type != NodeTypeFunction {
		return fmt.Errorf("无效的节点类型: %d", t.Type)
	}

	// 设置默认显示标签
	if t.Label == "" {
		t.Label = t.CnName
	}

	// 函数类型节点自动设为叶子节点
	if t.Type == NodeTypeFunction {
		t.IsLeaf = true
		t.Leaf = true
	}

	return nil
}

// AfterCreate 创建记录后的处理
// 主要用于生成和更新完整路径，确保路径包含节点自身ID
func (t *TreeNode) AfterCreate(tx *gorm.DB) error {
	// 初始化空路径
	if t.FullPath == "" {
		t.FullPath = "/"
	}

	// 添加节点ID到路径末尾
	pathSuffix := fmt.Sprintf("/%v", t.ID)
	if !strings.HasSuffix(t.FullPath, pathSuffix) {
		t.FullPath += pathSuffix
	}

	// 更新路径到数据库
	return tx.Model(t).UpdateColumn("full_path", t.FullPath).Error
}

// AfterFind 查询记录后的处理
// 设置运行时字段的值
func (t *TreeNode) AfterFind(tx *gorm.DB) error {
	// 设置显示标签为中文名
	t.Label = t.CnName

	// 函数类型节点设置为叶子节点
	if t.Type == NodeTypeFunction {
		t.IsLeaf = true
		t.Leaf = true
	}

	return nil
}

// IsFork 判断节点是否为fork来源
// 当SourceUser和SourceNodeID都有值时表示这是fork的节点
func (t *TreeNode) IsFork() bool {
	return t.SourceUser != "" && t.SourceNodeID > 0
}

// GetPath 获取包含用户的完整路径
// 用于唯一标识节点位置
func (t *TreeNode) GetPath() string {
	return fmt.Sprintf("/%s%s", t.User, t.FullPath)
}

// GetAPIPath 获取API调用路径
// 仅函数类型节点有API路径
func (t *TreeNode) GetAPIPath() string {
	if t.Type == NodeTypeFunction {
		return fmt.Sprintf("/api/%s%s", t.User, t.FullPath)
	}
	return ""
}

// 递归获取目录树
// 从指定根节点开始构建完整的树形结构
func GetDirectoryTree(db *gorm.DB, user string, rootNodeID uint) ([]*TreeNode, error) {
	var nodes []*TreeNode

	// 查询指定用户下根节点的直接子节点
	if err := db.Model(&TreeNode{}).
		Where("user = ? AND (parent_id = ? OR parent_id IS NULL)", user, rootNodeID).
		Order("sort"). // 按排序值排序
		Find(&nodes).Error; err != nil {
		return nil, err
	}

	// 递归获取每个包类型节点的子节点
	for _, node := range nodes {
		if node.Type == NodeTypePackage {
			children, err := GetDirectoryTree(db, user, node.ID)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}
	}

	return nodes, nil
}

// MoveNode 移动节点到新位置
// 同时更新其所有子节点的路径
func MoveNode(db *gorm.DB, nodeID, targetParentID uint) error {
	// 1. 获取要移动的节点
	var node TreeNode
	if err := db.First(&node, nodeID).Error; err != nil {
		return err
	}

	// 2. 获取目标父节点
	var parentNode TreeNode
	if err := db.First(&parentNode, targetParentID).Error; err != nil {
		return err
	}

	// 3. 验证：不能将节点移动到其自身的子节点下(避免循环引用)
	if strings.Contains(parentNode.FullPath, fmt.Sprintf("/%d/", node.ID)) {
		return fmt.Errorf("不能将节点移动到其子节点下")
	}

	// 4. 计算节点新路径
	node.ParentID = &targetParentID
	node.FullPath = parentNode.FullPath + "/" + fmt.Sprint(node.ID)

	// 5. 事务处理：更新节点及其所有子节点的路径
	return db.Transaction(func(tx *gorm.DB) error {
		// 保存当前节点
		if err := tx.Save(&node).Error; err != nil {
			return err
		}

		// 查找所有子节点
		var children []*TreeNode
		if err := tx.Where("full_path LIKE ?", node.FullPath+"/%").Find(&children).Error; err != nil {
			return err
		}

		// 更新每个子节点的路径
		for _, child := range children {
			// 替换路径前缀
			child.FullPath = strings.Replace(child.FullPath,
				node.FullPath, parentNode.FullPath+"/"+fmt.Sprint(node.ID), 1)
			if err := tx.Save(child).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
