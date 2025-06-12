package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yunhanshu-net/pkg/logger"
)

// QwenConfig 千问配置结构体
type QwenConfig struct {
	APIKey  string        `json:"api_key"`  // API密钥
	BaseURL string        `json:"base_url"` // API基础URL
	Model   string        `json:"model"`    // 模型名称
	Timeout time.Duration `json:"timeout"`  // 请求超时时间
}

// QwenClient 千问客户端
type QwenClient struct {
	config     *QwenConfig
	httpClient *http.Client
}

// QwenMessage 千问消息结构体
type QwenMessage struct {
	Role    string `json:"role"`    // 角色：system, user, assistant
	Content string `json:"content"` // 消息内容
}

// QwenRequest 千问请求结构体
type QwenRequest struct {
	Model       string        `json:"model"`                 // 模型名称
	Messages    []QwenMessage `json:"messages"`              // 消息列表
	Temperature float32       `json:"temperature,omitempty"` // 温度参数
	MaxTokens   int           `json:"max_tokens,omitempty"`  // 最大令牌数
	Stream      bool          `json:"stream,omitempty"`      // 是否流式输出
}

// QwenChoice 千问选择结构体
type QwenChoice struct {
	Index   int         `json:"index"`
	Message QwenMessage `json:"message"`
	Reason  string      `json:"finish_reason"`
}

// QwenUsage 千问使用情况统计
type QwenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// QwenResponse 千问响应结构体
type QwenResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []QwenChoice `json:"choices"`
	Usage   QwenUsage    `json:"usage"`
}

// QwenErrorResponse 千问错误响应结构体
type QwenErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// NewQwenClient 创建新的千问客户端
func NewQwenClient(config *QwenConfig) *QwenClient {
	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://dashscope.aliyuncs.com/api/v1"
	}
	if config.Model == "" {
		config.Model = "qwen-turbo"
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &QwenClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Chat 执行千问对话
func (c *QwenClient) Chat(ctx context.Context, messages []QwenMessage) (*QwenResponse, error) {
	// 构建请求
	req := &QwenRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      false,
	}

	return c.doRequest(ctx, req)
}

// ChatWithJSON 执行千问对话并返回JSON格式响应
func (c *QwenClient) ChatWithJSON(ctx context.Context, messages []QwenMessage) (string, error) {
	// 添加JSON格式要求到系统消息
	jsonSystemMessage := QwenMessage{
		Role:    "system",
		Content: "请确保你的回复是有效的JSON格式，不要包含任何额外的文字或格式标记如```json等。直接返回纯JSON内容。",
	}

	// 将JSON格式要求插入到消息列表开头
	allMessages := append([]QwenMessage{jsonSystemMessage}, messages...)

	resp, err := c.Chat(ctx, allMessages)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("千问返回空的响应选择")
	}

	content := resp.Choices[0].Message.Content
	logger.Debugf(ctx, "千问JSON响应: %s", content)

	return content, nil
}

// doRequest 执行HTTP请求
func (c *QwenClient) doRequest(ctx context.Context, req *QwenRequest) (*QwenResponse, error) {
	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化千问请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/services/aigc/text-generation/generation", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建千问HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("User-Agent", "Function-Server/1.0")

	logger.Debugf(ctx, "千问请求: %s", string(reqBody))

	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("千问HTTP请求失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取千问响应失败: %w", err)
	}

	logger.Debugf(ctx, "千问原始响应 (状态码: %d): %s", httpResp.StatusCode, string(respBody))

	// 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		var errorResp QwenErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return nil, fmt.Errorf("千问API错误 (状态码: %d): %s - %s", httpResp.StatusCode, errorResp.Error.Code, errorResp.Error.Message)
		}
		return nil, fmt.Errorf("千问HTTP错误 (状态码: %d): %s", httpResp.StatusCode, string(respBody))
	}

	// 解析响应
	var qwenResp QwenResponse
	if err := json.Unmarshal(respBody, &qwenResp); err != nil {
		return nil, fmt.Errorf("解析千问响应失败: %w, 响应内容: %s", err, string(respBody))
	}

	// 验证响应
	if len(qwenResp.Choices) == 0 {
		return nil, fmt.Errorf("千问返回空的响应选择")
	}

	logger.Debugf(ctx, "千问响应解析成功，使用令牌: %d", qwenResp.Usage.TotalTokens)

	return &qwenResp, nil
}

// GetConfig 获取客户端配置
func (c *QwenClient) GetConfig() *QwenConfig {
	return c.config
}

// UpdateConfig 更新客户端配置
func (c *QwenClient) UpdateConfig(config *QwenConfig) {
	c.config = config
	c.httpClient.Timeout = config.Timeout
}
