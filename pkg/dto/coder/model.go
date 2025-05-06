package coder

import "strings"

type CodeApi struct {
	Language       string `json:"language"`
	Code           string `json:"code"`
	Package        string `json:"package"`
	AbsPackagePath string `json:"abs_package_path"`
	//FilePath       string `json:"file_path"`
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`
	Desc   string `json:"desc"`
}

type CodeApiCreateInfo struct {
	Language       string `json:"language"`
	Package        string `json:"package"`
	AbsPackagePath string `json:"abs_package_path"`
	//FilePath       string `json:"file_path"`
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`

	Msg    string `json:"msg"`
	Status string `json:"status"`
}

func (c *CodeApi) GetFileSaveFullPath(sourceCodeDir string) (fullPath string, absFilePath string) {
	fullPath = strings.TrimSuffix(sourceCodeDir, "/") + "/api/" + strings.Trim(c.AbsPackagePath, "/")
	absFilePath = fullPath + "/" + c.GetFileName()
	return fullPath, absFilePath
}

func (c *CodeApi) GetFileName() string {
	return c.EnName + "." + c.Language
}
