package dto

import (
	"github.com/yunhanshu-net/api-server/pkg/utils"
	"reflect"
)

// BaseRequest 基础请求结构体
type BaseRequest struct {
	TraceID string `json:"-"` // 请求跟踪ID，不序列化到JSON
}

// BaseResponse 基础响应结构体
type BaseResponse struct {
	Code    int         `json:"code"`    // 错误码，0表示成功
	Message string      `json:"message"` // 错误消息
	Data    interface{} `json:"data"`    // 响应数据
}

// BasePaginatedRequest 基础分页请求结构体
type BasePaginatedRequest struct {
	BaseRequest
	utils.PageInfo // 嵌入分页信息
}

// BasePaginatedResponse 基础分页响应结构体
type BasePaginatedResponse struct {
	Code       int         `json:"code"`    // 错误码，0表示成功
	Message    string      `json:"message"` // 错误消息
	Data       interface{} `json:"data"`    // 数据项列表
	Pagination struct {
		CurrentPage int   `json:"current_page"` // 当前页码
		PageSize    int   `json:"page_size"`    // 每页大小
		TotalCount  int64 `json:"total_count"`  // 总记录数
		TotalPages  int   `json:"total_pages"`  // 总页数
	} `json:"pagination"` // 分页信息
}

// FrontendPaginatedResponse 前端期望的分页响应结构体
type FrontendPaginatedResponse struct {
	Items []interface{} `json:"items"` // 数据项列表
	Total int64         `json:"total"` // 总记录数
	Page  int           `json:"page"`  // 当前页码
	Size  int           `json:"size"`  // 每页大小
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *BaseResponse {
	return &BaseResponse{
		Code:    0,
		Message: "成功",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *BaseResponse {
	return &BaseResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(paginatedData *utils.Paginated) *FrontendPaginatedResponse {
	// 将数据从interface{}类型转换为[]interface{}类型
	var items []interface{}
	if dataSlice, ok := paginatedData.Items.([]interface{}); ok {
		items = dataSlice
	} else if reflect.TypeOf(paginatedData.Items).Kind() == reflect.Slice {
		// 如果数据是切片类型，但不是[]interface{}，则进行转换
		v := reflect.ValueOf(paginatedData.Items)
		items = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			items[i] = v.Index(i).Interface()
		}
	} else {
		// 如果不是切片类型，则创建包含该元素的切片
		items = []interface{}{paginatedData.Items}
	}

	return &FrontendPaginatedResponse{
		Items: items,
		Total: paginatedData.TotalCount,
		Page:  paginatedData.CurrentPage,
		Size:  paginatedData.PageSize,
	}
}
