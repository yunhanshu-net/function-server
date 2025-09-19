package repo

import (
	"context"
	"time"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type ChatSessionRepo struct {
	db *gorm.DB
}

func NewChatSessionRepo(db *gorm.DB) *ChatSessionRepo {
	return &ChatSessionRepo{db: db}
}

// Create 创建会话
func (r *ChatSessionRepo) Create(ctx context.Context, session *model.ChatSession) error {
	logger.Debugf(ctx, "开始创建会话，session_id: %s", session.SessionID)
	return r.db.Create(session).Error
}

// GetBySessionID 根据会话ID获取会话
func (r *ChatSessionRepo) GetBySessionID(ctx context.Context, sessionID, user string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("session_id = ? AND user = ?", sessionID, user).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update 更新会话
func (r *ChatSessionRepo) Update(ctx context.Context, session *model.ChatSession) error {
	return r.db.Save(session).Error
}

// UpdateLastMessage 更新最后消息时间和消息数量
func (r *ChatSessionRepo) UpdateLastMessage(ctx context.Context, sessionID, user string) error {
	now := time.Now()
	return r.db.Model(&model.ChatSession{}).
		Where("session_id = ? AND user = ?", sessionID, user).
		Updates(map[string]interface{}{
			"last_message_at": now,
			"message_count":   gorm.Expr("message_count + 1"),
			"updated_by":      contextx.GetRequestUserName(ctx),
		}).Error
}

// GetSessions 获取用户会话列表
func (r *ChatSessionRepo) GetSessions(ctx context.Context, user string, router string, limit, offset int) ([]*model.ChatSession, error) {
	var sessions []*model.ChatSession
	query := r.db.Where("user = ?", user)
	if router != "" {
		query = query.Where("router = ?", router)
	}
	err := query.Order("last_message_at DESC").
		Limit(limit).Offset(offset).
		Find(&sessions).Error
	return sessions, err
}

// GetSessionCount 获取用户会话总数
func (r *ChatSessionRepo) GetSessionCount(ctx context.Context, user string, router string) (int64, error) {
	var count int64
	query := r.db.Model(&model.ChatSession{}).Where("user = ?", user)
	if router != "" {
		query = query.Where("router = ?", router)
	}
	err := query.Count(&count).Error
	return count, err
}

// UpdateTitle 更新会话标题
func (r *ChatSessionRepo) UpdateTitle(ctx context.Context, sessionID, user, title string) error {
	return r.db.Model(&model.ChatSession{}).
		Where("session_id = ? AND user = ?", sessionID, user).
		Update("title", title).Error
}

// Delete 删除会话
func (r *ChatSessionRepo) Delete(ctx context.Context, sessionID, user string) error {
	return r.db.Where("session_id = ? AND user = ?", sessionID, user).Delete(&model.ChatSession{}).Error
}
