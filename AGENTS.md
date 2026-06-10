# Nostalgia Agent Context

## 项目目标

Nostalgia 是一个个人博客系统，提供公开博客展示、`/backend` 后台内容管理、文件上传、任务队列、权限管理、邮件验证等能力。项目包含 Go 后端服务、gRPC/Gateway 接口、PostgreSQL/Redis 基础设施，以及一个统一的 Vue 3 前端应用。

## 技术栈

- 后端：Go、Gin、gRPC、gRPC-Gateway、sqlc、golang-migrate、Swagger/statik。
- 数据：PostgreSQL，使用 PGroonga 支持搜索；Redis 用于缓存或异步任务相关能力。
- 前端：Vue 3、Vite、TypeScript、Tailwind CSS、Reka UI、Pinia、CKEditor 5，并通过 Bun 安装依赖与构建；公开页面与 owner-only `/backend` 后台在 `web/frontend/` 内统一维护。
- 部署：Docker、Docker Compose、Cloudflare、Nginx、GitHub Actions。

## 目录说明

- `api/`：Gin HTTP API 与相关测试。
- `gapi/`：gRPC 服务实现与相关测试。
- `pb/`：Protocol Buffers 生成代码。
- `proto/`：`.proto` 接口定义。
- `db/`：数据库迁移、SQL 查询、sqlc 生成代码和数据库测试。
- `internal/`：内部模块，例如 Redis 缓存封装。
- `mail/`：邮件发送模块。
- `token/`：JWT/Paseto 认证与令牌逻辑。
- `util/`：配置、文件、密码、随机数、角色等工具。
- `validator/`：自定义参数校验。
- `worker/`：异步任务分发与处理。
- `doc/`：数据库/Swagger 等项目文档。
- `web/frontend/`：统一 Vue 应用，包含公开博客页面与 `/backend` 后台内容管理页面。
- `.github/workflows/`：测试与部署工作流。

## 常用命令

后端与数据库：

```bash
make postgres
make redis
make createdb
make migrateup
make sqlc
make test
make server
make proto
make swag
```

Docker：

```bash
docker compose up --build
docker compose -f docker-compose.dev.yaml up --build
```

前台：

```bash
cd web/frontend
bun install
bun run dev
bun run build
bun run type-check
bun run lint
```

## 测试、构建与验证

- Go 单元测试：`make test`，等价于 `go test -v -cover -short -count=1 ./...`。
- 数据库相关测试通常需要 PostgreSQL、迁移和可用的 `.env` 配置。
- 修改 SQL 查询后运行 `make sqlc`，确认生成代码同步。
- 修改 proto 后运行 `make proto`，确认 `pb/` 生成代码同步。
- 修改 Swagger 注释或 API 文档后运行 `make swag`。
- 前端修改应至少运行 `cd web/frontend && bun run type-check && bun run build`。
- `/api` 是公开 Gin HTTP API，`/v1` 是后台 `/backend` 使用的 gRPC-Gateway API surface。

## 代码规范

- Go 代码提交前运行 `gofmt`/`go test`，保持包边界清晰。
- 数据访问优先通过 `db/query/*.sql` 与 sqlc 生成层，不随意手写重复 SQL。
- gRPC 接口变更要同步更新 `proto/` 与 `pb/`。
- 前端遵循各自 `package.json` 中的 lint、format、type-check 脚本。
- 避免无关重构；只改动与当前任务直接相关的文件。

## 安全边界

- 不要提交密钥、令牌、真实密码或私密凭据。
- `.env` 可能包含本地敏感配置，Agent 不应复制其内容到文档、日志或提交信息中。
- 可参考 `.env.example` 了解配置项名称，但不要把真实值写入上下文文档。
- 加密环境文件通过 `make decrypt_env` / `make encrypt_env` 管理，使用前确认权限与密钥来源。
- 部署工作流依赖 GitHub Secrets，不要在仓库中硬编码替代值。
- Docker 镜像构建阶段不得解密、复制或打包生产 `.env`；生产配置通过 Compose `env_file` 或等价运行时环境变量注入。

## Git 与提交规范

- 仓库在 `master` 分支时禁止直接修改代码；需要变更先创建 `feature/*`、`fix/*` 或 `chore/*` 分支。
- 提交信息使用 Conventional Commits，例如：
  - `feat: add article search`
  - `fix: handle expired refresh token`
  - `chore: add agent context`
- 不要回滚或覆盖用户未明确要求处理的改动。

## 协作约定

- 用户称呼为“老大”，默认使用中文回复。
- 开始任务时先读取相关目录、README、Makefile、配置和测试，尊重已有实现。
- 发现缺失上下文、命令不确定或测试需要外部服务时，要说明依据与剩余人工确认项。
- 优先给出已验证结果；无法运行的验证要明确原因。

## 信息来源

本文件依据 `README.md`、`Makefile`、`go.mod`、`web/frontend/package.json`、`sqlc.yaml`、`.github/workflows/*.yml` 和项目目录结构整理。

## 待人工确认

- 本地开发时推荐使用 `docker compose.yaml` 还是 `docker-compose.dev.yaml` 作为默认入口。
- 是否需要为数据库集成测试提供固定的本地测试数据或 seed 流程。
