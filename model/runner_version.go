package model

// RunnerVersion 表示runner版本记录
type RunnerVersion struct {
	Base

	RunnerID  int64  `json:"runner_id"`
	Version   string `json:"version"`
	Comment   string `json:"comment"`
	CreatedBy string `json:"created_by"`
	CreatedAt Time   `json:"created_at"`
}

// TableName 表名
func (RunnerVersion) TableName() string {
	return "runner_version"
}
