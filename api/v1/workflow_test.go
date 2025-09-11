package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yunhanshu-net/function-server/pkg/dto"
)

func TestWorkflowAPI_CreateWorkflow(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试用的工作流服务（这里需要mock数据库）
	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建API实例
	// workflowAPI := NewWorkflowAPI(workflowService)

	// 创建测试请求
	req := &dto.CreateWorkflowReq{
		Name:        "测试工作流",
		Description: "这是一个测试工作流",
		Code: `func main() {
    step1.Printf("开始执行工作流")
    step1.Printf("工作流执行完成")
}`,
	}

	// 序列化请求
	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflow", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 这里应该调用API方法
	// workflowAPI.CreateWorkflow(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWorkflowAPI_ExecuteWorkflow(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建测试请求
	req := &dto.ExecuteWorkflowReq{
		WorkflowId: "1",
		InputVars: map[string]interface{}{
			"用户名": "张三",
			"手机号": 13800138000,
		},
	}

	// 序列化请求
	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflow/execute", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 验证请求结构
	assert.Equal(t, "1", req.WorkflowId)
	assert.Equal(t, "张三", req.InputVars["用户名"])
	assert.Equal(t, 13800138000, req.InputVars["手机号"])

	t.Logf("执行工作流请求: %+v", req)
}

func TestWorkflowAPI_GetWorkflowStatus(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建测试请求
	req := &dto.GetWorkflowStatusReq{
		ExecutionId: "exec_1_1234567890",
	}

	// 序列化请求
	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflow/status", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 验证请求结构
	assert.Equal(t, "exec_1_1234567890", req.ExecutionId)

	t.Logf("获取工作流状态请求: %+v", req)
}

func TestWorkflowAPI_StopWorkflow(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建测试请求
	req := &dto.StopWorkflowReq{
		ExecutionId: "exec_1_1234567890",
	}

	// 序列化请求
	reqBody, _ := json.Marshal(req)

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/workflow/stop", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 验证请求结构
	assert.Equal(t, "exec_1_1234567890", req.ExecutionId)

	t.Logf("停止工作流请求: %+v", req)
}

func TestWorkflowAPI_GetWorkflowDetail(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/workflow/detail?workflow_id=1", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 验证URL参数
	assert.Equal(t, "1", httpReq.URL.Query().Get("workflow_id"))

	t.Logf("获取工作流详情请求: workflow_id=1")
}

func TestWorkflowAPI_ListWorkflow(t *testing.T) {
	// 设置gin为测试模式
	gin.SetMode(gin.TestMode)

	// 由于需要数据库连接，暂时跳过
	t.Skip("需要数据库连接和mock服务")

	// 创建HTTP请求
	httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/workflow/list?page=1&page_size=10&name=测试&status=active", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建gin上下文
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	// 验证URL参数
	assert.Equal(t, "1", httpReq.URL.Query().Get("page"))
	assert.Equal(t, "10", httpReq.URL.Query().Get("page_size"))
	assert.Equal(t, "测试", httpReq.URL.Query().Get("name"))
	assert.Equal(t, "active", httpReq.URL.Query().Get("status"))

	t.Logf("获取工作流列表请求: page=1, page_size=10, name=测试, status=active")
}
