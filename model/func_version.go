package model

// FuncVersion 表示函数版本记录
type FuncVersion struct {
	Base

	FuncID    int64  `json:"func_id"`
	Version   string `json:"version"`
	Comment   string `json:"comment"`
	CreatedBy string `json:"created_by"`
	CreatedAt Time   `json:"created_at"`
}

// TableName 表名
func (FuncVersion) TableName() string {
	return "func_version"
}
