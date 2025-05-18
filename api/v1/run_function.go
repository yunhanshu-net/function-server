package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/api-server/model"
	"github.com/yunhanshu-net/api-server/pkg/db"
	"github.com/yunhanshu-net/api-server/pkg/dto/runcher"
	"github.com/yunhanshu-net/api-server/pkg/logger"
	"github.com/yunhanshu-net/api-server/pkg/response"
	"github.com/yunhanshu-net/api-server/service"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Functions struct {
	runcher service.RuncherService
	runner  *service.Runner
}

func NewFunctions(db *gorm.DB) *Functions {
	return &Functions{
		runcher: service.GetRuncherService(),
		runner:  service.NewRunner(db),
	}
}

func (r *Functions) Run(c *gin.Context) {

	req := &runcher.RunFunctionReq{
		User:   c.Param("user"),
		Method: c.Request.Method,
		Runner: c.Param("runner"),
		Router: c.Param("router"),
	}
	if req.Method == http.MethodGet {
		req.RawQuery = c.Request.URL.RawQuery
	} else {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}
		defer c.Request.Body.Close()
		req.Body = string(b)
	}
	rn, err := r.runner.GetByUserName(c, req.User, req.Runner)
	if err != nil {
		response.ParamError(c, fmt.Sprintf("获取runner失败：%s", err.Error()))
		return
	}
	req.Version = rn.Version
	log := model.FuncRunRecord{
		Base: model.Base{
			CreatedBy: c.GetString("user"),
			UpdatedBy: c.GetString("user"),
		},
		FuncId:  1,
		Request: json.RawMessage(req.Body),
		StartTs: time.Now().UnixMilli(),
	}
	function2, err := r.runcher.RunFunction2(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	log.EndTs = time.Now().UnixMilli()
	log.Cost = log.EndTs - log.StartTs

	get := c.Request.Header.Get("X-Function-ID")
	funcId, err := strconv.Atoi(get)
	if err != nil {
		logger.Errorf(c, "函数id获取失败！")
	}
	log.FuncId = int64(funcId)

	log.Response = function2.Data
	var res runcher.Body
	err = json.Unmarshal(function2.Data, &res)
	if err != nil {
		log.Status = "fail"
		log.Message = err.Error()
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
	log.Status = "success"
	go func() {
		marshal, err2 := json.Marshal(res)
		if err2 != nil {
			logger.Errorf(c, "<UNK>")
		} else {
			log.Response = marshal
		}
		db.GetDB().Model(&model.FuncRunRecord{}).Create(&log)
	}()
	c.JSON(http.StatusOK, res)
}
