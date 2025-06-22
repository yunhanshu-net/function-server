package v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
)

// FileAPI 文件API控制器
type FileAPI struct {
	ossService service.OSSService
}

// NewFileAPI 创建文件API控制器
func NewFileAPI(ossService service.OSSService) *FileAPI {
	return &FileAPI{
		ossService: ossService,
	}
}

// GetUploadTokenReq 获取上传token请求
type GetUploadTokenReq struct {
	PathPrefix string `form:"path_prefix" json:"path_prefix"` // 可选的路径前缀
}

// GetUploadTokenResp 获取上传token响应
type GetUploadTokenResp struct {
	Token      string `json:"token"`       // 七牛上传token
	Domain     string `json:"domain"`      // CDN域名
	Bucket     string `json:"bucket"`      // 存储桶名称
	PathPrefix string `json:"path_prefix"` // 建议的文件路径前缀（包含用户ID）
	ExpiresAt  int64  `json:"expires_at"`  // token过期时间戳
	Region     string `json:"region"`      // 存储区域
}

// GetUploadToken 获取文件上传token
// @Summary 获取文件上传token
// @Description 获取七牛云文件上传token，前端使用此token直接上传文件到七牛云
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param path_prefix query string false "文件路径前缀"
// @Success 200 {object} response.Response{data=GetUploadTokenResp} "成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/file/upload-token [get]
func (f *FileAPI) GetUploadToken(c *gin.Context) {
	var req GetUploadTokenReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	userID := c.GetString("user")
	if userID == "" {
		response.Unauthorized(c, "用户未登录")
		return
	}

	// 生成上传token
	token, err := f.ossService.GetUploadToken(c)
	if err != nil {
		response.ServerError(c, "获取上传token失败: "+err.Error())
		return
	}

	// 从配置获取七牛云信息
	qiniuConfig := f.ossService.(*service.QiNiuOSSService).GetConfig()

	// 构建文件路径前缀
	pathPrefix := userID
	if req.PathPrefix != "" {
		pathPrefix = userID + "/" + req.PathPrefix
	}

	resp := &GetUploadTokenResp{
		Token:      token,
		Domain:     qiniuConfig.Domain,
		Bucket:     qiniuConfig.Bucket,
		PathPrefix: pathPrefix,
		ExpiresAt:  time.Now().Add(time.Hour).Unix(), // token 1小时后过期
		Region:     "z0",                             // 华东区域，可以从配置中获取
	}

	response.Success(c, resp)
}
