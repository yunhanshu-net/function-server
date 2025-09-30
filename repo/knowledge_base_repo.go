package repo

import (
	"context"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type KnowledgeBaseRepo struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepo(db *gorm.DB) *KnowledgeBaseRepo {
	return &KnowledgeBaseRepo{db: db}
}

// Create 创建知识库
func (r *KnowledgeBaseRepo) Create(ctx context.Context, kb *model.KnowledgeBase) error {
	logger.Debugf(ctx, "开始创建知识库，kb_key: %s", kb.KBKey)
	return r.db.Create(kb).Error
}

// GetByKBKey 根据知识库Key获取知识库
func (r *KnowledgeBaseRepo) GetByKBKey(ctx context.Context, kbKey, user string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	err := r.db.Where("kb_key = ? AND user = ?", kbKey, user).First(&kb).Error
	if err != nil {
		return nil, err
	}
	return &kb, nil
}

// Update 更新知识库
func (r *KnowledgeBaseRepo) Update(ctx context.Context, kb *model.KnowledgeBase) error {
	return r.db.Save(kb).Error
}

// Delete 删除知识库
func (r *KnowledgeBaseRepo) Delete(ctx context.Context, kbKey, user string) error {
	return r.db.Where("kb_key = ? AND user = ?", kbKey, user).Delete(&model.KnowledgeBase{}).Error
}

// List 获取知识库列表
func (r *KnowledgeBaseRepo) List(ctx context.Context, user string, name, router, status string, limit, offset int) ([]*model.KnowledgeBase, error) {
	var kbs []*model.KnowledgeBase
	query := r.db.Where("user = ?", user)

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if router != "" {
		query = query.Where("router = ?", router)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&kbs).Error
	return kbs, err
}

// Count 获取知识库总数
func (r *KnowledgeBaseRepo) Count(ctx context.Context, user string, name, router, status string) (int64, error) {
	var count int64
	query := r.db.Model(&model.KnowledgeBase{}).Where("user = ?", user)

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if router != "" {
		query = query.Where("router = ?", router)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}

// UpdateDocumentCount 更新文档数量
func (r *KnowledgeBaseRepo) UpdateDocumentCount(ctx context.Context, kbKey, user string) error {
	var count int64
	err := r.db.Model(&model.KnowledgeDocument{}).Where("kb_key = ? AND user = ?", kbKey, user).Count(&count).Error
	if err != nil {
		return err
	}

	return r.db.Model(&model.KnowledgeBase{}).
		Where("kb_key = ? AND user = ?", kbKey, user).
		Update("document_count", count).Error
}

// UpdateContentHash 更新知识库内容哈希值
func (r *KnowledgeBaseRepo) UpdateContentHash(ctx context.Context, kbKey, user, contentHash string) error {
	logger.Debugf(ctx, "更新知识库内容哈希值，kb_key: %s, hash: %s", kbKey, contentHash)
	return r.db.Model(&model.KnowledgeBase{}).
		Where("kb_key = ? AND user = ?", kbKey, user).
		Update("content_hash", contentHash).Error
}
