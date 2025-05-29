package model

import "encoding/json"

type FuncRunRecord struct {
	Base
	FuncId   int64           `json:"func_id"`
	Request  json.RawMessage `json:"request" gorm:"type:json"`
	Response json.RawMessage `json:"response" gorm:"type:json"`
	Status   string          `json:"status"` //success,fail_run
	Message  string          `json:"message"`
	StartTs  int64           `json:"start_ts" gorm:"column:start_ts"`
	EndTs    int64           `json:"end_ts" gorm:"column:end_ts"`
	Cost     int64           `json:"cost" gorm:"column:cost"`
}

func (FuncRunRecord) TableName() string {
	return "func_run_record"
}
