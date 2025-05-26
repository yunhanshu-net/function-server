package v1

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/dto/base"
	"github.com/yunhanshu-net/pkg/query"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/dto"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/pkg/response"
	"github.com/yunhanshu-net/api-server/pkg/utils"
	"github.com/yunhanshu-net/api-server/service"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

// RunnerFuncAPI RunnerFunc API控制器
type RunnerFuncAPI struct {
	service *service.RunnerFunc
}

// NewRunnerFuncAPI 创建RunnerFunc API控制器
func NewRunnerFuncAPI(db *gorm.DB) *RunnerFuncAPI {
	return &RunnerFuncAPI{
		service: service.NewRunnerFunc(db),
	}
}

// Create 创建函数
func (api *RunnerFuncAPI) Create(c *gin.Context) {
	logger.Debug(c, "开始处理RunnerFunc创建请求")

	// 使用CreateRunnerFuncReq DTO接收请求参数
	var req dto.CreateRunnerFuncReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析RunnerFunc参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 直接创建模型对象
	runnerFunc := &model.RunnerFunc{
		Name:        req.Name,
		Title:       req.Title,
		RunnerID:    req.RunnerID,
		TreeID:      req.TreeID,
		Description: req.Desc,
		IsPublic:    req.IsPublic,
		Code:        req.Code,
		User:        c.GetString("user"),
		// Type, Status, Content, Config字段在模型中不存在，暂时移除
	}

	// 设置创建者信息，实际项目中应从JWT或Session获取
	runnerFunc.CreatedBy = c.GetString("user")
	runnerFunc.UpdatedBy = c.GetString("user")

	now := time.Now()
	// 调用服务层创建函数
	if err := api.service.Create(c, runnerFunc); err != nil {
		logger.Error(c, "创建RunnerFunc失败", err)
		response.ServerError(c, "创建函数失败: "+err.Error())
		return
	}
	logger.Infof(c, "创建函数耗时：cost:%s", time.Since(now))

	// 直接创建响应DTO
	resp := dto.CreateRunnerFuncResp{
		ID:        runnerFunc.ID,
		Name:      runnerFunc.Name,
		Title:     runnerFunc.Title,
		RunnerID:  runnerFunc.RunnerID,
		TreeID:    runnerFunc.TreeID,
		CreatedAt: time.Time(runnerFunc.CreatedAt),
	}

	logger.Info(c, "创建RunnerFunc成功", zap.Int64("id", runnerFunc.ID), zap.String("name", runnerFunc.Name))
	response.Success(c, resp)
}

// List 获取函数列表
func (api *RunnerFuncAPI) List(c *gin.Context) {
	logger.Debug(c, "开始处理RunnerFunc列表请求")

	// 使用ListRunnerFuncReq DTO接收请求参数

	var req dto.ListRunnerFuncReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.ParamError(c, err.Error())
		return
	}

	getDB := db.GetDB().Where("user = ?", c.GetString("user"))
	var runnerFunctions []*model.RunnerFunc
	paginate, err := query.AutoPaginateTable(c, getDB, &model.RunnerFunc{}, &runnerFunctions, &req.PageInfoReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, paginate)
}

// Get 获取函数详情
func (api *RunnerFuncAPI) Get(c *gin.Context) {
	// 使用GetRunnerFuncReq DTO
	var req dto.GetRunnerFuncReq
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始处理RunnerFunc详情请求", zap.Int64("id", id))

	// 调用服务层获取函数详情
	runnerFunc, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取RunnerFunc详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取函数详情失败")
		return
	}

	if runnerFunc == nil {
		logger.Info(c, "RunnerFunc不存在", zap.Int64("id", id))
		response.NotFound(c, "函数不存在")
		return
	}

	// 直接创建响应DTO

	logger.Info(c, "获取RunnerFunc详情成功", zap.Int64("id", id))
	response.Success(c, runnerFunc)
}
func (api *RunnerFuncAPI) GetByTreeId(c *gin.Context) {
	// 使用GetRunnerFuncReq DTO
	var req dto.GetRunnerFuncReq
	id, err := strconv.ParseInt(c.Param("tree_id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始处理RunnerFunc详情请求", zap.Int64("id", id))

	// 调用服务层获取函数详情
	runnerFunc, err := api.service.GetByTreeId(c, id)
	if err != nil {
		logger.Error(c, "获取RunnerFunc详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取函数详情失败")
		return
	}

	if runnerFunc == nil {
		logger.Info(c, "RunnerFunc不存在", zap.Int64("id", id))
		response.NotFound(c, "函数不存在")
		return
	}

	// 直接创建响应DTO

	logger.Info(c, "获取RunnerFunc详情成功", zap.Int64("id", id))
	response.Success(c, runnerFunc)
}

func (api *RunnerFuncAPI) GetByFullPath(c *gin.Context) {
	// 使用GetRunnerFuncReq DTO
	var req dto.GetRunnerFuncByFullPath
	err := c.ShouldBindUri(&req)
	if err != nil {
		logger.Errorf(c, "解析RunnerFunc 失败:%s", err)
		response.ParamError(c, "无效的ID")
		return
	}
	var r model.ServiceTree
	err = db.GetDB().Model(&model.ServiceTree{}).Where("full_name_path = ? AND method = ?", strings.TrimPrefix(req.FullPath, "/"), strings.ToUpper(c.Query("method"))).First(&r).Error
	if err != nil {
		response.ServerError(c, err.Error())
		return

	}
	f := model.RunnerFunc{}
	err = db.GetDB().Model(&model.RunnerFunc{}).Where("id = ?", r.RefID).First(&f).Error
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	//runnerFunc, err := api.service.GetByFullPath(c, c.Query("method"), req.FullPath)
	// 调用服务层获取函数详情

	//if err != nil {
	//	logger.Errorf(c, "函数不存在:%s req:%+v", err, req)
	//	response.NotFound(c, "函数不存在")
	//	return
	//}
	response.Success(c, f)
}
func (api *RunnerFuncAPI) GetFuncRecord(c *gin.Context) {
	// 使用GetRunnerFuncReq DTO
	var req dto.GetFuncRecord
	err := c.ShouldBindQuery(&req)
	if err != nil {
		logger.Errorf(c, "解析RunnerFunc 失败:%s", err)
		response.ParamError(c, "无效的ID")
		return
	}

	f := c.Param("func_id")
	d := db.GetDB().Where("func_id = ?", f)

	var list []model.FuncRunRecord
	paginate, err := query.AutoPaginateTable(c, d, &model.FuncRunRecord{}, &list, &req.PageInfoReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, paginate)
}

// Update 更新函数
func (api *RunnerFuncAPI) Update(c *gin.Context) {
	// 使用UpdateRunnerFuncReq DTO
	var req dto.UpdateRunnerFuncReq

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始处理RunnerFunc更新请求", zap.Int64("id", id))

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析RunnerFunc参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 直接创建模型对象
	updateData := &model.RunnerFunc{
		Name:        req.Name,
		Title:       req.Title,
		TreeID:      req.TreeID,
		Description: req.Desc,
		IsPublic:    req.IsPublic,
		// Type, Status, Content, Config字段在模型中不存在，暂时移除
	}

	// 设置更新者信息，实际项目中应从JWT或Session获取
	updateData.UpdatedBy = "admin"

	// 调用服务层更新函数
	if err := api.service.Update(c, id, updateData); err != nil {
		logger.Error(c, "更新RunnerFunc失败", err, zap.Int64("id", id))
		response.ServerError(c, "更新函数失败: "+err.Error())
		return
	}

	// 重新获取最新数据
	runnerFunc, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取更新后的RunnerFunc详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取更新后的函数详情失败")
		return
	}

	// 直接创建响应DTO
	resp := dto.UpdateRunnerFuncResp{
		ID:        runnerFunc.ID,
		Name:      runnerFunc.Name,
		Title:     runnerFunc.Title,
		UpdatedAt: time.Time(runnerFunc.UpdatedAt),
	}

	logger.Info(c, "更新RunnerFunc成功", zap.Int64("id", id))
	response.Success(c, resp)
}

// Delete 删除函数
func (api *RunnerFuncAPI) Delete(c *gin.Context) {
	// 使用DeleteRunnerFuncReq DTO
	var req dto.DeleteRunnerFuncReq

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始处理RunnerFunc删除请求", zap.Int64("id", id))

	// 设置删除者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	// 调用服务层删除函数
	if err := api.service.Delete(c, id, operator); err != nil {
		logger.Error(c, "删除RunnerFunc失败", err, zap.Int64("id", id))
		response.ServerError(c, "删除函数失败: "+err.Error())
		return
	}

	// 创建响应DTO
	resp := dto.DeleteRunnerFuncResp{
		Success: true,
	}

	logger.Info(c, "删除RunnerFunc成功", zap.Int64("id", id))
	response.Success(c, resp)
}

// Fork 复制函数
func (api *RunnerFuncAPI) Fork(c *gin.Context) {
	// 使用ForkRunnerFuncReq DTO
	var req dto.ForkRunnerFuncReq
	req.TraceID = utils.GetTraceIDFromContext(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始处理RunnerFunc复制请求", zap.Int64("id", id))

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析Fork参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 设置操作者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	// 调用服务层复制函数
	newFunc, err := api.service.Fork(c, id, req.TargetTreeID, req.TargetRunnerID, req.NewName, operator)
	if err != nil {
		logger.Error(c, "Fork RunnerFunc失败", err, zap.Int64("id", id))
		response.ServerError(c, "Fork函数失败: "+err.Error())
		return
	}

	// 直接创建响应DTO
	resp := dto.ForkRunnerFuncResp{
		ID:        newFunc.ID,
		Name:      newFunc.Name,
		Title:     newFunc.Title,
		RunnerID:  newFunc.RunnerID,
		TreeID:    newFunc.TreeID,
		ForkFrom:  *newFunc.ForkFromID,
		CreatedAt: time.Time(newFunc.CreatedAt),
	}

	logger.Info(c, "Fork RunnerFunc成功", zap.Int64("source_id", id), zap.Int64("new_id", newFunc.ID), zap.String("name", newFunc.Name))
	response.Success(c, resp)
}

// GetByRunner 获取Runner下的所有函数
func (api *RunnerFuncAPI) GetByRunner(c *gin.Context) {
	// 使用GetByRunnerReq DTO
	var req dto.GetByRunnerReq
	req.TraceID = utils.GetTraceIDFromContext(c)

	runnerID, err := strconv.ParseInt(c.Param("runner_id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("runner_id_param", c.Param("runner_id")))
		response.ParamError(c, "无效的Runner ID")
		return
	}
	req.RunnerID = runnerID

	logger.Debug(c, "开始获取Runner下的函数", zap.Int64("runner_id", runnerID))

	// 调用服务层获取Runner下的函数列表
	funcs, err := api.service.GetByRunner(c, runnerID)
	if err != nil {
		logger.Error(c, "获取Runner下的函数失败", err, zap.Int64("runner_id", runnerID))
		response.ServerError(c, "获取Runner下的函数失败")
		return
	}

	// 将模型列表转换为DTO列表
	respItems := make([]dto.GetByRunnerResp, 0, len(funcs))
	for _, runnerFunc := range funcs {
		item := dto.GetByRunnerResp{
			ID:     runnerFunc.ID,
			Name:   runnerFunc.Name,
			Title:  runnerFunc.Title,
			TreeID: runnerFunc.TreeID,
			// Type和Status字段在模型中不存在，暂时填0
			Type:      0,
			Status:    0,
			IsPublic:  runnerFunc.IsPublic,
			CreatedAt: time.Time(runnerFunc.CreatedAt),
		}
		respItems = append(respItems, item)
	}

	logger.Info(c, "获取Runner下的函数成功", zap.Int64("runner_id", runnerID), zap.Int("func_count", len(funcs)))
	response.Success(c, respItems)
}

// GetVersionHistory 获取函数版本历史
func (api *RunnerFuncAPI) GetVersionHistory(c *gin.Context) {
	// 使用GetVersionHistoryReq DTO
	var req dto.GetVersionHistoryReq

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	logger.Debug(c, "开始获取RunnerFunc版本历史", zap.Int64("id", id))

	// 调用服务层获取函数版本历史
	versions, err := api.service.GetVersionHistory(c, id)
	if err != nil {
		logger.Error(c, "获取RunnerFunc版本历史失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取函数版本历史失败")
		return
	}

	// 将模型列表转换为DTO列表
	respItems := make([]dto.GetVersionHistoryResp, 0, len(versions))
	for _, version := range versions {
		item := dto.GetVersionHistoryResp{
			ID:        version.ID,
			FuncID:    version.FuncID,
			Version:   version.Version,
			Comment:   version.Comment,
			CreatedBy: version.CreatedBy,
			CreatedAt: time.Time(version.CreatedAt),
		}
		respItems = append(respItems, item)
	}

	logger.Info(c, "获取RunnerFunc版本历史成功", zap.Int64("id", id), zap.Int("count", len(versions)))
	response.Success(c, respItems)
}

// SaveVersion 保存函数版本
func (api *RunnerFuncAPI) SaveVersion(c *gin.Context) {
	// 使用SaveVersionReq DTO
	var req dto.SaveVersionReq

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析版本参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	logger.Debug(c, "开始保存RunnerFunc版本", zap.Int64("id", id), zap.String("version", req.Version))

	// 设置操作者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	err = api.service.SaveVersion(c, id, req.Version, req.Comment, operator)
	if err != nil {
		logger.Error(c, "保存RunnerFunc版本失败", err, zap.Int64("id", id))
		response.ServerError(c, "保存函数版本失败: "+err.Error())
		return
	}

	// 创建响应
	resp := dto.SaveVersionResp{
		ID:        id,
		Version:   req.Version,
		CreatedAt: time.Now(), // 设置当前时间，实际应从保存的版本记录中获取
	}

	logger.Info(c, "保存RunnerFunc版本成功", zap.Int64("id", id), zap.String("version", req.Version))
	response.Success(c, resp)
}

// UpdateStatus 更新函数状态
func (api *RunnerFuncAPI) UpdateStatus(c *gin.Context) {
	// 使用UpdateStatusReq DTO
	var req dto.UpdateStatusReq

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析RunnerFunc ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.ID = id

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析状态参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	logger.Debug(c, "开始更新RunnerFunc状态", zap.Int64("id", id), zap.Int("status", req.Status))

	// 调用服务层更新函数状态
	err = api.service.UpdateStatus(c, id, req.Status)
	if err != nil {
		logger.Error(c, "更新RunnerFunc状态失败", err, zap.Int64("id", id))
		response.ServerError(c, "更新函数状态失败: "+err.Error())
		return
	}

	// 创建响应
	resp := dto.UpdateStatusResp{
		ID:      id,
		Status:  req.Status,
		Success: true,
	}

	logger.Info(c, "更新RunnerFunc状态成功", zap.Int64("id", id), zap.Int("status", req.Status))
	response.Success(c, resp)
}

// GetUserRecentFuncRecords 获取用户最近执行过的函数记录（去重）
// @Summary 获取用户最近执行函数记录
// @Description 获取当前用户最近执行过的函数记录，每个函数只显示最新的一次执行记录，按执行时间倒序排列
// @Tags RunnerFunc
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} base.Paginated[dto.GetUserRecentFuncRecordsResp] "成功"
// @Failure 400 {object} response.ErrorResponse "参数错误"
// @Failure 500 {object} response.ErrorResponse "服务器错误"
// @Router /api/v1/runner-func/recent-records [get]
func (api *RunnerFuncAPI) GetUserRecentFuncRecords(c *gin.Context) {
	logger.Debug(c, "开始处理获取用户最近执行函数记录请求")

	// 解析请求参数
	var req dto.GetUserRecentFuncRecordsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(c, "解析请求参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 从中间件获取用户信息
	user := c.GetString("user")
	if user == "" {
		logger.Error(c, "获取用户信息失败", fmt.Errorf("用户信息为空"))
		response.ParamError(c, "用户信息获取失败")
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	logger.Debug(c, "开始获取用户最近执行函数记录",
		zap.String("user", user),
		zap.Int("page", req.Page),
		zap.Int("pageSize", req.PageSize))

	// 调用服务层获取详细信息
	records, total, err := api.service.GetUserRecentFuncRecordsWithDetails(c, user, req.Page, req.PageSize)
	if err != nil {
		logger.Error(c, "获取用户最近执行函数记录失败", err, zap.String("user", user))
		response.ServerError(c, "获取用户最近执行函数记录失败: "+err.Error())
		return
	}

	// 构建分页响应
	pageResp := base.Paginated[interface{}]{
		Items:       records,
		CurrentPage: req.Page,
		TotalCount:  total,
		TotalPages:  int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
		PageSize:    req.PageSize,
	}

	logger.Info(c, "获取用户最近执行函数记录成功",
		zap.String("user", user),
		zap.Int("count", len(records)),
		zap.Int64("total", total))

	response.Success(c, pageResp)
}
