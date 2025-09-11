package model

import (
	"encoding/json"
	"time"
)

// Workflow 工作流定义
type Workflow struct {
	Base
	Name        string          `gorm:"column:name;comment:工作流名称" json:"name"`
	Description string          `gorm:"column:description;comment:工作流描述" json:"description"`
	Code        string          `gorm:"column:code;type:text;comment:DSL代码" json:"code"`
	JsonData    json.RawMessage `gorm:"column:json_data;type:json;comment:解析后的JSON数据" json:"json_data"`
	Status      string          `gorm:"column:status;default:active;comment:工作流状态" json:"status"`
	User        string          `gorm:"column:user;comment:创建用户" json:"user"`
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	Base
	WorkflowId   string          `gorm:"column:workflow_id;comment:工作流ID" json:"workflow_id"`
	ExecutionId  string          `gorm:"column:execution_id;comment:执行ID" json:"execution_id"`
	Status       string          `gorm:"column:status;comment:执行状态" json:"status"`
	InputVars    json.RawMessage `gorm:"column:input_vars;type:json;comment:输入变量" json:"input_vars"`
	CurrentState json.RawMessage `gorm:"column:current_state;type:json;comment:当前状态" json:"current_state"`
	User         string          `gorm:"column:user;comment:执行用户" json:"user"`
	StartTime    *time.Time      `gorm:"column:start_time;comment:开始时间" json:"start_time"`
	EndTime      *time.Time      `gorm:"column:end_time;comment:结束时间" json:"end_time"`
}

// GetJsonData 获取解析后的JSON数据
func (w *Workflow) GetJsonData() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(w.JsonData, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetJsonData 设置JSON数据
func (w *Workflow) SetJsonData(data map[string]interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	w.JsonData = json.RawMessage(jsonBytes)
	return nil
}

// GetInputVars 获取输入变量
func (e *WorkflowExecution) GetInputVars() (map[string]interface{}, error) {
	var inputVars map[string]interface{}
	if err := json.Unmarshal(e.InputVars, &inputVars); err != nil {
		return nil, err
	}
	return inputVars, nil
}

// SetInputVars 设置输入变量
func (e *WorkflowExecution) SetInputVars(inputVars map[string]interface{}) error {
	jsonBytes, err := json.Marshal(inputVars)
	if err != nil {
		return err
	}
	e.InputVars = json.RawMessage(jsonBytes)
	return nil
}

// GetCurrentState 获取当前状态
func (e *WorkflowExecution) GetCurrentState() (map[string]interface{}, error) {
	var currentState map[string]interface{}
	if err := json.Unmarshal(e.CurrentState, &currentState); err != nil {
		return nil, err
	}
	return currentState, nil
}

// SetCurrentState 设置当前状态
func (e *WorkflowExecution) SetCurrentState(state map[string]interface{}) error {
	jsonBytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	e.CurrentState = json.RawMessage(jsonBytes)
	return nil
}
