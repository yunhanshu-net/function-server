package coder

// AddApiResp API添加响应结构
type AddApiResp struct {
	Version string `json:"version"` // API版本
	ID      uint   `json:"id"`      // API ID
	Name    string `json:"name"`    // API名称
	Status  string `json:"status"`  // 状态
}

// AddApisResp 批量添加API响应结构
type AddApisResp struct {
	Version string               `json:"version"`  // API版本
	ErrList []*CodeApiCreateInfo `json:"err_list"` // 错误列表
}

// BizPackageResp 业务包响应结构
type BizPackageResp struct {
	Version string `json:"version"` // 版本
}

// CreateProjectResp 创建项目响应结构
type CreateProjectResp struct {
	Version string `json:"version"` // 版本
}
