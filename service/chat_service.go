package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-go/pkg/llms"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/config"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/dto/base"
	"github.com/yunhanshu-net/function-server/pkg/utils"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/function-server/repo"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type ChatService struct {
	db                *gorm.DB
	sessionRepo       *repo.ChatSessionRepo
	messageRepo       *repo.ChatMessageRepo
	knowledgeBaseRepo *repo.KnowledgeBaseRepo
	documentRepo      *repo.KnowledgeDocumentRepo
	glmClient         *llms.GLMClient
}

func NewChatService(db *gorm.DB) *ChatService {
	// 创建带长超时配置的GLM客户端
	options := llms.DefaultClientOptions() // 使用默认的1200秒超时配置

	// 从配置文件获取API Key
	cfg := config.Get()
	apiKey := cfg.GLMConfig.APIKey
	if apiKey == "" {
		logger.Warnf(context.Background(), "[ChatService] GLM API Key未配置，将使用空字符串")
	} else {
		logger.Infof(context.Background(), "[ChatService] 从配置文件获取GLM API Key，长度: %d", len(apiKey))
	}

	glmClient := llms.NewGLMClientWithOptions(apiKey, options)

	return &ChatService{
		db:                db,
		sessionRepo:       repo.NewChatSessionRepo(db),
		messageRepo:       repo.NewChatMessageRepo(db),
		knowledgeBaseRepo: repo.NewKnowledgeBaseRepo(db),
		documentRepo:      repo.NewKnowledgeDocumentRepo(db),
		glmClient:         glmClient,
	}
}

// SendMessage 发送消息
func (s *ChatService) SendMessage(ctx context.Context, req *dto.ChatReq) (*dto.ChatResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 获取或创建会话
	session, err := s.getOrCreateSession(ctx, req.SessionID, user, req.Model, req.Router, req.KnowledgeKey)
	if err != nil {
		return nil, errors.Wrap(err, "获取或创建会话失败")
	}

	// 保存用户消息
	userMessage := &model.ChatMessage{
		SessionID:    session.SessionID,
		Role:         "user",
		Content:      req.Message,
		Model:        req.Model,
		Router:       req.Router,
		KnowledgeKey: req.KnowledgeKey,
		User:         user,
	}
	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, errors.Wrap(err, "保存用户消息失败")
	}

	// 如果是新会话且只有一条用户消息，生成智能标题
	if session.MessageCount == 0 {
		newTitle := s.generateSessionTitleFromMessage(ctx, req.Message)
		if newTitle != session.Title {
			if err := s.sessionRepo.UpdateTitle(ctx, session.SessionID, user, newTitle); err != nil {
				logger.Warnf(ctx, "更新会话标题失败: %v", err)
			} else {
				session.Title = newTitle
			}
		}
	}

	// 获取历史消息用于LLM
	historyMessages, err := s.messageRepo.GetMessagesForLLM(ctx, session.SessionID, user, 20)
	if err != nil {
		return nil, errors.Wrap(err, "获取历史消息失败")
	}

	// 智能知识库检索：第一次问话检索知识库，后续问话检查历史记录避免重复加载
	var knowledgeContext string
	if req.KnowledgeKey != "" {
		knowledgeContext = s.retrieveKnowledgeWithHistory(ctx, req.KnowledgeKey, req.Message, user, session.SessionID)
	}

	// 调用GLM
	llmMessages := s.convertToGLMMessages(historyMessages)

	// 如果有知识库上下文，添加到系统消息中
	if knowledgeContext != "" {
		systemMessage := llms.Message{
			Role:    "system",
			Content: fmt.Sprintf("以下是相关知识库信息，请基于这些信息回答用户问题：\n\n%s\n\n请根据以上信息回答用户的问题。", knowledgeContext),
		}
		llmMessages = append([]llms.Message{systemMessage}, llmMessages...)
	}

	// 设置普通请求的超时时间为60秒
	chatTimeout := 60 * time.Second
	llmReq := &llms.ChatRequest{
		Messages:    llmMessages,
		Temperature: 0.7,
		Model:       req.Model,
		Timeout:     &chatTimeout, // 设置60秒超时
	}

	llmResp, err := s.glmClient.Chat(ctx, llmReq)
	if err != nil {
		return nil, errors.Wrap(err, "调用GLM失败")
	}

	// 保存AI回复
	aiMessage := &model.ChatMessage{
		SessionID:    session.SessionID,
		Role:         "assistant",
		Content:      llmResp.Content,
		Model:        req.Model,
		Router:       req.Router,
		KnowledgeKey: req.KnowledgeKey,
		User:         user,
	}
	// 处理usage字段
	var promptTokens, completionTokens, totalTokens int
	if llmResp.Usage != nil {
		promptTokens = llmResp.Usage.PromptTokens
		completionTokens = llmResp.Usage.CompletionTokens
		totalTokens = llmResp.Usage.TotalTokens
	}
	if err := s.messageRepo.CreateWithUsage(ctx, aiMessage, promptTokens, completionTokens, totalTokens); err != nil {
		return nil, errors.Wrap(err, "保存AI回复失败")
	}

	// 更新会话
	if err := s.sessionRepo.UpdateLastMessage(ctx, session.SessionID, user); err != nil {
		logger.Warnf(ctx, "更新会话失败: %v", err)
	}

	return &dto.ChatResp{
		SessionID: session.SessionID,
		Content:   llmResp.Content,
		Model:     req.Model,
		Usage:     llmResp.Usage,
	}, nil
}

// SendMessageStream 发送流式消息
func (s *ChatService) SendMessageStream(ctx context.Context, req *dto.ChatReq) (<-chan *dto.StreamChunk, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 获取或创建会话
	session, err := s.getOrCreateSession(ctx, req.SessionID, user, req.Model, req.Router, req.KnowledgeKey)
	if err != nil {
		return nil, errors.Wrap(err, "获取或创建会话失败")
	}

	// 保存用户消息
	userMessage := &model.ChatMessage{
		SessionID:    session.SessionID,
		Role:         "user",
		Content:      req.Message,
		Model:        req.Model,
		Router:       req.Router,
		KnowledgeKey: req.KnowledgeKey,
		User:         user,
	}
	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, errors.Wrap(err, "保存用户消息失败")
	}

	// 获取历史消息用于LLM
	historyMessages, err := s.messageRepo.GetMessagesForLLM(ctx, session.SessionID, user, 20)
	if err != nil {
		return nil, errors.Wrap(err, "获取历史消息失败")
	}

	// 智能知识库检索：第一次问话检索知识库，后续问话检查历史记录避免重复加载
	var knowledgeContext string
	if req.KnowledgeKey != "" {
		knowledgeContext = s.retrieveKnowledgeWithHistory(ctx, req.KnowledgeKey, req.Message, user, session.SessionID)
	}

	// 创建响应通道
	chunkChan := make(chan *dto.StreamChunk, 100)

	go func() {
		defer close(chunkChan)
		var content strings.Builder

		// 调用GLM流式接口
		llmMessages := s.convertToGLMMessages(historyMessages)

		// 如果有知识库上下文，添加到系统消息中
		if knowledgeContext != "" {
			systemMessage := llms.Message{
				Role:    "system",
				Content: fmt.Sprintf("以下是相关知识库信息，请基于这些信息回答用户问题：\n\n%s\n\n请根据以上信息回答用户的问题。", knowledgeContext),
			}
			llmMessages = append([]llms.Message{systemMessage}, llmMessages...)
		}

		// 设置流式请求的超时时间为60秒
		streamTimeout := 600 * time.Second
		llmReq := &llms.ChatRequest{
			Messages:    llmMessages,
			Temperature: 0.7,
			Model:       req.Model,
			MaxTokens:   32000,
			Timeout:     &streamTimeout, // 设置60秒超时
		}

		// 添加调试日志
		logger.Infof(ctx, "[ChatStream] 开始调用GLM流式接口，模型: %s, 消息数量: %d", req.Model, len(llmMessages))
		logger.Infof(ctx, "[ChatStream] GLM请求参数: %+v", llmReq)

		stream, streamErr := s.glmClient.ChatStream(ctx, llmReq)
		if streamErr != nil {
			logger.Errorf(ctx, "[ChatStream] GLM流式接口调用失败: %v", streamErr)
			chunkChan <- &dto.StreamChunk{
				Error: streamErr.Error(),
				Done:  true,
			}
			return
		}

		logger.Infof(ctx, "[ChatStream] GLM流式接口调用成功，开始处理流式响应")

		var usage interface{}
		for chunk := range stream {
			if chunk.Error != "" {
				chunkChan <- &dto.StreamChunk{
					Error: chunk.Error,
					Done:  true,
				}
				return
			}

			if chunk.Content != "" {
				content.WriteString(chunk.Content)
				chunkChan <- &dto.StreamChunk{
					Content: chunk.Content,
					Done:    false,
				}
			}

			if chunk.Usage != nil {
				usage = chunk.Usage
			}

			if chunk.Done {
				break
			}
		}

		// 保存完整的AI回复
		aiMessage := &model.ChatMessage{
			SessionID:    session.SessionID,
			Role:         "assistant",
			Content:      content.String(),
			Model:        req.Model,
			Router:       req.Router,
			KnowledgeKey: req.KnowledgeKey,
			User:         user,
		}
		// 处理usage字段
		var promptTokens, completionTokens, totalTokens int
		if usage != nil {
			if usageStruct, ok := usage.(*llms.Usage); ok {
				promptTokens = usageStruct.PromptTokens
				completionTokens = usageStruct.CompletionTokens
				totalTokens = usageStruct.TotalTokens
			}
		}
		if err := s.messageRepo.CreateWithUsage(ctx, aiMessage, promptTokens, completionTokens, totalTokens); err != nil {
			logger.Errorf(ctx, "保存AI回复失败: %v", err)
		}

		// 更新会话
		if err := s.sessionRepo.UpdateLastMessage(ctx, session.SessionID, user); err != nil {
			logger.Warnf(ctx, "更新会话失败: %v", err)
		}

		chunkChan <- &dto.StreamChunk{
			Done:      true,
			Usage:     usage,
			SessionID: session.SessionID,
			Title:     session.Title,
		}
	}()

	return chunkChan, nil
}

// GetSessions 获取会话列表
func (s *ChatService) GetSessions(ctx context.Context, req *dto.ChatSessionListReq) (*base.Paginated[[]*dto.ChatSessionWithDetails], error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 构建查询
	db := s.db.Model(&model.ChatSession{}).Where("user = ?", user)

	if req.Title != "" {
		db = db.Where("title LIKE ?", "%"+req.Title+"%")
	}

	if req.Router != "" {
		db = db.Where("router = ?", req.Router)
	}
	db = db.Order("id DESC")

	// 使用AutoPaginate处理分页
	var sessions []*model.ChatSession
	pageInfo := &utils.PageInfo{
		Page:     req.Page,
		PageSize: req.PageSize,
		Sorts:    req.Sorts,
	}
	result, err := utils.AutoPaginate(ctx, db, &model.ChatSession{}, &sessions, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "查询会话列表失败")
	}

	// 转换为带详情的结构
	var resultList []*dto.ChatSessionWithDetails
	for _, session := range sessions {
		item := &dto.ChatSessionWithDetails{
			ChatSession: session,
		}

		// 获取最后一条消息
		lastMessage, err := s.messageRepo.GetLastMessage(ctx, session.SessionID, user)
		if err == nil {
			item.LastMessage = lastMessage
		}

		resultList = append(resultList, item)
	}

	return &base.Paginated[[]*dto.ChatSessionWithDetails]{
		Items:       resultList,
		CurrentPage: result.CurrentPage,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		PageSize:    result.PageSize,
	}, nil
}

// GetMessages 获取消息历史
func (s *ChatService) GetMessages(ctx context.Context, req *dto.ChatMessageListReq) (*base.Paginated[[]*dto.ChatMessageWithDetails], error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 构建查询
	db := s.db.Model(&model.ChatMessage{}).Where("session_id = ? AND user = ?", req.SessionID, user)

	if req.Router != "" {
		db = db.Where("router = ?", req.Router)
	}

	// 使用AutoPaginate处理分页
	var messages []*model.ChatMessage
	pageInfo := &utils.PageInfo{
		Page:     req.Page,
		PageSize: req.PageSize,
		Sorts:    req.Sorts,
	}
	result, err := utils.AutoPaginate(ctx, db, &model.ChatMessage{}, &messages, pageInfo)
	if err != nil {
		return nil, errors.Wrap(err, "查询消息历史失败")
	}

	// 转换为带详情的结构
	var resultList []*dto.ChatMessageWithDetails
	for _, message := range messages {
		item := &dto.ChatMessageWithDetails{
			ChatMessage: message,
		}
		resultList = append(resultList, item)
	}

	return &base.Paginated[[]*dto.ChatMessageWithDetails]{
		Items:       resultList,
		CurrentPage: result.CurrentPage,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		PageSize:    result.PageSize,
	}, nil
}

// UpdateSessionTitle 更新会话标题
func (s *ChatService) UpdateSessionTitle(ctx context.Context, req *dto.UpdateSessionTitleReq) error {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return errors.New("用户未登录")
	}

	// 验证会话是否存在
	_, err := s.sessionRepo.GetBySessionID(ctx, req.SessionID, user)
	if err != nil {
		return errors.Wrap(err, "会话不存在")
	}

	// 更新标题
	if err := s.sessionRepo.UpdateTitle(ctx, req.SessionID, user, req.Title); err != nil {
		return errors.Wrap(err, "更新会话标题失败")
	}

	return nil
}

// getOrCreateSession 获取或创建会话
func (s *ChatService) getOrCreateSession(ctx context.Context, sessionID, user, modelName, router, knowledgeKey string) (*model.ChatSession, error) {
	if sessionID != "" {
		// 尝试获取现有会话
		session, err := s.sessionRepo.GetBySessionID(ctx, sessionID, user)
		if err == nil {
			// 如果传入了新的knowledge_key且与现有会话不同，更新会话
			if knowledgeKey != "" && session.KnowledgeKey != knowledgeKey {
				session.KnowledgeKey = knowledgeKey
				if err := s.sessionRepo.Update(ctx, session); err != nil {
					logger.Warnf(ctx, "更新会话知识库失败: %v", err)
				}
			}
			return session, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// 创建新会话
	newSessionID := uuid.New().String()
	// 注意：这里先使用默认标题，后续会在保存用户消息后更新标题
	title := s.generateSessionTitle(ctx, user)

	session := &model.ChatSession{
		Base: model.Base{
			CreatedBy: user,
			UpdatedBy: user,
		},
		SessionID:    newSessionID,
		Title:        title,
		Model:        modelName,
		Router:       router,
		KnowledgeKey: knowledgeKey,
		MessageCount: 0,
		User:         user,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// convertToGLMMessages 转换为GLM消息格式
func (s *ChatService) convertToGLMMessages(messages []*model.ChatMessage) []llms.Message {
	var llmMessages []llms.Message
	for _, msg := range messages {
		llmMessages = append(llmMessages, llms.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return llmMessages
}

// generateSessionTitle 生成会话标题
func (s *ChatService) generateSessionTitle(ctx context.Context, user string) string {
	now := time.Now()
	return fmt.Sprintf("对话 %s", now.Format("2006-01-02 15:04"))
}

// generateSessionTitleFromMessage 根据用户消息生成会话标题
func (s *ChatService) generateSessionTitleFromMessage(ctx context.Context, message string) string {
	// 如果消息为空，使用默认标题
	if message == "" {
		return s.generateSessionTitle(ctx, "")
	}

	// 直接使用用户的第一条消息作为标题，限制长度
	title := strings.TrimSpace(message)

	// 如果标题太长，截取前50个字符并加上省略号
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	return title
}

// retrieveKnowledgeWithHistory 智能知识库检索（带历史记录检查）
func (s *ChatService) retrieveKnowledgeWithHistory(ctx context.Context, knowledgeKey, query, user, sessionID string) string {
	logger.Infof(ctx, "[知识库检索] 开始检查知识库: %s, 会话: %s", knowledgeKey, sessionID)

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, knowledgeKey, user)
	if err != nil {
		logger.Warnf(ctx, "[知识库检索] 知识库不存在或无权访问: %s, error: %v", knowledgeKey, err)
		return ""
	}
	logger.Infof(ctx, "[知识库检索] 知识库存在，继续检查历史记录")

	// 检查历史消息中是否已经使用过相同的知识库
	hasUsed := s.hasKnowledgeBeenUsedInHistory(ctx, sessionID, user, knowledgeKey)
	logger.Infof(ctx, "[知识库检索] 历史记录检查结果: %v", hasUsed)

	if hasUsed {
		logger.Infof(ctx, "[知识库检索] 知识库 %s 已在历史记录中使用过，跳过重复加载", knowledgeKey)
		return ""
	}

	// 第一次使用该知识库，进行检索
	logger.Infof(ctx, "[知识库检索] 首次使用知识库 %s，开始检索内容", knowledgeKey)
	knowledgeContext := s.retrieveKnowledge(ctx, knowledgeKey, query, user)
	logger.Infof(ctx, "[知识库检索] 检索完成，内容长度: %d", len(knowledgeContext))
	return knowledgeContext
}

// hasKnowledgeBeenUsedInHistory 检查历史记录中是否已经使用过指定的知识库
func (s *ChatService) hasKnowledgeBeenUsedInHistory(ctx context.Context, sessionID, user, knowledgeKey string) bool {
	// 获取会话的历史消息（排除当前正在处理的消息）
	messages, err := s.messageRepo.GetMessagesForLLM(ctx, sessionID, user, 50) // 获取最近50条消息
	if err != nil {
		logger.Warnf(ctx, "获取历史消息失败: %v", err)
		return false
	}

	logger.Infof(ctx, "[知识库检索] 检查历史消息，共 %d 条", len(messages))

	// 检查历史消息中是否已经使用过相同的knowledgeKey
	// 注意：这里需要排除当前正在处理的消息，只检查assistant角色的消息
	for _, msg := range messages {
		logger.Infof(ctx, "[知识库检索] 检查消息: role=%s, knowledgeKey=%s", msg.Role, msg.KnowledgeKey)
		// 只检查assistant角色的消息，因为只有assistant消息才会包含知识库内容
		if msg.Role == "assistant" && msg.KnowledgeKey == knowledgeKey {
			logger.Infof(ctx, "[知识库检索] 找到匹配的assistant消息，knowledgeKey=%s", msg.KnowledgeKey)
			return true
		}
	}

	logger.Infof(ctx, "[知识库检索] 历史消息中未找到匹配的知识库")
	return false
}

// retrieveKnowledge 检索知识库内容
func (s *ChatService) retrieveKnowledge(ctx context.Context, knowledgeKey, query, user string) string {
	logger.Infof(ctx, "[知识库检索] 开始检索知识库内容: %s", knowledgeKey)

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, knowledgeKey, user)
	if err != nil {
		logger.Warnf(ctx, "[知识库检索] 知识库不存在或无权访问: %s, error: %v", knowledgeKey, err)
		return ""
	}

	// 直接获取知识库中的所有文档，不基于用户问题搜索
	docs, err := s.documentRepo.List(ctx, knowledgeKey, user, "", 5, 0) // 获取前5个文档
	if err != nil {
		logger.Warnf(ctx, "[知识库检索] 获取知识库文档失败: %v", err)
		return ""
	}

	logger.Infof(ctx, "[知识库检索] 找到 %d 个文档", len(docs))

	if len(docs) == 0 {
		logger.Infof(ctx, "[知识库检索] 知识库 %s 中没有文档", knowledgeKey)
		return ""
	}

	// 构建知识库上下文
	var contextBuilder strings.Builder
	contextBuilder.WriteString("【知识库内容】\n")

	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, doc.Title))
		// 使用完整的文档内容，不进行截断
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n")
	}

	context := contextBuilder.String()
	logger.Infof(ctx, "[知识库检索] 成功构建知识库上下文，长度: %d", len(context))
	return context
}
