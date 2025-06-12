package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/pkg/llm"
	_ "github.com/yunhanshu-net/pkg/llm/deepseek"
	_ "github.com/yunhanshu-net/pkg/llm/qwen"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/x/httpx"
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

//func (s *RunnerFunc) Create(gen *model.FunctionGen) error {
//	return db.GetDB().Create(gen).Error
//}

func (s *RunnerFunc) FunctionGen(ctx context.Context, req *dto.FunctionGenReq) (*model.FunctionGen, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	//var messages []llm.Message
	var aiResp AICodeResponse
	var ragResp RagResp
	get, err := s.serviceTree.Get(ctx, req.TreeID)
	if err != nil {
		return nil, err
	}
	pkgPath := get.GetPackagePath() //服务目录
	bd := RagReq{Limit: 10, Role: "all"}
	post, err := httpx.Post("http://localhost:8080/function/run/beiluo/llm_gen_function/knowledge/get/", bd, &ragResp) //调用知识库知识
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
		Classify:   "代码示例"}
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
		messages := ragResp.DecodeData()
		messages = append(messages, llm.Message{Role: "user", Content: fmt.Sprintf("<message>%s</message>", req.Message+ss)})
		err = llm.ChatWithStructMessages(ctx, llm.ProviderQwen, messages, &aiResp)
		cost := time.Since(now)
		if err != nil {
			logger.Infof(ctx, "函数生成失败 req：%s： err:%s cost：%s", req.Message, err.Error(), cost)
			return err
		}

		logger.Infof(ctx, "函数生成成功 req：%s：cost：%s", req.Message, cost)

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
			err = s.Create(ctx, rf)
			if err != nil {
				lastError = err
				// 检查是否是编译失败错误
				if strings.Contains(err.Error(), "go build failed") && retry < maxRetries {
					logger.Infof(ctx, "代码编译失败，开始第%d次重试修正，错误：%s", retry+1, err.Error())

					// 进行代码修正
					fixedCode, fixErr := s.fixCodeWithDeepSeek(ctx, aiResp.Code, err.Error(), messages)
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
			Status:     "待审核"}

		mysqlDb.Where("id = ?", fg.ID).Updates(up)
		return nil
	}
	if req.Async {
		go func() {
			err = task()
			if err != nil {
				logger.Errorf(ctx, "task err:%s", err.Error())
			}
		}()
		return fg, nil
	} else {
		err = task()
		if err != nil {
			logger.Errorf(ctx, "task err:%s", err.Error())
			return fg, err
		}
	}

	return fg, nil

}

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
