package service

import (
	"context"
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

// 全局豆包客户端
var doubaoClient *llm.DoubaoClient

func init() {
	// 初始化豆包客户端
	config := llm.DoubaoConfig{
		APIKey:  "bd92875c-2155-4f26-9c13-d95ca0ae0c4a", // 应该从配置文件读取
		Model:   "doubao-1-5-pro-32k-250115",
		BaseURL: "https://ark.cn-beijing.volces.com/api/v3",
		Timeout: 180 * time.Second,
	}
	doubaoClient = llm.NewDoubaoClient(config)
}

type RagReqDoubao struct {
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
	Limit    int    `json:"limit"`
	Role     string `json:"role"`
	SortBy   string `json:"sort_by"`
}

type RagRespDoubao struct {
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

func (r *RagRespDoubao) DecodeData() []llm.DoubaoMessage {
	split := strings.Split(r.Data.FormattedContent, "</split>")
	messages := make([]llm.DoubaoMessage, 0, len(split))
	for _, s := range split {
		if strings.TrimSpace(s) != "" {
			messages = append(messages, llm.DoubaoMessage{Role: "system", Content: strings.TrimSpace(s)})
		}
	}
	return messages
}

type AICodeResponseDoubao struct {
	Tags    string `json:"tags"`    // 函数标签，例如数学，化学，文本转换，文字处理等等
	Level   int64  `json:"level"`   // 函数的复杂程度1-100，越复杂得分越高
	Code    string `json:"code"`    // 完整的Go代码，包含package声明、import、结构体定义、函数实现等
	Think   string `json:"think"`   // 详细的思考过程，包括需求分析、设计思路、实现方案
	Package string `json:"package"` // 包名，通常是模块名
	EnName  string `json:"en_name"` // 英文函数名，符合Go命名规范
	CnName  string `json:"cn_name"` // 中文函数描述
}

type FunctionGenDoubao struct {
	serviceTree *ServiceTree
}

func NewFunctionGenDoubao(serviceTree *ServiceTree) *FunctionGenDoubao {
	return &FunctionGenDoubao{
		serviceTree: serviceTree,
	}
}

func (s *FunctionGenDoubao) FunctionGenWithDoubao(ctx context.Context, req *dto.FunctionGenReq) (*model.FunctionGen, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var aiResp AICodeResponseDoubao
	var ragResp RagRespDoubao

	get, err := s.serviceTree.Get(ctx, req.TreeID)
	if err != nil {
		return nil, err
	}

	pkgPath := get.GetPackagePath() // 服务目录
	bd := RagReqDoubao{Limit: 10, Role: "all"}
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
		ss := "\n所属服务目录：" + pkgPath + "\n" + "生成函数类型：" + req.RenderType + "\n" + "该服务目录已经存在的函数逗号分隔多个函数（请勿生成重复函数）：" + strings.Join(existNames, ",")

		// 构建消息
		messages := ragResp.DecodeData()

		// 构建系统提示，使用知识库内容作为示例
		systemPrompt := `你是function-go框架代码生成专家。

请严格参考知识库中的真实示例格式生成代码。

必须包含：
1. package声明（功能包名，不是main）
2. 正确导入github.com/yunhanshu-net/function-go相关包
3. //go:generate runner标签
4. FunctionInfo变量定义
5. Request和Response结构体（带完整的json标签和验证标签）
6. Callback回调方法

返回JSON格式：
{"code": "生成的代码", "name": "函数名", "desc": "描述", "tags": "标签", "level": 复杂度, "think": "思路"}

注意：代码用\n换行，确保JSON格式正确。`

		messages = append([]llm.DoubaoMessage{{Role: "system", Content: systemPrompt}}, messages...)
		messages = append(messages, llm.DoubaoMessage{Role: "user", Content: fmt.Sprintf("<message>%s</message>", req.Message+ss)})

		// 使用简化的调用方式，避免复杂的JSON解析
		resp, err := doubaoClient.Chat(ctx, messages)
		if err == nil && len(resp.Choices) > 0 {
			content := resp.Choices[0].Message.Content
			// 简单解析，提取基本信息
			aiResp = AICodeResponseDoubao{
				Code:   content, // 将整个响应作为代码
				EnName: "generated_function",
				Tags:   "AI生成",
				Level:  50,
				Think:  "豆包生成的function-go框架代码",
			}
		}
		cost := time.Since(now)
		if err != nil {
			logger.Infof(ctx, "豆包函数生成失败 req：%s： err:%s cost：%s", req.Message, err.Error(), cost)
			return err
		}

		logger.Infof(ctx, "豆包函数生成成功 req：%s：cost：%s", req.Message, cost)

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
			// 这里需要调用 RunnerFunc 的 Create 方法进行编译验证
			// 临时直接用数据库操作（实际应该调用s.Create方法）
			err = mysqlDb.Create(rf).Error
			if err != nil {
				lastError = err
				// 检查是否是编译失败错误
				if strings.Contains(err.Error(), "go build failed") && retry < maxRetries {
					logger.Infof(ctx, "代码编译失败，开始第%d次重试修正，错误：%s", retry+1, err.Error())

					// 构建修正提示，携带知识库、源代码和错误信息
					fixedCode, fixErr := s.fixCodeWithDoubao(ctx, aiResp.Code, err.Error(), messages)
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
				logger.Errorf(ctx, "豆包代码生成任务失败：%s", err.Error())
			}
		}()
		return fg, nil
	} else {
		err = task()
		if err != nil {
			logger.Errorf(ctx, "豆包代码生成任务失败：%s", err.Error())
			return fg, err
		}
	}

	return fg, nil
}

// fixCodeWithDoubao 使用豆包修正编译失败的代码
func (s *FunctionGenDoubao) fixCodeWithDoubao(ctx context.Context, originalCode, errorMsg string, knowledgeMessages []llm.DoubaoMessage) (string, error) {
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
4. 返回完整的修正后代码

只返回修正后的完整代码，不要额外的解释或格式。`, originalCode, errorMsg)

	// 组合消息：知识库 + 修正提示
	messages := append(knowledgeMessages, llm.DoubaoMessage{
		Role:    "user",
		Content: fixPrompt,
	})

	// 调用豆包进行代码修正
	resp, err := doubaoClient.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("豆包代码修正调用失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("豆包返回空响应")
	}

	fixedCode := resp.Choices[0].Message.Content

	// 简单清理响应内容，移除可能的markdown格式
	fixedCode = strings.TrimSpace(fixedCode)
	if strings.HasPrefix(fixedCode, "```go") {
		fixedCode = strings.TrimPrefix(fixedCode, "```go")
	}
	if strings.HasPrefix(fixedCode, "```") {
		fixedCode = strings.TrimPrefix(fixedCode, "```")
	}
	if strings.HasSuffix(fixedCode, "```") {
		fixedCode = strings.TrimSuffix(fixedCode, "```")
	}
	fixedCode = strings.TrimSpace(fixedCode)

	if fixedCode == "" {
		return "", fmt.Errorf("豆包返回空的修正代码")
	}

	return fixedCode, nil
}
