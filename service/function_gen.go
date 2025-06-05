package service

import (
	"context"
	"fmt"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/pkg/llm"
	_ "github.com/yunhanshu-net/pkg/llm/deepseek"
	"github.com/yunhanshu-net/pkg/logger"
	"github.com/yunhanshu-net/pkg/x/httpx"
	"strings"
	"time"
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
	config := llm.GetDefaultConfig(llm.ProviderDeepSeek)
	config.APIKey = "sk-1ad584ba060842cebd7cf18fbaee701f" // 这里应该从配置文件读取
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
	now := time.Now()
	ss := "\n所属服务目录：" + pkgPath + "\n" + "生成函数类型：" + req.RenderType + "\n"
	messages := ragResp.DecodeData()
	messages = append(messages, llm.Message{Role: "user", Content: fmt.Sprintf("<message>%s</message>", req.Message+ss)})
	err = llm.ChatWithStructMessages(ctx, llm.ProviderDeepSeek, messages, &aiResp)
	cost := time.Since(now)
	if err != nil {
		logger.Infof(ctx, "函数生成失败 req：%s： err:%s cost：%s", req.Message, err.Error(), cost)
		return nil, err
	}

	logger.Infof(ctx, "函数生成成功 req：%s：cost：%s", req.Message, cost)

	rf := model.RunnerFunc{
		Name:     aiResp.EnName,
		Title:    req.Title,
		Code:     aiResp.Code,
		RunnerID: req.RunnerID,
		TreeID:   req.TreeID,
		User:     req.User,
	}
	err = s.Create(ctx, &rf)
	if err != nil {
		return nil, err
	}

	fg := &model.FunctionGen{
		Base: model.Base{
			CreatedBy: req.User,
		},
		CostMill:   cost.Milliseconds(),
		FunctionID: rf.ID,
		TreeID:     req.TreeID,
		Tags:       aiResp.Tags,
		Code:       aiResp.Code,
		Message:    req.Message,
		Level:      aiResp.Level,
		Length:     len(aiResp.Code),
		RenderType: req.RenderType,
		Thinking:   aiResp.Think,
		Enable:     1,
		Status:     "未审核",
		Classify:   "代码示例"}

	db.GetDB().Create(fg)
	return fg, nil

}
