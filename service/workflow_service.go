package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/query"
	"github.com/yunhanshu-net/pkg/workflow"
	"gorm.io/gorm"
)

type WorkflowService struct {
	db       *gorm.DB
	executor *workflow.Executor
	funcCase *FunctionExecuteCase // 复用现有的函数执行用例服务
}

func NewWorkflowService(db *gorm.DB) *WorkflowService {
	service := &WorkflowService{
		db:       db,
		executor: workflow.NewExecutor(),
		funcCase: NewFunctionExecuteCase(db),
	}

	// 设置执行回调
	service.setupCallbacks()

	return service
}

// CreateWorkflow 创建工作流
func (w *WorkflowService) CreateWorkflow(ctx context.Context, req *dto.CreateWorkflowReq) (*dto.CreateWorkflowResp, error) {
	// 1. 解析DSL代码或JSON数据
	var parseResult map[string]interface{}

	if len(req.JsonData) > 0 {
		// 直接使用提供的JSON数据
		if err := json.Unmarshal(req.JsonData, &parseResult); err != nil {
			return nil, fmt.Errorf("解析JSON数据失败: %v", err)
		}
	} else if req.Code != "" {
		// 解析DSL代码
		parser := workflow.NewSimpleParser()
		workflowResult := parser.ParseWorkflow(req.Code)
		if !workflowResult.Success {
			return nil, fmt.Errorf("DSL解析失败: %s", workflowResult.Error)
		}

		// 转换为map[string]interface{}
		jsonBytes, err := json.Marshal(workflowResult)
		if err != nil {
			return nil, fmt.Errorf("序列化工作流结果失败: %v", err)
		}
		if err := json.Unmarshal(jsonBytes, &parseResult); err != nil {
			return nil, fmt.Errorf("转换工作流结果失败: %v", err)
		}
	} else {
		return nil, fmt.Errorf("必须提供DSL代码或JSON数据")
	}

	// 2. 创建工作流记录
	workflowModel := &model.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		Status:      "active",
		User:        contextx.GetRequestUserName(ctx),
	}

	// 3. 设置解析后的JSON数据
	if err := workflowModel.SetJsonData(parseResult); err != nil {
		return nil, fmt.Errorf("设置JSON数据失败: %v", err)
	}

	// 4. 保存到数据库
	if err := w.db.Create(workflowModel).Error; err != nil {
		return nil, fmt.Errorf("创建工作流失败: %v", err)
	}

	logger.Infof(ctx, "创建工作流成功，ID: %d, 名称: %s", workflowModel.ID, workflowModel.Name)

	return &dto.CreateWorkflowResp{
		WorkflowId:  fmt.Sprintf("%d", workflowModel.ID),
		Name:        workflowModel.Name,
		Description: workflowModel.Description,
		Status:      workflowModel.Status,
		CreatedAt:   workflowModel.CreatedAt.String(),
		JsonData:    workflowModel.JsonData, // 返回完整的解析后JSON数据
	}, nil
}

// ExecuteWorkflow 执行工作流
func (w *WorkflowService) ExecuteWorkflow(ctx context.Context, req *dto.ExecuteWorkflowReq) (*dto.ExecuteWorkflowResp, error) {
	// 1. 获取工作流定义
	var workflowModel model.Workflow
	if err := w.db.Where("id = ?", req.WorkflowId).First(&workflowModel).Error; err != nil {
		return nil, fmt.Errorf("工作流不存在: %v", err)
	}

	// 2. 获取解析后的JSON数据
	jsonData, err := workflowModel.GetJsonData()
	if err != nil {
		return nil, fmt.Errorf("获取工作流数据失败: %v", err)
	}

	// 3. 转换为工作流引擎需要的格式
	workflowResult, err := w.convertToWorkflowResult(jsonData)
	if err != nil {
		return nil, fmt.Errorf("转换工作流数据失败: %v", err)
	}

	// 4. 生成执行ID
	executionId := fmt.Sprintf("exec_%d_%d", workflowModel.ID, time.Now().Unix())

	// 5. 设置FlowID为executionId
	workflowResult.FlowID = executionId

	// 6. 设置输入变量
	if req.InputVars != nil {
		workflowResult.InputVars = req.InputVars
	}

	// 7. 创建执行记录
	execution := &model.WorkflowExecution{
		WorkflowId:  req.WorkflowId,
		ExecutionId: executionId,
		Status:      "running",
		User:        contextx.GetRequestUserName(ctx),
		StartTime:   &[]time.Time{time.Now()}[0],
	}

	// 设置输入变量
	if err := execution.SetInputVars(req.InputVars); err != nil {
		return nil, fmt.Errorf("设置输入变量失败: %v", err)
	}

	if err := w.db.Create(execution).Error; err != nil {
		return nil, fmt.Errorf("创建执行记录失败: %v", err)
	}

	// 8. 异步执行工作流
	go func() {
		execCtx := context.Background()
		if err := w.executor.Start(execCtx, workflowResult); err != nil {
			// 更新执行状态为失败 - 解析当前SimpleParseResult并更新状态
			var currentWorkflowResult workflow.SimpleParseResult
			if unmarshalErr := json.Unmarshal(execution.CurrentState, &currentWorkflowResult); unmarshalErr == nil {
				// 更新状态为失败
				currentWorkflowResult.Success = false
				currentWorkflowResult.Error = err.Error()

				// 序列化更新后的状态
				if jsonBytes, marshalErr := json.Marshal(currentWorkflowResult); marshalErr == nil {
					execution.CurrentState = json.RawMessage(jsonBytes)
				}
			}

			w.db.Model(&execution).Updates(map[string]interface{}{
				"status":        "failed",
				"end_time":      time.Now(),
				"current_state": execution.CurrentState,
			})
			logger.Errorf(execCtx, "工作流执行失败: %v", err)
		} else {
			// 更新执行状态为成功 - 解析当前SimpleParseResult并更新状态
			var currentWorkflowResult workflow.SimpleParseResult
			if unmarshalErr := json.Unmarshal(execution.CurrentState, &currentWorkflowResult); unmarshalErr == nil {
				// 更新状态为成功
				currentWorkflowResult.Success = true

				// 序列化更新后的状态
				if jsonBytes, marshalErr := json.Marshal(currentWorkflowResult); marshalErr == nil {
					execution.CurrentState = json.RawMessage(jsonBytes)
				}
			}

			w.db.Model(&execution).Updates(map[string]interface{}{
				"status":        "completed",
				"end_time":      time.Now(),
				"current_state": execution.CurrentState,
			})
			logger.Infof(execCtx, "工作流执行完成，执行ID: %s", executionId)
		}
	}()

	return &dto.ExecuteWorkflowResp{
		ExecutionId: executionId,
		Status:      "running",
		Message:     "工作流开始执行",
	}, nil
}

// GetWorkflowStatus 获取工作流执行状态
func (w *WorkflowService) GetWorkflowStatus(ctx context.Context, req *dto.GetWorkflowStatusReq) (*dto.GetWorkflowStatusResp, error) {
	// 1. 获取执行记录
	var execution model.WorkflowExecution
	if err := w.db.Where("execution_id = ?", req.ExecutionId).First(&execution).Error; err != nil {
		return nil, fmt.Errorf("执行记录不存在: %v", err)
	}

	// 2. 解析当前状态 - 直接解析为SimpleParseResult
	var workflowResult workflow.SimpleParseResult
	if err := json.Unmarshal(execution.CurrentState, &workflowResult); err != nil {
		logger.Warnf(ctx, "解析工作流状态失败: %v", err)
		return nil, fmt.Errorf("解析工作流状态失败: %v", err)
	}

	// 提取状态信息
	currentState := map[string]interface{}{
		"variables": workflowResult.Variables,
		"steps":     workflowResult.Steps,
		"status":    "running",
	}

	// 3. 构建响应
	resp := &dto.GetWorkflowStatusResp{
		ExecutionId: execution.ExecutionId,
		WorkflowId:  execution.WorkflowId,
		Status:      execution.Status,
		StartTime:   execution.StartTime.String(),
	}

	if execution.EndTime != nil {
		resp.EndTime = execution.EndTime.String()
	}

	// 4. 计算进度和步骤状态
	if variables, ok := currentState["variables"].(map[string]interface{}); ok {
		resp.Variables = variables
	}
	if steps, ok := currentState["steps"].([]interface{}); ok {
		resp.Steps = w.parseStepsStatus(steps)
	}
	resp.Progress = w.calculateProgress(currentState)
	resp.CurrentStep = w.getCurrentStep(currentState)

	return resp, nil
}

// StopWorkflow 停止工作流
func (w *WorkflowService) StopWorkflow(ctx context.Context, req *dto.StopWorkflowReq) (*dto.StopWorkflowResp, error) {
	// 1. 获取执行记录
	var execution model.WorkflowExecution
	if err := w.db.Where("execution_id = ?", req.ExecutionId).First(&execution).Error; err != nil {
		return nil, fmt.Errorf("执行记录不存在: %v", err)
	}

	// 2. 检查状态
	if execution.Status != "running" {
		return nil, fmt.Errorf("工作流未在运行中，当前状态: %s", execution.Status)
	}

	// 3. 停止执行器 - 使用execution_id作为FlowID
	if err := w.executor.Stop(ctx, execution.ExecutionId); err != nil {
		return nil, fmt.Errorf("停止工作流失败: %v", err)
	}

	// 4. 更新执行状态 - 解析当前SimpleParseResult并更新状态
	var workflowResult workflow.SimpleParseResult
	if err := json.Unmarshal(execution.CurrentState, &workflowResult); err != nil {
		logger.Warnf(ctx, "解析工作流状态失败: %v", err)
		// 如果解析失败，创建一个基本的状态
		workflowResult = workflow.SimpleParseResult{
			FlowID: execution.ExecutionId,
		}
	}

	// 更新状态为stopped
	workflowResult.Success = false // 停止的工作流标记为不成功

	// 序列化更新后的状态
	jsonBytes, err := json.Marshal(workflowResult)
	if err != nil {
		return nil, fmt.Errorf("序列化工作流状态失败: %v", err)
	}

	execution.CurrentState = json.RawMessage(jsonBytes)

	if err := w.db.Model(&execution).Updates(map[string]interface{}{
		"status":        "stopped",
		"end_time":      time.Now(),
		"current_state": execution.CurrentState,
	}).Error; err != nil {
		return nil, fmt.Errorf("更新执行状态失败: %v", err)
	}

	logger.Infof(ctx, "工作流已停止，执行ID: %s", req.ExecutionId)

	return &dto.StopWorkflowResp{
		ExecutionId: req.ExecutionId,
		Status:      "stopped",
		Message:     "工作流已成功停止",
	}, nil
}

// GetWorkflowDetail 获取工作流详情
func (w *WorkflowService) GetWorkflowDetail(ctx context.Context, req *dto.GetWorkflowDetailReq) (*dto.GetWorkflowDetailResp, error) {
	// 1. 获取工作流ID
	workflowId, err := strconv.ParseUint(req.WorkflowId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("无效的工作流ID: %s", req.WorkflowId)
	}

	// 2. 查询工作流
	var workflowModel model.Workflow
	if err := w.db.Where("id = ?", workflowId).First(&workflowModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("工作流不存在: %s", req.WorkflowId)
		}
		return nil, fmt.Errorf("查询工作流失败: %v", err)
	}

	// 3. 构建响应
	resp := &dto.GetWorkflowDetailResp{
		WorkflowId:  fmt.Sprintf("%d", workflowModel.ID),
		Name:        workflowModel.Name,
		Description: workflowModel.Description,
		Status:      workflowModel.Status,
		CreatedAt:   workflowModel.CreatedAt.String(),
		UpdatedAt:   workflowModel.UpdatedAt.String(),
		JsonData:    workflowModel.JsonData, // 返回完整的解析后JSON数据
	}

	logger.Infof(ctx, "获取工作流详情成功，ID: %d, 名称: %s", workflowModel.ID, workflowModel.Name)

	return resp, nil
}

// ListWorkflow 获取工作流列表
func (w *WorkflowService) ListWorkflow(ctx context.Context, req *dto.ListWorkflowReq) (*dto.ListWorkflowResp, error) {
	// 构建基础查询
	db := w.db.Model(&model.Workflow{})

	// 添加查询条件
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	// 使用AutoPaginate处理分页和查询条件
	var list []model.Workflow

	result, err := query.AutoPaginateTable(ctx, db, &model.Workflow{}, &list, &req.SearchFilterPageReq)
	if err != nil {
		return nil, fmt.Errorf("查询工作流列表失败: %v", err)
	}

	// 转换为响应格式
	var workflowList []*dto.WorkflowListItem
	for _, workflow := range list {
		workflowList = append(workflowList, &dto.WorkflowListItem{
			WorkflowId:  fmt.Sprintf("%d", workflow.ID),
			Name:        workflow.Name,
			Description: workflow.Description,
			Status:      workflow.Status,
			CreatedAt:   workflow.CreatedAt.String(),
			UpdatedAt:   workflow.UpdatedAt.String(),
		})
	}

	logger.Infof(ctx, "查询工作流列表成功，共%d条记录", len(workflowList))

	return &dto.ListWorkflowResp{
		Items:       workflowList,
		CurrentPage: result.CurrentPage,
		TotalCount:  result.TotalCount,
		TotalPages:  result.TotalPages,
		PageSize:    result.PageSize,
	}, nil
}

// setupCallbacks 设置执行回调
func (w *WorkflowService) setupCallbacks() {
	// 函数执行回调 - 根据WantParams mock返回相应的数据
	w.executor.OnFunctionCall = func(ctx context.Context, step workflow.SimpleStep, in *workflow.ExecutorIn) (*workflow.ExecutorOut, error) {
		logger.Infof(ctx, "执行步骤: %s, 描述: %s", step.Name, in.StepDesc)
		logger.Infof(ctx, "输入参数: %+v", in.RealInput)
		logger.Infof(ctx, "预期返回参数: %+v", in.WantParams)

		// 根据WantParams mock返回相应的数据
		wantOutput := make(map[string]interface{})

		// 遍历WantParams，为每个期望的参数生成mock数据
		for _, paramInfo := range in.WantParams {
			logger.Infof(ctx, "处理期望参数: %s, 类型: %s, 描述: %s", paramInfo.Name, paramInfo.Type, paramInfo.Desc)

			// 根据参数类型生成mock数据
			switch paramInfo.Name {
			case "result":
				wantOutput[paramInfo.Name] = fmt.Sprintf("步骤 %s 执行成功", step.Name)
			case "err":
				wantOutput[paramInfo.Name] = nil
			case "data":
				wantOutput[paramInfo.Name] = map[string]interface{}{
					"step_name": step.Name,
					"status":    "completed",
					"timestamp": time.Now().Unix(),
				}
			case "message":
				wantOutput[paramInfo.Name] = fmt.Sprintf("步骤 %s 处理完成", step.Name)
			case "code":
				wantOutput[paramInfo.Name] = 200
			case "success":
				wantOutput[paramInfo.Name] = true
			default:
				// 对于未知参数，生成默认值
				wantOutput[paramInfo.Name] = fmt.Sprintf("mock_%s_value", paramInfo.Name)
			}
		}

		// 如果没有WantParams，返回默认结果
		if len(wantOutput) == 0 {
			wantOutput = map[string]interface{}{
				"result": "模拟执行结果",
				"err":    nil,
			}
		}

		logger.Infof(ctx, "Mock返回数据: %+v", wantOutput)

		return &workflow.ExecutorOut{
			Success:    true,
			WantOutput: wantOutput,
			Error:      "",
			Logs:       []string{fmt.Sprintf("步骤 %s 执行成功", step.Name)},
		}, nil
	}

	// 工作流状态更新回调
	w.executor.OnWorkFlowUpdate = func(ctx context.Context, current *workflow.SimpleParseResult) error {
		// 将完整的SimpleParseResult序列化到数据库
		jsonBytes, err := json.Marshal(current)
		if err != nil {
			logger.Errorf(ctx, "序列化工作流状态失败: %v", err)
			return err
		}

		// 更新数据库 - 使用FlowID作为execution_id查询
		var execution model.WorkflowExecution
		if err := w.db.Where("execution_id = ?", current.FlowID).First(&execution).Error; err == nil {
			// 直接更新current_state为完整的SimpleParseResult JSON
			execution.CurrentState = json.RawMessage(jsonBytes)

			// 更新执行状态
			if err := w.db.Model(&execution).Updates(map[string]interface{}{
				"current_state": execution.CurrentState,
				"status":        "running",
			}).Error; err != nil {
				logger.Errorf(ctx, "更新工作流状态失败: %v", err)
				return err
			}

			logger.Debugf(ctx, "更新工作流完整状态: execution_id=%s", current.FlowID)
		} else {
			logger.Warnf(ctx, "未找到执行记录: execution_id=%s, error=%v", current.FlowID, err)
		}

		return nil
	}

	// 工作流正常结束回调
	w.executor.OnWorkFlowExit = func(ctx context.Context, current *workflow.SimpleParseResult) error {
		// 更新工作流状态为完成
		var execution model.WorkflowExecution
		if err := w.db.Where("execution_id = ?", current.FlowID).First(&execution).Error; err == nil {
			// 解析当前状态
			var workflowResult workflow.SimpleParseResult
			if unmarshalErr := json.Unmarshal(execution.CurrentState, &workflowResult); unmarshalErr == nil {
				// 更新状态为完成
				workflowResult.Success = true

				// 序列化更新后的状态
				if jsonBytes, marshalErr := json.Marshal(workflowResult); marshalErr == nil {
					execution.CurrentState = json.RawMessage(jsonBytes)
				}
			}

			// 更新数据库
			w.db.Model(&execution).Updates(map[string]interface{}{
				"status":        "completed",
				"end_time":      time.Now(),
				"current_state": execution.CurrentState,
			})

			logger.Infof(ctx, "工作流正常结束: %s", current.FlowID)
		} else {
			logger.Warnf(ctx, "未找到执行记录: execution_id=%s, error=%v", current.FlowID, err)
		}

		return nil
	}
}

// convertToWorkflowResult 将JSON数据转换为工作流引擎需要的格式
func (w *WorkflowService) convertToWorkflowResult(jsonData map[string]interface{}) (*workflow.SimpleParseResult, error) {
	// 将map转换为JSON，再转换为SimpleParseResult
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	var result workflow.SimpleParseResult
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	// 确保FlowID不为空，如果为空则生成一个
	if result.FlowID == "" {
		result.FlowID = fmt.Sprintf("flow_%d", time.Now().UnixNano())
	}

	return &result, nil
}

// 辅助方法
func (w *WorkflowService) parseStepsStatus(steps []interface{}) []dto.WorkflowStepStatus {
	// 解析步骤状态的实现
	return []dto.WorkflowStepStatus{}
}

func (w *WorkflowService) calculateProgress(currentState map[string]interface{}) float64 {
	// 计算执行进度的实现
	return 0.0
}

func (w *WorkflowService) getCurrentStep(currentState map[string]interface{}) string {
	// 获取当前步骤的实现
	return ""
}
