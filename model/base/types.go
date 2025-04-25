// types.go
// 提供模型包中使用的所有公共类型和常量
package base

// Pagination 分页参数结构
// 用于API请求中的分页控制
type Pagination struct {
	Page     int `json:"page" form:"page" binding:"required,min=1"`                   // 页码，从1开始
	PageSize int `json:"page_size" form:"page_size" binding:"required,min=1,max=100"` // 每页条数，最大100
}

// SearchParams 查询参数结构
// 扩展分页参数，增加排序和关键词搜索
type SearchParams struct {
	Pagination        // 嵌入分页参数
	SortBy     string `json:"sort_by" form:"sort_by"`                                // 排序字段
	SortOrder  string `json:"sort_order" form:"sort_order" binding:"oneof=asc desc"` // 排序方向
	Keyword    string `json:"keyword" form:"keyword"`                                // 搜索关键词
}

// PagedResult 分页结果结构
// 用于返回分页查询结果
type PagedResult struct {
	Total     int64       `json:"total"`      // 总记录数
	Page      int         `json:"page"`       // 当前页码
	PageSize  int         `json:"page_size"`  // 每页条数
	TotalPage int         `json:"total_page"` // 总页数
	Data      interface{} `json:"data"`       // 数据列表
}

// EntityType 实体类型
// 用于识别不同实体类型
type EntityType string

const (
	EntityTypePackage  EntityType = "package"  // 包
	EntityTypeFunction EntityType = "function" // 函数
	EntityTypeNode     EntityType = "node"     // 节点
)

// PermissionType 权限类型
// 控制实体的访问权限
type PermissionType int8

const (
	PermissionPrivate PermissionType = 1 // 私有权限，仅创建者可访问
	PermissionPublic  PermissionType = 2 // 公开权限，所有人可访问
)
