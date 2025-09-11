package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"github.com/yunhanshu-net/pkg/logger"
)

// WorkflowAPI 工作流API
type WorkflowAPI struct {
	workflowService *service.WorkflowService
}

// NewWorkflowAPI 创建工作流API
func NewWorkflowAPI(workflowService *service.WorkflowService) *WorkflowAPI {
	return &WorkflowAPI{
		workflowService: workflowService,
	}
}

// CreateWorkflow 创建工作流
// @Summary 创建工作流
// @Description 创建新的工作流定义，支持DSL代码或JSON数据
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param request body dto.CreateWorkflowReq true "创建工作流请求"
// @Success 200 {object} response.Response{data=dto.CreateWorkflowResp} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow [post]
func (w *WorkflowAPI) CreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := w.workflowService.CreateWorkflow(c, &req)
	if err != nil {
		logger.Errorf(c, "[CreateWorkflow] 创建工作流失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ExecuteWorkflow 执行工作流
// @Summary 执行工作流
// @Description 执行指定的工作流，支持输入参数
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param request body dto.ExecuteWorkflowReq true "执行工作流请求"
// @Success 200 {object} response.Response{data=dto.ExecuteWorkflowResp} "执行成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow/execute [post]
func (w *WorkflowAPI) ExecuteWorkflow(c *gin.Context) {
	var req dto.ExecuteWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := w.workflowService.ExecuteWorkflow(c, &req)
	if err != nil {
		logger.Errorf(c, "[ExecuteWorkflow] 执行工作流失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetWorkflowStatus 获取工作流状态
// @Summary 获取工作流状态
// @Description 获取工作流执行状态和进度信息
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param request body dto.GetWorkflowStatusReq true "获取状态请求"
// @Success 200 {object} response.Response{data=dto.GetWorkflowStatusResp} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow/status [post]
func (w *WorkflowAPI) GetWorkflowStatus(c *gin.Context) {
	var req dto.GetWorkflowStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := w.workflowService.GetWorkflowStatus(c, &req)
	if err != nil {
		logger.Errorf(c, "[GetWorkflowStatus] 获取工作流状态失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// StopWorkflow 停止工作流
// @Summary 停止工作流
// @Description 停止正在执行的工作流
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param request body dto.StopWorkflowReq true "停止工作流请求"
// @Success 200 {object} response.Response{data=dto.StopWorkflowResp} "停止成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow/stop [post]
func (w *WorkflowAPI) StopWorkflow(c *gin.Context) {
	var req dto.StopWorkflowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := w.workflowService.StopWorkflow(c, &req)
	if err != nil {
		logger.Errorf(c, "[StopWorkflow] 停止工作流失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetWorkflowDetail 获取工作流详情
// @Summary 获取工作流详情
// @Description 获取指定工作流的详细信息，包括完整的JSON数据
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param workflow_id query string true "工作流ID"
// @Success 200 {object} response.Response{data=dto.GetWorkflowDetailResp} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "工作流不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow/detail [get]
func (w *WorkflowAPI) GetWorkflowDetail(c *gin.Context) {
	workflowId := c.Query("workflow_id")
	if workflowId == "" {
		response.ParamError(c, "工作流ID不能为空")
		return
	}

	req := &dto.GetWorkflowDetailReq{
		WorkflowId: workflowId,
	}

	resp, err := w.workflowService.GetWorkflowDetail(c, req)
	if err != nil {
		logger.Errorf(c, "[GetWorkflowDetail] 获取工作流详情失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListWorkflow 获取工作流列表
// @Summary 获取工作流列表
// @Description 获取工作流列表，支持分页和条件查询
// @Tags 工作流管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param name query string false "工作流名称（模糊查询）"
// @Param status query string false "工作流状态"
// @Success 200 {object} response.Response{data=dto.ListWorkflowResp} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/workflow/list [get]
func (w *WorkflowAPI) ListWorkflow(c *gin.Context) {
	var req dto.ListWorkflowReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := w.workflowService.ListWorkflow(c, &req)
	if err != nil {
		logger.Errorf(c, "[ListWorkflow] 获取工作流列表失败: %v", err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}
