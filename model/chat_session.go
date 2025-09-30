package model

// ChatSession 对话会话表
type ChatSession struct {
	Base
	SessionID     string `json:"session_id" gorm:"column:session_id;type:varchar(255);uniqueIndex;comment:会话ID"`
	Title         string `json:"title" gorm:"column:title;comment:会话标题"`
	Model         string `json:"model" gorm:"column:model;comment:使用的模型"`
	Router        string `json:"router" gorm:"column:router;comment:路由标识"`
	KnowledgeKey  string `json:"knowledge_key" gorm:"column:knowledge_key;comment:知识库标识"`
	LastMessageAt *Time  `json:"last_message_at" gorm:"column:last_message_at;comment:最后消息时间"`
	MessageCount  int    `json:"message_count" gorm:"column:message_count;comment:消息数量"`
	User          string `json:"user" gorm:"column:user;comment:用户标识"`
}
