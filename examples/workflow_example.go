package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yunhanshu-net/function-server/pkg/config"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/service"
)

func main() {
	// 初始化数据库
	cfg := config.DBConfig{
		Type:         "mysql",
		Host:         "localhost",
		Port:         3306,
		User:         "root",
		Password:     "password",
		Name:         "function_server",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  3600,
	}

	if err := db.Init(cfg); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 创建工作流服务
	workflowService := service.NewWorkflowService(db.GetDB())

	// 创建上下文
	ctx := context.Background()

	// 示例1: 使用JSON数据创建工作流
	fmt.Println("=== 示例1: 使用JSON数据创建工作流 ===")
	createWorkflowWithJSON(workflowService, ctx)

	// 示例2: 使用DSL代码创建工作流
	fmt.Println("\n=== 示例2: 使用DSL代码创建工作流 ===")
	createWorkflowWithDSL(workflowService, ctx)

	// 示例3: 执行工作流
	fmt.Println("\n=== 示例3: 执行工作流 ===")
	executeWorkflow(workflowService, ctx)
}

func createWorkflowWithJSON(workflowService *service.WorkflowService, ctx context.Context) {
	// 准备JSON数据
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
				"input_params": []interface{}{
					map[string]interface{}{
						"name": "username",
						"type": "string",
						"desc": "用户名",
					},
					map[string]interface{}{
						"name": "phone",
						"type": "int",
						"desc": "手机号",
					},
				},
				"output_params": []interface{}{
					map[string]interface{}{
						"name": "workId",
						"type": "string",
						"desc": "工号",
					},
					map[string]interface{}{
						"name": "username",
						"type": "string",
						"desc": "用户名",
					},
					map[string]interface{}{
						"name": "err",
						"type": "error",
						"desc": "是否失败",
					},
				},
			},
		},
		"main_func": map[string]interface{}{
			"statements": []interface{}{
				map[string]interface{}{
					"type":     "function-call",
					"function": "step1",
					"args": []interface{}{
						map[string]interface{}{
							"value":    "input[\"用户名\"]",
							"type":     "input",
							"is_input": true,
						},
						map[string]interface{}{
							"value":    "input[\"手机号\"]",
							"type":     "input",
							"is_input": true,
						},
					},
					"returns": []interface{}{
						map[string]interface{}{
							"value": "工号",
						},
						map[string]interface{}{
							"value": "用户名",
						},
						map[string]interface{}{
							"value": "step1Err",
						},
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Printf("序列化JSON失败: %v", err)
		return
	}

	req := &dto.CreateWorkflowReq{
		Name:        "JSON测试工作流",
		Description: "使用JSON数据创建的工作流",
		JsonData:    json.RawMessage(jsonBytes),
	}

	resp, err := workflowService.CreateWorkflow(ctx, req)
	if err != nil {
		log.Printf("创建工作流失败: %v", err)
		return
	}

	fmt.Printf("✅ 工作流创建成功!\n")
	fmt.Printf("   工作流ID: %s\n", resp.WorkflowId)
	fmt.Printf("   名称: %s\n", resp.Name)
	fmt.Printf("   状态: %s\n", resp.Status)
	fmt.Printf("   创建时间: %s\n", resp.CreatedAt)
}

func createWorkflowWithDSL(workflowService *service.WorkflowService, ctx context.Context) {
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

	resp, err := workflowService.CreateWorkflow(ctx, req)
	if err != nil {
		log.Printf("创建工作流失败: %v", err)
		return
	}

	fmt.Printf("✅ 工作流创建成功!\n")
	fmt.Printf("   工作流ID: %s\n", resp.WorkflowId)
	fmt.Printf("   名称: %s\n", resp.Name)
	fmt.Printf("   状态: %s\n", resp.Status)
	fmt.Printf("   创建时间: %s\n", resp.CreatedAt)
}

func executeWorkflow(workflowService *service.WorkflowService, ctx context.Context) {
	// 执行工作流
	req := &dto.ExecuteWorkflowReq{
		WorkflowId: "1", // 假设工作流ID为1
		InputVars: map[string]interface{}{
			"用户名": "李四",
			"手机号": 13900139000,
		},
	}

	resp, err := workflowService.ExecuteWorkflow(ctx, req)
	if err != nil {
		log.Printf("执行工作流失败: %v", err)
		return
	}

	fmt.Printf("✅ 工作流开始执行!\n")
	fmt.Printf("   执行ID: %s\n", resp.ExecutionId)
	fmt.Printf("   状态: %s\n", resp.Status)
	fmt.Printf("   消息: %s\n", resp.Message)
	fmt.Printf("   注意: 执行ID就是工作流引擎的FlowID\n")

	// 等待一段时间后查询状态
	fmt.Println("\n等待3秒后查询状态...")
	// time.Sleep(3 * time.Second)

	// 查询状态
	statusReq := &dto.GetWorkflowStatusReq{
		ExecutionId: resp.ExecutionId,
	}

	statusResp, err := workflowService.GetWorkflowStatus(ctx, statusReq)
	if err != nil {
		log.Printf("查询工作流状态失败: %v", err)
		return
	}

	fmt.Printf("📊 工作流状态:\n")
	fmt.Printf("   执行ID: %s\n", statusResp.ExecutionId)
	fmt.Printf("   状态: %s\n", statusResp.Status)
	fmt.Printf("   进度: %.2f%%\n", statusResp.Progress)
	fmt.Printf("   当前步骤: %s\n", statusResp.CurrentStep)
	fmt.Printf("   开始时间: %s\n", statusResp.StartTime)
	if statusResp.EndTime != "" {
		fmt.Printf("   结束时间: %s\n", statusResp.EndTime)
	}
}
