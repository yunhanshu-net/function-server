package model

// ChatMessage 对话消息表
type ChatMessage struct {
	Base
	SessionID        string `json:"session_id" gorm:"column:session_id;type:varchar(255);index;comment:会话ID"`
	Role             string `json:"role" gorm:"column:role;comment:角色(user/assistant/system)"`
	Content          string `json:"content" gorm:"column:content;type:text;comment:消息内容"`
	Model            string `json:"model" gorm:"column:model;comment:使用的模型"`
	Router           string `json:"router" gorm:"column:router;comment:路由标识"`
	KnowledgeKey     string `json:"knowledge_key" gorm:"column:knowledge_key;type:varchar(255);index;comment:知识库Key"`
	PromptTokens     int    `json:"prompt_tokens" gorm:"column:prompt_tokens;comment:输入token数"`
	CompletionTokens int    `json:"completion_tokens" gorm:"column:completion_tokens;comment:输出token数"`
	TotalTokens      int    `json:"total_tokens" gorm:"column:total_tokens;comment:总token数"`
	User             string `json:"user" gorm:"column:user;comment:用户标识"`
}
