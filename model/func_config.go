package model

import (
	"encoding/json"
	"reflect"
)

// FuncConfig 表示 func_config 表
type FuncConfig struct {
	Base

	FuncID       int64           `json:"func_id" gorm:"column:func_id;comment:函数ID"`
	ConfigKey    string          `json:"config_key" gorm:"column:config_key;type:varchar(255);comment:配置键"`
	ConfigType   string          `json:"config_type" gorm:"column:config_type;type:varchar(50);comment:配置类型"`
	ConfigStruct json.RawMessage `json:"config_struct" gorm:"type:json;column:config_struct;comment:配置结构体定义"`
	ConfigData   json.RawMessage `json:"config_data" gorm:"type:json;column:config_data;comment:配置初始值"`
	Description  string          `json:"description" gorm:"column:description;type:varchar(500);comment:配置描述"`
	Version      string          `json:"version" gorm:"column:version;type:varchar(50);comment:配置版本"`
	IsActive     bool            `json:"is_active" gorm:"column:is_active;comment:是否激活"`
}

func (FuncConfig) TableName() string {
	return "func_config"
}

func (f *FuncConfig) DiffStruct(new json.RawMessage) bool {
	return reflect.DeepEqual(f.ConfigStruct, new)
}
