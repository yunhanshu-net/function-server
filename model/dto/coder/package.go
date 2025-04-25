package coder

import (
	"strings"

	"github.com/yunhanshu-net/api-server/model/runner"
)

// BizPackage 业务包结构体
type BizPackage struct {
	Runner *runner.Runner `json:"runner"` // Runner信息

	AbsPackagePath string `json:"abs_package_path"` // 绝对包路径
	Language       string `json:"language"`         // 编程语言
	EnName         string `json:"en_name"`          // 英文名称
	CnName         string `json:"cn_name"`          // 中文名称
	Desc           string `json:"desc"`             // 描述
}

// GetPackageSaveFullPath 获取包保存的完整路径
func (c *BizPackage) GetPackageSaveFullPath(sourceCodeDir string) (savePath string, absPackagePath string) {
	savePath = strings.TrimSuffix(sourceCodeDir, "/") + "/api"
	absPackagePath = savePath + "/" + c.AbsPackagePath
	return savePath, absPackagePath
}

// GetPackageName 获取包名
func (c *BizPackage) GetPackageName() string {
	return c.EnName
}

// CreateProjectReq 创建项目请求结构
type CreateProjectReq struct {
	runner.Runner
}
