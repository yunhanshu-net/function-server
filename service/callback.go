package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/function-server/pkg/dto/callback"
	"github.com/yunhanshu-net/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CallbackService 回调服务
type CallbackService struct {
	runnerFuncService *RunnerFunc
}

// NewCallbackService 创建回调服务
func NewCallbackService(db *gorm.DB) *CallbackService {
	return &CallbackService{
		runnerFuncService: NewRunnerFunc(db),
	}
}

// ProcessCallback 处理回调业务逻辑
func (s *CallbackService) ProcessCallback(ctx context.Context, callbackCtx *callback.CallbackContext) error {
	switch callbackCtx.Type {
	case "OnUpdateConfig":
		return s.handleConfigUpdate(ctx, callbackCtx)
	case "OnGetConfig":
		return s.handleConfigGet(ctx, callbackCtx)
	// 可以继续添加其他回调类型的处理
	default:
		// 未知的回调类型，不做特殊处理
		return nil
	}
}

// handleConfigUpdate 处理配置更新业务逻辑
func (s *CallbackService) handleConfigUpdate(ctx context.Context, callbackCtx *callback.CallbackContext) error {
	// 检查回调是否成功
	if callbackCtx.Response.Code != 0 {
		return nil
	}

	// 检查是否有函数ID
	if callbackCtx.FuncID == 0 {
		return fmt.Errorf("缺少函数ID")
	}

	// 解析配置数据
	updateConfigBody, err := s.parseUpdateConfigBody(callbackCtx.Body)
	if err != nil {
		return fmt.Errorf("解析配置更新数据失败: %w", err)
	}

	// 更新数据库中的配置
	if err := s.updateConfigInDB(ctx, callbackCtx.FuncID, updateConfigBody.ConfigData); err != nil {
		return fmt.Errorf("更新数据库配置失败: %w", err)
	}

	logger.Info(ctx, "配置更新成功",
		zap.Int64("func_id", callbackCtx.FuncID),
		zap.String("router", updateConfigBody.Router),
		zap.Any("config_data", updateConfigBody.ConfigData),
	)
	return nil
}

// handleConfigGet 处理配置获取业务逻辑
func (s *CallbackService) handleConfigGet(ctx context.Context, callbackCtx *callback.CallbackContext) error {
	// 解析配置获取数据
	getConfigBody, err := s.parseGetConfigBody(callbackCtx.Body)
	if err != nil {
		return fmt.Errorf("解析配置获取数据失败: %w", err)
	}

	// 配置获取的业务逻辑
	logger.Info(ctx, "配置获取回调",
		zap.String("type", callbackCtx.Type),
		zap.String("router", getConfigBody.Router),
		zap.String("config_key", getConfigBody.ConfigKey),
	)
	return nil
}

// parseUpdateConfigBody 解析配置更新body
func (s *CallbackService) parseUpdateConfigBody(body interface{}) (*callback.UpdateConfigBody, error) {
	var updateConfigBody callback.UpdateConfigBody

	if bodyBytes, ok := body.([]byte); ok {
		if err := json.Unmarshal(bodyBytes, &updateConfigBody); err != nil {
			return nil, fmt.Errorf("解析配置更新数据失败: %w", err)
		}
	} else if bodyMap, ok := body.(map[string]interface{}); ok {
		// 将map转换为JSON，再解析为结构体
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("序列化body数据失败: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, &updateConfigBody); err != nil {
			return nil, fmt.Errorf("解析配置更新数据失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("不支持的body数据类型")
	}

	return &updateConfigBody, nil
}

// parseGetConfigBody 解析配置获取body
func (s *CallbackService) parseGetConfigBody(body interface{}) (*callback.GetConfigBody, error) {
	var getConfigBody callback.GetConfigBody

	if bodyBytes, ok := body.([]byte); ok {
		if err := json.Unmarshal(bodyBytes, &getConfigBody); err != nil {
			return nil, fmt.Errorf("解析配置获取数据失败: %w", err)
		}
	} else if bodyMap, ok := body.(map[string]interface{}); ok {
		// 将map转换为JSON，再解析为结构体
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("序列化body数据失败: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, &getConfigBody); err != nil {
			return nil, fmt.Errorf("解析配置获取数据失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("不支持的body数据类型")
	}

	return &getConfigBody, nil
}

// updateConfigInDB 更新数据库中的配置
func (s *CallbackService) updateConfigInDB(ctx context.Context, funcID int64, configData map[string]interface{}) error {
	// 直接使用动态的配置数据更新数据库
	return s.runnerFuncService.UpdateFuncConfig(ctx, funcID, configData)
}
