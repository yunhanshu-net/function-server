package dto

import (
	"github.com/yunhanshu-net/api-server/pkg/utils"
)

// BaseRequest 基础请求结构体
type BaseRequest struct {
	TraceID string `json:"-"` // 请求跟踪ID，不序列化到JSON
}

// BasePaginatedRequest 基础分页请求结构体
type BasePaginatedRequest struct {
	BaseRequest
	utils.PageInfo // 嵌入分页信息
}
