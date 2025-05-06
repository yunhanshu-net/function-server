package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/dto/coder"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"go.uber.org/zap"
)

// RuncherOptions Runcher服务配置选项
type RuncherOptions struct {
	NatsURL string        // NATS服务器URL
	Timeout time.Duration // 超时时间
}

// RuncherService Runcher服务接口
type RuncherService interface {
	// RunFunction 执行Runner函数
	RunFunction(ctx context.Context, runner, funcName string, params map[string]interface{}) (interface{}, error)
	// DeployFunction 部署Runner函数
	DeployFunction(ctx context.Context, runnerID int64, language string, code string) error
	// GetFunctionStatus 获取函数状态
	GetFunctionStatus(ctx context.Context, funcID string) (string, error)

	AddAPI2(ctx context.Context, req *coder.AddApiReq) (rsp *coder.AddApiResp, err error)
	// AddAPI 添加API
	AddAPI(ctx context.Context, runnerID int64, funcName, funcTitle, packageName, code string, isPublic bool) (string, interface{}, error)
	// AddAPIs 批量添加API
	AddAPIs(ctx context.Context, runnerID int64, apis []map[string]interface{}) (string, error)
	// CreateProject 创建项目
	CreateProject(ctx context.Context, runner *model.Runner) (string, error)
	// AddBizPackage 添加业务包
	AddBizPackage(ctx context.Context, runnerID int64, packageName, packageTitle, packageDesc string, treeID int64, isPublic bool) (string, error)

	AddBizPackage2(ctx context.Context, bizPackage *coder.BizPackage) (*coder.BizPackageResp, error)

	// Close 关闭服务
	Close() error
}

// runcherService Runcher服务实现
type runcherService struct {
	nc      *nats.Conn
	timeout time.Duration
}

var (
	globalRuncherService RuncherService
	runcherMutex         sync.Mutex
)

// SetGlobalRuncherService 设置全局RuncherService实例
func SetGlobalRuncherService(service RuncherService) {
	runcherMutex.Lock()
	defer runcherMutex.Unlock()
	globalRuncherService = service
}

// GetRuncherService 获取全局RuncherService实例
func GetRuncherService() RuncherService {
	runcherMutex.Lock()
	defer runcherMutex.Unlock()
	return globalRuncherService
}

// NewRuncherService 创建Runcher服务
func NewRuncherService(opts RuncherOptions) (RuncherService, error) {
	if opts.NatsURL == "" {
		opts.NatsURL = nats.DefaultURL
	}
	// 连接NATS服务器
	nc, err := nats.Connect(opts.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("连接NATS服务器失败: %w", err)
	}

	return &runcherService{
		nc:      nc,
		timeout: opts.Timeout,
	}, nil
}

// RunFunction 执行Runner函数
func (s *runcherService) RunFunction(ctx context.Context, runner, funcName string, params map[string]interface{}) (interface{}, error) {
	logger.Debug(ctx, "开始执行Runner函数", zap.String("runner", runner), zap.String("funcName", funcName))

	// 构建请求数据
	reqData := map[string]interface{}{
		"runner": runner,
		"func":   funcName,
		"params": params,
	}

	// 序列化请求数据
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 发送请求并等待响应
	msg := nats.NewMsg("runner.exec")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "执行Runner函数失败", err)
		return nil, fmt.Errorf("执行Runner函数失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "Runner函数执行返回错误", nil, zap.String("errMsg", errMsg))
		return nil, fmt.Errorf("Runner函数执行错误: %s", errMsg)
	}

	// 解析响应数据
	var result interface{}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	logger.Debug(ctx, "Runner函数执行成功", zap.String("runner", runner), zap.String("funcName", funcName))
	return result, nil
}

// DeployFunction 部署Runner函数
func (s *runcherService) DeployFunction(ctx context.Context, runnerID int64, language string, code string) error {
	logger.Debug(ctx, "开始部署Runner函数", zap.Int64("runnerID", runnerID))

	// 部署逻辑实现
	// ...

	logger.Debug(ctx, "部署Runner函数成功", zap.Int64("runnerID", runnerID))
	return nil
}

// GetFunctionStatus 获取函数状态
func (s *runcherService) GetFunctionStatus(ctx context.Context, funcID string) (string, error) {
	logger.Debug(ctx, "开始获取函数状态", zap.String("funcID", funcID))

	// 获取状态逻辑实现
	// ...

	logger.Debug(ctx, "获取函数状态成功", zap.String("funcID", funcID))
	return "running", nil
}

// AddAPI 添加API
func (s *runcherService) AddAPI(ctx context.Context, runnerID int64, funcName, funcTitle, packageName, code string, isPublic bool) (string, interface{}, error) {
	logger.Debug(ctx, "开始添加API", zap.Int64("runnerID", runnerID), zap.String("funcName", funcName))

	// 获取Runner信息
	runnerService := NewRunner(db.GetDB()) // 这里需要依赖注入DB实例
	runner, err := runnerService.Get(ctx, runnerID)
	if err != nil {
		logger.Error(ctx, "获取Runner信息失败", err, zap.Int64("runnerID", runnerID))
		return "", nil, fmt.Errorf("获取Runner信息失败: %w", err)
	}

	if runner == nil {
		logger.Error(ctx, "Runner不存在", nil, zap.Int64("runnerID", runnerID))
		return "", nil, fmt.Errorf("runner不存在")
	}

	// 构建请求数据
	reqData := map[string]interface{}{
		"runner": map[string]interface{}{
			"name":     runner.Name,
			"user":     runner.User,
			"language": runner.Language,
		},
		"code_api": map[string]interface{}{
			"language":         runner.Language,
			"package":          packageName,
			"abs_package_path": packageName,
			"en_name":          funcName,
			"cn_name":          funcTitle,
			"code":             code,
			"is_public":        isPublic,
		},
	}

	// 序列化请求数据
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return "", nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addApi")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "添加API失败", err)
		return "", nil, fmt.Errorf("添加API失败: %w", err)
	}

	// 解析响应码
	code = resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "添加API返回错误", nil, zap.String("errMsg", errMsg))
		return "", nil, fmt.Errorf("添加API错误: %s", errMsg)
	}

	// 解析响应数据
	var result struct {
		Version string      `json:"version"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return "", nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	logger.Info(ctx, "添加API成功", zap.Int64("runnerID", runnerID), zap.String("funcName", funcName), zap.String("version", result.Version))
	return result.Version, result.Data, nil
}
func (s *runcherService) AddAPI2(ctx context.Context, req *coder.AddApiReq) (rsp *coder.AddApiResp, err error) {

	// 序列化请求数据
	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addApi")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "添加API失败", err)
		return nil, fmt.Errorf("添加API失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "添加API返回错误", nil, zap.String("errMsg", errMsg))
		return nil, fmt.Errorf("添加API错误: %s", errMsg)
	}

	// 解析响应数据
	var result coder.AddApiResp
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &result, nil
}

// AddAPIs 批量添加API
func (s *runcherService) AddAPIs(ctx context.Context, runnerID int64, apis []map[string]interface{}) (string, error) {
	logger.Debug(ctx, "开始批量添加API", zap.Int64("runnerID", runnerID))

	// 获取Runner信息
	runnerService := NewRunner(db.GetDB()) // 这里需要依赖注入DB实例
	runner, err := runnerService.Get(ctx, runnerID)
	if err != nil {
		logger.Error(ctx, "获取Runner信息失败", err, zap.Int64("runnerID", runnerID))
		return "", fmt.Errorf("获取Runner信息失败: %w", err)
	}

	if runner == nil {
		logger.Error(ctx, "Runner不存在", nil, zap.Int64("runnerID", runnerID))
		return "", fmt.Errorf("Runner不存在")
	}

	// 构建API数组
	codeApis := make([]map[string]interface{}, 0, len(apis))
	for _, api := range apis {
		codeApi := map[string]interface{}{
			"language":         runner.Language,
			"package":          api["package"],
			"abs_package_path": api["package"],
			"en_name":          api["name"],
			"cn_name":          api["title"],
			"code":             api["code"],
			"is_public":        api["is_public"],
		}
		codeApis = append(codeApis, codeApi)
	}

	// 构建请求数据
	reqData := map[string]interface{}{
		"runner": map[string]interface{}{
			"name":     runner.Name,
			"user":     runner.User,
			"language": runner.Language,
		},
		"code_apis": codeApis,
	}

	// 序列化请求数据
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return "", fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addApis")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "批量添加API失败", err)
		return "", fmt.Errorf("批量添加API失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "批量添加API返回错误", nil, zap.String("errMsg", errMsg))
		return "", fmt.Errorf("批量添加API错误: %s", errMsg)
	}

	// 解析响应数据
	var result struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return "", fmt.Errorf("解析响应数据失败: %w", err)
	}

	logger.Info(ctx, "批量添加API成功", zap.Int64("runnerID", runnerID), zap.String("version", result.Version))
	return result.Version, nil
}

// CreateProject 创建项目
func (s *runcherService) CreateProject(ctx context.Context, runner *model.Runner) (string, error) {
	logger.Debug(ctx, "开始创建项目", zap.String("name", runner.Name), zap.String("user", runner.User))

	// 构建请求数据
	reqData := map[string]interface{}{
		"name":     runner.Name,
		"user":     runner.User,
		"language": runner.Language,
	}

	// 序列化请求数据
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return "", fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.createProject")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "创建项目失败", err)
		return "", fmt.Errorf("创建项目失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "创建项目返回错误", nil, zap.String("errMsg", errMsg))
		return "", fmt.Errorf("创建项目错误: %s", errMsg)
	}

	// 解析响应数据
	var result struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return "", fmt.Errorf("解析响应数据失败: %w", err)
	}

	logger.Info(ctx, "创建项目成功", zap.String("name", runner.Name), zap.String("user", runner.User), zap.String("version", result.Version))
	return result.Version, nil
}

func (s *runcherService) AddBizPackage2(ctx context.Context, bizPackage *coder.BizPackage) (*coder.BizPackageResp, error) {

	// 序列化请求数据
	reqBytes, err := json.Marshal(bizPackage)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addBizPackage")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "添加业务包失败", err)
		return nil, fmt.Errorf("添加业务包失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "添加业务包返回错误", nil, zap.String("errMsg", errMsg))
		return nil, fmt.Errorf("添加业务包错误: %s", errMsg)
	}

	// 解析响应数据
	var result coder.BizPackageResp

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}
	return &result, nil
}

// AddBizPackage 添加业务包
func (s *runcherService) AddBizPackage(ctx context.Context, runnerID int64, packageName, packageTitle, packageDesc string, treeID int64, isPublic bool) (string, error) {
	logger.Debug(ctx, "开始添加业务包", zap.Int64("runnerID", runnerID), zap.String("packageName", packageName))

	// 获取Runner信息
	runnerService := NewRunner(db.GetDB()) // 这里需要依赖注入DB实例
	runner, err := runnerService.Get(ctx, runnerID)
	if err != nil {
		logger.Error(ctx, "获取Runner信息失败", err, zap.Int64("runnerID", runnerID))
		return "", fmt.Errorf("获取Runner信息失败: %w", err)
	}

	if runner == nil {
		logger.Error(ctx, "Runner不存在", nil, zap.Int64("runnerID", runnerID))
		return "", fmt.Errorf("runner不存在")
	}

	// 构建请求数据
	reqData := map[string]interface{}{
		"runner": map[string]interface{}{
			"name":     runner.Name,
			"user":     runner.User,
			"language": runner.Language,
		},
		"language":         runner.Language,
		"abs_package_path": packageName,
		"en_name":          packageName,
		"cn_name":          packageTitle,
		"desc":             packageDesc,
	}

	// 序列化请求数据
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return "", fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addBizPackage")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "添加业务包失败", err)
		return "", fmt.Errorf("添加业务包失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "添加业务包返回错误", nil, zap.String("errMsg", errMsg))
		return "", fmt.Errorf("添加业务包错误: %s", errMsg)
	}

	// 解析响应数据
	var result struct {
		Version string `json:"version"`
	}

	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return "", fmt.Errorf("解析响应数据失败: %w", err)
	}

	logger.Info(ctx, "添加业务包成功", zap.Int64("runnerID", runnerID), zap.String("packageName", packageName), zap.String("version", result.Version))
	return result.Version, nil
}

// Close 关闭服务
func (s *runcherService) Close() error {
	if s.nc != nil {
		s.nc.Close()
	}
	return nil
}

// 获取追踪ID
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}

	return ""
}
