# SimpleBank

一个基于 Go + PostgreSQL + gRPC + gRPC-Gateway + sqlc 的后端项目。

当前仓库默认启动：
- gRPC 服务（`GRPC_SERVER_ADDRESS`，默认 `:9090`）
- HTTP Gateway（`HTTP_SERVER_ADDRESS`，默认 `:8080`）
- Asynq 任务处理器（依赖 Redis）

项目包含用户创建、登录、更新用户信息，并在创建用户后投递“发送邮箱验证码”异步任务。

## 技术栈

- Go
- PostgreSQL
- gRPC / Protocol Buffers
- gRPC-Gateway
- sqlc
- golang-migrate
- Asynq + Redis
- Zerolog
- Swagger(OpenAPI v2)

## 目录结构

- `main.go`：应用入口（加载配置、跑迁移、启动 gRPC/Gateway/Worker）
- `util/`：配置与通用工具
- `gapi/`：gRPC API 实现
- `db/migration/`：SQL 迁移
- `db/query/`：sqlc 查询定义
- `db/sqlc/`：sqlc 生成代码与事务逻辑（含 `CreateUserTx`）
- `worker/`：Asynq 任务分发与处理
- `proto/`：proto 定义
- `pb/`：protobuf / gateway 生成代码
- `doc/swagger/`：Swagger JSON 与静态资源
- `doc/statik/`：嵌入 Swagger 资源（由 `statik` 生成）

## 环境要求

- Go 1.26+
- PostgreSQL
- Redis
- protoc
- `protoc-gen-go`、`protoc-gen-go-grpc`、`protoc-gen-grpc-gateway`、`protoc-gen-openapiv2`
- `sqlc`
- `migrate` CLI
- `mockgen`
- `statik`

## 配置

本地默认读取 `app.env`，容器默认读取 `app.docker.env`（复制到容器内 `/app/app.env`）。

关键配置项：

- `ENVIRONMENT`：`development/dev` 时使用 zerolog ConsoleWriter，其他环境输出 JSON 日志
- `DB_DRIVER`
- `DB_SOURCE`
- `MIGRATION_URL`（本地通常 `file://db/migration`，容器通常 `file:///app/db/migration`）
- `REDIS_ADDRESS`（例如 `localhost:6379`）
- `HTTP_SERVER_ADDRESS`
- `GRPC_SERVER_ADDRESS`
- `TOKEN_SYMMETRIC_KEY`（至少 32 位）
- `ACCESS_TOKEN_DURATION`
- `REFRESH_TOKEN_DURATION`

## 本地快速启动

1. 启动 PostgreSQL

```bash
make postgres
make createdb
```

2. 启动 Redis

```bash
make redis
```

3. 启动服务

```bash
make server
```

说明：
- `main.go` 会在启动时自动执行 `migrate up`。
- 没有新迁移时会打印 `no new db migration`。

## Docker Compose

```bash
docker compose up --build
```

当前 `docker-compose.yaml` 只包含 `postgres` 与 `api` 服务。
如果你要跑异步任务链路，请确保 Redis 可达并给 `api` 配置 `REDIS_ADDRESS`。

## API（当前）

Service: `pb.SimpleBank`

- `CreateUser`
  - gRPC: `/pb.SimpleBank/CreateUser`
  - HTTP: `POST /v1/create_user`
- `UpdateUser`
  - gRPC: `/pb.SimpleBank/UpdateUser`
  - HTTP: `PATCH /v1/update_user`
- `LoginUser`
  - gRPC: `/pb.SimpleBank/LoginUser`
  - HTTP: `POST /v1/login_user`

## Swagger

先生成 swagger 文件和嵌入资源：

```bash
make proto
```

再启动服务后访问：

- `http://localhost:8080/swagger/`

说明：
- Gateway 使用嵌入的静态资源（`doc/statik`），不是运行时外部 CDN。
- `make proto` 会先生成 `doc/swagger/*.swagger.json`，再通过 `statik` 生成 `doc/statik`。

## 常用命令

数据库与迁移：

```bash
make migrateup
make migrateup1
make migratedown
make migratedown1
```

代码生成：

```bash
make sqlc
make mock
make proto
```

测试：

```bash
make test
go test ./... -count=1
```

## 一致性说明（创建用户 + 异步任务）

`CreateUser` 目前通过 `db/sqlc` 的 `CreateUserTx` 处理：

- 在同一个数据库事务流程中先创建用户
- 再执行任务分发回调（投递 Asynq 任务）
- 如果任务分发失败，则返回错误并回滚用户创建

这能避免“用户已落库但任务未分发”的窗口问题。

## 常见问题

- `failed to create session`
  - 请先检查 `sessions` 表约束与 `client_ip` 冲突（历史数据/旧逻辑可能导致）。
- `invalid parameters`
  - 检查请求 JSON 是否合法（特别是末尾多余逗号）。
- `cannot create new migrate instance: URL cannot be empty`
  - 检查 `MIGRATION_URL` 是否正确配置。
- `swagger 404`
  - 确认服务启动成功，并访问 `http://localhost:8080/swagger/`。
