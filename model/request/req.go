// request/runner.go
package request

// CreateRunnerRequest 创建Runner请求
type CreateRunnerRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50,alphanum"`
	Title       string `json:"title" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=1000"`
	User        string `json:"user" binding:"required,min=2,max=50,alphanum"`
	Language    string `json:"language" binding:"required,oneof=go python java nodejs"`
	Kind        string `json:"kind" binding:"required,oneof=cmd lib so"`
	Visibility  string `json:"visibility" binding:"required,oneof=private public"`
}

// UpdateRunnerRequest 更新Runner请求
type UpdateRunnerRequest struct {
	Title       string `json:"title" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Visibility  string `json:"visibility" binding:"omitempty,oneof=private public"`
}

// CreateAPIRequest 创建API请求
type CreateAPIRequest struct {
	RunnerID       uint   `json:"runner_id" binding:"required"`
	Route          string `json:"route" binding:"required,min=1,max=200"`
	Method         string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
	Name           string `json:"name" binding:"required,min=2,max=100"`
	Description    string `json:"description" binding:"max=1000"`
	Category       string `json:"category" binding:"max=50"`
	RequestSchema  string `json:"request_schema" binding:"omitempty,json"`
	ResponseSchema string `json:"response_schema" binding:"omitempty,json"`
}

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	RunnerID    uint   `json:"runner_id" binding:"required"`
	Version     string `json:"version" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=1000"`
	FilePath    string `json:"file_path" binding:"required,max=500"`
}
