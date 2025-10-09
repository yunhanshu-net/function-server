package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/function-server/pkg/x/contextx"
	"sync"
	"time"

	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/function-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/x/jsonx"

	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/pkg/logger"
	"go.uber.org/zap"
)

// RuncherOptions Runcher服务配置选项
type RuncherOptions struct {
	NatsURL string        // NATS服务器URL
	Timeout time.Duration // 超时时间
}

// RuncherService Runcher服务接口
type RuncherService interface {
	RunFunction2(ctx context.Context, req *runcher.RunFunctionReq) (*nats.Msg, error)
	RunFunction3(ctx context.Context, req *runcher.RunFunctionReq) (*nats.Msg, error)

	AddAPI2(ctx context.Context, req *coder.AddApisReq) (rsp *coder.AddApisResp, err error)
	PushApis(ctx context.Context, req *coder.PushApisReq) (rsp *coder.PushApisResp, err error)
	DeleteAPIs(ctx context.Context, req *coder.DeleteAPIsReq) (rsp *coder.DeleteAPIsResp, err error)
	// CreateProject 创建项目
	CreateProject(ctx context.Context, runner *model.Runner) (string, error)
	DeleteProject(ctx context.Context, req *coder.DeleteProjectReq) (rsp *coder.DeleteProjectResp, err error)
	AddBizPackage2(ctx context.Context, bizPackage *coder.BizPackage) (*coder.BizPackageResp, error)
	RebuildProject(ctx context.Context, req *coder.RebuildProjectReq) (rsp *coder.RebuildProjectResp, err error)

	PublishFunctionRuntime(msg string) error
	// Close 关闭服务
	Close() error
}

// runcherService Runcher服务实现
type runcherService struct {
	nc              *nats.Conn
	sub             *nats.Subscription
	subRunner       *nats.Subscription
	subFunctionCall *nats.Subscription
	timeout         time.Duration
	messageMap      map[string]chan *nats.Msg
	mu              sync.RWMutex // 保护 messageMap 的并发访问
}

func (s *runcherService) PublishFunctionRuntime(msg string) error {
	err := s.nc.Publish("function-runtime.receiver", []byte(msg))
	if err != nil {
		return err
	}
	fmt.Printf("Function Runtime Publish Success msg:%s\n", msg)
	return nil
}

func (s *runcherService) PublishFunctionRuntimeMsg(msg *nats.Msg) error {
	if msg.Subject == "" {
		msg.Subject = fmt.Sprintf("function.run.%s.%s.%s", msg.Header.Get("user"), msg.Header.Get("name"), msg.Header.Get("version"))
	}
	err := s.nc.PublishMsg(msg)
	if err != nil {
		return err
	}
	fmt.Printf("Function Runtime Publish Success msg:%+v\n", msg)
	return nil
}

func (s *runcherService) SubFunctionServerReceiver(handler nats.MsgHandler) error {
	subscribe, err := s.nc.Subscribe("function-server.receiver", handler)
	if err != nil {
		return err
	}
	s.sub = subscribe

	return nil
}

// SubFunctionRunner 消费结果数据
func (s *runcherService) SubFunctionRunner() error {
	subscribe, err := s.nc.Subscribe("function-runner.sub", func(msg *nats.Msg) {
		fmt.Printf("Function server Subscribe msg:%s\n", msg.Data)
		fmt.Printf("Function server Subscribe msg:%s\n", msg.Data)
		tid := msg.Header.Get(constants.TraceID)

		s.mu.Lock()
		if s.messageMap[tid] == nil {
			s.messageMap[tid] = make(chan *nats.Msg, 1)
		}
		ch := s.messageMap[tid]
		s.mu.Unlock()

		ch <- msg
		fmt.Printf("Function Runtime Subscribe Success msg:%s\n", msg.Data)
	})
	if err != nil {
		return err
	}
	s.subRunner = subscribe

	return nil
}

type FunctionCallReq struct {
	Name   string              `json:"name"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}
type FunctionCallResp struct {
	Name   string              `json:"name"`
	Header map[string][]string `json:"header"`
	Body   []byte              `json:"body"`
}

// SubFunctionCall 消费runner的rpc/函数调用
func (s *runcherService) SubFunctionCall() error {
	subscribe, err := s.nc.Subscribe("function-runner.function-call", func(msg *nats.Msg) {
		fmt.Printf("Function server Subscribe msg:%s\n", msg.Data)
		fmt.Printf("Function server Subscribe msg:%s\n", msg.Data)
		//tid := msg.Header.Get(constants.TraceID)

		//todo 这里调用function-call

		//
	})
	if err != nil {
		return err
	}
	s.subFunctionCall = subscribe

	return nil
}

func (s *runcherService) getConn() *nats.Conn {
	connect, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	return connect
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

	r := &runcherService{
		nc:         nc,
		messageMap: make(map[string]chan *nats.Msg),
		timeout:    opts.Timeout,
	}
	err = r.SubFunctionRunner()
	if err != nil {
		panic(err)
	}
	err = r.SubFunctionServerReceiver(func(msg *nats.Msg) {
		fmt.Printf("接收到来自 function-runtime：%s\n", string(msg.Data))
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func newNatsMsg(ctx context.Context, subject, user, runner, version, method, router string) *nats.Msg {

	header := nats.Header{}
	msg := nats.NewMsg(subject)

	header.Set("trace_id", getTraceID(ctx))
	header.Set("user", user)
	header.Set("runner", runner)
	header.Set("version", version)
	header.Set("method", method)
	header.Set("router", router)
	msg.Header = header
	return msg
}

func (s *runcherService) RunFunction2(ctx context.Context, req *runcher.RunFunctionReq) (*nats.Msg, error) {

	if req == nil {
		return nil, fmt.Errorf("<UNK>nil")
	}
	if req.User == "" {
		return nil, fmt.Errorf("user 不能为空")
	}
	if req.Runner == "" {
		return nil, fmt.Errorf("runner 不能为空")
	}
	if req.Router == "" {
		return nil, fmt.Errorf("router 不能为空")
	}
	if req.Version == "" {
		return nil, fmt.Errorf("version 不能为空")
	}
	// 发送请求并等待响应
	msg := nats.NewMsg(fmt.Sprintf("function.run.%s.%s.%s", req.User, req.Runner, req.Version))
	msg.Data = []byte(req.Body)
	header := nats.Header{}
	header.Set("trace_id", getTraceID(ctx))
	header.Set("user", req.User)
	header.Set("runner", req.Runner)
	header.Set("version", req.Version)
	header.Set("method", req.Method)
	header.Set("router", req.Router)
	header.Set("url_query", req.RawQuery)
	msg.Header = header

	//err := s.PublishFunctionRuntimeMsg(msg)
	//if err != nil {
	//	panic(err)
	//}
	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, time.Second*1000)
	if err != nil {
		logger.Error(ctx, "执行Runner函数失败", err)
		return nil, err
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "Runner函数执行返回错误", nil, zap.String("errMsg", errMsg))
		return nil, err
	}

	return resp, nil
}
func (s *runcherService) RunFunction3(ctx context.Context, req *runcher.RunFunctionReq) (*nats.Msg, error) {

	if req == nil {
		return nil, fmt.Errorf("<UNK>nil")
	}
	if req.User == "" {
		return nil, fmt.Errorf("user 不能为空")
	}
	if req.Runner == "" {
		return nil, fmt.Errorf("runner 不能为空")
	}
	if req.Router == "" {
		return nil, fmt.Errorf("router 不能为空")
	}
	if req.Version == "" {
		return nil, fmt.Errorf("version 不能为空")
	}
	// 发送请求并等待响应
	msg := nats.NewMsg(fmt.Sprintf("function.run.%s.%s.%s", req.User, req.Runner, req.Version))
	msg.Data = []byte(req.Body)
	header := nats.Header{}
	tid := getTraceID(ctx)
	header.Set("trace_id", tid)
	header.Set("user", req.User)
	header.Set(constants.RequestUserInfo, contextx.GetRequestUserName(ctx))
	header.Set("runner", req.Runner)
	header.Set("version", req.Version)
	header.Set("method", req.Method)
	header.Set("router", req.Router)
	header.Set("url_query", req.RawQuery)
	msg.Header = header

	s.mu.Lock()
	s.messageMap[tid] = make(chan *nats.Msg, 1)
	ch := s.messageMap[tid]
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		close(s.messageMap[tid])
		delete(s.messageMap, tid)
		s.mu.Unlock()
	}()
	err := s.PublishFunctionRuntimeMsg(msg)
	if err != nil {
		panic(err)
	}
	// 发送请求并等待响应
	//resp, err := s.nc.RequestMsg(msg, time.Second*1000)
	//if err != nil {
	//	logger.Error(ctx, "执行Runner函数失败", err)
	//	return nil, fmt.Errorf("执行Runner函数失败: %w", err)
	//}

	select {
	case <-time.After(time.Second * 1000):
		return nil, fmt.Errorf("timeout")

	case resp := <-ch:
		// 解析响应码
		code := resp.Header.Get("code")
		if code != "0" {
			errMsg := resp.Header.Get("msg")
			logger.Error(ctx, "Runner函数执行返回错误", nil, zap.String("errMsg", errMsg))
			return nil, fmt.Errorf(errMsg)
		}
		return resp, nil
	}

	return nil, nil
}

// RebuildProject 重新编译整个项目，然后返回全量api信息，平台侧可以全量更新
func (s *runcherService) RebuildProject(ctx context.Context, req *coder.RebuildProjectReq) (rsp *coder.RebuildProjectResp, err error) {

	msg := newNatsMsg(ctx, "coder.rebuildProject", req.User, req.Name, req.Version, "", "")
	msg.Data = []byte(jsonx.JSONString(req))
	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, time.Second*100)
	if err != nil {
		logger.Error(ctx, "RebuildProject 执行Runner函数失败", err)
		return nil, fmt.Errorf("RebuildProject 执行Runner函数失败: %w", err)
	}
	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "RebuildProject Runner函数执行返回错误", nil, zap.String("errMsg", errMsg))
		return nil, fmt.Errorf("RebuildProject runner函数执行错误: %s", errMsg)
	}

	rsp = &coder.RebuildProjectResp{}
	err = json.Unmarshal(resp.Data, &rsp)
	if err != nil {
		return nil, fmt.Errorf("RebuildProject 解析响应数据失败: %w", err)
	}

	return rsp, nil
}

func (s *runcherService) DeleteProject(ctx context.Context, req *coder.DeleteProjectReq) (rsp *coder.DeleteProjectResp, err error) {

	if req == nil {
		return nil, fmt.Errorf("<UNK>nil")
	}

	// 发送请求并等待响应
	msg := nats.NewMsg(fmt.Sprintf("coder.deleteProject"))
	msg.Data = []byte(jsonx.String(req))
	header := nats.Header{}
	header.Set("trace_id", getTraceID(ctx))
	header.Set("user", req.User)
	header.Set("runner", req.Runner)
	header.Set("version", req.Version)
	header.Set("method", req.Method)
	header.Set("router", req.Router)
	msg.Header = header

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
		return nil, err
	}
	rsp = &coder.DeleteProjectResp{}
	err = json.Unmarshal(resp.Data, &rsp)
	if err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return rsp, nil
}

func (s *runcherService) AddAPI2(ctx context.Context, req *coder.AddApisReq) (rsp *coder.AddApisResp, err error) {

	// 序列化请求数据
	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.addApis")
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
	var result coder.AddApisResp
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &result, nil
}
func (s *runcherService) PushApis(ctx context.Context, req *coder.PushApisReq) (rsp *coder.PushApisResp, err error) {

	// 序列化请求数据
	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.pushApis")
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
	var result coder.PushApisResp
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &result, nil
}

func (s *runcherService) DeleteAPIs(ctx context.Context, req *coder.DeleteAPIsReq) (rsp *coder.DeleteAPIsResp, err error) {

	// 序列化请求数据
	reqBytes, err := json.Marshal(req)
	if err != nil {
		logger.Error(ctx, "序列化请求数据失败", err)
		return nil, fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求消息
	msg := nats.NewMsg("coder.deleteApis")
	msg.Data = reqBytes
	msg.Header = nats.Header{}
	msg.Header.Set("trace_id", getTraceID(ctx))

	// 发送请求并等待响应
	resp, err := s.nc.RequestMsg(msg, s.timeout)
	if err != nil {
		logger.Error(ctx, "删除api失败", err)
		return nil, fmt.Errorf("添加API失败: %w", err)
	}

	// 解析响应码
	code := resp.Header.Get("code")
	if code != "0" {
		errMsg := resp.Header.Get("msg")
		logger.Error(ctx, "删除api失败", nil, zap.String("errMsg", errMsg))
		return nil, fmt.Errorf("添加API错误: %s", errMsg)
	}

	// 解析响应数据
	var result coder.DeleteAPIsResp
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		logger.Error(ctx, "解析响应数据失败", err)
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &result, nil
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

// Close 关闭服务
func (s *runcherService) Close() error {
	if s.nc != nil {
		s.nc.Close()
	}
	s.sub.Unsubscribe()
	s.subRunner.Unsubscribe()
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
