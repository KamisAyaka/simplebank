# SimpleBank

一个基于 Go + Gin + PostgreSQL + sqlc 的后端练习项目，支持用户、账户、流水、转账等功能，并包含鉴权中间件与测试。

## 运行环境

- Go `1.26+`
- Docker / Docker Compose
- PostgreSQL（本地模式可选）
- `migrate` CLI（本地迁移用）
- `sqlc`（生成 SQL 对应 Go 代码）
- `mockgen`（生成接口 mock）

## 配置文件

- 本地运行使用：`app.env`
- 镜像内默认配置：`app.docker.env`

主要配置项：

- `DB_DRIVER`
- `DB_SOURCE`
- `SERVER_ADDRESS`
- `TOKEN_SYMMETRIC_KEY`（至少 32 位）
- `ACCESS_TOKEN_DURATION`
- `REFRESH_TOKEN_DURATION`

## 快速开始（推荐：Docker Compose）

```bash
docker compose up --build
```

启动后：

- API: `http://localhost:8080`
- Postgres: `localhost:5432`（用户名/密码/库见 `docker-compose.yaml`）

停止：

```bash
docker compose down
```

删除数据卷（重置数据库）：

```bash
docker compose down -v
```

## 本地开发启动

1. 启动 PostgreSQL 容器（Makefile 方案）：

```bash
make postgres
make createdb
```

2. 执行迁移：

```bash
make migrateup
```

3. 启动服务：

```bash
make server
```

## 常用命令

### 数据库

```bash
make createdb
make dropdb
make migrateup
make migrateup1
make migratedown
make migratedown1
```

### 迁移（migrate）进阶指令

安装 `migrate`（macOS）：

```bash
brew install golang-migrate
```

创建新的迁移文件（会生成 up/down 两个文件）：

```bash
migrate create -ext sql -dir db/migration -seq add_xxx_table
```

手动执行迁移（不走 Makefile）：

```bash
migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up
migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down
```

只迁移/回滚一步：

```bash
migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up 1
migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down 1
```

### 代码生成

```bash
make sqlc
make mock
```

### 测试

```bash
make test
go test ./... -count=1
```

### Docker

```bash
docker build -t simplebank:latest .
docker run --rm --name simplebank -p 8080:8080 simplebank:latest
```

### Docker Compose 迁移指令

`api` 容器启动时会自动执行迁移（见 `start.sh`）。

手动在容器里执行迁移：

```bash
docker compose exec api /app/migrate -path /app/migration -database "postgresql://root:secret@postgres:5432/simple_bank?sslmode=disable" -verbose up
docker compose exec api /app/migrate -path /app/migration -database "postgresql://root:secret@postgres:5432/simple_bank?sslmode=disable" -verbose down 1
```

## 常用 API 调试

### 注册用户

```bash
curl -i -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username":"tom",
    "password":"123456",
    "full_name":"Tom Jerry",
    "email":"tom@example.com"
  }'
```

### 登录

```bash
curl -i -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username":"tom",
    "password":"123456"
  }'
```

登录后拿到 `access_token`，访问受保护接口时加：

```bash
Authorization: Bearer <access_token>
```

## 常见问题

- `404 /users/login`：你发的是 `GET`，正确是 `POST /users/login`。
- `invalid key size`：`TOKEN_SYMMETRIC_KEY` 长度不足 32。
- `port is already allocated`：端口冲突，先停掉占用端口的容器/服务。
