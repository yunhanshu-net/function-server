// types.go
// 包含目录结构中使用的类型定义
package directory

// NodeType 节点类型
// 用于区分TreeNode表示的是包(目录)还是函数(可执行单元)
type NodeType int8

const (
	NodeTypePackage  NodeType = 1 // 包/目录类型
	NodeTypeFunction NodeType = 2 // 函数/可执行单元类型
)

// FunctionType 函数类型枚举
// 定义不同类型的可执行单元
type FunctionType string

const (
	FunctionTypeRegular  FunctionType = "regular"  // 普通函数
	FunctionTypeWorkflow FunctionType = "workflow" // 工作流函数
	FunctionTypeAPI      FunctionType = "api"      // API函数
	FunctionTypeUIForm   FunctionType = "ui_form"  // 表单函数
)
