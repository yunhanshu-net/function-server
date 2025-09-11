package v1

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-server/pkg/dto"
	"github.com/yunhanshu-net/function-server/pkg/response"
	"github.com/yunhanshu-net/function-server/service"
	"github.com/yunhanshu-net/pkg/logger"
	"go.uber.org/zap"
)

// RunnerFuncConfig API处理器
type RunnerFuncConfig struct {
	runnerFuncService *service.RunnerFunc
}

// NewRunnerFuncConfig 创建配置API处理器
func NewRunnerFuncConfig(runnerFuncService *service.RunnerFunc) *RunnerFuncConfig {
	return &RunnerFuncConfig{
		runnerFuncService: runnerFuncService,
	}
}

// GetFuncConfig 获取函数配置
// @Summary 获取函数配置
// @Description 根据函数ID获取函数配置信息
// @Tags 函数配置
// @Accept json
// @Produce json
// @Param func_id query int true "函数ID"
// @Success 200 {object} dto.GetFuncConfigResp
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/runner-func/config [get]
func (r *RunnerFuncConfig) GetFuncConfig(c *gin.Context) {
	// 参数验证
	var req dto.GetFuncConfigReq
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(c, "参数验证失败", err, zap.String("user", c.GetString("user")))
		response.ParamError(c, err.Error())
		return
	}

	// 获取函数配置
	funcConfig, err := r.runnerFuncService.GetFuncConfig(c, req.FuncID)
	if err != nil {
		logger.Error(c, "获取函数配置失败", err, zap.Int64("func_id", req.FuncID))
		response.ServerError(c, "获取函数配置失败: "+err.Error())
		return
	}

	if funcConfig == nil {
		response.NotFound(c, "函数配置不存在")
		return
	}

	// 解析配置结构体
	var configStruct interface{}
	if funcConfig.ConfigStruct != nil {
		if err := json.Unmarshal(funcConfig.ConfigStruct, &configStruct); err != nil {
			logger.Errorf(c, "解析配置结构体失败: %v, func_id: %d", err, req.FuncID)
		}
	}

	// 解析配置数据
	var configData interface{}
	if funcConfig.ConfigData != nil {
		if err := json.Unmarshal(funcConfig.ConfigData, &configData); err != nil {
			logger.Errorf(c, "解析配置数据失败: %v, func_id: %d", err, req.FuncID)
		}
	}

	// 构建响应
	resp := &dto.GetFuncConfigResp{
		FuncID:       funcConfig.FuncID,
		ConfigKey:    funcConfig.ConfigKey,
		ConfigType:   funcConfig.ConfigType,
		ConfigStruct: configStruct,
		ConfigData:   configData,
		Description:  funcConfig.Description,
		Version:      funcConfig.Version,
		IsActive:     funcConfig.IsActive,
	}

	response.Success(c, resp)
}

// UpdateFuncConfig 更新函数配置
// @Summary 更新函数配置
// @Description 根据函数ID更新函数配置数据
// @Tags 函数配置
// @Accept json
// @Produce json
// @Param request body dto.UpdateFuncConfigReq true "更新配置请求"
// @Success 200 {object} dto.UpdateFuncConfigResp
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/runner-func/config [put]
func (r *RunnerFuncConfig) UpdateFuncConfig(c *gin.Context) {
	// 参数验证
	var req dto.UpdateFuncConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error(c, "参数验证失败", err, zap.String("user", c.GetString("user")))
		response.ParamError(c, err.Error())
		return
	}

	// 更新函数配置
	err := r.runnerFuncService.UpdateFuncConfig(c, req.FuncID, req.ConfigData)
	if err != nil {
		logger.Error(c, "更新函数配置失败", err, zap.Int64("func_id", req.FuncID))
		response.ServerError(c, "更新函数配置失败: "+err.Error())
		return
	}

	// 构建响应
	resp := &dto.UpdateFuncConfigResp{
		Success: true,
		Message: "配置更新成功",
	}

	response.Success(c, resp)
}
