package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"github.com/yunhanshu-net/pkg/query"
)

type FunctionGen struct {
	functionGen *service.FunctionGen
}

func NewFunctionGen(functionGen *service.FunctionGen) *FunctionGen {
	return &FunctionGen{
		functionGen: functionGen,
	}
}

func (f *FunctionGen) List(c *gin.Context) {
	var req dto.FunctionGenListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	var list []model.FunctionGen
	table, err := query.AutoPaginateTable(c, db.GetDB(), &model.FunctionGen{}, &list, &req.PageInfoReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, table)
}

func (api *RunnerFuncAPI) FunctionGen(c *gin.Context) {
	var req dto.FunctionGenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	req.User = c.GetString("user")
	gen, err := api.service.FunctionGen(c, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gen)
}
func (api *RunnerFuncAPI) GeneratingList(c *gin.Context) {
	var req dto.FunctionGenListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	var list []model.FunctionGen
	table, err := query.AutoPaginateTable(c,
		db.GetDB().
			Where("runner_id = ?", req.RunnerID),
		&model.FunctionGen{}, &list, &req.PageInfoReq)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, table)
}

func (api *RunnerFuncAPI) GeneratingCount(c *gin.Context) {
	var req dto.GeneratingCount
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}
	var count int64
	db.GetDB().
		Model(&model.FunctionGen{}).
		Where("runner_id = ?", req.RunnerID).
		Count(&count)

	response.Success(c, map[string]int64{"count": count})
}
