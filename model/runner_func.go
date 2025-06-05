package model

import (
	"encoding/json"
)

// RunnerFunc 表示 runner_func 表
type RunnerFunc struct {
	Base

	Title           string          `json:"title"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Tags            string          `json:"tags"`
	Request         json.RawMessage `json:"request" gorm:"type:json"`
	Response        json.RawMessage `json:"response" gorm:"type:json"`
	Callbacks       string          `json:"callbacks"`
	UseTables       string          `json:"use_tables"`
	Timeout         int             `json:"timeout"`
	AutoRun         bool            `json:"auto_run"`       //是否自动运行，默认false，如果为true，则在用户访问这个函数时候，会自动运行一次
	Async           bool            `json:"async"`          //是否异步，比较耗时的api，或者需要后台慢慢处理的api
	FunctionType    string          `json:"function_type"`  //函数类型 默认：dynamic_function
	RenderType      string          `json:"widget"`         // 渲染类型
	CreateTables    string          `json:"create_tables"`  //创建该api时候会自动帮忙创建这个数据库表gorm的model列表
	OperateTables   json.RawMessage `json:"operate_tables"` //用到了哪些表，对表进行了哪些操作方便梳理引用关系
	IsPublic        bool            `json:"is_public"`
	User            string          `json:"user"`
	TreeID          int64           `json:"tree_id"`
	RunnerID        int64           `json:"runner_id"`
	ForkFromUser    string          `json:"fork_from_user,omitempty"`
	ForkFromVersion string          `json:"fork_from_version"`
	ForkFromID      *int64          `json:"fork_from_id"`
	Method          string          `json:"method" gorm:"type:varchar(255);column:method"`
	Path            string          `json:"path" gorm:"type:varchar(255);column:path"`
	Code            string          `json:"-" gorm:"-"`
}

func (RunnerFunc) TableName() string {
	return "runner_func"
}
