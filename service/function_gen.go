package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yunhanshu-net/function-runtime/pkg/dto/coder"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"github.com/yunhanshu-net/pkg/fmtx"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/pkg/llm"
	_ "github.com/yunhanshu-net/pkg/llm/deepseek"

	// _ "github.com/yunhanshu-net/pkg/llm/qwen"  // 暂时注释掉，包不存在
	"github.com/yunhanshu-net/pkg/logger"
)

type RagReq struct {
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
	Limit    int    `json:"limit"`
	Role     string `json:"role"`
	SortBy   string `json:"sort_by"`
}

type RagResp struct {
	MetaData struct {
		Cost       string `json:"cost"`
		CostMemory string `json:"cost_memory"`
		Memory     string `json:"memory"`
		Version    string `json:"version"`
	} `json:"meta_data"`
	Headers    interface{} `json:"headers"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	TraceId    string      `json:"trace_id"`
	RenderType string      `json:"render_type"`
	Data       struct {
		Categories       string `json:"categories"`
		FormattedContent string `json:"formatted_content"`
		TotalCount       int    `json:"total_count"`
	} `json:"data"`
	DataList interface{} `json:"data_list"`
	Multiple bool        `json:"multiple"`
}

func (r *RagResp) DecodeData() []llm.Message {
	split := strings.Split(r.Data.FormattedContent, "</split>")
	messages := make([]llm.Message, 0, len(split))
	for _, s := range split {
		messages = append(messages, llm.Message{Role: "system", Content: s})
	}
	return messages
}

func init() {
	//config := llm.GetDefaultConfig(llm.ProviderDeepSeek)
	//config.APIKey = "sk-1ad584ba060842cebd7cf18fbaee701f" // 这里应该从配置文件读取
	//config.Timeout = 180 * time.Second
	//
	//_, err := llm.GetOrCreateClient(config)
	//if err != nil {
	//	// 这里不能使用logger.Errorf，因为没有ctx，使用panic或log.Fatal
	//	panic("初始化LLM客户端失败: " + err.Error())
	//}

	config := llm.GetDefaultConfig(llm.ProviderQwen)
	config.APIKey = "sk-7834e5ded5964b14b29c59af2ae9298a" // 这里应该从配置文件读取
	config.Timeout = 180 * time.Second

	_, err := llm.GetOrCreateClient(config)
	if err != nil {
		// 这里不能使用logger.Errorf，因为没有ctx，使用panic或log.Fatal
		panic("初始化LLM客户端失败: " + err.Error())
	}
}

type AICodeResponse struct {
	Tags    string `json:"tags" llm:"desc:函数标签，例如数学，化学，文本转换，文字处理等等"`
	Level   int64  `json:"level" llm:"desc:函数的复杂程度1-100，越复杂得分越高"`
	Code    string `json:"code" llm:"desc:完整的Go代码，包含package声明、import、结构体定义、函数实现等"`
	Think   string `json:"think" llm:"desc:详细的思考过程，包括需求分析、设计思路、实现方案"`
	Package string `json:"package" llm:"desc:包名，通常是模块名"`
	EnName  string `json:"en_name" llm:"desc:英文函数名，符合Go命名规范"`
	CnName  string `json:"cn_name" llm:"desc:中文函数描述"`
}

type FunctionGen struct {
}

// GenCode 调用LLM Agent接口生成代码
func (fg *FunctionGen) GenCode(ctx context.Context, question string) (*LLMAgentChatResp, error) {
	return genCode(ctx, 1, question, "")
}

// LLMAgentChatReq LLM Agent聊天请求参数
type LLMAgentChatReq struct {
	AgentID        int    `json:"agent_id"`
	Question       string `json:"question"`
	Context        string `json:"context"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// LLMAgentChatResp LLM Agent聊天响应参数
type LLMAgentChatResp struct {
	MetaData struct {
		Cost       string `json:"cost"`
		CostMemory string `json:"cost_memory"`
		Memory     string `json:"memory"`
		TraceID    string `json:"trace_id"`
		Version    string `json:"version"`
	} `json:"meta_data"`
	Headers    interface{} `json:"headers"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	TraceID    string      `json:"trace_id"`
	RenderType string      `json:"render_type"`
	Data       struct {
		Answer         string  `json:"answer"`
		Confidence     string  `json:"confidence"`
		Feedback       string  `json:"feedback"`
		KnowledgeCount int     `json:"knowledge_count"`
		ModelName      string  `json:"model_name"`
		ResponseTime   float64 `json:"response_time"`
		TokensUsed     int     `json:"tokens_used"`
		UsedKnowledge  []struct {
			Category     string `json:"category"`
			Content      string `json:"content"`
			CreatedAt    int64  `json:"created_at"`
			DeletedAt    *int64 `json:"deleted_at"`
			ID           int    `json:"id"`
			LastUsedTime int64  `json:"last_used_time"`
			Priority     int    `json:"priority"`
			Status       string `json:"status"`
			Tags         string `json:"tags"`
			Title        string `json:"title"`
			UpdatedAt    int64  `json:"updated_at"`
			UseCount     int    `json:"use_count"`
			ViewCount    int    `json:"view_count"`
		} `json:"used_knowledge"`
	} `json:"data"`
	DataList interface{} `json:"data_list"`
	Multiple bool        `json:"multiple"`
}

func (r *LLMAgentChatResp) DecodeCode() (string, error) {
	if r.Data.Answer == "" {
		return "", fmt.Errorf("响应内容为空")
	}

	// 使用字符串分隔的方式提取代码块
	// 先尝试匹配 ```go 格式
	startMarker := "```go"
	endMarker := "```"

	// 查找开始标记
	startIndex := strings.Index(r.Data.Answer, startMarker)
	if startIndex == -1 {
		// 如果没有找到 ```go，尝试查找 ``` 格式
		startMarker = "```"
		startIndex = strings.Index(r.Data.Answer, startMarker)
		if startIndex == -1 {
			return "", fmt.Errorf("未找到有效的代码块标记，请确保代码被 ```go 和 ``` 包围")
		}
	}

	// 跳过开始标记
	startIndex += len(startMarker)

	// 查找结束标记
	endIndex := strings.Index(r.Data.Answer[startIndex:], endMarker)
	if endIndex == -1 {
		return "", fmt.Errorf("未找到代码块结束标记 ```")
	}

	// 提取代码内容
	code := r.Data.Answer[startIndex : startIndex+endIndex]
	code = strings.TrimSpace(code)

	if code == "" {
		return "", fmt.Errorf("提取的代码块为空")
	}

	return code, nil
}

//func (s *RunnerFunc) Create(gen *model.FunctionGen) error {
//	return db.GetDB().Create(gen).Error
//}

// genCode 调用LLM Agent接口生成代码（私有方法）
func genCode(ctx context.Context, agentID int, question string, context string) (*LLMAgentChatResp, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 构建请求参数
	req := LLMAgentChatReq{
		AgentID:        agentID,
		Question:       question,
		Context:        context,
		TimeoutSeconds: 700, // 默认120秒超时
	}

	// 序列化请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		logger.Errorf(ctx, "序列化请求参数失败: %v", err)
		return nil, fmt.Errorf("序列化请求参数失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/function/run/beiluo/demo6/ai/llm/llm_agent_chat", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Errorf(ctx, "创建HTTP请求失败: %v", err)
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 1000 * time.Second,
	}

	// 发送请求
	httpResp, err := client.Do(httpReq)
	if err != nil {
		logger.Errorf(ctx, "调用LLM Agent接口失败: %v", err)
		return nil, fmt.Errorf("调用LLM Agent接口失败: %w", err)
	}
	defer httpResp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		logger.Errorf(ctx, "读取响应体失败: %v", err)
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 检查HTTP状态码
	if httpResp.StatusCode != 200 {
		logger.Errorf(ctx, "LLM Agent接口返回错误状态码: %d, 响应: %s", httpResp.StatusCode, string(respBody))
		return nil, fmt.Errorf("LLM Agent接口返回错误状态码: %d", httpResp.StatusCode)
	}

	// 解析响应
	var resp LLMAgentChatResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		logger.Errorf(ctx, "解析响应失败: %v, 响应内容: %s", err, string(respBody))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != 0 {
		logger.Errorf(ctx, "LLM Agent业务错误: %s", resp.Msg)
		return nil, fmt.Errorf("LLM Agent业务错误: %s", resp.Msg)
	}

	logger.Infof(ctx, "LLM Agent调用成功，响应时间: %s, 使用Token: %d", resp.MetaData.Cost, resp.Data.TokensUsed)
	return &resp, nil
}

func (s *RunnerFunc) FunctionGen(ctx context.Context, req *dto.FunctionGenReq) (*model.FunctionGen, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	//var messages []llm.Message
	getTree, err := s.serviceTree.Get(ctx, req.TreeID)
	if err != nil {
		return nil, err
	}
	runner, err := s.runnerRepo.Get(ctx, req.RunnerID)

	mysqlDb := db.GetDB()
	fg := &model.FunctionGen{
		Base: model.Base{
			CreatedBy: req.User,
		},
		RunnerID:   req.RunnerID,
		TreeID:     req.TreeID,
		Message:    req.Message,
		RenderType: req.RenderType,
		Enable:     -1,
		Status:     "生成中",
		Classify:   "代码示例"}
	mysqlDb.Create(fg)
	var funcs []model.ServiceTree
	var existNames []string
	mysqlDb.Model(&model.ServiceTree{}).Where("parent_id = ? AND type = ?",
		req.TreeID, model.ServiceTreeTypeFunction).Find(&funcs)
	for _, v := range funcs {
		existNames = append(existNames, v.Name)
	}
	ss := "\n当前package：" + getTree.Name + "\n"

	if len(existNames) > 0 {
		ss += "\n" + "该目录已经存在的函数逗号分隔多个函数（请勿生成重复函数）：" + strings.Join(existNames, ",")
	}
	go func() {
		now := time.Now()
		funcGen := &FunctionGen{}
		resp, err := funcGen.GenCode(ctx, req.Message+"\n"+ss)
		if err != nil {
			mysqlDb.Where("id=?", fg.ID).Updates(map[string]interface{}{
				"status":  "失败",
				"comment": err.Error(),
			})
			return
		}

		// 提取代码
		code, codeErr := resp.DecodeCode()
		if codeErr != nil {
			mysqlDb.Model(&model.FunctionGen{}).Where("id=?", fg.ID).Updates(map[string]interface{}{
				"status":  "失败",
				"code":    code,
				"comment": fmt.Sprintf("代码提取失败: %v", codeErr),
			})
			return
		}
		deleteVar := fmtx.DeleteVar("RouterGroup", code)
		fmtCode := fmtx.ConvertToRouterGroup(deleteVar)
		up := model.FunctionGen{
			Code:       code,
			UpdateCode: fmtCode,
			Status:     "待审核",
		}

		//todo
		//GetRuncherService().AddAPI2(ctx,)
		runnerInfo, err := runnerproject.NewRunner(runner.User, runner.Name, runner.Version)
		if err != nil {
			return
		}
		r := &coder.PushApisReq{
			Runner: runnerInfo,
			CodeApis: []*coder.CodeApi{
				{
					Code:           fmtCode,
					Package:        getTree.Name,
					AbsPackagePath: getTree.GetPackagePath(),
				},
			},
		}
		rsp, err := GetRuncherService().PushApis(ctx, r)
		if err != nil {
			logger.Errorf(ctx, "FunctionGen PushApis:%s", err.Error())
		}
		logger.Infof(ctx, "FunctionGen PushApis: %+v", rsp)

		if err != nil {
			logger.Errorf(ctx, "FunctionGen NewRunner RebuildProject:%s", err.Error())

			// 检查是否是构建失败，如果是则进行重试
			if isBuildFailed(err) {
				logger.Infof(ctx, "检测到构建失败，开始重试修复，FunctionGen ID: %d", fg.ID)
				retryErr := s.retry(ctx, fg, code, err.Error(), getTree, existNames)
				if retryErr != nil {
					logger.Errorf(ctx, "重试修复失败: %v", retryErr)
					// 记录最终失败状态和错误信息
					mysqlDb.Model(&model.FunctionGen{}).Where("id=?", fg.ID).Updates(map[string]interface{}{
						"status":    "重试失败",
						"cost_mill": time.Now().Sub(now).Milliseconds(),
						"comment":   fmt.Sprintf("重试修复失败: %v", retryErr),
					})
				}
			} else {
				// 非构建失败的其他错误，直接标记为失败
				mysqlDb.Model(&model.FunctionGen{}).Where("id=?", fg.ID).Updates(map[string]interface{}{
					"status":    "失败",
					"cost_mill": time.Now().Sub(now).Milliseconds(),
					"comment":   fmt.Sprintf("构建失败: %v", err),
				})
			}
			return
		}
		logger.Infof(ctx, "FunctionGen NewRunner RebuildProject: %+v success", req)
		_, err = NewRunner(db.GetDB()).RebuildProject(ctx, req.RunnerID)
		up.CostMill = time.Now().Sub(now).Milliseconds()
		if err != nil {
			logger.Errorf(ctx, "FunctionGen NewRunner RebuildProject:%s", err.Error())
			up.Status = "失败"
		} else {
			mysqlDb.Where("id=?", fg.ID).Updates(up)
		}
	}()

	return fg, nil

}

func isBuildFailed(err error) bool {
	//2025-09-02 14:16:22.256	ERROR	function-server/service/function_gen.go:326 [func1]		{"msg": "FunctionGen NewRunner RebuildProject:RebuildProject runner函数执行错误: DeleteApis err: go build failed: exit status 1 # github.com/yunhanshu-net/function-go/soft/beiluo/demo7/code/api/tools/image ../../api/tools/image/image_convert.go:86:53: undefined: imaging.TIFFCompression ../../api/tools/image/image_convert.go:86:77: undefined: imaging.LZW ../../api/tools/image/image_convert.go:88:53: undefined: imaging.WebPQuality", "trace_id": "20250902141435-18c5a258-3183-48c5-9548-6e6483df8f79"}
	return strings.Contains(err.Error(), "exit status 1")
}

func (s *RunnerFunc) retry(ctx context.Context, fg *model.FunctionGen, originalCode string, errorMsg string, getTree *model.ServiceTree, existNames []string) error {
	mysqlDb := db.GetDB()

	// 获取当前重试次数
	var retryCount int64
	mysqlDb.Model(&model.FunctionGenRetry{}).Where("function_gen_id = ?", fg.ID).Count(&retryCount)

	// 设置最大重试次数
	maxRetries := 3
	if int(retryCount) >= maxRetries {
		return fmt.Errorf("已达到最大重试次数 %d，停止重试", maxRetries)
	}

	retryIndex := int(retryCount) + 1
	startTime := time.Now()

	logger.Infof(ctx, "开始第 %d 次重试修复，FunctionGen ID: %d", retryIndex, fg.ID)

	// 构建重试提示
	retryPrompt := fmt.Sprintf(`你是function-go框架代码修复专家。

以下代码构建失败，请修复错误：

原始代码：
%s

构建错误信息：
%s

当前package：%s

请根据错误信息修复代码，确保：
1. 完全符合function-go框架规范
2. 修复所有编译错误
3. 保持原有功能逻辑不变
4. 添加缺失的import语句
5. 修正语法错误和类型错误

只返回修复后的完整代码，不要额外的解释或格式。`, originalCode, errorMsg, getTree.Name)

	// 调用LLM进行代码修复
	funcGen := &FunctionGen{}
	resp, err := funcGen.GenCode(ctx, retryPrompt)
	if err != nil {
		// 记录重试失败
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, "", false, fmt.Sprintf("重试调用LLM失败: %v", err), time.Since(startTime).Milliseconds())
		return fmt.Errorf("重试调用LLM失败: %w", err)
	}

	// 提取修复后的代码
	fixedCode, codeErr := resp.DecodeCode()
	if codeErr != nil {
		// 记录重试失败
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, "", false, fmt.Sprintf("重试代码提取失败: %v", codeErr), time.Since(startTime).Milliseconds())
		return fmt.Errorf("重试代码提取失败: %w", codeErr)
	}

	// 处理修复后的代码
	deleteVar := fmtx.DeleteVar("RouterGroup", fixedCode)
	fmtCode := fmtx.ConvertToRouterGroup(deleteVar)

	// 获取runner信息
	runner, err := s.runnerRepo.Get(ctx, fg.RunnerID)
	if err != nil {
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, fixedCode, false, fmt.Sprintf("获取runner信息失败: %v", err), time.Since(startTime).Milliseconds())
		return fmt.Errorf("获取runner信息失败: %w", err)
	}

	// 推送到runtime
	runnerInfo, err := runnerproject.NewRunner(runner.User, runner.Name, runner.Version)
	if err != nil {
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, fixedCode, false, fmt.Sprintf("创建runner信息失败: %v", err), time.Since(startTime).Milliseconds())
		return fmt.Errorf("创建runner信息失败: %w", err)
	}

	r := &coder.PushApisReq{
		Runner: runnerInfo,
		CodeApis: []*coder.CodeApi{
			{
				Code:           fmtCode,
				Package:        getTree.Name,
				AbsPackagePath: getTree.GetPackagePath(),
			},
		},
	}

	rsp, err := GetRuncherService().PushApis(ctx, r)
	if err != nil {
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, fixedCode, false, fmt.Sprintf("重试推送API失败: %v", err), time.Since(startTime).Milliseconds())
		logger.Errorf(ctx, "重试推送API失败: %+v err:%+v", rsp, err)
	} else {
		logger.Infof(ctx, "重试推送API成功: %+v", rsp)

	}

	// 重新构建项目
	_, err = NewRunner(db.GetDB()).RebuildProject(ctx, fg.RunnerID)
	if err != nil {
		// 记录重试失败
		s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, fixedCode, false, fmt.Sprintf("重试后重新构建项目失败: %v", err), time.Since(startTime).Milliseconds())

		// 如果还有重试机会，继续重试
		if int(retryCount+1) < maxRetries {
			logger.Infof(ctx, "第 %d 次重试失败，准备进行第 %d 次重试", retryIndex, retryIndex+1)
			return s.retry(ctx, fg, fixedCode, err.Error(), getTree, existNames)
		} else {
			// 达到最大重试次数，更新状态为最终失败
			mysqlDb.Model(&model.FunctionGen{}).Where("id=?", fg.ID).Updates(map[string]interface{}{
				"status":    "重试失败",
				"cost_mill": time.Now().Sub(time.Time(fg.CreatedAt)).Milliseconds(),
				"comment":   fmt.Sprintf("重试 %d 次后仍然构建失败: %v", maxRetries, err),
			})
			return fmt.Errorf("重试 %d 次后仍然构建失败: %w", maxRetries, err)
		}
	}

	// 重试成功，记录成功信息
	s.recordRetryAttempt(ctx, fg.ID, retryIndex, originalCode, errorMsg, fixedCode, true, "", time.Since(startTime).Milliseconds())

	// 更新FunctionGen状态
	up := model.FunctionGen{
		CostMill:   time.Now().Sub(time.Time(fg.CreatedAt)).Milliseconds(),
		Code:       fixedCode,
		UpdateCode: fmtCode,
		Status:     "重试成功",
		Comment:    fmt.Sprintf("第 %d 次重试修复成功", retryIndex),
	}
	err = mysqlDb.Model(&model.FunctionGen{}).Where("id=?", fg.ID).Updates(up).Error
	if err != nil {
		logger.Errorf(ctx, "更新重试成功状态失败: %v", err)
	}

	logger.Infof(ctx, "第 %d 次重试修复成功，FunctionGen ID: %d", retryIndex, fg.ID)
	return nil
}

// recordRetryAttempt 记录重试尝试
func (s *RunnerFunc) recordRetryAttempt(ctx context.Context, functionGenID int64, retryIndex int, originalCode, errorMsg, fixedCode string, success bool, finalError string, costMill int64) {
	mysqlDb := db.GetDB()

	retryRecord := &model.FunctionGenRetry{
		FunctionGenID: functionGenID,
		RetryIndex:    retryIndex,
		OriginalCode:  originalCode,
		ErrorMsg:      errorMsg,
		FixedCode:     fixedCode,
		Success:       success,
		FinalError:    finalError,
		RetryTime:     time.Now().UnixMilli(),
		CostMill:      costMill,
	}

	err := mysqlDb.Create(retryRecord).Error
	if err != nil {
		logger.Errorf(ctx, "记录重试尝试失败: %v", err)
	} else {
		logger.Infof(ctx, "记录重试尝试成功，FunctionGen ID: %d, 重试次数: %d, 成功: %v", functionGenID, retryIndex, success)
	}
}

//func (s *RunnerFunc) FunctionGen(ctx context.Context, req *dto.FunctionGenReq) (*model.FunctionGen, error) {
//	if ctx.Err() != nil {
//		return nil, ctx.Err()
//	}
//	//var messages []llm.Message
//	var aiResp AICodeResponse
//	var ragResp RagResp
//	get, err := s.serviceTree.Get(ctx, req.TreeID)
//	if err != nil {
//		return nil, err
//	}
//	pkgPath := get.GetPackagePath() //服务目录
//	bd := RagReq{Limit: 10, Role: "all"}
//	post, err := httpx.Post("http://localhost:8080/function/run/beiluo/llm_gen_function/knowledge/get/").Body(bd).Do(&ragResp)
//	if err != nil {
//		return nil, err
//	}
//	if post.Code != 200 {
//		return nil, fmt.Errorf(post.ResBodyString)
//	}
//	if ragResp.Code != 0 {
//		return nil, fmt.Errorf(ragResp.Msg)
//	}
//	mysqlDb := db.GetDB()
//	fg := &model.FunctionGen{
//		Base: model.Base{
//			CreatedBy: req.User,
//		},
//		RunnerID:   req.RunnerID,
//		TreeID:     req.TreeID,
//		Message:    req.Message,
//		RenderType: req.RenderType,
//		Enable:     -1,
//		Status:     "生成中",
//		Classify:   "代码示例"}
//	mysqlDb.Create(fg)
//	var funcs []model.ServiceTree
//	var existNames []string
//	mysqlDb.Model(&model.ServiceTree{}).Where("parent_id = ? AND type = ?",
//		req.TreeID, model.ServiceTreeTypeFunction).Find(&funcs)
//	for _, v := range funcs {
//		existNames = append(existNames, v.Name)
//	}
//
//	rf := &model.RunnerFunc{}
//	task := func() error {
//		now := time.Now()
//		ss := "\n所属服务目录：" + pkgPath + "\n" + "生成函数类型：" + req.RenderType + "\n" + "该服务目录已经存在的函数逗号分隔多个函数（请勿生成重复函数）：" + strings.Join(existNames, ",")
//		messages := ragResp.DecodeData()
//		messages = append(messages, llm.Message{Role: "user", Content: fmt.Sprintf("<message>%s</message>", req.Message+ss)})
//		err = llm.ChatWithStructMessages(ctx, llm.ProviderQwen, messages, &aiResp)
//		cost := time.Since(now)
//		if err != nil {
//			logger.Infof(ctx, "函数生成失败 req：%s： err:%s cost：%s", req.Message, err.Error(), cost)
//			return err
//		}
//
//		logger.Infof(ctx, "函数生成成功 req：%s：cost：%s", req.Message, cost)
//
//		rf = &model.RunnerFunc{
//			Name:     aiResp.EnName,
//			Title:    req.Title,
//			Code:     aiResp.Code,
//			RunnerID: req.RunnerID,
//			TreeID:   req.TreeID,
//			User:     req.User,
//		}
//
//		// 实现编译失败重试逻辑，最多重试4次
//		maxRetries := 4
//		var lastError error
//
//		for retry := 0; retry <= maxRetries; retry++ {
//			err = s.Create(ctx, rf)
//			if err != nil {
//				lastError = err
//				// 检查是否是编译失败错误
//				if strings.Contains(err.Error(), "go build failed") && retry < maxRetries {
//					logger.Infof(ctx, "代码编译失败，开始第%d次重试修正，错误：%s", retry+1, err.Error())
//
//					// 进行代码修正
//					fixedCode, fixErr := s.fixCodeWithDeepSeek(ctx, aiResp.Code, err.Error(), messages)
//					if fixErr != nil {
//						logger.Errorf(ctx, "代码修正失败：%s", fixErr.Error())
//						continue
//					}
//
//					// 更新代码并重试
//					aiResp.Code = fixedCode
//					rf.Code = fixedCode
//					logger.Infof(ctx, "第%d次代码修正完成，重新编译中...", retry+1)
//					continue
//				} else {
//					// 非编译错误或达到最大重试次数
//					return err
//				}
//			} else {
//				// 编译成功，跳出重试循环
//				if retry > 0 {
//					logger.Infof(ctx, "代码修正成功！经过%d次重试后编译通过", retry)
//				}
//				break
//			}
//		}
//
//		// 如果所有重试都失败了
//		if lastError != nil && err != nil {
//			logger.Errorf(ctx, "代码编译失败，已重试%d次仍无法修正：%s", maxRetries, lastError.Error())
//			return lastError
//		}
//		up := &model.FunctionGen{
//			Base: model.Base{
//				ID: fg.ID,
//			},
//			CostMill:   cost.Milliseconds(),
//			FunctionID: rf.ID,
//			Tags:       aiResp.Tags,
//			Code:       aiResp.Code,
//			Level:      aiResp.Level,
//			Length:     len(aiResp.Code),
//			Thinking:   aiResp.Think,
//			Status:     "待审核"}
//
//		mysqlDb.Where("id = ?", fg.ID).Updates(up)
//		return nil
//	}
//	if req.Async {
//		go func() {
//			err = task()
//			if err != nil {
//				logger.Errorf(ctx, "task err:%s", err.Error())
//			}
//		}()
//		return fg, nil
//	} else {
//		err = task()
//		if err != nil {
//			logger.Errorf(ctx, "task err:%s", err.Error())
//			return fg, err
//		}
//	}
//
//	return fg, nil
//
//}

// fixCodeWithDeepSeek 使用DeepSeek修正编译失败的代码
func (s *RunnerFunc) fixCodeWithDeepSeek(ctx context.Context, originalCode, errorMsg string, knowledgeMessages []llm.Message) (string, error) {
	// 构建修正提示
	fixPrompt := fmt.Sprintf(`你是function-go框架代码修正专家。

以下代码编译失败，请修正错误：

原始代码：
%s

编译错误信息：
%s

请根据知识库示例和错误信息修正代码，确保：
1. 完全符合function-go框架规范
2. 修正编译错误
3. 保持原有功能逻辑不变
4. 添加缺失的import语句
5. 修正语法错误

只返回修正后的完整代码，不要额外的解释或格式。`, originalCode, errorMsg)

	// 组合消息：知识库 + 修正提示
	messages := append(knowledgeMessages, llm.Message{
		Role:    "user",
		Content: fixPrompt,
	})

	// 调用DeepSeek进行代码修正
	var fixResp AICodeResponse
	err := llm.ChatWithStructMessages(ctx, llm.ProviderDeepSeek, messages, &fixResp)
	if err != nil {
		return "", fmt.Errorf("代码修正调用失败: %w", err)
	}

	fixedCode := fixResp.Code
	if fixedCode == "" {
		return "", fmt.Errorf("返回空的修正代码")
	}

	return fixedCode, nil
}
