# QQ邮件发送Caller

这是一个基于QQ邮箱SMTP服务的邮件发送功能模块，集成到function-server的caller系统中。

## 功能特性

- ✅ 支持QQ邮箱SMTP发送
- ✅ 支持纯文本和HTML格式邮件
- ✅ 支持多收件人
- ✅ 支持TLS加密连接
- ✅ 完整的错误处理
- ✅ RESTful API接口
- ✅ 配置管理功能

## 快速开始

### 1. 获取QQ邮箱授权码

1. 登录QQ邮箱网页版
2. 进入"设置" -> "账户"
3. 找到"POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务"
4. 开启"IMAP/SMTP服务"
5. 按照提示获取授权码（不是QQ密码）

### 2. 配置QQ邮箱

```bash
curl -X POST http://localhost:8080/api/v1/qq-email/config \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your_qq_email@qq.com",
    "password": "your_authorization_code",
    "from_email": "your_qq_email@qq.com",
    "from_name": "Your Name"
  }'
```

### 3. 发送邮件

```bash
curl -X POST http://localhost:8080/api/v1/qq-email/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["recipient@example.com"],
    "subject": "测试邮件",
    "body": "这是一封测试邮件",
    "is_html": false
  }'
```

## API接口

### 设置配置

**POST** `/api/v1/qq-email/config`

```json
{
  "username": "your_qq_email@qq.com",
  "password": "your_authorization_code", 
  "from_email": "your_qq_email@qq.com",
  "from_name": "Your Name"
}
```

### 发送邮件

**POST** `/api/v1/qq-email/send`

```json
{
  "to": ["recipient1@example.com", "recipient2@example.com"],
  "subject": "邮件主题",
  "body": "邮件正文",
  "is_html": false
}
```

### 获取配置

**GET** `/api/v1/qq-email/config`

## 代码集成

### 在代码中使用

```go
// 创建函数调用服务
functionCallService := service.NewFunctionCallService()

// 设置QQ邮箱配置
err := functionCallService.SetQQEmailConfig(
    "your_email@qq.com",
    "your_auth_code",
    "your_email@qq.com", 
    "Your Name",
)

// 发送邮件
emailData := map[string]interface{}{
    "to":      []string{"recipient@example.com"},
    "subject": "测试邮件",
    "body":    "邮件内容",
    "is_html": false,
}

emailBody, _ := json.Marshal(emailData)
callReq := &interfaces.FunctionCallReq{
    Name: "qq_email",
    Body: emailBody,
}

resp, err := functionCallService.CreateFunctionCall(callReq)
```

### 直接使用Caller

```go
// 创建QQ邮件caller
qqCaller := caller.NewQQEmailCaller()

// 设置配置
qqCaller.SetConfig(
    "your_email@qq.com",
    "your_auth_code", 
    "your_email@qq.com",
    "Your Name",
)

// 发送邮件
emailReq := map[string]interface{}{
    "to":      []string{"recipient@example.com"},
    "subject": "测试邮件",
    "body":    "邮件内容",
    "is_html": false,
}

reqBody, _ := json.Marshal(emailReq)
req := &interfaces.FunctionCallReq{
    Name: "qq_email",
    Body: reqBody,
}

resp, err := qqCaller.Call(req)
```

## 技术细节

### SMTP配置

- **服务器**: smtp.qq.com
- **端口**: 587
- **加密**: STARTTLS
- **认证**: PLAIN

### 安全注意事项

1. 使用授权码而不是QQ密码
2. 不要在代码中硬编码敏感信息
3. 建议通过环境变量管理配置
4. 定期更换授权码

### 错误处理

所有API都会返回统一的错误格式：

```json
{
  "code": 400,
  "message": "错误描述",
  "data": {
    "success": false,
    "message": "详细错误信息",
    "error": "具体错误原因"
  }
}
```

## 测试

运行测试：

```bash
cd function-server
go test ./test/qq_email_test.go -v
```

## 故障排除

### 常见问题

1. **SMTP认证失败**
   - 确认使用的是授权码而不是QQ密码
   - 检查邮箱账号是否正确

2. **连接超时**
   - 检查网络连接
   - 确认SMTP服务器地址和端口

3. **邮件发送失败**
   - 检查收件人邮箱格式
   - 确认邮件内容符合规范

### 调试模式

可以通过日志查看详细的SMTP交互过程：

```go
// 在代码中添加日志
logger.Debugf(ctx, "SMTP连接信息: %+v", config)
```

## 扩展功能

### 支持其他邮箱服务

可以基于相同的接口实现其他邮箱服务的caller：

- Gmail SMTP
- Outlook SMTP  
- 企业邮箱SMTP

### 批量发送

支持批量发送邮件，只需在`to`字段中提供多个收件人：

```json
{
  "to": [
    "user1@example.com",
    "user2@example.com", 
    "user3@example.com"
  ],
  "subject": "群发邮件",
  "body": "这是群发邮件内容"
}
```

## 许可证

本项目遵循MIT许可证。
