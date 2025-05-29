package model

import "encoding/json"

// RunnerVersion 表示runner版本记录
type RunnerVersion struct {
	Base
	Hash     string          `json:"hash"`
	RunnerID int64           `json:"runner_id"`
	Version  string          `json:"version"`
	Comment  string          `json:"comment"` //ai评价
	Log      string          `json:"log"`     //变更日志
	Desc     string          `json:"desc"`    //用户描述
	MetaData json.RawMessage `json:"meta_data"`
}

// TableName 表名
func (RunnerVersion) TableName() string {
	return "runner_version"
}
