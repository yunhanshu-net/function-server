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
	Request         json.RawMessage `json:"request"`
	Response        json.RawMessage `json:"response"`
	Callbacks       string          `json:"callbacks"`
	UseTables       string          `json:"use_tables"`
	IsPublic        bool            `json:"is_public"`
	User            string          `json:"user"`
	TreeID          int64           `json:"tree_id"`
	RunnerID        int64           `json:"runner_id"`
	ForkFromUser    string          `json:"fork_from_user,omitempty"`
	ForkFromVersion string          `json:"fork_from_version"`
	ForkFromID      *int64          `json:"fork_from_id"`
}

func (RunnerFunc) TableName() string {
	return "runner_func"
}
