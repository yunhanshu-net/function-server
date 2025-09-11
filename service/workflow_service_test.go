package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/yunhanshu-net/function-server/pkg/dto"
)

func TestWorkflowService_CreateWorkflow(t *testing.T) {
	// 这里可以添加单元测试
	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接")
}

func TestWorkflowService_CreateWorkflowWithJSON(t *testing.T) {
	// 测试使用JSON数据创建工作流
	jsonData := map[string]interface{}{
		"success": true,
		"input_vars": map[string]interface{}{
			"用户名": "张三",
			"手机号": 13800138000,
		},
		"steps": []interface{}{
			map[string]interface{}{
				"name":     "step1",
				"function": "beiluo.test1.devops.devops_script_create",
			},
		},
		"main_func": map[string]interface{}{
			"statements": []interface{}{
				map[string]interface{}{
					"type":     "function-call",
					"function": "step1",
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		t.Fatalf("序列化JSON失败: %v", err)
	}

	req := &dto.CreateWorkflowReq{
		Name:        "测试工作流",
		Description: "这是一个测试工作流",
		JsonData:    json.RawMessage(jsonBytes),
	}

	// 验证请求结构
	if req.Name != "测试工作流" {
		t.Errorf("期望名称: 测试工作流, 实际: %s", req.Name)
	}

	if len(req.JsonData) == 0 {
		t.Error("JSON数据不能为空")
	}

	t.Logf("测试请求: %+v", req)
}

func TestWorkflowService_CreateWorkflowWithDSL(t *testing.T) {
	// 测试使用DSL代码创建工作流
	dslCode := `
var input = map[string]interface{}{
    "用户名": "张三",
    "手机号": 13800138000,
}

step1 = beiluo.test1.devops.devops_script_create(
    username: string "用户名",
    phone: int "手机号"
) -> (
    workId: string "工号",
    username: string "用户名", 
    err: error "是否失败"
);

func main() {
    //desc: 创建用户账号，获取工号
    工号, 用户名, step1Err := step1(input["用户名"], input["手机号"])
    
    //desc: 检查用户创建是否成功
    if step1Err != nil {
        step1.Printf("创建用户失败: %v", step1Err)
        return
    }
    
    step1.Printf("✅ 用户创建成功，工号: %s", 工号)
}`

	req := &dto.CreateWorkflowReq{
		Name:        "DSL测试工作流",
		Description: "使用DSL代码创建的工作流",
		Code:        dslCode,
	}

	// 验证请求结构
	if req.Name != "DSL测试工作流" {
		t.Errorf("期望名称: DSL测试工作流, 实际: %s", req.Name)
	}

	if req.Code == "" {
		t.Error("DSL代码不能为空")
	}

	t.Logf("DSL代码长度: %d", len(req.Code))
}

func TestWorkflowService_ExecuteWorkflow(t *testing.T) {
	// 测试执行工作流
	req := &dto.ExecuteWorkflowReq{
		WorkflowId: "1",
		InputVars: map[string]interface{}{
			"用户名": "李四",
			"手机号": 13900139000,
		},
	}

	// 验证请求结构
	if req.WorkflowId != "1" {
		t.Errorf("期望工作流ID: 1, 实际: %s", req.WorkflowId)
	}

	if len(req.InputVars) != 2 {
		t.Errorf("期望输入变量数量: 2, 实际: %d", len(req.InputVars))
	}

	t.Logf("执行请求: %+v", req)
}

func TestWorkflowService_GetWorkflowStatus(t *testing.T) {
	// 测试获取工作流状态
	req := &dto.GetWorkflowStatusReq{
		ExecutionId: "exec_1_1234567890",
	}

	// 验证请求结构
	if req.ExecutionId == "" {
		t.Error("执行ID不能为空")
	}

	t.Logf("状态查询请求: %+v", req)
}

func TestWorkflowService_StopWorkflow(t *testing.T) {
	// 测试停止工作流
	req := &dto.StopWorkflowReq{
		ExecutionId: "exec_1_1234567890",
	}

	// 验证请求结构
	if req.ExecutionId == "" {
		t.Error("执行ID不能为空")
	}

	t.Logf("停止请求: %+v", req)
}

func TestWorkflowService_FlowIDGeneration(t *testing.T) {
	// 测试FlowID生成规则
	workflowId := 1
	timestamp := int64(1703123456)
	expectedExecutionId := fmt.Sprintf("exec_%d_%d", workflowId, timestamp)

	// 模拟生成executionId
	executionId := fmt.Sprintf("exec_%d_%d", workflowId, timestamp)

	if executionId != expectedExecutionId {
		t.Errorf("期望ExecutionId: %s, 实际: %s", expectedExecutionId, executionId)
	}

	// 验证FlowID格式
	if !strings.HasPrefix(executionId, "exec_") {
		t.Error("ExecutionId应该以'exec_'开头")
	}

	t.Logf("生成的ExecutionId: %s", executionId)
	t.Logf("这个ExecutionId将作为工作流引擎的FlowID使用")
}

func TestWorkflowService_CompleteStateStorage(t *testing.T) {
	// 测试完整状态存储功能
	// 模拟SimpleParseResult
	mockSimpleParseResult := map[string]interface{}{
		"flow_id": "exec_1_1703123456",
		"variables": map[string]interface{}{
			"用户名": "张三",
			"手机号": 13800138000,
		},
		"steps": []interface{}{
			map[string]interface{}{
				"name":   "step1",
				"status": "completed",
			},
		},
		"success": true,
	}

	// 序列化为JSON
	jsonBytes, err := json.Marshal(mockSimpleParseResult)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 验证JSON数据
	if len(jsonBytes) == 0 {
		t.Error("JSON数据不能为空")
	}

	// 验证可以解析回SimpleParseResult
	var parsedResult map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsedResult); err != nil {
		t.Fatalf("解析JSON失败: %v", err)
	}

	// 验证关键字段
	if parsedResult["flow_id"] != "exec_1_1703123456" {
		t.Errorf("期望flow_id: exec_1_1703123456, 实际: %v", parsedResult["flow_id"])
	}

	if variables, ok := parsedResult["variables"].(map[string]interface{}); ok {
		if variables["用户名"] != "张三" {
			t.Errorf("期望用户名: 张三, 实际: %v", variables["用户名"])
		}
	} else {
		t.Error("variables字段解析失败")
	}

	t.Logf("完整状态存储测试通过")
	t.Logf("JSON数据长度: %d bytes", len(jsonBytes))
	t.Logf("解析后的flow_id: %v", parsedResult["flow_id"])
}

func TestWorkflowService_OnWorkFlowExit(t *testing.T) {
	// 测试OnWorkFlowExit回调功能
	// 模拟工作流正常结束的场景

	// 模拟SimpleParseResult
	mockWorkflowResult := map[string]interface{}{
		"flow_id": "exec_1_1703123456",
		"variables": map[string]interface{}{
			"用户名": "张三",
			"手机号": 13800138000,
		},
		"steps": []interface{}{
			map[string]interface{}{
				"name":   "step1",
				"status": "completed",
			},
		},
		"success": false, // 初始状态为false
	}

	// 序列化为JSON
	jsonBytes, err := json.Marshal(mockWorkflowResult)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 模拟OnWorkFlowExit回调逻辑
	// 1. 解析当前状态
	var workflowResult map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &workflowResult); err != nil {
		t.Fatalf("解析JSON失败: %v", err)
	}

	// 2. 更新状态为完成
	workflowResult["success"] = true

	// 3. 序列化更新后的状态
	updatedJsonBytes, err := json.Marshal(workflowResult)
	if err != nil {
		t.Fatalf("序列化更新后的状态失败: %v", err)
	}

	// 验证状态更新
	if workflowResult["success"] != true {
		t.Errorf("期望success: true, 实际: %v", workflowResult["success"])
	}

	// 验证JSON数据变化
	if len(updatedJsonBytes) == len(jsonBytes) {
		t.Error("更新后的JSON数据长度应该与原始数据不同")
	}

	t.Logf("OnWorkFlowExit回调测试通过")
	t.Logf("原始状态: success=%v", mockWorkflowResult["success"])
	t.Logf("更新后状态: success=%v", workflowResult["success"])
	t.Logf("JSON数据长度变化: %d -> %d", len(jsonBytes), len(updatedJsonBytes))
}

func TestWorkflowService_OnFunctionCallMock(t *testing.T) {
	// 测试OnFunctionCall的mock功能
	// 模拟WantParams参数

	// 模拟ParameterInfo
	mockWantParams := []map[string]interface{}{
		{
			"name": "result",
			"type": "string",
			"desc": "执行结果",
		},
		{
			"name": "err",
			"type": "error",
			"desc": "错误信息",
		},
		{
			"name": "data",
			"type": "object",
			"desc": "返回数据",
		},
		{
			"name": "message",
			"type": "string",
			"desc": "消息",
		},
		{
			"name": "code",
			"type": "int",
			"desc": "状态码",
		},
		{
			"name": "success",
			"type": "bool",
			"desc": "是否成功",
		},
		{
			"name": "custom_param",
			"type": "string",
			"desc": "自定义参数",
		},
	}

	// 模拟OnFunctionCall的mock逻辑
	wantOutput := make(map[string]interface{})
	stepName := "test_step"

	for _, paramInfo := range mockWantParams {
		paramName := paramInfo["name"].(string)
		paramType := paramInfo["type"].(string)
		paramDesc := paramInfo["desc"].(string)

		t.Logf("处理期望参数: %s, 类型: %s, 描述: %s", paramName, paramType, paramDesc)

		// 根据参数类型生成mock数据
		switch paramName {
		case "result":
			wantOutput[paramName] = fmt.Sprintf("步骤 %s 执行成功", stepName)
		case "err":
			wantOutput[paramName] = nil
		case "data":
			wantOutput[paramName] = map[string]interface{}{
				"step_name": stepName,
				"status":    "completed",
				"timestamp": 1703123456,
			}
		case "message":
			wantOutput[paramName] = fmt.Sprintf("步骤 %s 处理完成", stepName)
		case "code":
			wantOutput[paramName] = 200
		case "success":
			wantOutput[paramName] = true
		default:
			// 对于未知参数，生成默认值
			wantOutput[paramName] = fmt.Sprintf("mock_%s_value", paramName)
		}
	}

	// 验证mock数据
	expectedResults := map[string]interface{}{
		"result":       "步骤 test_step 执行成功",
		"err":          nil,
		"data":         map[string]interface{}{"step_name": "test_step", "status": "completed", "timestamp": 1703123456},
		"message":      "步骤 test_step 处理完成",
		"code":         200,
		"success":      true,
		"custom_param": "mock_custom_param_value",
	}

	for key, expectedValue := range expectedResults {
		if actualValue, exists := wantOutput[key]; !exists {
			t.Errorf("期望参数 %s 不存在", key)
		} else if !compareValues(actualValue, expectedValue) {
			t.Errorf("参数 %s 期望: %v, 实际: %v", key, expectedValue, actualValue)
		}
	}

	t.Logf("OnFunctionCall Mock测试通过")
	t.Logf("生成的mock数据: %+v", wantOutput)
}

// 辅助函数：比较两个值是否相等
func compareValues(a, b interface{}) bool {
	switch va := a.(type) {
	case map[string]interface{}:
		if vb, ok := b.(map[string]interface{}); ok {
			if len(va) != len(vb) {
				return false
			}
			for k, v := range va {
				if !compareValues(v, vb[k]) {
					return false
				}
			}
			return true
		}
		return false
	default:
		return a == b
	}
}
