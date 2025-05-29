package dto

import (
	"time"

	"github.com/yunhanshu-net/function-server/model"
)

// ===========================================================================
// 创建Runner
// ===========================================================================

// CreateRunnerReq 创建Runner请求
type CreateRunnerReq struct {
	BaseRequest
	Name      string `json:"name" binding:"required"`  // Runner名称
	Title     string `json:"title" binding:"required"` // Runner标题
	Desc      string `json:"desc"`                     // Runner描述
	Status    int    `json:"status"`                   // 状态
	IsPublic  bool   `json:"is_public"`                // 是否公开
	Variables string `json:"variables"`                // 环境变量
}

// CreateRunnerResp 创建Runner响应
type CreateRunnerResp struct {
	ID        int64     `json:"id"`         // Runner ID
	Name      string    `json:"name"`       // Runner名称
	Title     string    `json:"title"`      // Runner标题
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// ===========================================================================
// 获取Runner详情
// ===========================================================================

// GetRunnerReq 获取Runner详情请求
type GetRunnerReq struct {
	BaseRequest
	ID int64 `json:"-"` // Runner ID，从路径参数获取
}

// GetRunnerResp 获取Runner详情响应
type GetRunnerResp struct {
	ID         int64     `json:"id"`           // Runner ID
	Name       string    `json:"name"`         // Runner名称
	Title      string    `json:"title"`        // Runner标题
	Desc       string    `json:"desc"`         // Runner描述
	Status     int       `json:"status"`       // 状态
	IsPublic   bool      `json:"is_public"`    // 是否公开
	Variables  string    `json:"variables"`    // 环境变量
	User       string    `json:"user"`         // 所属用户
	ForkFromID *int64    `json:"fork_from_id"` // Fork来源ID
	CreatedBy  string    `json:"created_by"`   // 创建者
	CreatedAt  time.Time `json:"created_at"`   // 创建时间
	UpdatedBy  string    `json:"updated_by"`   // 更新者
	UpdatedAt  time.Time `json:"updated_at"`   // 更新时间
}

// ===========================================================================
// 更新Runner
// ===========================================================================

// UpdateRunnerReq 更新Runner请求
type UpdateRunnerReq struct {
	BaseRequest
	ID        int64  `json:"-"`                        // Runner ID，从路径参数获取
	Name      string `json:"name"`                     // Runner名称
	Title     string `json:"title" binding:"required"` // Runner标题
	Desc      string `json:"desc"`                     // Runner描述
	Status    int    `json:"status"`                   // 状态
	IsPublic  bool   `json:"is_public"`                // 是否公开
	Variables string `json:"variables"`                // 环境变量
}

// UpdateRunnerResp 更新Runner响应
type UpdateRunnerResp struct {
	ID        int64     `json:"id"`         // Runner ID
	Name      string    `json:"name"`       // Runner名称
	Title     string    `json:"title"`      // Runner标题
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// ===========================================================================
// 删除Runner
// ===========================================================================

// DeleteRunnerReq 删除Runner请求
type DeleteRunnerReq struct {
	BaseRequest
	ID int64 `json:"-"` // Runner ID，从路径参数获取
}

// DeleteRunnerResp 删除Runner响应
type DeleteRunnerResp struct {
	Success bool `json:"success"` // 是否成功
}

// ===========================================================================
// Runner列表
// ===========================================================================

// ListRunnerReq 获取Runner列表请求
type ListRunnerReq struct {
	BasePaginatedRequest
	User     string `json:"user" form:"user"`           // 用户名过滤
	Status   int    `json:"status" form:"status"`       // 状态过滤
	IsPublic *bool  `json:"is_public" form:"is_public"` // 是否公开过滤
}

// ListRunnerResp 获取Runner列表响应（单个项）
type ListRunnerResp struct {
	ID        int64     `json:"id"`         // Runner ID
	Name      string    `json:"name"`       // Runner名称
	Title     string    `json:"title"`      // Runner标题
	Status    int       `json:"status"`     // 状态
	IsPublic  bool      `json:"is_public"`  // 是否公开
	User      string    `json:"user"`       // 所属用户
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// ToConditions 转换为查询条件
func (req *ListRunnerReq) ToConditions() map[string]interface{} {
	conditions := make(map[string]interface{})

	if req.User != "" {
		conditions["user"] = req.User
	}

	if req.Status != 0 {
		conditions["status"] = req.Status
	}

	if req.IsPublic != nil {
		conditions["is_public"] = *req.IsPublic
	}

	return conditions
}

// FromModel 从模型转换
func (resp *ListRunnerResp) FromModel(runner *model.Runner) {
	resp.ID = runner.ID
	resp.Name = runner.Name
	resp.Title = runner.Title
	resp.Status = int(runner.Status)
	resp.IsPublic = runner.IsPublic
	resp.User = runner.User
	resp.CreatedAt = time.Time(runner.CreatedAt)
}

// ===========================================================================
// Runner Fork
// ===========================================================================

// ForkRunnerReq Fork Runner请求
type ForkRunnerReq struct {
	BaseRequest
	ID int64 `json:"-"` // 源Runner ID，从路径参数获取
}

// ForkRunnerResp Fork Runner响应
type ForkRunnerResp struct {
	ID        int64     `json:"id"`         // 新Runner ID
	Name      string    `json:"name"`       // 新Runner名称
	Title     string    `json:"title"`      // 新Runner标题
	ForkFrom  int64     `json:"fork_from"`  // Fork来源ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *ForkRunnerResp) FromModel(runner *model.Runner) {
	resp.ID = runner.ID
	resp.Name = runner.Name
	resp.Title = runner.Title
	//if runner.ForkFromID != nil {
	//	resp.ForkFrom = *runner.ForkFromID
	//}
	resp.CreatedAt = time.Time(runner.CreatedAt)
}

// ===========================================================================
// Runner版本
// ===========================================================================

// VersionHistoryItem Runner版本历史项
type VersionHistoryItem struct {
	Version   string    `json:"version"`    // 版本号
	Comment   string    `json:"comment"`    // 版本说明
	CreatedBy string    `json:"created_by"` // 创建者
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
