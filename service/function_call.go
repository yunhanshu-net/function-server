package service

import (
	"fmt"

	"github.com/yunhanshu-net/function-server/pkg/interfaces"
	emailcaller "github.com/yunhanshu-net/function-server/service/caller"
)

type FunctionCallService struct {
	callers map[string]interfaces.Caller
}

func NewFunctionCallService() *FunctionCallService {
	service := &FunctionCallService{
		callers: make(map[string]interfaces.Caller),
	}

	// 注册QQ邮件发送caller
	qqEmailCaller := emailcaller.NewQQEmailCaller()
	service.callers["qq_email"] = qqEmailCaller

	return service
}

func (service *FunctionCallService) CreateFunctionCall(call *interfaces.FunctionCallReq) (*interfaces.FunctionCallResp, error) {
	caller, ok := service.callers[call.Name]
	if !ok {
		return nil, fmt.Errorf("not found caller")
	}
	rsp, err := caller.Call(call)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

// SetQQEmailConfig 设置QQ邮箱配置
func (service *FunctionCallService) SetQQEmailConfig(username, password, fromEmail, fromName string) error {
	caller, ok := service.callers["qq_email"]
	if !ok {
		return fmt.Errorf("QQ邮件caller未注册")
	}

	qqEmailCaller, ok := caller.(*emailcaller.QQEmailCaller)
	if !ok {
		return fmt.Errorf("caller类型错误")
	}

	qqEmailCaller.SetConfig(username, password, fromEmail, fromName)
	return nil
}

// GetQQEmailConfig 获取QQ邮箱配置
func (service *FunctionCallService) GetQQEmailConfig() (map[string]string, error) {
	caller, ok := service.callers["qq_email"]
	if !ok {
		return nil, fmt.Errorf("QQ邮件caller未注册")
	}

	qqEmailCaller, ok := caller.(*emailcaller.QQEmailCaller)
	if !ok {
		return nil, fmt.Errorf("caller类型错误")
	}

	return qqEmailCaller.GetConfig(), nil
}
