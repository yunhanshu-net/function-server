package model

import "encoding/json"

// FuncVersion 表示函数版本记录
type FuncVersion struct {
	Base
	RunnerID int64           `json:"runner_id"`
	FuncID   int64           `json:"func_id"`
	Version  string          `json:"version"`
	Comment  string          `json:"comment"`
	MetaData json.RawMessage `json:"metadata"`
	Hash     string          `json:"hash"`
}

// TableName 表名
func (FuncVersion) TableName() string {
	return "func_version"
}
