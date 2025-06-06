package dto

import (
	"github.com/yunhanshu-net/pkg/query"
)

// BaseRequest 基础请求结构体
type BaseRequest struct {
	TraceID string `json:"-"` // 请求跟踪ID，不序列化到JSON
}

// BasePaginatedRequest 基础分页请求结构体
type BasePaginatedRequest struct {
	BaseRequest
	query.PageInfoReq // 嵌入分页信息
}
