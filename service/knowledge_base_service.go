package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/dto/base"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/function-server/repo"
	"github.com/yunhanshu-net/pkg/logger"
	"gorm.io/gorm"
)

type KnowledgeBaseService struct {
	db                *gorm.DB
	knowledgeBaseRepo *repo.KnowledgeBaseRepo
	documentRepo      *repo.KnowledgeDocumentRepo
}

func NewKnowledgeBaseService(db *gorm.DB) *KnowledgeBaseService {
	return &KnowledgeBaseService{
		db:                db,
		knowledgeBaseRepo: repo.NewKnowledgeBaseRepo(db),
		documentRepo:      repo.NewKnowledgeDocumentRepo(db),
	}
}

// CreateKnowledgeBase 创建知识库
func (s *KnowledgeBaseService) CreateKnowledgeBase(ctx context.Context, req *dto.KnowledgeBaseCreateReq) (*dto.KnowledgeBaseResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 检查知识库Key是否已存在
	existing, err := s.knowledgeBaseRepo.GetByKBKey(ctx, req.Name, user)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrap(err, "检查知识库Key失败")
	}
	if existing != nil {
		return nil, errors.New("知识库Key已存在")
	}

	// 生成知识库Key（基于名称）
	kbKey := s.generateKBKey(req.Name)

	kb := &model.KnowledgeBase{
		Base: model.Base{
			CreatedBy: user,
			UpdatedBy: user,
		},
		KBKey:         kbKey,
		Name:          req.Name,
		Description:   req.Description,
		Router:        req.Router,
		Status:        "active",
		DocumentCount: 0,
		User:          user,
	}

	if err := s.knowledgeBaseRepo.Create(ctx, kb); err != nil {
		return nil, errors.Wrap(err, "创建知识库失败")
	}

	return &dto.KnowledgeBaseResp{KnowledgeBase: kb}, nil
}

// UpdateKnowledgeBase 更新知识库
func (s *KnowledgeBaseService) UpdateKnowledgeBase(ctx context.Context, req *dto.KnowledgeBaseUpdateReq) error {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return errors.New("用户未登录")
	}

	// 获取现有知识库
	kb, err := s.knowledgeBaseRepo.GetByKBKey(ctx, req.KBKey, user)
	if err != nil {
		return errors.Wrap(err, "知识库不存在")
	}

	// 更新字段
	if req.Name != "" {
		kb.Name = req.Name
	}
	if req.Description != "" {
		kb.Description = req.Description
	}
	if req.Status != "" {
		kb.Status = req.Status
	}

	if err := s.knowledgeBaseRepo.Update(ctx, kb); err != nil {
		return errors.Wrap(err, "更新知识库失败")
	}

	return nil
}

// DeleteKnowledgeBase 删除知识库
func (s *KnowledgeBaseService) DeleteKnowledgeBase(ctx context.Context, kbKey string) error {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return errors.New("用户未登录")
	}

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, kbKey, user)
	if err != nil {
		return errors.Wrap(err, "知识库不存在")
	}

	// 删除知识库（级联删除相关文档）
	if err := s.knowledgeBaseRepo.Delete(ctx, kbKey, user); err != nil {
		return errors.Wrap(err, "删除知识库失败")
	}

	return nil
}

// GetKnowledgeBase 获取知识库详情
func (s *KnowledgeBaseService) GetKnowledgeBase(ctx context.Context, kbKey string) (*dto.KnowledgeBaseResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	kb, err := s.knowledgeBaseRepo.GetByKBKey(ctx, kbKey, user)
	if err != nil {
		return nil, errors.Wrap(err, "获取知识库失败")
	}

	return &dto.KnowledgeBaseResp{KnowledgeBase: kb}, nil
}

// ListKnowledgeBases 获取知识库列表
func (s *KnowledgeBaseService) ListKnowledgeBases(ctx context.Context, req *dto.KnowledgeBaseListReq) (*base.Paginated[[]*dto.KnowledgeBaseResp], error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 构建查询
	limit := req.PageSize
	if limit <= 0 {
		limit = 10
	}
	offset := (req.Page - 1) * limit

	kbs, err := s.knowledgeBaseRepo.List(ctx, user, req.Name, req.Router, req.Status, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "查询知识库列表失败")
	}

	// 获取总数
	total, err := s.knowledgeBaseRepo.Count(ctx, user, req.Name, req.Router, req.Status)
	if err != nil {
		return nil, errors.Wrap(err, "查询知识库总数失败")
	}

	// 转换为响应格式
	var resultList []*dto.KnowledgeBaseResp
	for _, kb := range kbs {
		resultList = append(resultList, &dto.KnowledgeBaseResp{KnowledgeBase: kb})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &base.Paginated[[]*dto.KnowledgeBaseResp]{
		Items:       resultList,
		CurrentPage: req.Page,
		TotalCount:  total,
		TotalPages:  totalPages,
		PageSize:    limit,
	}, nil
}

// UploadDocument 上传文档
func (s *KnowledgeBaseService) UploadDocument(ctx context.Context, req *dto.KnowledgeDocumentUploadReq) (*dto.KnowledgeDocumentResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, req.KBKey, user)
	if err != nil {
		return nil, errors.Wrap(err, "知识库不存在")
	}

	// 生成文档ID
	docID := uuid.New().String()

	// 创建文档
	doc := &model.KnowledgeDocument{
		Base: model.Base{
			CreatedBy: user,
			UpdatedBy: user,
		},
		KBKey:    req.KBKey,
		DocID:    docID,
		Title:    req.Title,
		Content:  req.Content,
		FileType: req.FileType,
		FileSize: int64(len(req.Content)),
		Status:   "completed",
		User:     user,
	}

	if err := s.documentRepo.Create(ctx, doc); err != nil {
		return nil, errors.Wrap(err, "创建文档失败")
	}

	// 更新知识库文档数量
	if err := s.knowledgeBaseRepo.UpdateDocumentCount(ctx, req.KBKey, user); err != nil {
		logger.Warnf(ctx, "更新知识库文档数量失败: %v", err)
	}

	return &dto.KnowledgeDocumentResp{KnowledgeDocument: doc}, nil
}

// UpdateDocument 更新文档
func (s *KnowledgeBaseService) UpdateDocument(ctx context.Context, req *dto.KnowledgeDocumentUpdateReq) (*dto.KnowledgeDocumentUpdateResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 获取现有文档
	doc, err := s.documentRepo.GetByDocID(ctx, req.DocID, user)
	if err != nil {
		return nil, errors.Wrap(err, "文档不存在")
	}

	// 更新字段
	if req.Title != "" {
		doc.Title = req.Title
	}
	if req.Content != "" {
		doc.Content = req.Content
		doc.FileSize = int64(len(req.Content))
	}
	if req.FileType != "" {
		doc.FileType = req.FileType
	}

	// 更新文档
	if err := s.documentRepo.Update(ctx, doc); err != nil {
		return nil, errors.Wrap(err, "更新文档失败")
	}

	return &dto.KnowledgeDocumentUpdateResp{
		DocID:     doc.DocID,
		Title:     doc.Title,
		UpdatedAt: doc.UpdatedAt.String(),
	}, nil
}

// ListDocuments 获取文档列表
func (s *KnowledgeBaseService) ListDocuments(ctx context.Context, req *dto.KnowledgeDocumentListReq) (*base.Paginated[[]*dto.KnowledgeDocumentResp], error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, req.KBKey, user)
	if err != nil {
		return nil, errors.Wrap(err, "知识库不存在")
	}

	// 构建查询
	limit := req.PageSize
	if limit <= 0 {
		limit = 20
	}
	offset := (req.Page - 1) * limit

	docs, err := s.documentRepo.List(ctx, req.KBKey, user, req.Title, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "查询文档列表失败")
	}

	// 获取总数
	total, err := s.documentRepo.Count(ctx, req.KBKey, user, req.Title)
	if err != nil {
		return nil, errors.Wrap(err, "查询文档总数失败")
	}

	// 转换为响应格式
	var resultList []*dto.KnowledgeDocumentResp
	for _, doc := range docs {
		resultList = append(resultList, &dto.KnowledgeDocumentResp{KnowledgeDocument: doc})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &base.Paginated[[]*dto.KnowledgeDocumentResp]{
		Items:       resultList,
		CurrentPage: req.Page,
		TotalCount:  total,
		TotalPages:  totalPages,
		PageSize:    limit,
	}, nil
}

// SearchKnowledge 搜索知识库
func (s *KnowledgeBaseService) SearchKnowledge(ctx context.Context, req *dto.KnowledgeSearchReq) (*dto.KnowledgeSearchResp, error) {
	user := contextx.GetRequestUserName(ctx)
	if user == "" {
		return nil, errors.New("用户未登录")
	}

	// 检查知识库是否存在
	_, err := s.knowledgeBaseRepo.GetByKBKey(ctx, req.KBKey, user)
	if err != nil {
		return nil, errors.Wrap(err, "知识库不存在")
	}

	// 搜索文档
	docs, err := s.documentRepo.Search(ctx, req.KBKey, user, req.Query, req.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "搜索知识库失败")
	}

	// 转换为响应格式
	var chunks []dto.KnowledgeChunkResp
	for _, doc := range docs {
		chunks = append(chunks, dto.KnowledgeChunkResp{
			ChunkID:    doc.DocID,
			Content:    doc.Content,
			DocTitle:   doc.Title,
			ChunkIndex: 0,
		})
	}

	return &dto.KnowledgeSearchResp{
		Chunks: chunks,
		Total:  len(chunks),
	}, nil
}

// generateKBKey 生成知识库Key
func (s *KnowledgeBaseService) generateKBKey(name string) string {
	// 将中文名称转换为拼音或英文标识
	// 这里简化处理，实际项目中可以使用拼音库
	key := strings.ToLower(name)
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "知识库", "")
	key = strings.ReplaceAll(key, "系统", "_system")
	key = strings.ReplaceAll(key, "管理", "_manage")

	// 如果为空，使用默认值
	if key == "" {
		key = "knowledge_base"
	}

	return key
}
