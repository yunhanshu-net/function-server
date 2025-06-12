package dto

import (
	"time"

	"github.com/yunhanshu-net/pkg/query"

	"github.com/yunhanshu-net/function-server/model"
)

// ===========================================================================
// 创建RunnerFunc
// ===========================================================================

// CreateRunnerFuncReq 创建函数请求
type CreateRunnerFuncReq struct {
	BaseRequest
	Name     string `json:"name" binding:"required"`    // 函数名称
	Title    string `json:"title" binding:"required"`   // 函数标题
	RunnerID int64  `json:"runner_id"`                  // 所属Runner ID
	TreeID   int64  `json:"tree_id" binding:"required"` // 所属目录 ID
	Desc     string `json:"desc"`                       // 函数描述
	Type     int    `json:"type"`                       // 函数类型
	Status   int    `json:"status"`                     // 状态
	IsPublic bool   `json:"is_public"`                  // 是否公开
	Content  string `json:"content"`                    // 函数内容
	Config   string `json:"config"`                     // 函数配置

	Code string `json:"code"`
}

// ToModel 转换为模型
func (req *CreateRunnerFuncReq) ToModel() *model.RunnerFunc {
	return &model.RunnerFunc{
		Name:        req.Name,
		Title:       req.Title,
		RunnerID:    req.RunnerID,
		TreeID:      req.TreeID,
		Description: req.Desc,
		IsPublic:    req.IsPublic,
	}
}

// CreateRunnerFuncResp 创建函数响应
type CreateRunnerFuncResp struct {
	ID        int64     `json:"id"`         // 函数 ID
	Name      string    `json:"name"`       // 函数名称
	Title     string    `json:"title"`      // 函数标题
	RunnerID  int64     `json:"runner_id"`  // 所属Runner ID
	TreeID    int64     `json:"tree_id"`    // 所属目录 ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *CreateRunnerFuncResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.RunnerID = runnerFunc.RunnerID
	resp.TreeID = runnerFunc.TreeID
	resp.CreatedAt = time.Time(runnerFunc.CreatedAt)
}

// ===========================================================================
// 获取RunnerFunc详情
// ===========================================================================

type GetFuncRecord struct {
	//base.PageInfoReq // 嵌入分页信息
	query.PageInfoReq
}

// GetRunnerFuncReq 获取函数详情请求
type GetRunnerFuncReq struct {
	BaseRequest
	ID int64 `json:"-"` // 函数 ID，从路径参数获取
}

type GetRunnerFuncByFullPath struct {
	FullPath string `json:"full_path" form:"full_path" uri:"full_path"`
	Method   string `json:"method" form:"method" uri:"method"`
}

// GetRunnerFuncResp 获取函数详情响应
type GetRunnerFuncResp struct {
	ID         int64     `json:"id"`           // 函数 ID
	Name       string    `json:"name"`         // 函数名称
	Title      string    `json:"title"`        // 函数标题
	RunnerID   int64     `json:"runner_id"`    // 所属Runner ID
	TreeID     int64     `json:"tree_id"`      // 所属目录 ID
	Desc       string    `json:"desc"`         // 函数描述
	Type       int       `json:"type"`         // 函数类型
	Status     int       `json:"status"`       // 状态
	IsPublic   bool      `json:"is_public"`    // 是否公开
	Content    string    `json:"content"`      // 函数内容
	Config     string    `json:"config"`       // 函数配置
	User       string    `json:"user"`         // 所属用户
	ForkFromID *int64    `json:"fork_from_id"` // Fork来源ID
	CreatedBy  string    `json:"created_by"`   // 创建者
	CreatedAt  time.Time `json:"created_at"`   // 创建时间
	UpdatedBy  string    `json:"updated_by"`   // 更新者
	UpdatedAt  time.Time `json:"updated_at"`   // 更新时间
}

// FromModel 从模型转换
func (resp *GetRunnerFuncResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.RunnerID = runnerFunc.RunnerID
	resp.TreeID = runnerFunc.TreeID
	resp.Desc = runnerFunc.Description
	resp.IsPublic = runnerFunc.IsPublic
	resp.User = runnerFunc.User
	resp.ForkFromID = runnerFunc.ForkFromID
	resp.CreatedBy = runnerFunc.CreatedBy
	resp.CreatedAt = time.Time(runnerFunc.CreatedAt)
	resp.UpdatedBy = runnerFunc.UpdatedBy
	resp.UpdatedAt = time.Time(runnerFunc.UpdatedAt)
}

// ===========================================================================
// 更新RunnerFunc
// ===========================================================================

// UpdateRunnerFuncReq 更新函数请求
type UpdateRunnerFuncReq struct {
	BaseRequest
	ID       int64  `json:"-"`                        // 函数 ID，从路径参数获取
	Name     string `json:"name"`                     // 函数名称
	Title    string `json:"title" binding:"required"` // 函数标题
	TreeID   int64  `json:"tree_id"`                  // 所属目录 ID
	Desc     string `json:"desc"`                     // 函数描述
	Type     int    `json:"type"`                     // 函数类型
	Status   int    `json:"status"`                   // 状态
	IsPublic bool   `json:"is_public"`                // 是否公开
	Content  string `json:"content"`                  // 函数内容
	Config   string `json:"config"`                   // 函数配置
}

// ToModel 转换为模型
func (req *UpdateRunnerFuncReq) ToModel() *model.RunnerFunc {
	return &model.RunnerFunc{
		Name:        req.Name,
		Title:       req.Title,
		TreeID:      req.TreeID,
		Description: req.Desc,
		IsPublic:    req.IsPublic,
	}
}

// UpdateRunnerFuncResp 更新函数响应
type UpdateRunnerFuncResp struct {
	ID        int64     `json:"id"`         // 函数 ID
	Name      string    `json:"name"`       // 函数名称
	Title     string    `json:"title"`      // 函数标题
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// FromModel 从模型转换
func (resp *UpdateRunnerFuncResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.UpdatedAt = time.Time(runnerFunc.UpdatedAt)
}

// ===========================================================================
// 删除RunnerFunc
// ===========================================================================

// DeleteRunnerFuncReq 删除函数请求
type DeleteRunnerFuncReq struct {
	BaseRequest
	ID int64 `json:"-"` // 函数 ID，从路径参数获取
}

type DeleteRunnerFuncByIds struct {
	Ids []int64 `json:"ids"`
}

// DeleteRunnerFuncResp 删除函数响应
type DeleteRunnerFuncResp struct {
	Success bool `json:"success"` // 是否成功
}

// ===========================================================================
// RunnerFunc列表
// ===========================================================================

// ListRunnerFuncReq 获取函数列表请求
type ListRunnerFuncReq struct {
	//BasePaginatedRequest
	query.PageInfoReq
	User     string `json:"user" form:"user"`           // 用户名过滤
	RunnerID int64  `json:"runner_id" form:"runner_id"` // 所属Runner ID过滤
	TreeID   int64  `json:"tree_id" form:"tree_id"`     // 所属目录 ID过滤
	IsPublic *bool  `json:"is_public" form:"is_public"` // 是否公开过滤
}

// ListRunnerFuncResp 获取函数列表响应（单个项）
type ListRunnerFuncResp struct {
	ID        int64     `json:"id"`         // 函数 ID
	Name      string    `json:"name"`       // 函数名称
	Title     string    `json:"title"`      // 函数标题
	RunnerID  int64     `json:"runner_id"`  // 所属Runner ID
	TreeID    int64     `json:"tree_id"`    // 所属目录 ID
	Type      int       `json:"type"`       // 函数类型
	Status    int       `json:"status"`     // 状态
	IsPublic  bool      `json:"is_public"`  // 是否公开
	User      string    `json:"user"`       // 所属用户
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *ListRunnerFuncResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.RunnerID = runnerFunc.RunnerID
	resp.TreeID = runnerFunc.TreeID
	resp.IsPublic = runnerFunc.IsPublic
	resp.User = runnerFunc.User
	resp.CreatedAt = time.Time(runnerFunc.CreatedAt)
}

// ToConditions 转换为查询条件
func (req *ListRunnerFuncReq) ToConditions() map[string]interface{} {
	conditions := make(map[string]interface{})

	if req.User != "" {
		conditions["user"] = req.User
	}

	if req.RunnerID > 0 {
		conditions["runner_id"] = req.RunnerID
	}

	if req.TreeID > 0 {
		conditions["tree_id"] = req.TreeID
	}

	if req.IsPublic != nil {
		conditions["is_public"] = *req.IsPublic
	}

	return conditions
}

// ===========================================================================
// Fork RunnerFunc
// ===========================================================================

// ForkRunnerFuncReq 复制函数请求
type ForkRunnerFuncReq struct {
	BaseRequest
	ID             int64  `json:"-"`                                   // 源函数 ID，从路径参数获取
	TargetTreeID   int64  `json:"target_tree_id" binding:"required"`   // 目标目录 ID
	TargetRunnerID int64  `json:"target_runner_id" binding:"required"` // 目标Runner ID
	NewName        string `json:"new_name"`                            // 新函数名称
}

// ForkRunnerFuncResp 复制函数响应
type ForkRunnerFuncResp struct {
	ID        int64     `json:"id"`         // 新函数 ID
	Name      string    `json:"name"`       // 新函数名称
	Title     string    `json:"title"`      // 新函数标题
	RunnerID  int64     `json:"runner_id"`  // 所属Runner ID
	TreeID    int64     `json:"tree_id"`    // 所属目录 ID
	ForkFrom  int64     `json:"fork_from"`  // Fork来源ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *ForkRunnerFuncResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.RunnerID = runnerFunc.RunnerID
	resp.TreeID = runnerFunc.TreeID
	if runnerFunc.ForkFromID != nil {
		resp.ForkFrom = *runnerFunc.ForkFromID
	}
	resp.CreatedAt = time.Time(runnerFunc.CreatedAt)
}

// ===========================================================================
// 获取Runner下的所有函数
// ===========================================================================

// GetByRunnerReq 获取Runner下的函数请求
type GetByRunnerReq struct {
	BaseRequest
	RunnerID int64 `json:"-"` // Runner ID，从路径参数获取
}

// GetByRunnerResp 获取Runner下的函数响应（单个项）
type GetByRunnerResp struct {
	ID        int64     `json:"id"`         // 函数 ID
	Name      string    `json:"name"`       // 函数名称
	Title     string    `json:"title"`      // 函数标题
	TreeID    int64     `json:"tree_id"`    // 所属目录 ID
	Type      int       `json:"type"`       // 函数类型
	Status    int       `json:"status"`     // 状态
	IsPublic  bool      `json:"is_public"`  // 是否公开
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *GetByRunnerResp) FromModel(runnerFunc *model.RunnerFunc) {
	resp.ID = runnerFunc.ID
	resp.Name = runnerFunc.Name
	resp.Title = runnerFunc.Title
	resp.TreeID = runnerFunc.TreeID
	resp.IsPublic = runnerFunc.IsPublic
	resp.CreatedAt = time.Time(runnerFunc.CreatedAt)
}

// ===========================================================================
// 获取函数版本历史
// ===========================================================================

// GetVersionHistoryReq 获取函数版本历史请求
type GetVersionHistoryReq struct {
	BaseRequest
	ID int64 `json:"-"` // 函数 ID，从路径参数获取
}

// GetVersionHistoryResp 获取函数版本历史响应（单个项）
type GetVersionHistoryResp struct {
	ID        int64     `json:"id"`         // 版本 ID
	FuncID    int64     `json:"func_id"`    // 函数 ID
	Version   string    `json:"version"`    // 版本号
	Comment   string    `json:"comment"`    // 版本注释
	CreatedBy string    `json:"created_by"` // 创建者
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *GetVersionHistoryResp) FromModel(version *model.FuncVersion) {
	resp.ID = version.ID
	resp.FuncID = version.FuncID
	resp.Version = version.Version
	resp.Comment = version.Comment
	resp.CreatedBy = version.CreatedBy
	resp.CreatedAt = time.Time(version.CreatedAt)
}

// ===========================================================================
// 保存函数版本
// ===========================================================================

// SaveVersionReq 保存函数版本请求
type SaveVersionReq struct {
	BaseRequest
	ID      int64  `json:"-"`                          // 函数 ID，从路径参数获取
	Version string `json:"version" binding:"required"` // 版本号
	Comment string `json:"comment"`                    // 版本注释
}

// SaveVersionResp 保存函数版本响应
type SaveVersionResp struct {
	ID        int64     `json:"id"`         // 版本 ID
	FuncID    int64     `json:"func_id"`    // 函数 ID
	Version   string    `json:"version"`    // 版本号
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// FromModel 从模型转换
func (resp *SaveVersionResp) FromModel(version *model.FuncVersion) {
	resp.ID = version.ID
	resp.FuncID = version.FuncID
	resp.Version = version.Version
	resp.CreatedAt = time.Time(version.CreatedAt)
}

// ===========================================================================
// 更新函数状态
// ===========================================================================

// UpdateStatusReq 更新函数状态请求
type UpdateStatusReq struct {
	BaseRequest
	ID     int64 `json:"-"`                         // 函数 ID，从路径参数获取
	Status int   `json:"status" binding:"required"` // 状态
}

// UpdateStatusResp 更新函数状态响应
type UpdateStatusResp struct {
	ID      int64 `json:"id"`      // 函数 ID
	Status  int   `json:"status"`  // 状态
	Success bool  `json:"success"` // 是否成功
}

// ===========================================================================
// 获取用户最近执行函数记录（去重）
// ===========================================================================

// GetUserRecentFuncRecordsReq 获取用户最近执行函数记录请求
type GetUserRecentFuncRecordsReq struct {
	BaseRequest
	query.PageInfoReq
	User string `json:"user" form:"user"` // 用户名，从中间件获取
}

// GetUserRecentFuncRecordsResp 获取用户最近执行函数记录响应（单个项）
type GetUserRecentFuncRecordsResp struct {
	FuncID       int64     `json:"func_id"`        // 函数ID
	FuncName     string    `json:"func_name"`      // 函数名称
	FuncTitle    string    `json:"func_title"`     // 函数标题
	RunnerID     int64     `json:"runner_id"`      // Runner ID
	RunnerName   string    `json:"runner_name"`    // Runner名称
	RunnerTitle  string    `json:"runner_title"`   // Runner标题
	TreeID       int64     `json:"tree_id"`        // 服务树ID
	FullNamePath string    `json:"full_name_path"` // 完整路径
	LastRunTime  time.Time `json:"last_run_time"`  // 最后执行时间
	Status       string    `json:"status"`         // 最后执行状态
	RunCount     int64     `json:"run_count"`      // 执行次数
}

// FromFuncRunRecord 从函数运行记录转换
func (resp *GetUserRecentFuncRecordsResp) FromFuncRunRecord(record *model.FuncRunRecord, runnerFunc *model.RunnerFunc, runner *model.Runner, serviceTree *model.ServiceTree) {
	resp.FuncID = record.FuncId
	resp.LastRunTime = time.Unix(record.EndTs/1000, 0) // 假设EndTs是毫秒时间戳
	resp.Status = record.Status

	if runnerFunc != nil {
		resp.FuncName = runnerFunc.Name
		resp.FuncTitle = runnerFunc.Title
		resp.RunnerID = runnerFunc.RunnerID
		resp.TreeID = runnerFunc.TreeID
	}

	if runner != nil {
		resp.RunnerName = runner.Name
		resp.RunnerTitle = runner.Title
	}

	if serviceTree != nil {
		resp.FullNamePath = serviceTree.FullNamePath
	}
}
