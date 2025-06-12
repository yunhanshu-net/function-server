package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yunhanshu-net/function-server/pkg/llm"
)

func main() {
	fmt.Println("测试千问(Qwen)客户端功能...")

	// 千问配置 - 需要替换为实际的API Key
	config := &llm.QwenConfig{
		APIKey:  "sk-your-qwen-api-key-here", // 请替换为实际的千问API Key
		BaseURL: "https://dashscope.aliyuncs.com/api/v1",
		Model:   "qwen-turbo", // 使用turbo模型进行测试
		Timeout: 60 * time.Second,
	}

	// 创建千问客户端
	client := llm.NewQwenClient(config)
	ctx := context.Background()

	fmt.Printf("千问配置: %+v\n\n", config)

	// 测试1: 基础对话功能
	fmt.Println("=== 测试1: 基础对话功能 ===")
	testBasicChat(ctx, client)

	// 测试2: JSON格式响应
	fmt.Println("\n=== 测试2: JSON格式响应 ===")
	testJSONResponse(ctx, client)

	// 测试3: function-go代码生成
	fmt.Println("\n=== 测试3: function-go代码生成 ===")
	testCodeGeneration(ctx, client)
}

func testBasicChat(ctx context.Context, client *llm.QwenClient) {
	messages := []llm.QwenMessage{
		{
			Role:    "user",
			Content: "你好，请简单介绍一下千问AI助手的特点",
		},
	}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		log.Printf("基础对话测试失败: %v", err)
		return
	}

	if len(resp.Choices) > 0 {
		fmt.Printf("千问回复: %s\n", resp.Choices[0].Message.Content)
		fmt.Printf("令牌使用: 输入=%d, 输出=%d, 总计=%d\n",
			resp.Usage.PromptTokens,
			resp.Usage.CompletionTokens,
			resp.Usage.TotalTokens)
	}
}

func testJSONResponse(ctx context.Context, client *llm.QwenClient) {
	messages := []llm.QwenMessage{
		{
			Role:    "user",
			Content: `请以JSON格式返回一个简单的用户信息示例，包含姓名、年龄、邮箱字段`,
		},
	}

	jsonContent, err := client.ChatWithJSON(ctx, messages)
	if err != nil {
		log.Printf("JSON响应测试失败: %v", err)
		return
	}

	fmt.Printf("千问JSON响应: %s\n", jsonContent)

	// 验证是否为有效JSON
	if strings.Contains(jsonContent, "{") && strings.Contains(jsonContent, "}") {
		fmt.Println("✅ JSON格式验证通过")
	} else {
		fmt.Println("❌ JSON格式验证失败")
	}
}

func testCodeGeneration(ctx context.Context, client *llm.QwenClient) {
	messages := []llm.QwenMessage{
		{
			Role: "system",
			Content: `你是一个专业的function-go框架代码生成专家。请严格按照以下JSON格式返回：

{
  "tags": "函数标签",
  "level": 复杂程度1-100,
  "code": "完整的Go代码",
  "think": "思考过程",
  "package": "包名",
  "en_name": "英文函数名",
  "cn_name": "中文描述"
}`,
		},
		{
			Role:    "user",
			Content: "生成一个简单的Hello World函数，基于function-go框架",
		},
	}

	jsonContent, err := client.ChatWithJSON(ctx, messages)
	if err != nil {
		log.Printf("代码生成测试失败: %v", err)
		return
	}

	fmt.Printf("千问代码生成响应: %s\n", jsonContent)

	// 验证代码生成质量
	validateCodeGeneration(jsonContent)
}

func validateCodeGeneration(jsonContent string) {
	fmt.Println("\n=== 千问代码生成质量验证 ===")

	checks := map[string]bool{
		"包含JSON格式":    strings.Contains(jsonContent, "{") && strings.Contains(jsonContent, "}"),
		"包含代码字段":      strings.Contains(jsonContent, "code"),
		"包含package声明": strings.Contains(jsonContent, "package"),
		"包含import语句":  strings.Contains(jsonContent, "import"),
		"包含函数名字段":     strings.Contains(jsonContent, "en_name"),
		"包含中文描述":      strings.Contains(jsonContent, "cn_name"),
		"包含复杂度等级":     strings.Contains(jsonContent, "level"),
		"包含思考过程":      strings.Contains(jsonContent, "think"),
	}

	passed := 0
	total := len(checks)

	for check, result := range checks {
		if result {
			fmt.Printf("✅ %s\n", check)
			passed++
		} else {
			fmt.Printf("❌ %s\n", check)
		}
	}

	fmt.Printf("\n千问代码生成质量: %.1f%% (%d/%d)\n", float64(passed)/float64(total)*100, passed, total)

	if passed >= 6 {
		fmt.Println("🎉 千问客户端测试成功！")
	} else {
		fmt.Println("⚠️  千问客户端需要进一步调优")
	}
}
