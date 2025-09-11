package model

import (
	"encoding/json"
)

//执行用例

type FunctionExecuteCase struct {
	Base
	RunnerId    int    `json:"runner_id" gorm:"column:runner_id"`
	FunctionId  int    `json:"function_id" gorm:"column:function_id"`
	Name        string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`

	DefineRequest json.RawMessage `json:"define_request" gorm:"column:define_request;type:json"` //保存当初保存用例时候的函数的请求参数，可以进行对比，防止函数变更导致参数不一致的用例执行失败

	Request       json.RawMessage `json:"request" gorm:"column:request;type:json"`
	CanBackground bool            `json:"can_background" gorm:"column:can_background"` //是否支持后台执行
	AutoRun       bool            `json:"auto_run" gorm:"column:auto_run"`             //是否自动运行
	ExecCount     int             `json:"exec_count" gorm:"column:exec_count"`

	LastUsedAt       *Time  `json:"last_used_at" gorm:"column:last_used_at"`             // 最后使用时间
	RunnerVersion    string `json:"runner_version" gorm:"column:runner_version"`         //用例的快照版本
	UseLatestVersion bool   `json:"use_latest_version" gorm:"column:use_latest_version"` //是否总是使用最新版本？
}

func (FunctionExecuteCase) TableName() string {
	return "function_execute_case"
}

type FunctionExecuteCaseRecord struct {
	Base
	CaseId     int `json:"case_id" gorm:"column:case_id"`
	RunnerId   int `json:"runner_id" gorm:"column:runner_id"`
	FunctionId int `json:"function_id" gorm:"column:function_id"`

	Status     string          `json:"status" gorm:"column:status"`   //running，sys_failed,biz_failed，success
	Message    string          `json:"message" gorm:"column:message"` //
	Remark     string          `json:"remark" gorm:"column:remark"`   //执行备注
	Request    json.RawMessage `json:"request" gorm:"column:request;type:json"`
	Response   json.RawMessage `json:"response" gorm:"column:response;type:json"`
	Background bool            `json:"background" gorm:"column:background"` //是否后台执行
	ExecCount  int             `json:"exec_count" gorm:"column:exec_count"`
	CostMillis int64           `json:"cost_millis" gorm:"column:cost_millis"`
	TraceId    string          `json:"traceId" gorm:"column:trace_id"`

	ErrorDetails string `json:"error_details" gorm:"column:error_details"` // 详细错误信息
	StartTime    Time   `json:"start_time" gorm:"column:start_time"`       // 开始时间
	EndTime      *Time  `json:"end_time" gorm:"column:end_time"`           // 结束时间
}

func (FunctionExecuteCaseRecord) TableName() string {
	return "function_execute_case_record"
}
