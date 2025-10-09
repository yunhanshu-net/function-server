package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/pkg/interfaces"
	"github.com/yunhanshu-net/function-server/service"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, message string, data ...interface{}) {
	resp := Response{
		Code:    0,
		Message: message,
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string, data ...interface{}) {
	resp := Response{
		Code:    code,
		Message: message,
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	c.JSON(code, resp)
}

// ParamError 参数错误响应
func ParamError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// QQEmailConfigReq QQ邮箱配置请求
type QQEmailConfigReq struct {
	Username  string `json:"username" binding:"required"`   // QQ邮箱账号
	Password  string `json:"password" binding:"required"`   // QQ邮箱授权码
	FromEmail string `json:"from_email" binding:"required"` // 发件人邮箱
	FromName  string `json:"from_name"`                     // 发件人名称（可选）
}

// QQEmailSendReq QQ邮件发送请求
type QQEmailSendReq struct {
	To      []string `json:"to" binding:"required"`      // 收件人邮箱列表
	Subject string   `json:"subject" binding:"required"` // 邮件主题
	Body    string   `json:"body" binding:"required"`    // 邮件正文
	IsHTML  bool     `json:"is_html"`                    // 是否为HTML格式
}

// QQEmailSendResp QQ邮件发送响应
type QQEmailSendResp struct {
	Success bool   `json:"success"` // 发送是否成功
	Message string `json:"message"` // 响应消息
	Error   string `json:"error"`   // 错误信息（如果有）
}

// QQEmailConfig 设置QQ邮箱配置
// @Summary 设置QQ邮箱配置
// @Description 配置QQ邮箱的SMTP认证信息
// @Tags QQ邮件
// @Accept json
// @Produce json
// @Param request body QQEmailConfigReq true "配置信息"
// @Success 200 {object} Response
// @Router /api/v1/qq-email/config [post]
func (f *Functions) QQEmailConfig(c *gin.Context) {
	var req QQEmailConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, err.Error())
		return
	}

	// 设置QQ邮箱配置
	functionCallService := service.NewFunctionCallService()
	if err := functionCallService.SetQQEmailConfig(req.Username, req.Password, req.FromEmail, req.FromName); err != nil {
		Error(c, http.StatusInternalServerError, "设置配置失败: "+err.Error())
		return
	}

	Success(c, "QQ邮箱配置设置成功")
}

// QQEmailSend 发送QQ邮件
// @Summary 发送QQ邮件
// @Description 通过QQ邮箱发送邮件
// @Tags QQ邮件
// @Accept json
// @Produce json
// @Param request body QQEmailSendReq true "邮件信息"
// @Success 200 {object} Response{data=QQEmailSendResp}
// @Router /api/v1/qq-email/send [post]
func (f *Functions) QQEmailSend(c *gin.Context) {
	var req QQEmailSendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		ParamError(c, err.Error())
		return
	}

	// 构建邮件请求数据
	emailData := map[string]interface{}{
		"to":      req.To,
		"subject": req.Subject,
		"body":    req.Body,
		"is_html": req.IsHTML,
	}

	emailBody, err := json.Marshal(emailData)
	if err != nil {
		Error(c, http.StatusInternalServerError, "构建请求数据失败: "+err.Error())
		return
	}

	// 构建函数调用请求
	callReq := &interfaces.FunctionCallReq{
		Name:   "qq_email",
		Header: make(map[string][]string),
		Body:   emailBody,
	}

	// 调用QQ邮件发送服务
	functionCallService := service.NewFunctionCallService()
	callResp, err := functionCallService.CreateFunctionCall(callReq)
	if err != nil {
		Error(c, http.StatusInternalServerError, "邮件发送失败: "+err.Error())
		return
	}

	// 解析响应
	var resp QQEmailSendResp
	if err := json.Unmarshal(callResp.Body, &resp); err != nil {
		Error(c, http.StatusInternalServerError, "解析响应失败: "+err.Error())
		return
	}

	if resp.Success {
		Success(c, resp.Message, resp)
	} else {
		Error(c, http.StatusBadRequest, resp.Message, resp)
	}
}

// QQEmailGetConfig 获取QQ邮箱配置
// @Summary 获取QQ邮箱配置
// @Description 获取当前QQ邮箱的配置信息
// @Tags QQ邮件
// @Produce json
// @Success 200 {object} Response{data=map[string]string}
// @Router /api/v1/qq-email/config [get]
func (f *Functions) QQEmailGetConfig(c *gin.Context) {
	functionCallService := service.NewFunctionCallService()
	config, err := functionCallService.GetQQEmailConfig()
	if err != nil {
		Error(c, http.StatusInternalServerError, "获取配置失败: "+err.Error())
		return
	}

	// 隐藏敏感信息
	safeConfig := make(map[string]string)
	for k, v := range config {
		if k == "password" {
			safeConfig[k] = "***"
		} else {
			safeConfig[k] = v
		}
	}

	Success(c, "获取配置成功", safeConfig)
}
