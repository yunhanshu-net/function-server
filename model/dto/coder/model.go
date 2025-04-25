package coder

import "strings"

// CodeApi API代码结构体
type CodeApi struct {
	Language       string `json:"language"`         // 编程语言
	Code           string `json:"code"`             // 代码内容
	Package        string `json:"package"`          // 包名
	AbsPackagePath string `json:"abs_package_path"` // 绝对包路径
	EnName         string `json:"en_name"`          // 英文名称
	CnName         string `json:"cn_name"`          // 中文名称
	Desc           string `json:"desc"`             // 描述
}

// CodeApiCreateInfo API创建信息
type CodeApiCreateInfo struct {
	Language       string `json:"language"`         // 编程语言
	Package        string `json:"package"`          // 包名
	AbsPackagePath string `json:"abs_package_path"` // 绝对包路径
	EnName         string `json:"en_name"`          // 英文名称
	CnName         string `json:"cn_name"`          // 中文名称
	Msg            string `json:"msg"`              // 消息
	Status         string `json:"status"`           // 状态
}

// GetFileSaveFullPath 获取文件保存的完整路径
func (c *CodeApi) GetFileSaveFullPath(sourceCodeDir string) (fullPath string, absFilePath string) {
	fullPath = strings.TrimSuffix(sourceCodeDir, "/") + "/api/" + strings.Trim(c.AbsPackagePath, "/")
	absFilePath = fullPath + "/" + c.GetFileName()
	return fullPath, absFilePath
}

// GetFileName 获取文件名
func (c *CodeApi) GetFileName() string {
	return c.EnName + "." + c.Language
}
