package dto

import (
	"encoding/json"
	"github.com/yunhanshu-net/pkg/query"
)

// CreateWorkflowReq 创建工作流请求
type CreateWorkflowReq struct {
	Name        string          `json:"name" validate:"required"` // 工作流名称
	Description string          `json:"description"`              // 工作流描述
	Code        string          `json:"code"`                     // DSL代码
	JsonData    json.RawMessage `json:"json_data,omitempty"`      // 可选：直接提供JSON数据
}

// CreateWorkflowResp 创建工作流响应
type CreateWorkflowResp struct {
	WorkflowId  string          `json:"workflow_id"` // 工作流ID
	Name        string          `json:"name"`        // 工作流名称
	Description string          `json:"description"` // 工作流描述
	Status      string          `json:"status"`      // 工作流状态
	CreatedAt   string          `json:"created_at"`  // 创建时间
	JsonData    json.RawMessage `json:"json_data"`   // 完整的解析后JSON数据
}

// ExecuteWorkflowReq 执行工作流请求
type ExecuteWorkflowReq struct {
	WorkflowId string                 `json:"workflow_id" validate:"required"` // 工作流ID
	InputVars  map[string]interface{} `json:"input_vars"`                      // 输入变量
}

// ExecuteWorkflowResp 执行工作流响应
type ExecuteWorkflowResp struct {
	ExecutionId string `json:"execution_id"` // 执行ID
	Status      string `json:"status"`       // 执行状态
	Message     string `json:"message"`      // 执行消息
}

// GetWorkflowStatusReq 获取工作流状态请求
type GetWorkflowStatusReq struct {
	ExecutionId string `json:"execution_id" validate:"required"` // 执行ID
}

// GetWorkflowStatusResp 获取工作流状态响应
type GetWorkflowStatusResp struct {
	ExecutionId  string                 `json:"execution_id"`  // 执行ID
	WorkflowId   string                 `json:"workflow_id"`   // 工作流ID
	Status       string                 `json:"status"`        // 执行状态
	Progress     float64                `json:"progress"`      // 执行进度
	CurrentStep  string                 `json:"current_step"`  // 当前步骤
	Variables    map[string]interface{} `json:"variables"`     // 当前变量
	Steps        []WorkflowStepStatus   `json:"steps"`         // 步骤状态
	StartTime    string                 `json:"start_time"`    // 开始时间
	EndTime      string                 `json:"end_time"`      // 结束时间
	ErrorMessage string                 `json:"error_message"` // 错误信息
}

// WorkflowStepStatus 工作流步骤状态
type WorkflowStepStatus struct {
	StepName  string                 `json:"step_name"`  // 步骤名称
	Status    string                 `json:"status"`     // 步骤状态
	StartTime string                 `json:"start_time"` // 开始时间
	EndTime   string                 `json:"end_time"`   // 结束时间
	Duration  int64                  `json:"duration"`   // 执行时长(ms)
	Error     string                 `json:"error"`      // 错误信息
	Input     map[string]interface{} `json:"input"`      // 输入参数
	Output    map[string]interface{} `json:"output"`     // 输出参数
}

// StopWorkflowReq 停止工作流请求
type StopWorkflowReq struct {
	ExecutionId string `json:"execution_id" validate:"required"` // 执行ID
}

// StopWorkflowResp 停止工作流响应
type StopWorkflowResp struct {
	ExecutionId string `json:"execution_id"` // 执行ID
	Status      string `json:"status"`       // 停止状态
	Message     string `json:"message"`      // 停止消息
}

// GetWorkflowDetailReq 获取工作流详情请求
type GetWorkflowDetailReq struct {
	WorkflowId string `json:"workflow_id" validate:"required"` // 工作流ID
}

// GetWorkflowDetailResp 获取工作流详情响应
type GetWorkflowDetailResp struct {
	WorkflowId  string          `json:"workflow_id"` // 工作流ID
	Name        string          `json:"name"`        // 工作流名称
	Description string          `json:"description"` // 工作流描述
	Status      string          `json:"status"`      // 工作流状态
	CreatedAt   string          `json:"created_at"`  // 创建时间
	UpdatedAt   string          `json:"updated_at"`  // 更新时间
	JsonData    json.RawMessage `json:"json_data"`   // 完整的解析后JSON数据
}

// ListWorkflowReq 获取工作流列表请求
type ListWorkflowReq struct {
	query.SearchFilterPageReq
	Name   string `form:"name"`   // 工作流名称（模糊查询）
	Status string `form:"status"` // 工作流状态
}

// ListWorkflowResp 获取工作流列表响应
type ListWorkflowResp struct {
	Items       interface{} `json:"items" runner:"widget:table;type:array;code:items"` // 分页数据
	CurrentPage int         `json:"current_page" runner:"search_cond"`                 // 当前页码
	TotalCount  int64       `json:"total_count" runner:"search_cond"`                  // 总数据量
	TotalPages  int         `json:"total_pages" runner:"search_cond"`                  // 总页数
	PageSize    int         `json:"page_size" runner:"search_cond"`                    // 每页数量
}

// WorkflowListItem 工作流列表项
type WorkflowListItem struct {
	WorkflowId  string `json:"workflow_id"` // 工作流ID
	Name        string `json:"name"`        // 工作流名称
	Description string `json:"description"` // 工作流描述
	Status      string `json:"status"`      // 工作流状态
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
}
