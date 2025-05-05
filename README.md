# API Server

API Server 是一个用于管理和部署API的服务端应用，支持Fork功能，类似Git的工作方式。

## 功能特点

1. **Runner管理**：Runner代表一个完整的项目，可以包含多个服务和函数。
2. **服务树管理**：服务树（ServiceTree）代表项目中的目录结构，类似于Go的包结构。
3. **函数管理**：Runner中的函数（RunnerFunc）代表具体的API实现。
4. **Fork功能**：支持对Runner、ServiceTree、RunnerFunc进行Fork操作，实现代码共享和复用。
5. **版本控制**：支持API的版本管理和回滚。

## 项目结构

```
api-server/
├── api/           # API控制器
│   └── v1/        # API V1版本
├── configs/       # 配置文件
├── model/         # 数据模型
├── pkg/           # 公共包
│   ├── config/    # 配置管理
│   ├── db/        # 数据库连接
│   ├── logger/    # 日志管理
│   ├── middleware/# 中间件
│   └── response/  # 响应处理
├── repo/          # 数据访问层
├── router/        # 路由管理
├── service/       # 业务逻辑层
└── main.go        # 程序入口
```

## 快速开始

### 环境要求

- Go 1.16+
- MySQL 5.7+

### 安装

1. 克隆代码仓库

```bash
git clone https://github.com/yunhanshu-net/api-server.git
cd api-server
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置数据库

编辑 `configs/config.json` 文件，配置数据库连接参数。

4. 运行项目

```bash
go run main.go
```

## API文档

### Runner API

- `POST /api/v1/runner`：创建Runner
- `GET /api/v1/runner`：获取Runner列表
- `GET /api/v1/runner/:id`：获取Runner详情
- `PUT /api/v1/runner/:id`：更新Runner
- `DELETE /api/v1/runner/:id`：删除Runner
- `POST /api/v1/runner/:id/fork`：Fork Runner
- `GET /api/v1/runner/:id/version`：获取Runner版本历史

### ServiceTree API

- `POST /api/v1/service-tree`：创建目录
- `GET /api/v1/service-tree`：获取目录列表
- `GET /api/v1/service-tree/:id`：获取目录详情
- `PUT /api/v1/service-tree/:id`：更新目录
- `DELETE /api/v1/service-tree/:id`：删除目录
- `POST /api/v1/service-tree/:id/fork`：Fork目录
- `GET /api/v1/service-tree/children/:parent_id`：获取子目录列表

### RunnerFunc API

- `POST /api/v1/runner-func`：创建函数
- `GET /api/v1/runner-func`：获取函数列表
- `GET /api/v1/runner-func/:id`：获取函数详情
- `PUT /api/v1/runner-func/:id`：更新函数
- `DELETE /api/v1/runner-func/:id`：删除函数
- `POST /api/v1/runner-func/:id/fork`：Fork函数
- `GET /api/v1/runner-func/runner/:runner_id`：获取Runner下的函数列表

## Fork功能实现

本项目实现了类似Git的Fork功能，可以对以下资源进行Fork操作：

1. **Runner**：Fork整个项目
2. **ServiceTree**：Fork一个目录及其下的所有内容
3. **RunnerFunc**：Fork单个函数

Fork时会记录来源信息，包括来源用户、来源ID等，便于后续的更新和同步。 