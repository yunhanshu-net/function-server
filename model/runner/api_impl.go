package runner

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/api-server/model/dto/coder"
	"github.com/yunhanshu-net/api-server/utils/filex"
)

// AddApi 添加API的完整实现
// 此方法处理API添加的所有逻辑，包括：
// 1. 验证API参数
// 2. 创建必要的目录结构
// 3. 保存API代码到文件
// 4. 更新版本号
// 5. 返回API添加结果
func (r *Runner) AddApi(codeApi interface{}) (*coder.AddApiResp, error) {
	// 1. 参数验证和类型转换
	api, ok := codeApi.(*coder.CodeApi)
	if !ok {
		return nil, errors.New("无效的API参数类型")
	}

	// 校验必填字段
	if api.EnName == "" {
		return nil, errors.New("API英文名称不能为空")
	}
	if api.Language == "" {
		return nil, errors.New("API语言不能为空")
	}
	if api.Code == "" {
		return nil, errors.New("API代码不能为空")
	}

	// 2. 确定源码根目录
	sourceCodeRoot := r.getSourceCodeRoot()
	if err := filex.EnsureDir(sourceCodeRoot); err != nil {
		return nil, fmt.Errorf("创建源码目录失败: %w", err)
	}

	// 3. 创建新版本目录
	nextVersion := r.getNextVersion()
	versionPath := filepath.Join(sourceCodeRoot, nextVersion)
	if err := filex.EnsureDir(versionPath); err != nil {
		return nil, fmt.Errorf("创建版本目录失败: %w", err)
	}

	// 4. 确定API存储路径并创建目录
	apiDir, apiFilePath := api.GetFileSaveFullPath(versionPath)
	if err := filex.EnsureDir(apiDir); err != nil {
		return nil, fmt.Errorf("创建API目录失败: %w", err)
	}

	// 5. 将API代码写入文件
	if err := filex.WriteFile(apiFilePath, api.Code); err != nil {
		return nil, fmt.Errorf("写入API代码失败: %w", err)
	}
	logrus.Infof("API代码已保存到: %s", apiFilePath)

	// 6. 更新Runner版本信息
	r.CurrentVersion = nextVersion
	r.VersionCount++
	r.APICount++
	r.UpdatedAt = time.Now()

	// 7. 保存Runner更新到数据库
	// 这里应该有数据库更新代码，但作为示例先省略

	// 8. 返回结果
	return &coder.AddApiResp{
		Version: nextVersion,
		ID:      r.ID,
		Name:    api.EnName,
		Status:  "success",
	}, nil
}

// 获取源码根目录
func (r *Runner) getSourceCodeRoot() string {
	// 这里可以根据实际项目配置调整路径构建逻辑
	return fmt.Sprintf("./storage/runners/%s/%s", r.User, r.Name)
}

// 获取下一个版本号
func (r *Runner) getNextVersion() string {
	// 如果当前没有版本，返回v1
	if r.CurrentVersion == "" {
		return "v1"
	}

	// 解析当前版本号
	version := strings.TrimPrefix(r.CurrentVersion, "v")
	versionNum := 1
	fmt.Sscanf(version, "%d", &versionNum)

	// 返回下一个版本号
	return fmt.Sprintf("v%d", versionNum+1)
}
