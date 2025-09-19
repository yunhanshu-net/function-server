package dto

import (
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto/base"
)

// KnowledgeBaseCreateReq 创建知识库请求
type KnowledgeBaseCreateReq struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=500"`
	Router      string `json:"router" validate:"max=100"`
}

// KnowledgeBaseUpdateReq 更新知识库请求
type KnowledgeBaseUpdateReq struct {
	KBKey       string `json:"kb_key" validate:"required"`
	Name        string `json:"name" validate:"max=100"`
	Description string `json:"description" validate:"max=500"`
	Status      string `json:"status" validate:"oneof=active inactive"`
}

// KnowledgeBaseListReq 知识库列表请求
type KnowledgeBaseListReq struct {
	base.PageInfoReq
	Name   string `json:"name" form:"name"`     // 按名称搜索
	Router string `json:"router" form:"router"` // 按路由过滤
	Status string `json:"status" form:"status"` // 按状态过滤
}

// KnowledgeBaseResp 知识库响应
type KnowledgeBaseResp struct {
	*model.KnowledgeBase
}

// KnowledgeDocumentUploadReq 上传文档请求
type KnowledgeDocumentUploadReq struct {
	KBKey    string `json:"kb_key" validate:"required"`
	Title    string `json:"title" validate:"required,max=200"`
	Content  string `json:"content" validate:"required"`
	FileType string `json:"file_type" validate:"required,oneof=pdf txt doc md"`
}

// KnowledgeDocumentListReq 文档列表请求
type KnowledgeDocumentListReq struct {
	base.PageInfoReq
	KBKey string `json:"kb_key" form:"kb_key" validate:"required"`
	Title string `json:"title" form:"title"` // 按标题搜索
}

// KnowledgeDocumentUpdateReq 更新文档请求
type KnowledgeDocumentUpdateReq struct {
	DocID    string `json:"doc_id" validate:"required"`                // 文档ID
	Title    string `json:"title" validate:"max=200"`                  // 文档标题
	Content  string `json:"content"`                                   // 文档内容
	FileType string `json:"file_type" validate:"oneof=pdf txt doc md"` // 文件类型
}

// KnowledgeDocumentUpdateResp 更新文档响应
type KnowledgeDocumentUpdateResp struct {
	DocID     string `json:"doc_id"`     // 文档ID
	Title     string `json:"title"`      // 文档标题
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// KnowledgeDocumentResp 文档响应
type KnowledgeDocumentResp struct {
	*model.KnowledgeDocument
}

// KnowledgeSearchReq 知识库搜索请求
type KnowledgeSearchReq struct {
	KBKey string `json:"kb_key" validate:"required"`
	Query string `json:"query" validate:"required,max=500"`
	Limit int    `json:"limit" validate:"min=1,max=20" data:"default_value:5"`
}

// KnowledgeSearchResp 知识库搜索响应
type KnowledgeSearchResp struct {
	Chunks []KnowledgeChunkResp `json:"chunks"`
	Total  int                  `json:"total"`
}

// KnowledgeChunkResp 知识分块响应
type KnowledgeChunkResp struct {
	ChunkID    string  `json:"chunk_id"`
	Content    string  `json:"content"`
	DocTitle   string  `json:"doc_title"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score,omitempty"` // 相似度分数
}
