package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	resp "github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"io"
	"net/http"
)

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
	c.JSON(http.StatusOK, res)

}
