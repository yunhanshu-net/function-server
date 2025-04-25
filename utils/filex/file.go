package filex

import (
	"os"
	"path/filepath"
)

// PathExists 检查文件或目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// EnsureDir 确保目录存在，不存在则创建
func EnsureDir(dir string) error {
	exists, err := PathExists(dir)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// WriteFile 写入文件内容
func WriteFile(path string, content string) error {
	// 确保父目录存在
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFile 读取文件内容
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
