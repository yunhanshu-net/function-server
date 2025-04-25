# 目录树API

基于Gin和GORM的RESTful API服务，提供目录树、Runner和Runcher管理功能。

## 项目结构

```
api-server/
├── api/
│   └── v1/                    # API版本1
│       ├── directory/         # 目录结构模块
│       │   ├── tree.go        # 树形结构API
│       │   ├── package.go     # 包API
│       │   ├── function.go    # 函数API
│       │   └── routes.go      # 路由注册
│       ├── runner/            # Runner模块
│       │   ├── runner.go      # Runner API
│       │   ├── api.go         # Runner API管理
│       │   ├── deploy.go      # 部署管理
│       │   └── routes.go      # 路由注册
│       ├── runcher/           # Runcher模块
│       │   ├── runcher.go     # Runcher节点API
│       │   └── routes.go      # 路由注册
│       └── router.go          # API路由主文件
├── middleware/                # 中间件
│   └── auth.go                # 认证中间件
├── model/                     # 数据模型
│   ├── base/                  # 基础模型
│   ├── directory/             # 目录结构模型
│   ├── runner/                # Runner模型
│   ├── runcher/               # Runcher模型
│   ├── version/               # 版本模型
│   └── model.go               # 模型主入口
├── service/                   # 业务逻辑
├── main.go                    # 应用入口
└── go.mod                     # 项目依赖
```

## API 概览

### 目录结构

- `GET /api/v1/directory/tree` - 获取目录树
- `POST /api/v1/directory/tree/node` - 创建节点
- `GET /api/v1/directory/tree/node/:id` - 获取节点详情
- `PUT /api/v1/directory/tree/node/:id` - 更新节点
- `DELETE /api/v1/directory/tree/node/:id` - 删除节点
- `POST /api/v1/directory/tree/node/:id/move` - 移动节点

- `GET /api/v1/directory/packages` - 获取包列表
- `GET /api/v1/directory/packages/:id` - 获取包详情
- `POST /api/v1/directory/packages` - 创建包
- `PUT /api/v1/directory/packages/:id` - 更新包
- `DELETE /api/v1/directory/packages/:id` - 删除包

- `GET /api/v1/directory/functions` - 获取函数列表
- `GET /api/v1/directory/functions/:id` - 获取函数详情
- `POST /api/v1/directory/functions` - 创建函数
- `PUT /api/v1/directory/functions/:id` - 更新函数
- `DELETE /api/v1/directory/functions/:id` - 删除函数

### Runner管理

- `GET /api/v1/runners` - 获取Runner列表
- `GET /api/v1/runners/:id` - 获取Runner详情
- `POST /api/v1/runners` - 创建Runner
- `PUT /api/v1/runners/:id` - 更新Runner
- `DELETE /api/v1/runners/:id` - 删除Runner

- `GET /api/v1/runners/:runner_id/apis` - 获取Runner API列表
- `GET /api/v1/runners/:runner_id/apis/:id` - 获取Runner API详情
- `POST /api/v1/runners/:runner_id/apis` - 添加Runner API
- `PUT /api/v1/runners/:runner_id/apis/:id` - 更新Runner API
- `DELETE /api/v1/runners/:runner_id/apis/:id` - 删除Runner API

- `POST /api/v1/runners/:id/deploy` - 部署Runner
- `GET /api/v1/runners/:id/status` - 获取Runner状态
- `POST /api/v1/runners/:id/restart` - 重启Runner
- `POST /api/v1/runners/:id/stop` - 停止Runner

### Runcher节点管理

- `GET /api/v1/runchers` - 获取Runcher节点列表
- `GET /api/v1/runchers/:id` - 获取Runcher节点详情
- `POST /api/v1/runchers` - 创建Runcher节点
- `PUT /api/v1/runchers/:id` - 更新Runcher节点
- `DELETE /api/v1/runchers/:id` - 删除Runcher节点

## 运行

```bash
# 安装依赖
go mod tidy

# 运行服务器
go run main.go
```

服务器默认在 http://localhost:8080 启动。

## 认证

所有API请求需要在Header中提供认证信息：

```
Authorization: Bearer {token}
```

或者在URL中提供token参数：

```
?token={token}
```