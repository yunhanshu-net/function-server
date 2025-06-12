package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/llm"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/x/httpx"
)

// FunctionGenQwen 基于千问的函数生成服务
type FunctionGenQwen struct {
	qwenClient  *llm.QwenClient
	serviceTree *ServiceTree
}

// NewFunctionGenQwen 创建千问函数生成服务
func NewFunctionGenQwen() *FunctionGenQwen {
	// 千问配置
	config := &llm.QwenConfig{
		APIKey:  "sk-your-qwen-api-key-here", // 需要替换为实际的API Key
		BaseURL: "https://dashscope.aliyuncs.com/api/v1",
		Model:   "qwen-max", // 使用最强的模型
		Timeout: 180 * time.Second,
	}

	return &FunctionGenQwen{
		qwenClient:  llm.NewQwenClient(config),
		serviceTree: NewServiceTree(db.GetDB()),
	}
}

// QwenAICodeResponse 千问AI代码响应结构体
type QwenAICodeResponse struct {
	Tags    string `json:"tags"`    // 函数标签
	Level   int64  `json:"level"`   // 复杂度1-100
	Code    string `json:"code"`    // 完整Go代码
	Think   string `json:"think"`   // 思考过程
	Package string `json:"package"` // 包名
	EnName  string `json:"en_name"` // 英文函数名
	CnName  string `json:"cn_name"` // 中文函数描述
}

// FunctionGenWithQwen 使用千问生成函数代码
func (s *FunctionGenQwen) FunctionGenWithQwen(ctx context.Context, req *dto.FunctionGenReq) (*model.FunctionGen, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var aiResp QwenAICodeResponse
	var ragResp RagResp

	// 获取服务树信息
	get, err := s.serviceTree.Get(ctx, req.TreeID)
	if err != nil {
		return nil, err
	}
	pkgPath := get.GetPackagePath() // 服务目录

	// 调用知识库
	bd := RagReq{Limit: 10, Role: "all"}
	post, err := httpx.Post("http://localhost:8080/function/run/beiluo/llm_gen_function/knowledge/get/", bd, &ragResp)
	if err != nil {
		return nil, err
	}
	if post.Code != 200 {
		return nil, fmt.Errorf(post.ResBodyString)
	}
	if ragResp.Code != 0 {
		return nil, fmt.Errorf(ragResp.Msg)
	}

	// 创建函数生成记录
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
		Classify:   "代码示例",
	}
	mysqlDb.Create(fg)

	// 获取已存在函数名列表
	var funcs []model.ServiceTree
	var existNames []string
	mysqlDb.Model(&model.ServiceTree{}).Where("parent_id = ? AND type = ?",
		req.TreeID, model.ServiceTreeTypeFunction).Find(&funcs)
	for _, v := range funcs {
		existNames = append(existNames, v.Name)
	}

	rf := &model.RunnerFunc{}
	task := func() error {
		now := time.Now()

		// 构建提示信息
		contextInfo := "\n所属服务目录：" + pkgPath + "\n" + "生成函数类型：" + req.RenderType + "\n" + "该服务目录已经存在的函数逗号分隔多个函数（请勿生成重复函数）：" + strings.Join(existNames, ",")

		// 准备知识库消息
		knowledgeMessages := ragResp.DecodeData()

		// 构建千问消息
		var qwenMessages []llm.QwenMessage
		for _, msg := range knowledgeMessages {
			qwenMessages = append(qwenMessages, llm.QwenMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// 添加用户请求
		qwenMessages = append(qwenMessages, llm.QwenMessage{
			Role:    "user",
			Content: fmt.Sprintf("<message>%s</message>", req.Message+contextInfo),
		})

		// 调用千问生成代码
		err = s.generateCodeWithQwen(ctx, qwenMessages, &aiResp)
		cost := time.Since(now)

		if err != nil {
			logger.Infof(ctx, "千问函数生成失败 req：%s： err:%s cost：%s", req.Message, err.Error(), cost)
			return err
		}

		logger.Infof(ctx, "千问函数生成成功 req：%s：cost：%s", req.Message, cost)

		// 创建RunnerFunc
		rf = &model.RunnerFunc{
			Name:     aiResp.EnName,
			Title:    req.Title,
			Code:     aiResp.Code,
			RunnerID: req.RunnerID,
			TreeID:   req.TreeID,
			User:     req.User,
		}

		// 实现编译失败重试逻辑，最多重试4次
		maxRetries := 4
		var lastError error

		for retry := 0; retry <= maxRetries; retry++ {
			// 这里需要调用真实的Create方法，暂时模拟
			err = s.createRunnerFunc(ctx, rf)
			if err != nil {
				lastError = err
				// 检查是否是编译失败错误
				if strings.Contains(err.Error(), "go build failed") && retry < maxRetries {
					logger.Infof(ctx, "代码编译失败，开始第%d次重试修正，错误：%s", retry+1, err.Error())

					// 使用千问进行代码修正
					fixedCode, fixErr := s.fixCodeWithQwen(ctx, aiResp.Code, err.Error(), qwenMessages)
					if fixErr != nil {
						logger.Errorf(ctx, "代码修正失败：%s", fixErr.Error())
						continue
					}

					// 更新代码并重试
					aiResp.Code = fixedCode
					rf.Code = fixedCode
					logger.Infof(ctx, "第%d次代码修正完成，重新编译中...", retry+1)
					continue
				} else {
					// 非编译错误或达到最大重试次数
					return err
				}
			} else {
				// 编译成功，跳出重试循环
				if retry > 0 {
					logger.Infof(ctx, "代码修正成功！经过%d次重试后编译通过", retry)
				}
				break
			}
		}

		// 如果所有重试都失败了
		if lastError != nil && err != nil {
			logger.Errorf(ctx, "代码编译失败，已重试%d次仍无法修正：%s", maxRetries, lastError.Error())
			return lastError
		}

		// 更新函数生成记录
		up := &model.FunctionGen{
			Base: model.Base{
				ID: fg.ID,
			},
			CostMill:   cost.Milliseconds(),
			FunctionID: rf.ID,
			Tags:       aiResp.Tags,
			Code:       aiResp.Code,
			Level:      aiResp.Level,
			Length:     len(aiResp.Code),
			Thinking:   aiResp.Think,
			Status:     "待审核",
		}

		mysqlDb.Where("id = ?", fg.ID).Updates(up)
		return nil
	}

	if req.Async {
		go func() {
			err = task()
			if err != nil {
				logger.Errorf(ctx, "千问函数生成任务失败：%s", err.Error())
			}
		}()
		return fg, nil
	} else {
		err = task()
		if err != nil {
			logger.Errorf(ctx, "千问函数生成任务失败：%s", err.Error())
			return fg, err
		}
	}

	return fg, nil
}

// generateCodeWithQwen 使用千问生成代码
func (s *FunctionGenQwen) generateCodeWithQwen(ctx context.Context, messages []llm.QwenMessage, response *QwenAICodeResponse) error {
	// 添加结构化输出要求
	systemPrompt := llm.QwenMessage{
		Role: "system",
		Content: `你是一个专业的function-go框架代码生成专家。请严格按照以下JSON格式返回，不要添加任何额外文字或格式标记：

{
  "tags": "函数标签，例如数学，化学，文本转换，文字处理等等",
  "level": 复杂程度1-100的数字，
  "code": "完整的Go代码，包含package声明、import、结构体定义、函数实现等",
  "think": "详细的思考过程，包括需求分析、设计思路、实现方案",
  "package": "包名，通常是模块名",
  "en_name": "英文函数名，符合Go命名规范",
  "cn_name": "中文函数描述"
}

请确保生成的代码：
1. 完全符合function-go框架规范
2. 包含正确的package声明和import语句
3. 包含//go:generate runner标签
4. 定义FunctionInfo结构体
5. 实现Request和Response结构体
6. 实现Callback方法
7. 可以直接编译和运行`,
	}

	// 将系统提示添加到消息开头
	allMessages := append([]llm.QwenMessage{systemPrompt}, messages...)

	// 调用千问API
	jsonContent, err := s.qwenClient.ChatWithJSON(ctx, allMessages)
	if err != nil {
		return fmt.Errorf("千问API调用失败: %w", err)
	}

	// 清理JSON内容
	jsonContent = strings.TrimSpace(jsonContent)
	jsonContent = strings.Trim(jsonContent, "`")
	if strings.HasPrefix(jsonContent, "json\n") {
		jsonContent = strings.TrimPrefix(jsonContent, "json\n")
	}
	if strings.HasPrefix(jsonContent, "json") {
		jsonContent = strings.TrimPrefix(jsonContent, "json")
	}

	logger.Debugf(ctx, "千问清理后的JSON响应: %s", jsonContent)

	// 解析JSON响应
	if err := json.Unmarshal([]byte(jsonContent), response); err != nil {
		return fmt.Errorf("解析千问JSON响应失败: %w, 内容: %s", err, jsonContent)
	}

	// 验证必要字段
	if response.Code == "" {
		return fmt.Errorf("千问返回的代码为空")
	}
	if response.EnName == "" {
		return fmt.Errorf("千问返回的英文函数名为空")
	}

	return nil
}

// fixCodeWithQwen 使用千问修正编译失败的代码
func (s *FunctionGenQwen) fixCodeWithQwen(ctx context.Context, originalCode, errorMsg string, knowledgeMessages []llm.QwenMessage) (string, error) {
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

请以JSON格式返回：
{
  "code": "修正后的完整代码"
}`, originalCode, errorMsg)

	// 组合消息：知识库 + 修正提示
	messages := append(knowledgeMessages, llm.QwenMessage{
		Role:    "user",
		Content: fixPrompt,
	})

	// 调用千问进行代码修正
	jsonContent, err := s.qwenClient.ChatWithJSON(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("千问代码修正调用失败: %w", err)
	}

	// 解析修正结果
	var fixResp struct {
		Code string `json:"code"`
	}

	// 清理JSON内容
	jsonContent = strings.TrimSpace(jsonContent)
	jsonContent = strings.Trim(jsonContent, "`")
	if strings.HasPrefix(jsonContent, "json\n") {
		jsonContent = strings.TrimPrefix(jsonContent, "json\n")
	}

	if err := json.Unmarshal([]byte(jsonContent), &fixResp); err != nil {
		return "", fmt.Errorf("解析千问修正响应失败: %w, 内容: %s", err, jsonContent)
	}

	fixedCode := fixResp.Code
	if fixedCode == "" {
		return "", fmt.Errorf("千问返回空的修正代码")
	}

	return fixedCode, nil
}

// createRunnerFunc 创建RunnerFunc（模拟方法，实际需要调用真实的service）
func (s *FunctionGenQwen) createRunnerFunc(ctx context.Context, rf *model.RunnerFunc) error {
	// 这里应该调用真实的RunnerFunc服务的Create方法
	// 暂时模拟编译成功
	logger.Infof(ctx, "模拟创建RunnerFunc: %s", rf.Name)
	return nil
}
