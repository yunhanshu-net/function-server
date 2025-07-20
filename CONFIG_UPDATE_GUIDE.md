# 配置更新功能使用指南

## 概述

本功能允许用户通过界面更新函数的配置文件，实现了从 function-server 到 function-go 的完整配置更新链路。

## 架构设计

```
前端界面
    ↓
function-server (API网关)
    ↓ NATS消息
function-runtime (运行时管理)
    ↓ 命令行调用
function-go (函数执行器)
    ↓ 配置管理器
配置文件存储
```

## API 接口

### 1. 更新Runner配置

**接口地址**: `PUT /api/v1/runner-config`

**请求参数**:
```json
{
  "user": "用户名",
  "runner": "Runner名称",
  "version": "Runner版本",
  "router": "路由路径",
  "method": "HTTP方法",
  "config_data": {
    "api_key": "new_api_key_123",
    "timeout": 30,
    "retry_count": 3,
    "debug_mode": true
  }
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "配置更新成功"
}
```

### 2. 获取Runner配置

**接口地址**: `GET /api/v1/runner-config`

**请求参数**:
```
?user=用户名&runner=Runner名称&version=Runner版本&router=路由路径&method=HTTP方法
```

**响应示例**:
```json
{
  "success": true,
  "config": {
    "type": "json",
    "data": {
      "api_key": "new_api_key_123",
      "timeout": 30,
      "retry_count": 3,
      "debug_mode": true
    }
  }
}
```

## 使用示例

### 1. 更新配置

```bash
curl -X PUT http://localhost:8080/api/v1/runner-config \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "user": "testuser",
    "runner": "testrunner",
    "version": "v1",
    "router": "/api/test",
    "method": "POST",
    "config_data": {
      "api_key": "new_api_key_123",
      "timeout": 30,
      "retry_count": 3,
      "debug_mode": true
    }
  }'
```

### 2. 获取配置

```bash
curl -X GET "http://localhost:8080/api/v1/runner-config?user=testuser&runner=testrunner&version=v1&router=/api/test&method=POST" \
  -H "Authorization: Bearer your-token"
```

## 实现细节

### 1. function-server 层

- **API接口**: `function-server/api/v1/runner_config.go`
- **服务层**: `function-server/service/runcher.go`
- **DTO定义**: `function-server/pkg/dto/runner_func_config.go`

### 2. function-runtime 层

- **NATS消息处理**: `function-runtime/kernel/scheduler/config_manage.go`
- **命令行调用**: 通过命令行调用 function-go 的配置更新方法

### 3. function-go 层

- **命令行接口**: `function-go/runner/syscall.go`
- **配置管理**: `function-go/runner/default.go`
- **配置存储**: 使用配置管理器进行持久化存储

## 配置键生成规则

配置键按照以下规则生成：
```
function.{router}.{method}
```

例如：
- 路由: `/api/test`
- 方法: `POST`
- 配置键: `function.api.test.post`

## 安全考虑

1. **用户权限验证**: 确保用户只能更新自己的配置
2. **参数验证**: 对所有输入参数进行严格验证
3. **错误处理**: 完善的错误处理和日志记录
4. **数据隔离**: 通过用户字段实现多租户隔离

## 错误处理

### 常见错误码

- `400`: 参数验证失败
- `401`: 用户未登录
- `403`: 权限不足
- `500`: 服务器内部错误

### 错误响应格式

```json
{
  "success": false,
  "error": "错误描述信息"
}
```

## 测试

### 运行测试

```bash
cd function-server
go run test_config_update_flow.go
```

### 测试内容

1. **配置更新测试**: 验证配置更新功能
2. **配置获取测试**: 验证配置获取功能
3. **API接口测试**: 验证API接口的正确性

## 部署说明

### 1. 启动服务

```bash
# 启动 function-server
cd function-server
go run main.go

# 启动 function-runtime
cd function-runtime
go run main.go

# 启动 function-go (作为Runner)
cd function-go
go run main.go
```

### 2. 配置NATS

确保所有服务都连接到同一个NATS服务器：
```
nats://localhost:4222
```

### 3. 验证部署

```bash
# 测试配置更新
curl -X PUT http://localhost:8080/api/v1/runner-config \
  -H "Content-Type: application/json" \
  -d '{
    "user": "testuser",
    "runner": "testrunner",
    "version": "v1",
    "router": "/api/test",
    "method": "POST",
    "config_data": {
      "test_key": "test_value"
    }
  }'
```

## 扩展功能

### 1. 配置版本管理

可以扩展支持配置版本管理，记录配置变更历史。

### 2. 配置模板

可以支持配置模板，提供常用的配置模板供用户选择。

### 3. 配置验证

可以添加配置验证规则，确保配置数据的有效性。

### 4. 配置回滚

可以支持配置回滚功能，允许用户回滚到之前的配置版本。

## 注意事项

1. **配置数据大小**: 建议配置数据不要过大，避免影响性能
2. **并发安全**: 配置更新操作需要考虑并发安全
3. **备份策略**: 建议定期备份配置文件
4. **监控告警**: 建议添加配置更新的监控和告警

## 故障排查

### 1. 配置更新失败

- 检查NATS连接状态
- 检查Runner是否正常运行
- 查看日志文件获取详细错误信息

### 2. 配置获取失败

- 检查配置键是否正确
- 检查配置文件是否存在
- 检查权限设置

### 3. API调用失败

- 检查API服务是否正常启动
- 检查认证信息是否正确
- 检查请求参数格式

## 总结

本配置更新功能提供了完整的配置管理解决方案，支持用户通过界面动态更新函数配置，具有良好的扩展性和安全性。通过三层架构设计，确保了系统的稳定性和可维护性。 