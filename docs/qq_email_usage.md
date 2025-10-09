# QQ邮件发送Caller使用指南

## 概述

QQ邮件发送Caller是一个用于通过QQ邮箱SMTP服务发送邮件的功能模块。它支持发送纯文本和HTML格式的邮件，支持多个收件人。

## 配置

### 1. 获取QQ邮箱授权码

1. 登录QQ邮箱网页版
2. 进入"设置" -> "账户"
3. 找到"POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务"
4. 开启"IMAP/SMTP服务"
5. 按照提示获取授权码（不是QQ密码）

### 2. 设置配置

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

## API接口

### 1. 设置QQ邮箱配置

**POST** `/api/v1/qq-email/config`

**请求参数：**
```json
{
  "username": "your_qq_email@qq.com",    // QQ邮箱账号
  "password": "your_authorization_code", // QQ邮箱授权码
  "from_email": "your_qq_email@qq.com",  // 发件人邮箱
  "from_name": "Your Name"               // 发件人名称（可选）
}
```

**响应：**
```json
{
  "code": 0,
  "message": "QQ邮箱配置设置成功",
  "data": null
}
```

### 2. 发送邮件

**POST** `/api/v1/qq-email/send`

**请求参数：**
```json
{
  "to": ["recipient1@example.com", "recipient2@example.com"], // 收件人邮箱列表
  "subject": "邮件主题",                                       // 邮件主题
  "body": "邮件正文内容",                                      // 邮件正文
  "is_html": false                                           // 是否为HTML格式
}
```

**响应：**
```json
{
  "code": 0,
  "message": "邮件发送成功，收件人：recipient1@example.com, recipient2@example.com",
  "data": {
    "success": true,
    "message": "邮件发送成功，收件人：recipient1@example.com, recipient2@example.com",
    "error": ""
  }
}
```

### 3. 获取当前配置

**GET** `/api/v1/qq-email/config`

**响应：**
```json
{
  "code": 0,
  "message": "获取配置成功",
  "data": {
    "smtp_host": "smtp.qq.com",
    "smtp_port": "587",
    "username": "your_qq_email@qq.com",
    "from_email": "your_qq_email@qq.com",
    "from_name": "Your Name",
    "password": "***"
  }
}
```

## 使用示例

### 发送纯文本邮件

```bash
curl -X POST http://localhost:8080/api/v1/qq-email/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["test@example.com"],
    "subject": "测试邮件",
    "body": "这是一封测试邮件，请忽略。",
    "is_html": false
  }'
```

### 发送HTML邮件

```bash
curl -X POST http://localhost:8080/api/v1/qq-email/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["test@example.com"],
    "subject": "HTML测试邮件",
    "body": "<h1>标题</h1><p>这是一封<strong>HTML格式</strong>的测试邮件。</p>",
    "is_html": true
  }'
```

### 发送给多个收件人

```bash
curl -X POST http://localhost:8080/api/v1/qq-email/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["user1@example.com", "user2@example.com", "user3@example.com"],
    "subject": "群发邮件",
    "body": "这是一封群发邮件。",
    "is_html": false
  }'
```

## 错误处理

### 常见错误

1. **配置错误**
   ```json
   {
     "code": 500,
     "message": "设置配置失败: QQ邮件caller未注册"
   }
   ```

2. **参数错误**
   ```json
   {
     "code": 400,
     "message": "收件人不能为空"
   }
   ```

3. **SMTP认证失败**
   ```json
   {
     "code": 0,
     "message": "邮件发送失败",
     "data": {
       "success": false,
       "message": "邮件发送失败",
       "error": "SMTP认证失败: 535 Authentication failed"
     }
   }
   ```

## 技术细节

### SMTP配置

- **服务器地址**: smtp.qq.com
- **端口**: 587
- **加密方式**: STARTTLS
- **认证方式**: PLAIN

### 安全注意事项

1. 使用授权码而不是QQ密码
2. 不要在代码中硬编码邮箱密码
3. 建议通过环境变量或配置文件管理敏感信息
4. 定期更换授权码

### 限制说明

1. QQ邮箱有发送频率限制
2. 单次发送的收件人数量有限制
3. 邮件内容需要符合相关法律法规

## 集成到现有系统

### 在代码中调用

```go
// 设置配置
err := functionCallService.SetQQEmailConfig(
    "your_email@qq.com",
    "your_auth_code", 
    "your_email@qq.com",
    "Your Name",
)

// 发送邮件
callReq := &service.FunctionCallReq{
    Name: "qq_email",
    Body: []byte(`{
        "to": ["test@example.com"],
        "subject": "测试邮件",
        "body": "邮件内容",
        "is_html": false
    }`),
}

resp, err := functionCallService.CreateFunctionCall(callReq)
```

## 故障排除

1. **无法连接SMTP服务器**
   - 检查网络连接
   - 确认SMTP服务器地址和端口正确

2. **认证失败**
   - 确认使用的是授权码而不是QQ密码
   - 检查邮箱账号是否正确

3. **邮件发送失败**
   - 检查收件人邮箱地址格式
   - 确认邮件内容符合规范
   - 检查是否触发了发送频率限制
