// runner.go
// 定义Runner模型，代表一个运行器实例
package runner

import (
	"time"

	"github.com/yunhanshu-net/api-server/model/base"
	"gorm.io/gorm"
)

// Status Runner状态枚举
type Status string

const (
	StatusActive    Status = "active"    // 活跃状态
	StatusInactive  Status = "inactive"  // 不活跃
	StatusError     Status = "error"     // 错误状态
	StatusMigrating Status = "migrating" // 迁移中
	StatusUpdating  Status = "updating"  // 更新中
)

// Runner Runner模型，存储Runner的元数据信息
type Runner struct {
	base.BaseModel

	// 基本信息
	Name        string `json:"name" gorm:"index;not null;uniqueIndex:idx_user_name"` // Runner名称(英文标识)
	Title       string `json:"title" gorm:"size:100"`                                // 显示名称
	Description string `json:"description" gorm:"size:1000"`                         // 描述信息
	User        string `json:"user" gorm:"index;not null;uniqueIndex:idx_user_name"` // 所属用户/租户

	// 技术信息
	Language       string `json:"language" gorm:"size:20;default:'go'"` // 使用语言
	Kind           string `json:"kind" gorm:"size:20;default:'cmd'"`    // 类型(cmd/lib/so等)
	CurrentVersion string `json:"current_version" gorm:"size:20"`       // 当前版本

	// 权限和可见性
	Visibility string `json:"visibility" gorm:"size:20;default:'private'"` // 可见性(private/public)
	AccessType string `json:"access_type" gorm:"size:20;default:'user'"`   // 访问类型(user/team/all)

	// 状态信息
	Status        string    `json:"status" gorm:"size:20;default:'inactive'"` // 状态
	StatusMessage string    `json:"status_message" gorm:"size:200"`           // 状态信息
	LastActive    time.Time `json:"last_active"`                              // 最后活跃时间

	// 统计信息
	APICount      int   `json:"api_count" gorm:"default:0"`      // API数量
	VersionCount  int   `json:"version_count" gorm:"default:0"`  // 版本数量
	TotalRequests int64 `json:"total_requests" gorm:"default:0"` // 总请求数

	// 审计信息
	base.AuditInfo

	// 关联 - 这里只定义关系，不嵌入具体数据
	// APIs将通过RunnerID外键关联到Runner
	// Versions将通过RunnerID外键关联到Runner

	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Runner) TableName() string {
	return "runners"
}

// RunnerSummary Runner简要信息
// 用于API返回简要信息，避免返回过多数据
type RunnerSummary struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Title          string    `json:"title"`
	User           string    `json:"user"`
	CurrentVersion string    `json:"current_version"`
	Status         string    `json:"status"`
	Language       string    `json:"language"`
	LastActive     time.Time `json:"last_active"`
	APICount       int       `json:"api_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// GetRequestSubject 获取NATS请求主题
func (r *Runner) GetRequestSubject() string {
	return "runner." + r.User + "." + r.Name + "." + r.CurrentVersion + ".run"
}

// IsActive 判断Runner是否活跃
// 如果状态为活跃且最后活跃时间在5分钟内，则认为是活跃的
func (r *Runner) IsActive() bool {
	return r.Status == string(StatusActive) &&
		time.Since(r.LastActive) < 5*time.Minute
}

// NewRunner 创建Runner对象
func NewRunner(r Runner) *Runner {
	return &r
}

// CodeAPI API定义
type CodeAPI struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
}

// AddApiResp API响应结构
type AddApiResp struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// AddApi 添加API方法
// 这是一个示例实现，实际应该连接到数据库并创建API记录
func (r *Runner) AddApi(codeApi interface{}) (*AddApiResp, error) {
	// 转换参数为CodeAPI类型
	// 在实际实现中，此处可能需要更复杂的类型转换
	api, ok := codeApi.(CodeAPI)
	if !ok {
		// 假设传入的是map
		apiMap, ok := codeApi.(map[string]interface{})
		if !ok {
			return &AddApiResp{
				ID:      0,
				Name:    "unknown",
				Version: "v1",
				Status:  "error",
			}, nil
		}

		// 构建API对象
		api = CodeAPI{
			Name:    apiMap["name"].(string),
			Version: apiMap["version"].(string),
			Path:    apiMap["path"].(string),
			Method:  apiMap["method"].(string),
			Handler: apiMap["handler"].(string),
		}
	}

	// 创建响应
	// 在实际实现中，这里应该将API保存到数据库并返回实际的ID
	return &AddApiResp{
		ID:      1, // 示例ID
		Name:    api.Name,
		Version: api.Version,
		Status:  "created",
	}, nil
}
