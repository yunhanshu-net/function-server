package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	resp "github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-server/pkg/dto/callback"
	"github.com/yunhanshu-net/function-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"io"
	"net/http"
	"strconv"
)

// CallbackRequest 回调请求结构
type CallbackRequest struct {
	Method string      `json:"method"`
	Router string      `json:"router"`
	Type   string      `json:"type"`
	Body   interface{} `json:"body"`
}

// Functions API处理器
type Functions struct {
	runcher         service.RuncherService
	runner          *service.Runner
	callbackService *service.CallbackService
}

func (r *Functions) Callback(c *gin.Context) {
	req := &runcher.RunFunctionReq{
		User:   c.Param("user"),
		Method: http.MethodPost,
		Runner: c.Param("runner"),
		Router: "_callback",
	}
	rn, err := r.runner.GetByUserName(c, req.User, req.Runner)
	if err != nil {
		response.ParamError(c, fmt.Sprintf("获取runner失败：%s", err.Error()))
		return
	}
	req.Version = rn.Version
	all, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	req.Body = string(all)

	// 解析回调请求以获取回调类型
	var callbackReq CallbackRequest
	if err := json.Unmarshal(all, &callbackReq); err != nil {
		response.ParamError(c, fmt.Sprintf("解析回调请求失败：%s", err.Error()))
		return
	}

	function2, err := r.runcher.RunFunction2(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	var res resp.RunFunctionResp
	err = json.Unmarshal(function2.Data, &res)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	if res.MetaData == nil {
		res.MetaData = make(map[string]interface{})
	}
	for k, v := range function2.Header {
		if k != "code" {
			if len(v) > 0 {
				res.MetaData[k] = v[0]
			}
		}
	}
	res.MetaData["version"] = rn.Version

	// 构造回调上下文
	callbackCtx := &callback.CallbackContext{
		User:     req.User,
		Runner:   req.Runner,
		Type:     callbackReq.Type,
		Body:     callbackReq.Body,
		Response: &res,
	}

	// 从请求头获取函数ID
	if funcIDStr := c.GetHeader("x-function-id"); funcIDStr != "" {
		if funcID, err := strconv.ParseInt(funcIDStr, 10, 64); err == nil {
			callbackCtx.FuncID = funcID
		}
	}

	// 后置处理：调用service层处理业务逻辑
	if err := r.postCallbackProcess(c, callbackCtx); err != nil {
		// 记录错误但不影响回调响应
		fmt.Printf("后置处理失败 [类型:%s]: %v\n", callbackReq.Type, err)
	}

	c.JSON(http.StatusOK, res)
}

// postCallbackProcess 后置处理逻辑
func (r *Functions) postCallbackProcess(c *gin.Context, callbackCtx *callback.CallbackContext) error {
	// 调用service层处理回调业务逻辑，传递context.Context而不是gin.Context
	return r.callbackService.ProcessCallback(c, callbackCtx)
}
