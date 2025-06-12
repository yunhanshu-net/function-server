package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DoubaoConfig 豆包配置
type DoubaoConfig struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout time.Duration
}

// DoubaoClient 豆包客户端
type DoubaoClient struct {
	config     DoubaoConfig
	httpClient *http.Client
}

// DoubaoMessage 豆包消息格式
type DoubaoMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DoubaoRequest 豆包请求格式
type DoubaoRequest struct {
	Model          string          `json:"model"`
	Messages       []DoubaoMessage `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// ResponseFormat 响应格式
type ResponseFormat struct {
	Type string `json:"type"`
}

// DoubaoResponse 豆包响应格式
type DoubaoResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// NewDoubaoClient 创建豆包客户端
func NewDoubaoClient(config DoubaoConfig) *DoubaoClient {
	if config.BaseURL == "" {
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if config.Model == "" {
		config.Model = "doubao-1-5-pro-32k-250115"
	}
	if config.Timeout == 0 {
		config.Timeout = 180 * time.Second
	}

	return &DoubaoClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Chat 发送聊天请求
func (c *DoubaoClient) Chat(ctx context.Context, messages []DoubaoMessage) (*DoubaoResponse, error) {
	req := DoubaoRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   4000,
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	var doubaoResp DoubaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&doubaoResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if doubaoResp.Error != nil {
		return nil, fmt.Errorf("doubao api error: %s", doubaoResp.Error.Message)
	}

	if len(doubaoResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &doubaoResp, nil
}

// ChatWithJSON 发送聊天请求并解析JSON响应
func (c *DoubaoClient) ChatWithJSON(ctx context.Context, messages []DoubaoMessage, result interface{}) error {
	resp, err := c.Chat(ctx, messages)
	if err != nil {
		return err
	}

	content := resp.Choices[0].Message.Content

	// 处理豆包返回的JSON中可能包含反引号的问题
	// 将反引号包围的代码块转换为正确的JSON字符串格式
	content = fixJSONBackticks(content)

	if err := json.Unmarshal([]byte(content), result); err != nil {
		return fmt.Errorf("unmarshal response content failed: %w, content: %s", err, content)
	}

	return nil
}

// fixJSONBackticks 修复JSON中的反引号问题
func fixJSONBackticks(content string) string {
	// 更智能的JSON修复策略
	// 处理豆包可能返回的各种格式问题

	// 1. 处理反引号包围的代码块
	re1 := regexp.MustCompile(`"code":\s*` + "`" + `([^` + "`" + `]*)` + "`")
	content = re1.ReplaceAllStringFunc(content, func(match string) string {
		// 提取反引号中的内容
		parts := strings.Split(match, "`")
		if len(parts) >= 2 {
			codeContent := parts[1]
			// 转义JSON中的特殊字符
			codeContent = escapeJSONString(codeContent)
			return `"code": "` + codeContent + `"`
		}
		return match
	})

	// 2. 修复JSON标签格式问题
	// 处理 "json:"username"` 这种格式错误
	re2 := regexp.MustCompile(`"json:"([^"]*)"([^"]*)`)
	content = re2.ReplaceAllString(content, `"json:\"$1\"$2`)

	// 3. 修复更复杂的JSON标签问题
	// 处理 `json:"field"` 在字符串中的情况
	content = regexp.MustCompile(`(\w+)\s+string\s+"json:"([^"]*)"(\s*)`).ReplaceAllString(content, `$1 string $3`)
	content = regexp.MustCompile(`"json:"([^"]*)"([\s]*)`).ReplaceAllString(content, `\"json:\\\"$1\\\"\"$2`)

	// 4. 处理代码中的反引号
	content = strings.ReplaceAll(content, "`json:", "\\\"json:")
	content = strings.ReplaceAll(content, "`", "\\\"")

	// 5. 修复可能出现的其他格式错误
	content = strings.ReplaceAll(content, "if!emailRegex", "if !emailRegex")

	// 6. 最后处理，确保代码字段内的换行正确转义
	content = regexp.MustCompile(`"code":\s*"([^"]*(?:\\.[^"]*)*)"(?:,|\s*})`).ReplaceAllStringFunc(content, func(match string) string {
		// 找到code字段的值并重新转义
		parts := regexp.MustCompile(`"code":\s*"([^"]*(?:\\.[^"]*)*)"((?:,|\s*}))`).FindStringSubmatch(match)
		if len(parts) >= 3 {
			codeValue := parts[1]
			ending := parts[2]
			// 确保所有换行符都被正确转义
			codeValue = strings.ReplaceAll(codeValue, "\n", "\\n")
			codeValue = strings.ReplaceAll(codeValue, "\t", "\\t")
			return `"code": "` + codeValue + `"` + ending
		}
		return match
	})

	return content
}

// escapeJSONString 转义JSON字符串中的特殊字符
func escapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	return s
}
