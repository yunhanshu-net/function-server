package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应码
const (
	CodeSuccess      = 0
	CodeParamError   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// 响应消息
var msgMap = map[int]string{
	CodeSuccess:      "成功",
	CodeParamError:   "参数错误",
	CodeUnauthorized: "未授权",
	CodeForbidden:    "禁止访问",
	CodeNotFound:     "资源不存在",
	CodeServerError:  "服务器内部错误",
}

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msgMap[CodeSuccess],
		Data: data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, msg string) {
	// 如果没有提供错误消息，使用默认消息
	if msg == "" {
		if defaultMsg, ok := msgMap[code]; ok {
			msg = defaultMsg
		} else {
			msg = "未知错误"
		}
	}

	// 设置HTTP状态码
	httpStatus := http.StatusOK
	switch code {
	case CodeParamError:
		httpStatus = http.StatusBadRequest
	case CodeUnauthorized:
		httpStatus = http.StatusUnauthorized
	case CodeForbidden:
		httpStatus = http.StatusForbidden
	case CodeNotFound:
		httpStatus = http.StatusNotFound
	case CodeServerError:
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// ParamError 参数错误响应
func ParamError(c *gin.Context, msg string) {
	Fail(c, CodeParamError, msg)
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, CodeUnauthorized, msg)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, msg string) {
	Fail(c, CodeForbidden, msg)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, msg string) {
	Fail(c, CodeNotFound, msg)
}

// ServerError 服务器错误响应
func ServerError(c *gin.Context, msg string) {
	Fail(c, CodeServerError, msg)
}
