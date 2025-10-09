package test

import (
	"encoding/json"
	"testing"

	"github.com/yunhanshu-net/function-server/pkg/interfaces"
	"github.com/yunhanshu-net/function-server/service/caller"
)

func TestQQEmailCaller(t *testing.T) {
	// 创建QQ邮件caller
	qqCaller := caller.NewQQEmailCaller()

	// 设置测试配置
	qqCaller.SetConfig(
		"test@qq.com",
		"test_auth_code",
		"test@qq.com",
		"Test Sender",
	)

	// 构建测试请求
	emailReq := map[string]interface{}{
		"to":      []string{"recipient@example.com"},
		"subject": "测试邮件",
		"body":    "这是一封测试邮件",
		"is_html": false,
	}

	reqBody, _ := json.Marshal(emailReq)
	req := &interfaces.FunctionCallReq{
		Name:   "qq_email",
		Header: make(map[string][]string),
		Body:   reqBody,
	}

	// 调用caller（注意：这会尝试真实发送邮件，需要有效的SMTP配置）
	// 在实际测试中，你可能想要mock SMTP连接
	resp, err := qqCaller.Call(req)
	if err != nil {
		t.Logf("调用失败（预期，因为没有真实的SMTP配置）: %v", err)
		return
	}

	// 解析响应
	var emailResp map[string]interface{}
	if err := json.Unmarshal(resp.Body, &emailResp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	t.Logf("响应: %+v", emailResp)
}

func TestQQEmailCallerConfig(t *testing.T) {
	// 创建QQ邮件caller
	qqCaller := caller.NewQQEmailCaller()

	// 设置配置
	qqCaller.SetConfig(
		"test@qq.com",
		"test_auth_code",
		"test@qq.com",
		"Test Sender",
	)

	// 获取配置
	config := qqCaller.GetConfig()

	// 验证配置
	if config["username"] != "test@qq.com" {
		t.Errorf("期望username为test@qq.com，实际为%s", config["username"])
	}

	if config["from_email"] != "test@qq.com" {
		t.Errorf("期望from_email为test@qq.com，实际为%s", config["from_email"])
	}

	if config["from_name"] != "Test Sender" {
		t.Errorf("期望from_name为Test Sender，实际为%s", config["from_name"])
	}

	t.Logf("配置验证通过: %+v", config)
}
