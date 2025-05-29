package v1

import (
	"fmt"
	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/dto/base"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/dto"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/pkg/response"
	"github.com/yunhanshu-net/api-server/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RunnerAPI Runner API控制器
type RunnerAPI struct {
	service *service.Runner
}

// NewRunnerAPI 创建Runner API控制器
func NewRunnerAPI(db *gorm.DB) *RunnerAPI {
	return &RunnerAPI{
		service: service.NewRunner(db),
	}
}

// Create 创建Runner
func (api *RunnerAPI) Create(c *gin.Context) {
	// 使用gin.Context作为context
	// 不再需要 ctx := utils.FromGinContext(c)，直接使用c作为context

	logger.Debug(c, "开始处理Runner创建请求")

	var runner model.Runner
	if err := c.ShouldBindJSON(&runner); err != nil {
		logger.Error(c, "解析Runner参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 设置创建者信息，实际项目中应从JWT或Session获取
	runner.CreatedBy = c.GetString("user")
	runner.UpdatedBy = c.GetString("user")
	runner.User = c.GetString("user")
	if err := api.service.Create(c, &runner); err != nil {
		logger.Error(c, "创建Runner失败", err)
		response.ServerError(c, "创建Runner失败: "+err.Error())
		return
	}

	logger.Info(c, "创建Runner成功", zap.Int64("id", runner.ID), zap.String("title", runner.Title))
	response.Success(c, runner)
}

// List 获取Runner列表
func (api *RunnerAPI) List(c *gin.Context) {
	// 使用gin.Context作为context

	logger.Debug(c, "开始处理Runner列表请求")

	var req base.PageInfoReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.ParamError(c, err.Error())
		return
	}
	var list []model.Runner

	db := db.GetDB()
	db.Where("user = ?", c.GetString("user"))

	paginate, err := base.AutoPaginate(c, db, &model.Runner{}, &list, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, paginate)

}

// Get 获取Runner详情
func (api *RunnerAPI) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理Runner详情请求", zap.Int64("id", id))

	runner, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取Runner详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取Runner详情失败")
		return
	}

	if runner == nil {
		logger.Info(c, "Runner不存在", zap.Int64("id", id))
		response.NotFound(c, "Runner不存在")
		return
	}

	logger.Info(c, "获取Runner详情成功", zap.Int64("id", id))
	response.Success(c, runner)
}
func (api *RunnerAPI) GetByUserName(c *gin.Context) {
	user := c.Param("user")
	name := c.Param("name")
	if user == "" || name == "" {
		logger.Error(c, "user和name不能为空", fmt.Errorf("user和name不能为空"))
		response.ParamError(c, "无效的ID")
		return
	}

	runner, err := api.service.GetByUserName(c, user, name)
	if err != nil {
		response.ServerError(c, "获取Runner详情失败")
		return
	}

	if runner == nil {
		response.NotFound(c, "Runner不存在")
		return
	}
	response.Success(c, runner)
}

// Update 更新Runner
func (api *RunnerAPI) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理Runner更新请求", zap.Int64("id", id))

	var updateData model.Runner
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Error(c, "解析Runner参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	// 设置更新者信息，实际项目中应从JWT或Session获取
	updateData.UpdatedBy = "admin"

	// 更新Runner
	if err := api.service.Update(c, id, &updateData); err != nil {
		logger.Error(c, "更新Runner失败", err, zap.Int64("id", id))
		response.ServerError(c, "更新Runner失败: "+err.Error())
		return
	}

	// 重新获取最新数据
	runner, err := api.service.Get(c, id)
	if err != nil {
		logger.Error(c, "获取更新后的Runner详情失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取更新后的Runner详情失败")
		return
	}

	logger.Info(c, "更新Runner成功", zap.Int64("id", id))
	response.Success(c, runner)
}

// Delete 删除Runner
func (api *RunnerAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理Runner删除请求", zap.Int64("id", id))

	// 设置删除者信息，实际项目中应从JWT或Session获取
	operator := c.GetString("user")

	if err := api.service.Delete(c, id, operator); err != nil {
		logger.Error(c, "删除Runner失败", err, zap.Int64("id", id))
		response.ServerError(c, "删除Runner失败: "+err.Error())
		return
	}

	logger.Info(c, "删除Runner成功", zap.Int64("id", id))
	response.Success(c, nil)
}

// Fork 复制Runner
func (api *RunnerAPI) Fork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理Runner复制请求", zap.Int64("id", id))

	// 设置操作者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	newRunner, err := api.service.Fork(c, id, operator)
	if err != nil {
		logger.Error(c, "Fork Runner失败", err, zap.Int64("id", id))
		response.ServerError(c, "Fork Runner失败: "+err.Error())
		return
	}

	logger.Info(c, "Fork Runner成功",
		zap.Int64("source_id", id),
		zap.Int64("new_id", newRunner.ID),
		zap.String("title", newRunner.Title))
	response.Success(c, newRunner)
}

// Version 查看Runner版本历史
func (api *RunnerAPI) Version(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}

	logger.Debug(c, "开始处理Runner版本请求", zap.Int64("id", id))

	versions, err := api.service.GetVersionHistory(c, id)
	if err != nil {
		logger.Error(c, "获取Runner版本历史失败", err, zap.Int64("id", id))
		response.ServerError(c, "获取Runner版本历史失败")
		return
	}

	logger.Info(c, "获取Runner版本历史成功", zap.Int64("id", id), zap.Int("version_count", len(versions)))
	response.Success(c, versions)
}

// GetVersionHistory 获取Runner版本历史
func (api *RunnerAPI) GetVersionHistory(c *gin.Context) {
	// 使用GetRunnerVersionHistoryReq DTO
	var req dto.GetRunnerVersionHistoryReq

	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.RunnerID = runnerID

	logger.Debug(c, "开始获取Runner版本历史", zap.Int64("id", runnerID))

	// 调用服务层获取Runner版本历史
	versions, err := api.service.GetVersionHistory(c, runnerID)
	if err != nil {
		logger.Error(c, "获取Runner版本历史失败", err, zap.Int64("id", runnerID))
		response.ServerError(c, "获取Runner版本历史失败")
		return
	}

	// 将模型列表转换为DTO列表
	respItems := make([]dto.GetRunnerVersionHistoryResp, 0, len(versions))
	for _, version := range versions {
		item := dto.GetRunnerVersionHistoryResp{
			ID:        version.ID,
			RunnerID:  version.RunnerID,
			Version:   version.Version,
			Comment:   version.Comment,
			CreatedBy: version.CreatedBy,
			CreatedAt: time.Time(version.CreatedAt),
		}
		respItems = append(respItems, item)
	}

	logger.Info(c, "获取Runner版本历史成功", zap.Int64("id", runnerID), zap.Int("count", len(versions)))
	response.Success(c, respItems)
}

// SaveVersion 保存Runner版本
func (api *RunnerAPI) SaveVersion(c *gin.Context) {
	// 使用SaveRunnerVersionReq DTO
	var req dto.SaveRunnerVersionReq

	runnerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error(c, "解析Runner ID失败", err, zap.String("id_param", c.Param("id")))
		response.ParamError(c, "无效的ID")
		return
	}
	req.RunnerID = runnerID

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "解析版本参数失败", err)
		response.ParamError(c, "参数解析失败: "+err.Error())
		return
	}

	logger.Debug(c, "开始保存Runner版本", zap.Int64("id", runnerID), zap.String("version", req.Version))

	// 设置操作者信息，实际项目中应从JWT或Session获取
	operator := "admin"

	// 调用服务层保存Runner版本
	err = api.service.SaveVersion(c, runnerID, req.Version, req.Comment, operator)
	if err != nil {
		logger.Error(c, "保存Runner版本失败", err, zap.Int64("id", runnerID))
		response.ServerError(c, "保存Runner版本失败: "+err.Error())
		return
	}

	// 创建响应
	resp := dto.SaveRunnerVersionResp{
		RunnerID:  runnerID,
		Version:   req.Version,
		CreatedAt: time.Now(), // 设置当前时间，实际应从保存的版本记录中获取
	}

	logger.Info(c, "保存Runner版本成功", zap.Int64("id", runnerID), zap.String("version", req.Version))
	response.Success(c, resp)
}

// GetByName 通过用户名和Runner名称获取Runner
func (api *RunnerAPI) GetByName(c *gin.Context) {
	user := c.Param("user")
	name := c.Param("name")

	logger.Debug(c, "开始处理Runner详情请求", zap.String("user", user), zap.String("name", name))

	// 查询条件
	conditions := map[string]interface{}{
		"user": user,
		"name": name,
	}

	// 获取数据
	runners, total, err := api.service.List(c, 1, 1, conditions)
	if err != nil {
		logger.Error(c, "获取Runner详情失败", err)
		response.ServerError(c, "获取Runner详情失败")
		return
	}

	if total == 0 || len(runners) == 0 {
		logger.Info(c, "Runner不存在", zap.String("user", user), zap.String("name", name))
		response.NotFound(c, "Runner不存在")
		return
	}

	logger.Info(c, "获取Runner详情成功", zap.String("user", user), zap.String("name", name))
	response.Success(c, runners[0])
}
