package dto

// UserWorkCountListReq 用户工作数量列表请求
type UserWorkCountListReq struct {
	Fuzzy string `form:"fuzzy" json:"fuzzy"` // 模糊查询参数
}

// UserWorkCountListResp 用户工作数量列表响应
type UserWorkCountListResp struct {
	User  string `json:"user"`  // 用户名
	Count int    `json:"count"` // 工作数量
}
