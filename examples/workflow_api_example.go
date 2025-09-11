package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 工作流API使用示例
func main() {
	baseURL := "http://localhost:8080/api/v1"

	fmt.Println("=== 工作流API使用示例 ===")

	// 示例1: 创建工作流
	fmt.Println("\n1. 创建工作流")
	createWorkflowExample(baseURL)

	// 示例2: 执行工作流
	fmt.Println("\n2. 执行工作流")
	executeWorkflowExample(baseURL)

	// 示例3: 查询工作流状态
	fmt.Println("\n3. 查询工作流状态")
	getWorkflowStatusExample(baseURL)

	// 示例4: 获取工作流列表
	fmt.Println("\n4. 获取工作流列表")
	listWorkflowExample(baseURL)

	// 示例5: 获取工作流详情
	fmt.Println("\n5. 获取工作流详情")
	getWorkflowDetailExample(baseURL)

	// 示例6: 停止工作流
	fmt.Println("\n6. 停止工作流")
	stopWorkflowExample(baseURL)
}

// 创建工作流示例
func createWorkflowExample(baseURL string) {
	// 使用DSL代码创建工作流
	dslCode := `func main() {
    step1.Printf("开始执行用户创建流程")
    
    // 创建用户
    step1.Printf("正在创建用户: %s", input["用户名"])
    step1.Printf("用户手机号: %s", input["手机号"])
    
    step1.Printf("✅ 用户创建成功，工号: %s", 工号)
}`

	req := map[string]interface{}{
		"name":        "用户创建工作流",
		"description": "自动创建用户的工作流",
		"code":        dslCode,
	}

	// 发送请求
	resp, err := sendRequest("POST", baseURL+"/workflow", req)
	if err != nil {
		fmt.Printf("❌ 创建工作流失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流创建成功!\n")
	fmt.Printf("   响应: %s\n", resp)
}

// 执行工作流示例
func executeWorkflowExample(baseURL string) {
	req := map[string]interface{}{
		"workflow_id": "1", // 假设工作流ID为1
		"input_vars": map[string]interface{}{
			"用户名": "张三",
			"手机号": 13800138000,
		},
	}

	// 发送请求
	resp, err := sendRequest("POST", baseURL+"/workflow/execute", req)
	if err != nil {
		fmt.Printf("❌ 执行工作流失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流开始执行!\n")
	fmt.Printf("   响应: %s\n", resp)

	// 等待一段时间
	fmt.Println("   等待3秒后查询状态...")
	time.Sleep(3 * time.Second)
}

// 查询工作流状态示例
func getWorkflowStatusExample(baseURL string) {
	req := map[string]interface{}{
		"execution_id": "exec_1_1234567890", // 假设执行ID
	}

	// 发送请求
	resp, err := sendRequest("POST", baseURL+"/workflow/status", req)
	if err != nil {
		fmt.Printf("❌ 查询工作流状态失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流状态查询成功!\n")
	fmt.Printf("   响应: %s\n", resp)
}

// 获取工作流列表示例
func listWorkflowExample(baseURL string) {
	// 使用GET请求 + query参数
	url := baseURL + "/workflow/list?page=1&page_size=10&name=测试&status=active"

	// 发送请求
	resp, err := sendGetRequest(url)
	if err != nil {
		fmt.Printf("❌ 获取工作流列表失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流列表获取成功!\n")
	fmt.Printf("   响应: %s\n", resp)
}

// 获取工作流详情示例
func getWorkflowDetailExample(baseURL string) {
	// 使用GET请求 + query参数
	url := baseURL + "/workflow/detail?workflow_id=1"

	// 发送请求
	resp, err := sendGetRequest(url)
	if err != nil {
		fmt.Printf("❌ 获取工作流详情失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流详情获取成功!\n")
	fmt.Printf("   响应: %s\n", resp)
}

// 停止工作流示例
func stopWorkflowExample(baseURL string) {
	req := map[string]interface{}{
		"execution_id": "exec_1_1234567890", // 假设执行ID
	}

	// 发送请求
	resp, err := sendRequest("POST", baseURL+"/workflow/stop", req)
	if err != nil {
		fmt.Printf("❌ 停止工作流失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 工作流停止成功!\n")
	fmt.Printf("   响应: %s\n", resp)
}

// 发送HTTP请求的辅助函数
func sendRequest(method, url string, data interface{}) (string, error) {
	// 序列化请求数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// 创建HTTP请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", "test_user") // 模拟用户信息

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// 发送GET请求的辅助函数
func sendGetRequest(url string) (string, error) {
	// 创建HTTP请求
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("User", "test_user") // 模拟用户信息

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP错误: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
