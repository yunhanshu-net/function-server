package v1

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"github.com/yunhanshu-net/pkg/logger"
)

type KnowledgeBaseAPI struct {
	knowledgeBaseService *service.KnowledgeBaseService
}

func NewKnowledgeBaseAPI(knowledgeBaseService *service.KnowledgeBaseService) *KnowledgeBaseAPI {
	return &KnowledgeBaseAPI{
		knowledgeBaseService: knowledgeBaseService,
	}
}

// CreateKnowledgeBase 创建知识库
// @Summary 创建知识库
// @Description 创建新的知识库
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param req body dto.KnowledgeBaseCreateReq true "创建知识库请求"
// @Success 200 {object} response.Response{data=dto.KnowledgeBaseResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base [post]
func (api *KnowledgeBaseAPI) CreateKnowledgeBase(c *gin.Context) {
	var req dto.KnowledgeBaseCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.CreateKnowledgeBase(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 创建知识库失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// UpdateKnowledgeBase 更新知识库
// @Summary 更新知识库
// @Description 更新知识库信息
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param req body dto.KnowledgeBaseUpdateReq true "更新知识库请求"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base [put]
func (api *KnowledgeBaseAPI) UpdateKnowledgeBase(c *gin.Context) {
	var req dto.KnowledgeBaseUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	err := api.knowledgeBaseService.UpdateKnowledgeBase(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 更新知识库失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteKnowledgeBase 删除知识库
// @Summary 删除知识库
// @Description 删除指定知识库
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param kb_key query string true "知识库Key"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base [delete]
func (api *KnowledgeBaseAPI) DeleteKnowledgeBase(c *gin.Context) {
	kbKey := c.Query("kb_key")
	if kbKey == "" {
		response.ParamError(c, "知识库Key不能为空")
		return
	}

	ctx := context.Background()
	err := api.knowledgeBaseService.DeleteKnowledgeBase(ctx, kbKey)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 删除知识库失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetKnowledgeBase 获取知识库详情
// @Summary 获取知识库详情
// @Description 获取指定知识库的详细信息
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param kb_key query string true "知识库Key"
// @Success 200 {object} response.Response{data=dto.KnowledgeBaseResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/detail [get]
func (api *KnowledgeBaseAPI) GetKnowledgeBase(c *gin.Context) {
	kbKey := c.Query("kb_key")
	if kbKey == "" {
		response.ParamError(c, "知识库Key不能为空")
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.GetKnowledgeBase(ctx, kbKey)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 获取知识库详情失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListKnowledgeBases 获取知识库列表
// @Summary 获取知识库列表
// @Description 获取用户的知识库列表
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param name query string false "按名称搜索"
// @Param router query string false "按路由过滤"
// @Param status query string false "按状态过滤"
// @Success 200 {object} response.Response{data=base.Paginated[[]dto.KnowledgeBaseResp]} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/list [get]
func (api *KnowledgeBaseAPI) ListKnowledgeBases(c *gin.Context) {
	var req dto.KnowledgeBaseListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.ListKnowledgeBases(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 获取知识库列表失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// UploadDocument 上传文档
// @Summary 上传文档
// @Description 向知识库上传文档
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param req body dto.KnowledgeDocumentUploadReq true "上传文档请求"
// @Success 200 {object} response.Response{data=dto.KnowledgeDocumentResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/documents [post]
func (api *KnowledgeBaseAPI) UploadDocument(c *gin.Context) {
	var req dto.KnowledgeDocumentUploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.UploadDocument(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 上传文档失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// UpdateDocument 更新文档
// @Summary 更新文档
// @Description 更新知识库中的文档内容
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param req body dto.KnowledgeDocumentUpdateReq true "更新文档请求"
// @Success 200 {object} response.Response{data=dto.KnowledgeDocumentUpdateResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/documents [put]
func (api *KnowledgeBaseAPI) UpdateDocument(c *gin.Context) {
	var req dto.KnowledgeDocumentUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.UpdateDocument(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 更新文档失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// ListDocuments 获取文档列表
// @Summary 获取文档列表
// @Description 获取知识库中的文档列表
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param kb_key query string true "知识库Key"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param title query string false "按标题搜索"
// @Success 200 {object} response.Response{data=base.Paginated[[]dto.KnowledgeDocumentResp]} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/documents [get]
func (api *KnowledgeBaseAPI) ListDocuments(c *gin.Context) {
	var req dto.KnowledgeDocumentListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.ListDocuments(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 获取文档列表失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// SearchKnowledge 搜索知识库
// @Summary 搜索知识库
// @Description 在知识库中搜索相关内容
// @Tags 知识库管理
// @Accept json
// @Produce json
// @Param req body dto.KnowledgeSearchReq true "搜索请求"
// @Success 200 {object} response.Response{data=dto.KnowledgeSearchResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/knowledge-base/search [post]
func (api *KnowledgeBaseAPI) SearchKnowledge(c *gin.Context) {
	var req dto.KnowledgeSearchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	ctx := context.Background()
	resp, err := api.knowledgeBaseService.SearchKnowledge(ctx, &req)
	if err != nil {
		logger.Errorf(ctx, "[KnowledgeBase] 搜索知识库失败：根因：%s 完整错误跟踪:\n%+v\n", errors.Cause(err), err)
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}
