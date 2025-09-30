package dto

import (
	"encoding/json"

	"github.com/yunhanshu-net/pkg/query"

	"github.com/yunhanshu-net/function-server/model"
)

type FunctionExecuteCaseReq struct {
	CaseId     int    `json:"case_id"`
	Remark     string `json:"remark"`
	Background bool   `json:"background"`
}

type BatchDeleteFunctionExecuteCaseReq struct {
	CaseIds []int `json:"case_ids" binding:"required,min=1"`
}

type UpdateFunctionExecuteCaseReq struct {
	CaseId           int             `json:"case_id" binding:"required"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	Request          json.RawMessage `json:"request"`
	CanBackground    bool            `json:"can_background"`
	AutoRun          bool            `json:"auto_run"` //是否自动运行
	UseLatestVersion bool            `json:"use_latest_version"`
}

type QueryFunctionExecuteCaseReq struct {
	query.SearchFilterPageReq
	// 保留一些特殊字段，其他通过PageInfoReq的查询条件处理
}

type QueryFunctionExecuteCaseRecordReq struct {
	query.SearchFilterPageReq
	// 保留一些特殊字段，其他通过PageInfoReq的查询条件处理
}

// FunctionExecuteCaseWithDetails 带预加载信息的执行用例
type FunctionExecuteCaseWithDetails struct {
	*model.FunctionExecuteCase
	Runner   *RunnerInfo   `json:"runner,omitempty"`
	Function *FunctionInfo `json:"function,omitempty"`
	UrlPath  string        `json:"url_path,omitempty"`
}

// RunnerInfo Runner预加载信息
type RunnerInfo struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Version      string `json:"version"`
	Language     string `json:"language"`
	Status       int8   `json:"status"`
	User         string `json:"user"`
	FullNamePath string `json:"full_name_path"`
}

// FunctionInfo Function预加载信息
type FunctionInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Router      string `json:"router"`
	Tags        string `json:"tags"`
	User        string `json:"user"`
	HasConfig   bool   `json:"has_config"`
}

// FunctionExecuteCaseRecordWithDetails 带预加载信息的执行记录
type FunctionExecuteCaseRecordWithDetails struct {
	model.FunctionExecuteCaseRecord
	Runner   *RunnerInfo   `json:"runner,omitempty"`
	Function *FunctionInfo `json:"function,omitempty"`
}
