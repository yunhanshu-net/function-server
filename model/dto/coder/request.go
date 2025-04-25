package coder

import (
	"github.com/yunhanshu-net/api-server/model/runner"
)

// AddApiReq 添加API请求结构
type AddApiReq struct {
	Runner  *runner.Runner `json:"runner"`   // Runner信息
	CodeApi *CodeApi       `json:"code_api"` // API代码信息
}

// AddApisReq 批量添加API请求结构
type AddApisReq struct {
	Runner   *runner.Runner `json:"runner"`    // Runner信息
	CodeApis []*CodeApi     `json:"code_apis"` // 多个API代码信息
}
