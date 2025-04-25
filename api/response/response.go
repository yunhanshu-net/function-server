// response.go
// 统一响应处理
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var ERROR = -1

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 业务码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 数据
}

// Ok 成功响应
func Ok(c *gin.Context) {
	Success(c, nil)
}

// OkWithMessage 带消息的成功响应
func OkWithMessage(c *gin.Context, message string) {
	Success(c, gin.H{"message": message})
}

// OkWithData 带数据的成功响应
func OkWithData(c *gin.Context, data interface{}) {
	Success(c, data)
}

// Success 通用成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "操作成功",
		Data:    data,
	})
}

// Fail 失败响应
func Fail(c *gin.Context) {
	FailWithMessage(c, "操作失败")
}

// FailWithMessage 带消息的失败响应
func FailWithMessage(c *gin.Context, message string) {
	FailWithCode(c, http.StatusBadRequest, message)
}

// FailWithCode 带状态码的失败响应
func FailWithCode(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    -1,
		Message: message,
		Data:    nil,
	})
}

// FailWithDetailed 返回失败响应带详细数据
func FailWithDetailed(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    ERROR,
		Message: message,
		Data:    data,
	})
}

// FailWithCodeData 返回带状态码和数据的失败响应
func FailWithCodeData(c *gin.Context, httpCode int, message string, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    ERROR,
		Message: message,
		Data:    data,
	})
}
