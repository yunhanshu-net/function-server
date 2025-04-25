// function.go
// 定义Function模型，表示目录结构中的函数/可执行单元
package directory

import (
	"time"

	"gorm.io/gorm"

	"github.com/yunhanshu-net/api-server/model/base"
)

// Function 函数/可执行单元模型
// 代表一个可执行的代码单元
// 是TreeNode中RefID引用的实际内容之一
type Function struct {
	base.BaseModel
	// 基本信息
	Name        string       `json:"name" gorm:"size:100;not null;index"`   // 函数名称(英文标识)
	DisplayName string       `json:"display_name" gorm:"size:100"`          // 显示名称(可中文)
	Description string       `json:"description" gorm:"size:500"`           // 描述信息
	User        string       `json:"user" gorm:"index;not null"`            // 所属用户
	Type        FunctionType `json:"type" gorm:"size:20;default:'regular'"` // 函数类型

	// 分类与标签
	Category string `json:"category" gorm:"size:50"` // 分类
	Tags     string `json:"tags" gorm:"size:200"`    // 标签(逗号分隔)

	// 所属包信息
	PackageID uint `json:"package_id" gorm:"index"` // 所属包ID

	// 函数详细信息
	InputSchema  string `json:"input_schema" gorm:"type:text"`        // 输入参数Schema(JSON格式)
	OutputSchema string `json:"output_schema" gorm:"type:text"`       // 输出参数Schema(JSON格式)
	Code         string `json:"code" gorm:"type:text"`                // 函数代码
	Language     string `json:"language" gorm:"size:20;default:'go'"` // 实现语言

	// API相关
	Route  string `json:"route" gorm:"size:200"`               // API路由
	Method string `json:"method" gorm:"size:10;default:'GET'"` // HTTP方法

	// 权限与状态
	Visibility string `json:"visibility" gorm:"size:20;default:'private'"` // 可见性(private/public)
	Status     string `json:"status" gorm:"size:20;default:'active'"`      // 状态

	// 版本管理
	CurrentVersion string `json:"current_version" gorm:"size:20;default:'v1'"` // 当前版本

	// 统计信息
	ExecutionCount int64     `json:"execution_count" gorm:"default:0"`  // 执行次数
	AvgExecTimeMs  int       `json:"avg_exec_time_ms" gorm:"default:0"` // 平均执行时间(毫秒)
	ErrorCount     int64     `json:"error_count" gorm:"default:0"`      // 错误次数
	LastExecTime   time.Time `json:"last_exec_time"`                    // 最后执行时间
	ForkCount      int       `json:"fork_count" gorm:"default:0"`       // 被fork次数

	// Fork信息
	SourceID   uint   `json:"source_id" gorm:"default:0"` // 源函数ID(如果是fork的)
	SourceUser string `json:"source_user" gorm:"size:50"` // 源用户

	// 元数据
	Metadata string `json:"metadata" gorm:"type:text"` // 元数据(JSON格式)
}

// TableName 指定数据库表名
func (f *Function) TableName() string {
	return "directory_function"
}

// BeforeCreate 创建前钩子
func (f *Function) BeforeCreate(tx *gorm.DB) error {
	// 确保有默认显示名称
	if f.DisplayName == "" {
		f.DisplayName = f.Name
	}

	return nil
}

// IsFork 判断是否为fork
func (f *Function) IsFork() bool {
	return f.SourceUser != "" && f.SourceID > 0
}

// GetNodeType 获取节点类型
func (f *Function) GetNodeType() NodeType {
	return NodeTypeFunction
}

// GetFullAPIPath 获取完整API路径
func (f *Function) GetFullAPIPath() string {
	if f.Type == FunctionTypeAPI && f.Route != "" {
		prefix := "/api"
		if f.Route[0] != '/' {
			return prefix + "/" + f.Route
		}
		return prefix + f.Route
	}
	return ""
}

// FunctionCall 函数调用记录
// 记录每次函数调用的详细信息，用于统计和审计
type FunctionCall struct {
	ID        uint      `gorm:"primarykey" json:"id"` // 主键ID
	CreatedAt time.Time `json:"created_at"`           // 创建时间(调用时间)

	// 函数信息
	FunctionID   uint   `json:"function_id" gorm:"index;not null"` // 函数ID
	FunctionName string `json:"function_name" gorm:"size:100"`     // 函数名称
	User         string `json:"user" gorm:"size:50;index"`         // 所属用户
	Version      string `json:"version" gorm:"size:20"`            // 版本

	// 调用信息
	CallerUser string `json:"caller_user" gorm:"size:50;index"` // 调用者用户
	CallerIP   string `json:"caller_ip" gorm:"size:50"`         // 调用者IP
	TraceID    string `json:"trace_id" gorm:"size:100;index"`   // 追踪ID(请求链路)

	// 执行信息
	ExecutionTimeMs int    `json:"execution_time_ms"`             // 执行时间(毫秒)
	Status          string `json:"status" gorm:"size:20"`         // 执行状态
	ErrorMessage    string `json:"error_message" gorm:"size:500"` // 错误信息

	// 参数与结果
	InputParams  string `json:"input_params" gorm:"type:text"`  // 输入参数(JSON)
	OutputResult string `json:"output_result" gorm:"type:text"` // 输出结果(JSON)

	// 额外信息
	RuncherID     string `json:"runcher_id" gorm:"size:50"`       // 处理的Runcher节点
	ResourceUsage string `json:"resource_usage" gorm:"type:text"` // 资源使用情况(JSON)
}

// TableName 指定数据库表名
func (f *FunctionCall) TableName() string {
	return "function_call_logs"
}
