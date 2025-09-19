package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/function-server/service"
	"github.com/yunhanshu-net/pkg/logger"
)

type ChatAPI struct {
	chatService *service.ChatService
}

func NewChatAPI(chatService *service.ChatService) *ChatAPI {
	return &ChatAPI{
		chatService: chatService,
	}
}

// SendMessage 发送消息
// @Summary 发送消息
// @Description 发送消息到AI，支持普通和流式两种模式
// @Tags 对话
// @Accept json
// @Produce json
// @Param req body dto.ChatReq true "发送消息请求"
// @Success 200 {object} response.Response{data=dto.ChatResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/chat [post]
func (api *ChatAPI) SendMessage(c *gin.Context) {
	var req dto.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	// 从中间件获取用户信息
	user := c.GetString("user")
	if user == "" {
		response.Error(c, "用户未登录")
		return
	}

	// 将用户信息设置到context中
	ctx := contextx.WithRequestUserName(context.Background(), user)
	resp, err := api.chatService.SendMessage(ctx, &req)
	if err != nil {
		logger.Errorf(c, "[Chat] 发送消息失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// SendMessageStream 发送流式消息
// @Summary 发送流式消息
// @Description 发送消息到AI，返回流式响应
// @Tags 对话
// @Accept json
// @Produce json
// @Param req body dto.ChatReq true "发送消息请求"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/chat/stream [post]
func (api *ChatAPI) SendMessageStream(c *gin.Context) {
	var req dto.ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	// 强制设置为流式模式
	req.Stream = true

	// 从中间件获取用户信息
	user := c.GetString("user")
	if user == "" {
		response.Error(c, "用户未登录")
		return
	}

	// 将用户信息设置到context中
	ctx := contextx.WithRequestUserName(context.Background(), user)
	stream, err := api.chatService.SendMessageStream(ctx, &req)
	if err != nil {
		logger.Errorf(c, "[Chat] 发送流式消息失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 流式返回数据
	for chunk := range stream {
		if chunk.Error != "" {
			c.SSEvent("error", gin.H{"error": chunk.Error})
			c.Writer.Flush()
			return
		}

		if chunk.Content != "" {
			c.SSEvent("message", gin.H{"content": chunk.Content})
			c.Writer.Flush()
		}

		if chunk.Done {
			if chunk.Usage != nil {
				c.SSEvent("usage", gin.H{"usage": chunk.Usage})
			}
			c.SSEvent("done", gin.H{
				"done":       true,
				"session_id": chunk.SessionID,
				"title":      chunk.Title,
			})
			c.Writer.Flush()
			return
		}
	}
}

// GetSessions 获取会话列表
// @Summary 获取会话列表
// @Description 获取用户的对话会话列表
// @Tags 对话
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param title query string false "标题搜索"
// @Success 200 {object} response.Response{data=base.Paginated[[]dto.ChatSessionWithDetails]} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/chat/sessions [get]
func (api *ChatAPI) GetSessions(c *gin.Context) {
	var req dto.ChatSessionListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := api.chatService.GetSessions(c, &req)
	if err != nil {
		logger.Errorf(c, "[Chat] 获取会话列表失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetMessages 获取消息历史
// @Summary 获取消息历史
// @Description 获取指定会话的消息历史
// @Tags 对话
// @Accept json
// @Produce json
// @Param session_id query string true "会话ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=base.Paginated[[]dto.ChatMessageWithDetails]} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/chat/messages [get]
func (api *ChatAPI) GetMessages(c *gin.Context) {
	var req dto.ChatMessageListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := api.chatService.GetMessages(c, &req)
	if err != nil {
		logger.Errorf(c, "[Chat] 获取消息历史失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// UpdateSessionTitle 更新会话标题
// @Summary 更新会话标题
// @Description 更新指定会话的标题
// @Tags 对话
// @Accept json
// @Produce json
// @Param req body dto.UpdateSessionTitleReq true "更新会话标题请求"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/chat/session/title [put]
func (api *ChatAPI) UpdateSessionTitle(c *gin.Context) {
	var req dto.UpdateSessionTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	err := api.chatService.UpdateSessionTitle(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[Chat] 更新会话标题失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
