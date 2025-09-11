# 工作流服务使用指南

## 概述

工作流服务基于 `pkg/workflow` 工作流引擎，提供了完整的工作流管理功能，包括创建、执行、状态查询和停止等操作。

## 核心特性

- ✅ **DSL解析** - 支持基于Go语法的工作流DSL代码
- ✅ **JSON支持** - 支持直接使用JSON数据创建工作流
- ✅ **异步执行** - 工作流异步执行，不阻塞主线程
- ✅ **状态管理** - 完整的执行状态跟踪和持久化
- ✅ **回调机制** - 可扩展的函数执行回调
- ✅ **数据库存储** - 使用 `json.RawMessage` 高效存储JSON数据
- ✅ **FlowID管理** - 执行ID(ExecutionId)就是工作流引擎的FlowID

## 快速开始

### 1. 数据库初始化

工作流相关的表会在数据库初始化时自动创建：

```go
// 在 db.Init() 中已包含
&model.Workflow{},
&model.WorkflowExecution{},
```

### 2. 创建工作流服务

```go
import (
    "github.com/yunhanshu-net/function-server/service"
    "github.com/yunhanshu-net/function-server/pkg/db"
)

// 创建工作流服务
workflowService := service.NewWorkflowService(db.GetDB())
```

### 3. 创建工作流

#### 使用DSL代码创建

```go
dslCode := `
var input = map[string]interface{}{
    "用户名": "张三",
    "手机号": 13800138000,
}

step1 = beiluo.test1.devops.devops_script_create(
    username: string "用户名",
    phone: int "手机号"
) -> (
    workId: string "工号",
    username: string "用户名", 
    err: error "是否失败"
);

func main() {
    //desc: 创建用户账号，获取工号
    工号, 用户名, step1Err := step1(input["用户名"], input["手机号"])
    
    if step1Err != nil {
        step1.Printf("创建用户失败: %v", step1Err)
        return
    }
    
    step1.Printf("✅ 用户创建成功，工号: %s", 工号)
}`

req := &dto.CreateWorkflowReq{
    Name:        "用户注册工作流",
    Description: "处理用户注册和工号分配",
    Code:        dslCode,
}

resp, err := workflowService.CreateWorkflow(ctx, req)
```

#### 使用JSON数据创建

```go
jsonData := map[string]interface{}{
    "success": true,
    "input_vars": map[string]interface{}{
        "用户名": "张三",
        "手机号": 13800138000,
    },
    "steps": []interface{}{
        // 步骤定义...
    },
    "main_func": map[string]interface{}{
        // 主函数定义...
    },
}

jsonBytes, _ := json.Marshal(jsonData)

req := &dto.CreateWorkflowReq{
    Name:        "JSON工作流",
    Description: "使用JSON数据创建的工作流",
    JsonData:    json.RawMessage(jsonBytes),
}

resp, err := workflowService.CreateWorkflow(ctx, req)
```

### 4. 执行工作流

```go
req := &dto.ExecuteWorkflowReq{
    WorkflowId: "1", // 工作流ID
    InputVars: map[string]interface{}{
        "用户名": "李四",
        "手机号": 13900139000,
    },
}

resp, err := workflowService.ExecuteWorkflow(ctx, req)
// resp.ExecutionId 用于后续状态查询
// 注意: ExecutionId 就是工作流引擎的 FlowID
```

### 5. 查询工作流状态

```go
req := &dto.GetWorkflowStatusReq{
    ExecutionId: "exec_1_1234567890",
}

resp, err := workflowService.GetWorkflowStatus(ctx, req)
// resp.Status: running/completed/failed/stopped
// resp.Progress: 执行进度 0.0-1.0
// resp.Variables: 当前变量状态
// resp.Steps: 步骤执行状态
```

### 6. 获取工作流详情

```go
req := &dto.GetWorkflowDetailReq{
    WorkflowId: "1",
}

resp, err := workflowService.GetWorkflowDetail(ctx, req)
// resp.JsonData: 完整的解析后JSON数据
// resp.Name: 工作流名称
// resp.Description: 工作流描述
// resp.Status: 工作流状态
// resp.CreatedAt: 创建时间
// resp.UpdatedAt: 更新时间
```

### 7. 停止工作流

```go
req := &dto.StopWorkflowReq{
    ExecutionId: "exec_1_1234567890",
}

resp, err := workflowService.StopWorkflow(ctx, req)
```

## 数据结构

### Workflow 工作流定义

```go
type Workflow struct {
    Base
    Name        string          `json:"name"`         // 工作流名称
    Description string          `json:"description"`  // 工作流描述
    Code        string          `json:"code"`         // DSL代码
    JsonData    json.RawMessage `json:"json_data"`    // 解析后的JSON数据
    Status      string          `json:"status"`       // 工作流状态
    User        string          `json:"user"`         // 创建用户
}
```

### WorkflowExecution 工作流执行记录

```go
type WorkflowExecution struct {
    Base
    WorkflowId   string          `json:"workflow_id"`   // 工作流ID
    ExecutionId  string          `json:"execution_id"`  // 执行ID
    Status       string          `json:"status"`        // 执行状态
    InputVars    json.RawMessage `json:"input_vars"`    // 输入变量
    CurrentState json.RawMessage `json:"current_state"` // 当前状态
    User         string          `json:"user"`          // 执行用户
    StartTime    *time.Time      `json:"start_time"`    // 开始时间
    EndTime      *time.Time      `json:"end_time"`      // 结束时间
}
```

## 回调机制

工作流服务提供了三个主要的回调接口：

### 1. OnFunctionCall - 函数执行回调

```go
w.executor.OnFunctionCall = func(ctx context.Context, step workflow.SimpleStep, in *workflow.ExecutorIn) (*workflow.ExecutorOut, error) {
    // 这里实现具体的函数执行逻辑
    // 可以调用现有的FunctionExecuteCase.Exec方法
    
    logger.Infof(ctx, "执行步骤: %s, 描述: %s", step.Name, in.StepDesc)
    logger.Infof(ctx, "输入参数: %+v", in.RealInput)
    
    // 返回执行结果
    return &workflow.ExecutorOut{
        Success: true,
        WantOutput: map[string]interface{}{
            "result": "执行结果",
            "err":    nil,
        },
        Error: "",
        Logs:  []string{"步骤执行成功"},
    }, nil
}
```

### 2. OnWorkFlowUpdate - 工作流状态更新回调

```go
w.executor.OnWorkFlowUpdate = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    // 更新执行记录的状态到数据库
    currentState := map[string]interface{}{
        "variables": current.Variables,
        "steps":     current.Steps,
        "status":    "running",
    }
    
    // 更新数据库
    var execution model.WorkflowExecution
    if err := w.db.Where("execution_id = ?", current.FlowID).First(&execution).Error; err == nil {
        execution.SetCurrentState(currentState)
        w.db.Model(&execution).Update("current_state", execution.CurrentState)
    }
    
    return nil
}
```

### 3. OnWorkFlowExit - 工作流正常结束回调

```go
w.executor.OnWorkFlowExit = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    logger.Infof(ctx, "工作流正常结束: %s", current.FlowID)
    return nil
}
```

## 集成现有系统

### 与FunctionExecuteCase集成

工作流服务可以轻松集成现有的 `FunctionExecuteCase` 服务：

```go
// 在setupCallbacks中集成
w.executor.OnFunctionCall = func(ctx context.Context, step workflow.SimpleStep, in *workflow.ExecutorIn) (*workflow.ExecutorOut, error) {
    // 将工作流步骤映射到FunctionExecuteCase
    caseReq := dto.FunctionExecuteCaseReq{
        CaseId:     step.Name, // 使用步骤名作为用例ID
        Background: false,
        Remark:     in.StepDesc,
    }
    
    // 调用现有的函数执行逻辑
    record, rsp, err := w.funcCase.Exec(ctx, caseReq)
    if err != nil {
        return &workflow.ExecutorOut{
            Success: false,
            Error:   err.Error(),
        }, nil
    }
    
    // 转换返回结果
    return &workflow.ExecutorOut{
        Success:    rsp.Code == 0,
        WantOutput: map[string]interface{}{
            "result": rsp.Data,
            "err":    rsp.Msg,
        },
        Error: rsp.Msg,
        Logs:  []string{record.Message},
    }, nil
}
```

## 运行示例

```bash
# 运行工作流示例
cd function-server/examples
go run workflow_example.go
```

## FlowID 管理

工作流引擎使用 `FlowID` 来标识每个工作流实例，在我们的实现中：

- **ExecutionId** = **FlowID** - 执行ID就是工作流引擎的FlowID
- **生成规则** - `exec_{workflow_id}_{timestamp}` 格式
- **唯一性** - 每个执行实例都有唯一的FlowID
- **持久化** - FlowID存储在数据库的 `execution_id` 字段中

## 状态存储策略

我们采用**完整状态存储**策略，将整个 `*workflow.SimpleParseResult` 序列化到数据库：

- **OnWorkFlowUpdate** - 每次状态更新时，将完整的 `SimpleParseResult` 序列化为JSON存储
- **CurrentState字段** - 存储完整的 `json.RawMessage` 格式的工作流状态
- **状态解析** - 直接解析为 `SimpleParseResult` 格式
- **数据完整性** - 确保工作流的所有状态信息都被完整保存

### FlowID 使用示例

```go
// 执行工作流时生成FlowID
executionId := fmt.Sprintf("exec_%d_%d", workflowModel.ID, time.Now().Unix())
// 例如: "exec_1_1703123456"

// 设置到工作流结果中
workflowResult.FlowID = executionId

// 启动工作流引擎
w.executor.Start(ctx, workflowResult)

// 在回调中使用FlowID查询执行记录
w.db.Where("execution_id = ?", current.FlowID).First(&execution)
```

### 完整状态存储示例

```go
// OnWorkFlowUpdate 回调中
w.executor.OnWorkFlowUpdate = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    // 将完整的SimpleParseResult序列化到数据库
    jsonBytes, err := json.Marshal(current)
    if err != nil {
        return err
    }

    // 直接更新current_state为完整的SimpleParseResult JSON
    execution.CurrentState = json.RawMessage(jsonBytes)
    
    // 更新数据库
    w.db.Model(&execution).Updates(map[string]interface{}{
        "current_state": execution.CurrentState,
        "status":        "running",
    })
    
    return nil
}

// OnWorkFlowExit 回调中 - 工作流正常结束
w.executor.OnWorkFlowExit = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    // 更新工作流状态为完成
    var execution model.WorkflowExecution
    if err := w.db.Where("execution_id = ?", current.FlowID).First(&execution).Error; err == nil {
        // 解析当前状态并更新为完成
        var workflowResult workflow.SimpleParseResult
        if err := json.Unmarshal(execution.CurrentState, &workflowResult); err == nil {
            workflowResult.Success = true
            // 序列化并更新状态...
        }
        
        // 更新数据库状态为completed
        w.db.Model(&execution).Updates(map[string]interface{}{
            "status":        "completed",
            "end_time":      time.Now(),
            "current_state": execution.CurrentState,
        })
    }
    
    return nil
}

// 状态查询时解析完整状态
func (w *WorkflowService) GetWorkflowStatus(ctx context.Context, req *dto.GetWorkflowStatusReq) (*dto.GetWorkflowStatusResp, error) {
    // 直接解析为SimpleParseResult
    var workflowResult workflow.SimpleParseResult
    if err := json.Unmarshal(execution.CurrentState, &workflowResult); err != nil {
        return nil, fmt.Errorf("解析工作流状态失败: %v", err)
    }
    
    // 提取状态信息
    currentState := map[string]interface{}{
        "variables": workflowResult.Variables,
        "steps":     workflowResult.Steps,
        "status":    "running",
    }
    // 使用完整的状态信息...
}
```

## 回调函数说明

工作流引擎提供了三个关键回调函数：

### 1. OnFunctionCall - 函数执行回调
```go
w.executor.OnFunctionCall = func(ctx context.Context, step workflow.SimpleStep, in *workflow.ExecutorIn) (*workflow.ExecutorOut, error) {
    // 根据WantParams mock返回相应的数据
    wantOutput := make(map[string]interface{})
    
    // 遍历WantParams，为每个期望的参数生成mock数据
    for _, paramInfo := range in.WantParams {
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
    
    return &workflow.ExecutorOut{
        Success:   true,
        WantOutput: wantOutput,
        Error:     "",
        Logs:      []string{fmt.Sprintf("步骤 %s 执行成功", step.Name)},
    }, nil
}
```

### 2. OnWorkFlowUpdate - 状态更新回调
```go
w.executor.OnWorkFlowUpdate = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    // 每次状态更新时，将完整的SimpleParseResult序列化到数据库
    // 用于实时跟踪工作流执行进度
    return nil
}
```

### 3. OnWorkFlowExit - 工作流结束回调
```go
w.executor.OnWorkFlowExit = func(ctx context.Context, current *workflow.SimpleParseResult) error {
    // 工作流正常结束时，更新状态为completed
    // 设置结束时间和最终状态
    return nil
}
```

## Mock 功能说明

当前 `OnFunctionCall` 回调实现了智能 mock 功能：

### 支持的 Mock 参数类型

| 参数名 | 类型 | Mock 数据 | 说明 |
|--------|------|-----------|------|
| `result` | string | `"步骤 {stepName} 执行成功"` | 执行结果 |
| `err` | error | `nil` | 错误信息 |
| `data` | object | `{"step_name": "...", "status": "completed", "timestamp": ...}` | 返回数据 |
| `message` | string | `"步骤 {stepName} 处理完成"` | 消息 |
| `code` | int | `200` | 状态码 |
| `success` | bool | `true` | 是否成功 |
| `*` | any | `"mock_{paramName}_value"` | 自定义参数 |

### Mock 数据生成规则

1. **智能识别** - 根据 `WantParams` 中的参数名自动生成相应的 mock 数据
2. **类型匹配** - 生成的数据类型与期望的参数类型匹配
3. **动态内容** - 包含步骤名称、时间戳等动态信息
4. **默认处理** - 未知参数自动生成 `mock_{paramName}_value` 格式的默认值

### 使用示例

```go
// 工作流DSL中定义的步骤
step1 := workflow.SimpleStep{
    Name: "create_user",
    Function: "user.create",
    WantParams: []workflow.ParameterInfo{
        {Name: "result", Type: "string", Desc: "执行结果"},
        {Name: "user_id", Type: "int", Desc: "用户ID"},
        {Name: "success", Type: "bool", Desc: "是否成功"},
    },
}

// OnFunctionCall 会自动生成：
// {
//   "result": "步骤 create_user 执行成功",
//   "user_id": "mock_user_id_value",
//   "success": true
// }
```

## 注意事项

1. **数据库连接** - 确保数据库连接正常，工作流表会自动创建
2. **异步执行** - 工作流是异步执行的，需要通过状态查询接口获取执行结果
3. **回调实现** - 需要根据实际业务需求实现 `OnFunctionCall` 回调
4. **错误处理** - 工作流执行失败时会自动更新状态为 `failed`
5. **资源清理** - 长时间运行的工作流需要定期清理执行记录
6. **FlowID一致性** - 确保ExecutionId和FlowID保持一致
7. **状态更新** - OnWorkFlowUpdate 和 OnWorkFlowExit 都会更新数据库状态
8. **Mock测试** - 当前使用 mock 数据，便于测试工作流流程

## API接口列表

| 方法 | 路径 | 功能 | 说明 |
|------|------|------|------|
| POST | `/api/v1/workflow` | 创建工作流 | 支持DSL代码或JSON数据，返回完整JSON数据 |
| POST | `/api/v1/workflow/execute` | 执行工作流 | 支持输入参数 |
| POST | `/api/v1/workflow/status` | 获取工作流状态 | 查询执行进度和结果 |
| GET | `/api/v1/workflow/detail` | 获取工作流详情 | 获取工作流完整信息包括JSON数据 |
| GET | `/api/v1/workflow/list` | 获取工作流列表 | 支持分页和条件查询 |
| POST | `/api/v1/workflow/stop` | 停止工作流 | 停止正在执行的工作流 |

## 下一步

1. 完善回调机制与现有系统集成
2. 添加工作流列表查询接口
3. 实现工作流版本管理
3. 添加工作流可视化界面
4. 实现工作流版本管理
5. 添加工作流监控和告警
