package dto

import (
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto/base"
)

// ChatReq 发送消息请求
type ChatReq struct {
	SessionID    string `json:"session_id" form:"session_id"` // 可选，不传则创建新会话
	Message      string `json:"message" form:"message" validate:"required"`
	Model        string `json:"model" form:"model" data:"default_value:glm-4.5"`
	Router       string `json:"router" form:"router"`               // 路由标识，用于管理不同路由的会话
	KnowledgeKey string `json:"knowledge_key" form:"knowledge_key"` // 知识库key，用于检索相关知识
	Stream       bool   `json:"stream" form:"stream" data:"default_value:false"`
}

// ChatResp 发送消息响应
type ChatResp struct {
	SessionID string      `json:"session_id"`
	Content   string      `json:"content"`
	Model     string      `json:"model"`
	Usage     interface{} `json:"usage,omitempty"`
}

// ChatSessionListReq 获取会话列表请求
type ChatSessionListReq struct {
	base.PageInfoReq
	Title  string `json:"title" form:"title"`   // 按标题搜索
	Router string `json:"router" form:"router"` // 按路由过滤
}

// ChatMessageListReq 获取消息历史请求
type ChatMessageListReq struct {
	base.PageInfoReq
	SessionID string `json:"session_id" form:"session_id" validate:"required"`
	Router    string `json:"router" form:"router"` // 路由参数，用于管理不同路由的会话
}

// ChatSessionWithDetails 带详情的会话信息
type ChatSessionWithDetails struct {
	*model.ChatSession
	LastMessage *model.ChatMessage `json:"last_message,omitempty"` // 最后一条消息
}

// ChatMessageWithDetails 带详情的消息信息
type ChatMessageWithDetails struct {
	*model.ChatMessage
	Session *model.ChatSession `json:"session,omitempty"` // 所属会话
}

// StreamChunk 流式响应数据块
type StreamChunk struct {
	Content   string      `json:"content,omitempty"`
	Error     string      `json:"error,omitempty"`
	Done      bool        `json:"done"`
	Usage     interface{} `json:"usage,omitempty"`
	SessionID string      `json:"session_id,omitempty"`
	Title     string      `json:"title,omitempty"`
}

// UpdateSessionTitleReq 更新会话标题请求
type UpdateSessionTitleReq struct {
	SessionID string `json:"session_id" form:"session_id" validate:"required"`
	Title     string `json:"title" form:"title" validate:"required,max=100"`
}
