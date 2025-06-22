# API 测试示例

## 文件上传Token接口测试

### 基本请求
```bash
curl -X GET "http://localhost:8080/api/v1/file/upload-token" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json"
```

### 带路径前缀的请求
```bash
curl -X GET "http://localhost:8080/api/v1/file/upload-token?path_prefix=images" \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json"
```

### 预期响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
    "domain": "http://cdn.geeleo.com",
    "bucket": "geeleo",
    "path_prefix": "user123/images",
    "expires_at": 1703127456,
    "region": "z0"
  }
}
```

### 响应字段说明
- `token`: 七牛云上传token，有效期1小时
- `domain`: CDN域名，用于构建文件访问URL
- `bucket`: 存储桶名称
- `path_prefix`: 文件路径前缀（包含用户ID和指定前缀）
- `expires_at`: token过期时间戳
- `region`: 存储区域（z0=华东，z1=华北，z2=华南等）

### 前端使用示例
```javascript
// 获取上传配置
const response = await fetch('/api/v1/file/upload-token?path_prefix=uploads')
const result = await response.json()

if (result.code === 200) {
  const config = result.data
  console.log('上传配置:', config)
  
  // 使用配置进行文件上传
  // ...
}
``` 