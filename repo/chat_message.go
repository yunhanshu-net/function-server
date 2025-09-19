package repo

import (
	"context"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type ChatMessageRepo struct {
	db *gorm.DB
}

func NewChatMessageRepo(db *gorm.DB) *ChatMessageRepo {
	return &ChatMessageRepo{db: db}
}

// Create 创建消息
func (r *ChatMessageRepo) Create(ctx context.Context, message *model.ChatMessage) error {
	logger.Debugf(ctx, "开始创建消息，session_id: %s, role: %s", message.SessionID, message.Role)
	return r.db.Create(message).Error
}

// GetBySessionID 根据会话ID获取消息列表
func (r *ChatMessageRepo) GetBySessionID(ctx context.Context, sessionID, user string, limit, offset int) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := r.db.Where("session_id = ? AND user = ?", sessionID, user).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	return messages, err
}

// GetMessageCount 获取会话消息总数
func (r *ChatMessageRepo) GetMessageCount(ctx context.Context, sessionID, user string) (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatMessage{}).Where("session_id = ? AND user = ?", sessionID, user).Count(&count).Error
	return count, err
}

// GetLastMessage 获取会话最后一条消息
func (r *ChatMessageRepo) GetLastMessage(ctx context.Context, sessionID, user string) (*model.ChatMessage, error) {
	var message model.ChatMessage
	err := r.db.Where("session_id = ? AND user = ?", sessionID, user).
		Order("created_at DESC").
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetMessagesForLLM 获取用于LLM的消息历史（最近N条）
func (r *ChatMessageRepo) GetMessagesForLLM(ctx context.Context, sessionID, user string, limit int) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := r.db.Where("session_id = ? AND user = ?", sessionID, user).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	// 反转顺序，让消息按时间正序排列
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, err
}

// CreateWithUsage 创建消息并保存使用统计
func (r *ChatMessageRepo) CreateWithUsage(ctx context.Context, message *model.ChatMessage, promptTokens, completionTokens, totalTokens int) error {
	message.PromptTokens = promptTokens
	message.CompletionTokens = completionTokens
	message.TotalTokens = totalTokens
	return r.Create(ctx, message)
}

// DeleteBySessionID 删除会话的所有消息
func (r *ChatMessageRepo) DeleteBySessionID(ctx context.Context, sessionID, user string) error {
	return r.db.Where("session_id = ? AND user = ?", sessionID, user).Delete(&model.ChatMessage{}).Error
}
