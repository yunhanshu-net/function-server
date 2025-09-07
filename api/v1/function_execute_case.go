package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/model"
	"github.com/yunhanshu-net/function-server/pkg/db"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
)

// CreateFunctionExecuteCase 创建执行用例
func CreateFunctionExecuteCase(c *gin.Context) {
	var req model.FunctionExecuteCase
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	err = executeCase.Create(c, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, req)
}

// ExecFunctionExecuteCase 执行用例
func ExecFunctionExecuteCase(c *gin.Context) {
	var req dto.FunctionExecuteCaseReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	record, _, err := executeCase.Exec(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, record)
}

// BatchDeleteFunctionExecuteCase 批量删除执行用例
func BatchDeleteFunctionExecuteCase(c *gin.Context) {
	var req dto.BatchDeleteFunctionExecuteCaseReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	err = executeCase.BatchDelete(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": fmt.Sprintf("成功删除 %d 个执行用例", len(req.CaseIds))})
}

// UpdateFunctionExecuteCase 更新执行用例
func UpdateFunctionExecuteCase(c *gin.Context) {
	var req dto.UpdateFunctionExecuteCaseReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	err = executeCase.Update(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "更新执行用例成功"})
}

// QueryFunctionExecuteCase 查询执行用例
func QueryFunctionExecuteCase(c *gin.Context) {
	var req dto.QueryFunctionExecuteCaseReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	resp, err := executeCase.Query(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// QueryFunctionExecuteCaseRecords 查询执行记录
func QueryFunctionExecuteCaseRecords(c *gin.Context) {
	var req dto.QueryFunctionExecuteCaseRecordReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	resp, err := executeCase.QueryRecords(c, req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// GetFunctionExecuteCase 根据ID获取执行用例详情
// @Summary 获取执行用例详情
// @Description 根据用例ID获取执行用例的详细信息，包括关联的Runner和Function信息
// @Tags 执行用例管理
// @Accept json
// @Produce json
// @Param id query int true "执行用例ID"
// @Success 200 {object} response.Response{data=dto.FunctionExecuteCaseWithDetails} "获取成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "用例不存在"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/function-execute-case/get [get]
func GetFunctionExecuteCase(c *gin.Context) {
	// 从查询参数中获取用例ID
	caseIdStr := c.Query("id")
	if caseIdStr == "" {
		response.ParamError(c, "用例ID不能为空")
		return
	}

	// 解析用例ID
	var caseId int
	if _, err := fmt.Sscanf(caseIdStr, "%d", &caseId); err != nil {
		response.ParamError(c, "无效的用例ID格式")
		return
	}

	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	result, err := executeCase.Get(c, caseId)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// InitExecCaseCount 初始化所有函数的exec_case_count字段
// @Summary 初始化执行用例计数
// @Description 为所有函数重新统计并更新exec_case_count字段，用于修复存量数据
// @Tags 执行用例管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string} "初始化成功"
// @Failure 500 {object} response.Response "服务器错误"
// @Router /api/v1/function-execute-case/init-count [post]
func InitExecCaseCount(c *gin.Context) {
	executeCase := service.NewFunctionExecuteCase(db.GetDB())
	err := executeCase.InitExecCaseCount(c)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, "初始化执行用例计数成功")
}
