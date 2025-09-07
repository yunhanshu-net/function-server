package model

import (
	"encoding/json"
	"strings"

	"github.com/yunhanshu-net/function-go/pkg/dto/api"
	"github.com/yunhanshu-net/pkg/x/jsonx"
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
	Group           json.RawMessage `json:"group" gorm:"type:json"`
	IsPublic        bool            `json:"is_public"`
	User            string          `json:"user"`
	TreeID          int64           `json:"tree_id"`
	RunnerID        int64           `json:"runner_id"`
	ForkFromUser    string          `json:"fork_from_user,omitempty"`
	ForkFromVersion string          `json:"fork_from_version"`
	ForkFromID      *int64          `json:"fork_from_id"`
	Method          string          `json:"method" gorm:"type:varchar(255);column:method"`
	Path            string          `json:"path" gorm:"type:varchar(255);column:path"`
	Router          string          `json:"router" gorm:"-"`
	Code            string          `json:"-" gorm:"-"`
	HasConfig       bool            `json:"has_config" gorm:"column:has_config;comment:是否存在配置"` // 是否存在配置

	RunnerName    string `json:"runner_name" gorm:"-"`
	ExecCaseCount int    `json:"exec_case_count" gorm:"column:exec_case_count"`
}

func (r *RunnerFunc) GetRouter() string {
	trim := strings.Trim(r.Path, "/")
	split := strings.Split(trim, "/") // a/b/c/d
	router := split[2:]
	runner := split[1:2]
	r.Router = strings.Join(router, "/")
	r.RunnerName = runner[0]
	return r.Router
}

func (r *RunnerFunc) GetRunnerName() string {
	trim := strings.Trim(r.Path, "/")
	split := strings.Split(trim, "/") // a/b/c/d
	router := split[2:]
	runner := split[1:2]
	r.Router = strings.Join(router, "/")
	r.RunnerName = runner[0]
	return r.RunnerName
}

func (RunnerFunc) TableName() string {
	return "runner_func"
}

// FromAPIInfo 从api.Info结构体赋值数据到RunnerFunc,需要赋值User和Name
func (rf *RunnerFunc) FromAPIInfo(apiInfo *api.Info) {
	rf.Async = apiInfo.Async
	rf.Timeout = apiInfo.Timeout
	rf.Description = apiInfo.ApiDesc
	rf.RenderType = apiInfo.RenderType
	rf.FunctionType = apiInfo.FunctionType
	rf.Name = apiInfo.EnglishName
	rf.Title = apiInfo.ChineseName
	rf.Tags = strings.Join(apiInfo.Tags, ",")
	rf.Request = json.RawMessage(jsonx.String(apiInfo.ParamsIn))
	rf.Response = json.RawMessage(jsonx.String(apiInfo.ParamsOut))
	rf.Method = apiInfo.Method
	rf.Callbacks = strings.Join(apiInfo.Callbacks, ",")
	rf.UseTables = strings.Join(apiInfo.UseTables, ",")
	rf.CreateTables = strings.Join(apiInfo.CreateTables, ",")

	if apiInfo.OperateTables != nil {
		rf.OperateTables = json.RawMessage(jsonx.String(apiInfo.OperateTables))
	}

	if apiInfo.Group != nil {
		rf.Group = json.RawMessage(jsonx.String(apiInfo.Group))
	}

	// 设置默认值
	if rf.User == "" {
		rf.User = "admin"
	}

	// 生成路径
	path := "/" + apiInfo.User + "/" + apiInfo.Runner + "/" + strings.Trim(apiInfo.Router, "/") + "/"
	rf.Path = path

	// 检查是否有配置
	if apiInfo.ParamsConfig != nil && apiInfo.ParamsData != nil {
		rf.HasConfig = true
	}
}

// DiffWithAPIInfo 比较api.Info和现有的RunnerFunc，返回需要更新的字段
// 如果apiInfo中的字段与现有RunnerFunc不同，则以apiInfo为准
func (rf *RunnerFunc) DiffWithAPIInfo(apiInfo *api.Info) *RunnerFunc {
	// 创建一个新的RunnerFunc用于存储差异
	diff := &RunnerFunc{}

	// 比较并设置差异字段
	if rf.Async != apiInfo.Async {
		diff.Async = apiInfo.Async
	}

	if rf.Timeout != apiInfo.Timeout {
		diff.Timeout = apiInfo.Timeout
	}

	if rf.Description != apiInfo.ApiDesc {
		diff.Description = apiInfo.ApiDesc
	}

	if rf.RenderType != apiInfo.RenderType {
		diff.RenderType = apiInfo.RenderType
	}

	if rf.FunctionType != apiInfo.FunctionType {
		diff.FunctionType = apiInfo.FunctionType
	}

	if rf.Name != apiInfo.EnglishName {
		diff.Name = apiInfo.EnglishName
	}

	if rf.Title != apiInfo.ChineseName {
		diff.Title = apiInfo.ChineseName
	}

	// 比较Tags（数组转字符串）
	expectedTags := strings.Join(apiInfo.Tags, ",")
	if rf.Tags != expectedTags {
		diff.Tags = expectedTags
	}

	// 比较Request（JSON序列化后比较）
	//expectedRequest := json.RawMessage(jsonx.String(apiInfo.ParamsIn))
	//if string(rf.Request) != string(expectedRequest) {
	//	diff.Request = expectedRequest
	//}

	// 比较Response（JSON序列化后比较）
	expectedRequest := json.RawMessage(jsonx.String(apiInfo.ParamsIn))

	if !jsonx.EQRawMessage(rf.Request, expectedRequest) {
		diff.Request = expectedRequest
	}

	// 比较Response（JSON序列化后比较）
	expectedResponse := json.RawMessage(jsonx.String(apiInfo.ParamsOut))

	if !jsonx.EQRawMessage(rf.Response, expectedResponse) {
		diff.Response = expectedResponse
	}

	if rf.Method != apiInfo.Method {
		diff.Method = apiInfo.Method
	}

	// 比较Callbacks（数组转字符串）
	expectedCallbacks := strings.Join(apiInfo.Callbacks, ",")
	if rf.Callbacks != expectedCallbacks {
		diff.Callbacks = expectedCallbacks
	}

	//// 比较UseTables（数组转字符串）
	//expectedUseTables := strings.Join(apiInfo.UseTables, ",")
	//if rf.UseTables != expectedUseTables {
	//	diff.UseTables = expectedUseTables
	//}

	//// 比较CreateTables（数组转字符串）
	//expectedCreateTables := strings.Join(apiInfo.CreateTables, ",")
	//if rf.CreateTables != expectedCreateTables {
	//	diff.CreateTables = expectedCreateTables
	//}

	//// 比较OperateTables（JSON序列化后比较）
	//if apiInfo.OperateTables != nil {
	//	expectedOperateTables := json.RawMessage(jsonx.String(apiInfo.OperateTables))
	//	if string(rf.OperateTables) != string(expectedOperateTables) {
	//		diff.OperateTables = expectedOperateTables
	//	}
	//}

	// 比较Group（JSON序列化后比较）
	if apiInfo.Group != nil {
		eq := jsonx.EQRawMessage(json.RawMessage(jsonx.String(apiInfo.Group)), rf.Group)
		if !eq {
			expectedGroup := json.RawMessage(jsonx.String(apiInfo.Group))
			diff.Group = expectedGroup
		}
	}

	// 生成新的路径
	expectedPath := "/" + apiInfo.User + "/" + apiInfo.Runner + "/" + strings.Trim(apiInfo.Router, "/") + "/"
	if rf.Path != expectedPath {
		diff.Path = expectedPath
	}

	// 检查配置变化
	expectedHasConfig := apiInfo.ParamsConfig != nil && apiInfo.ParamsData != nil
	if rf.HasConfig != expectedHasConfig {
		diff.HasConfig = expectedHasConfig
	}

	// 如果没有任何差异，返回nil
	if diff.ID == 0 && diff.Title == "" && diff.Name == "" && diff.Description == "" &&
		diff.Tags == "" && diff.Request == nil && diff.Response == nil && diff.Callbacks == "" &&
		diff.UseTables == "" && diff.CreateTables == "" && diff.Timeout == 0 &&
		diff.Async == false && diff.FunctionType == "" && diff.RenderType == "" &&
		diff.OperateTables == nil && diff.Group == nil && diff.Path == "" &&
		diff.Method == "" && diff.HasConfig == false {
		return nil
	}

	return diff
}
