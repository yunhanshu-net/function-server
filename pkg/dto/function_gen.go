package dto

import "github.com/yunhanshu-net/pkg/query"

// FunctionGenReq 获取生成函数列表请求
type FunctionGenListReq struct {
	RunnerID int64 `json:"runner_id" form:"runner_id"`

	query.PageInfoReq
}

type FunctionGenReq struct {
	Title      string `json:"title"`
	RunnerID   int64  `json:"runner_id"`
	TreeID     int64  `json:"tree_id"`
	Message    string `json:"message"`
	RenderType string `json:"render_type"`
	Async      bool   `json:"async"`
	User       string `json:"-"`
}

type GeneratingCount struct {
	RunnerID int64 `json:"runner_id" form:"runner_id"`
}
