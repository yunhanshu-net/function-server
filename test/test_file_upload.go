package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// GetUploadTokenResponse 获取上传token响应
type GetUploadTokenResponse struct {
	Code int `json:"code"`
	Data struct {
		Token  string `json:"token"`
		Bucket string `json:"bucket"`
		Domain string `json:"domain"`
	} `json:"data"`
	Message string `json:"message"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	Code int `json:"code"`
	Data struct {
		Key  string `json:"key"`
		URL  string `json:"url"`
		Hash string `json:"hash"`
		Size int64  `json:"size"`
		Name string `json:"name"`
	} `json:"data"`
	Message string `json:"message"`
}

func main() {
	baseURL := "http://localhost:8080"
	
	// 1. 测试获取上传token
	fmt.Println("=== 测试获取上传token ===")
	tokenResp, err := getUploadToken(baseURL)
	if err != nil {
		fmt.Printf("获取上传token失败: %v\n", err)
		return
	}
	fmt.Printf("获取上传token成功: %+v\n", tokenResp.Data)
	
	// 2. 创建测试文件
	fmt.Println("\n=== 创建测试文件 ===")
	testFile := "test_upload.txt"
	testContent := "这是一个测试文件，用于验证文件上传功能。\n当前时间: " + fmt.Sprintf("%v", os.Getenv("USER"))
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		return
	}
	defer os.Remove(testFile) // 清理测试文件
	fmt.Printf("创建测试文件成功: %s\n", testFile)
	
	// 3. 测试文件上传
	fmt.Println("\n=== 测试文件上传 ===")
	uploadResp, err := uploadFile(baseURL, testFile, "test-uploads")
	if err != nil {
		fmt.Printf("文件上传失败: %v\n", err)
		return
	}
	fmt.Printf("文件上传成功: %+v\n", uploadResp.Data)
	
	// 4. 验证文件URL是否可访问
	fmt.Println("\n=== 验证文件URL ===")
	if uploadResp.Data.URL != "" {
		resp, err := http.Get(uploadResp.Data.URL)
		if err != nil {
			fmt.Printf("访问文件URL失败: %v\n", err)
		} else {
			defer resp.Body.Close()
			fmt.Printf("文件URL访问成功，状态码: %d\n", resp.StatusCode)
			if resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("文件内容预览: %s\n", string(body)[:min(100, len(body))])
			}
		}
	}
	
	fmt.Println("\n=== 测试完成 ===")
}

// getUploadToken 获取上传token
func getUploadToken(baseURL string) (*GetUploadTokenResponse, error) {
	url := baseURL + "/api/v1/file/upload-token"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// 添加用户信息头（模拟中间件设置）
	req.Header.Set("X-User", "test_user")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var tokenResp GetUploadTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}
	
	if tokenResp.Code != 200 {
		return nil, fmt.Errorf("获取token失败: %s", tokenResp.Message)
	}
	
	return &tokenResp, nil
}

// uploadFile 上传文件
func uploadFile(baseURL, filePath, pathPrefix string) (*UploadFileResponse, error) {
	url := baseURL + "/api/v1/file/upload"
	
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	
	// 添加文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	
	// 添加路径字段
	if pathPrefix != "" {
		err = writer.WriteField("path", pathPrefix)
		if err != nil {
			return nil, err
		}
	}
	
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User", "test_user") // 模拟用户信息
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var uploadResp UploadFileResponse
	err = json.Unmarshal(body, &uploadResp)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}
	
	if uploadResp.Code != 200 {
		return nil, fmt.Errorf("上传文件失败: %s", uploadResp.Message)
	}
	
	return &uploadResp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 