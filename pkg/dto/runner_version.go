package dto

import (
	"time"
)

// ===========================================================================
// 获取版本历史
// ===========================================================================

// GetRunnerVersionHistoryReq 获取Runner版本历史请求
type GetRunnerVersionHistoryReq struct {
	BaseRequest
	RunnerID int64 `json:"-"` // Runner ID，从路径参数获取
}

// GetRunnerVersionHistoryResp 获取Runner版本历史响应（单个项）
type GetRunnerVersionHistoryResp struct {
	ID        int64     `json:"id"`         // 版本 ID
	RunnerID  int64     `json:"runner_id"`  // Runner ID
	Version   string    `json:"version"`    // 版本号
	Comment   string    `json:"comment"`    // 版本注释
	CreatedBy string    `json:"created_by"` // 创建者
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// ===========================================================================
// 保存版本
// ===========================================================================

// SaveRunnerVersionReq 保存Runner版本请求
type SaveRunnerVersionReq struct {
	BaseRequest
	RunnerID int64  `json:"-"`                          // Runner ID，从路径参数获取
	Version  string `json:"version" binding:"required"` // 版本号
	Comment  string `json:"comment"`                    // 版本注释
}

// SaveRunnerVersionResp 保存Runner版本响应
type SaveRunnerVersionResp struct {
	ID        int64     `json:"id"`         // 版本 ID
	RunnerID  int64     `json:"runner_id"`  // Runner ID
	Version   string    `json:"version"`    // 版本号
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
