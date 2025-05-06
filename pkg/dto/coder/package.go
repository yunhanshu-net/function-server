package coder

import (
	"github.com/yunhanshu-net/api-server/pkg/dto/syscallback"
	"strings"
)

type Runner struct {
	Kind     string `json:"kind"`     //类型，可执行程序，so文件等等
	Language string `json:"language"` //编程语言
	Name     string `json:"name"`     //应用名称（英文标识）
	Version  string `json:"version"`  //应用版本
	User     string `json:"user"`     //所属租户
}

type BizPackage struct {
	Runner         *Runner `json:"runner"`
	AbsPackagePath string  `json:"abs_package_path"`
	Language       string  `json:"language"`
	EnName         string  `json:"en_name"`
	CnName         string  `json:"cn_name"`
	Desc           string  `json:"desc"`
}

func (c *BizPackage) GetPackageSaveFullPath(sourceCodeDir string) (savePath string, absPackagePath string) {
	savePath = strings.TrimSuffix(sourceCodeDir, "/") + "/api"
	absPackagePath = savePath + "/" + c.AbsPackagePath
	return savePath, absPackagePath
}

func (c *BizPackage) GetPackageName() string {
	return c.EnName
}

type CreateProjectReq struct {
	Runner
}
type CreateProjectResp struct {
	Version string `json:"version"`
}

type AddApisResp struct {
	Version              string               `json:"version"`
	ErrList              []*CodeApiCreateInfo `json:"err_list"`
	SyscallChangeVersion *syscallback.SysOnVersionChangeResp
}

type AddApiResp struct {
	Version              string                              `json:"version"`
	Data                 interface{}                         `json:"data"`
	SyscallChangeVersion *syscallback.SysOnVersionChangeResp `json:"syscall_change_version"`
}

type BizPackageResp struct {
	Version string `json:"version"`
}
