package repo

import (
	"context"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type KnowledgeDocumentRepo struct {
	db *gorm.DB
}

func NewKnowledgeDocumentRepo(db *gorm.DB) *KnowledgeDocumentRepo {
	return &KnowledgeDocumentRepo{db: db}
}

// Create 创建文档
func (r *KnowledgeDocumentRepo) Create(ctx context.Context, doc *model.KnowledgeDocument) error {
	logger.Debugf(ctx, "开始创建文档，kb_key: %s, doc_id: %s", doc.KBKey, doc.DocID)
	return r.db.Create(doc).Error
}

// GetByDocID 根据文档ID获取文档
func (r *KnowledgeDocumentRepo) GetByDocID(ctx context.Context, docID, user string) (*model.KnowledgeDocument, error) {
	var doc model.KnowledgeDocument
	err := r.db.Where("doc_id = ? AND user = ?", docID, user).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// List 获取文档列表
func (r *KnowledgeDocumentRepo) List(ctx context.Context, kbKey, user, title string, limit, offset int) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument
	query := r.db.Where("kb_key = ? AND user = ?", kbKey, user)

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&docs).Error
	return docs, err
}

// Count 获取文档总数
func (r *KnowledgeDocumentRepo) Count(ctx context.Context, kbKey, user, title string) (int64, error) {
	var count int64
	query := r.db.Model(&model.KnowledgeDocument{}).Where("kb_key = ? AND user = ?", kbKey, user)

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	err := query.Count(&count).Error
	return count, err
}

// Update 更新文档
func (r *KnowledgeDocumentRepo) Update(ctx context.Context, doc *model.KnowledgeDocument) error {
	logger.Debugf(ctx, "开始更新文档，doc_id: %s", doc.DocID)
	return r.db.Model(&model.KnowledgeDocument{}).
		Where("doc_id = ? AND user = ?", doc.DocID, doc.User).
		Updates(doc).Error
}

// Delete 删除文档
func (r *KnowledgeDocumentRepo) Delete(ctx context.Context, docID, user string) error {
	return r.db.Where("doc_id = ? AND user = ?", docID, user).Delete(&model.KnowledgeDocument{}).Error
}

// Search 搜索文档内容
func (r *KnowledgeDocumentRepo) Search(ctx context.Context, kbKey, user, query string, limit int) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument
	err := r.db.Where("kb_key = ? AND user = ? AND (title LIKE ? OR content LIKE ?)",
		kbKey, user, "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&docs).Error
	return docs, err
}
