package model

// KnowledgeBase 知识库表
type KnowledgeBase struct {
	Base
	KBKey         string `json:"kb_key" gorm:"column:kb_key;type:varchar(255);uniqueIndex;comment:知识库Key"`
	Name          string `json:"name" gorm:"column:name;comment:知识库名称"`
	Description   string `json:"description" gorm:"column:description;comment:知识库描述"`
	Router        string `json:"router" gorm:"column:router;comment:关联路由"`
	Status        string `json:"status" gorm:"column:status;comment:状态(active/inactive)"`
	DocumentCount int    `json:"document_count" gorm:"column:document_count;comment:文档数量"`
	User          string `json:"user" gorm:"column:user;comment:创建用户"`
}

// KnowledgeDocument 知识库文档表
type KnowledgeDocument struct {
	Base
	KBKey    string `json:"kb_key" gorm:"column:kb_key;type:varchar(255);index;comment:知识库Key"`
	DocID    string `json:"doc_id" gorm:"column:doc_id;type:varchar(255);comment:文档ID"`
	Title    string `json:"title" gorm:"column:title;comment:文档标题"`
	Content  string `json:"content" gorm:"column:content;type:longtext;comment:文档内容"`
	FileType string `json:"file_type" gorm:"column:file_type;comment:文件类型(pdf/txt/doc)"`
	FileSize int64  `json:"file_size" gorm:"column:file_size;comment:文件大小(字节)"`
	Status   string `json:"status" gorm:"column:status;comment:状态(processing/completed/failed)"`
	User     string `json:"user" gorm:"column:user;comment:上传用户"`
}

// KnowledgeChunk 知识库文档分块表
type KnowledgeChunk struct {
	Base
	KBKey      string `json:"kb_key" gorm:"column:kb_key;type:varchar(255);index;comment:知识库Key"`
	DocID      string `json:"doc_id" gorm:"column:doc_id;type:varchar(255);index;comment:文档ID"`
	ChunkID    string `json:"chunk_id" gorm:"column:chunk_id;type:varchar(255);comment:分块ID"`
	Content    string `json:"content" gorm:"column:content;type:longtext;comment:分块内容"`
	ChunkIndex int    `json:"chunk_index" gorm:"column:chunk_index;comment:分块序号"`
	User       string `json:"user" gorm:"column:user;comment:用户标识"`
}
