# Nostalgia 个人博客系统

Nostalgia 是一个基于 Golang + gRPC + Gin + Redis + PostgreSQL + Docker 的博客系统，支持公开博客展示与 `/admin` 后台内容管理，内置文件上传、任务队列、权限管理等功能。

> 🚀 适用于个人博客、中小型内容系统、全栈开发练习项目。

---

## 📂 项目结构
```text
NOSTALGIA/
├── .github/ # GitHub CI/CD workflows 配置
├── api/ # Gin HTTP 服务（博客前台 API）
├── db/ # 数据库迁移脚本与 SQLC 定义
├── doc/ # 项目文档（含 statik 嵌入）
├── gapi/ # gRPC 服务实现（博客后台 API）
├── internal/ # 内部服务模块（如 Redis 封装）
├── mail/ # 邮件服务模块
├── pb/ # Protocol Buffers 编译输出
├── proto/ # gRPC 接口定义（.proto 文件）
├── resources/ # 上传资源文件目录（图片、附件等）
├── token/ # JWT / Paseto 认证逻辑
├── util/ # 配置加载与工具函数
├── validator/ # 自定义参数校验器
├── web/ # 前端 Dockerfile
│   └── frontend/ # 统一 Vue3 前端，包含公开博客与 /admin 后台
└── worker/ # 异步任务处理模块（如邮件队列）
```

---

## ⚙️ 环境变量配置（`.env.example`）

```env
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173,http://localhost:80,...
DB_DRIVER=postgres
DB_USER=root
DB_PASSWORD=secret
DB_SOURCE=postgresql://root:secret@0.0.0.0:15432/nostalgia?sslmode=disable
MIGRATION_URL=file://db/migration
RESOURCE_PATH=./resources
DOMAIN=http://localhost:8080
HTTP_SERVER_ADDRESS=0.0.0.0:8080
GRPC_GATEWAY_ADDRESS=0.0.0.0:9091
GRPC_SERVER_ADDRESS=0.0.0.0:9090
TOKEN_SYMMETRIC_KEY=...
SETUP_TOKEN=replace-with-a-random-one-time-bootstrap-token
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=24h
REDIS_ADDRESS=redis:6379
EMAIL_SENDER_NAME=name
EMAIL_SENDER_ADDRESS=...
EMAIL_SENDER_PASSWORD=...
UPLOAD_FILE_SIZE_LIMIT=5242880
UPLOAD_FILE_ALLOWED_MIME=image/jpeg,image/png
HTTP_PROXY_ADDR=http://host.docker.internal:10808
DEFAULT_USER_ID=uuid
DEFAULT_USERNAME=username
DEFAULT_USER_PASSWORD=123456
DEFAULT_USER_FULLNAME=fullname
DEFAULT_USER_EMAIL=xxx@qq.com
```

`TOKEN_SYMMETRIC_KEY` 用于签发访问令牌，至少 32 字节；`SETUP_TOKEN` 只用于首次创建管理员账号，不是后台登录密码，也不要提交真实值。通过 Makefile 启动本地 PostgreSQL 时，宿主机端口是 `15432`；Docker Compose 内部服务仍通过 `postgres:5432` 互联。

## 🚀 快速部署

> 确保你已安装 Docker 和 Docker Compose

### 1. 克隆项目

```bash
git clone https://github.com/MonitorAllen/nostalgia.git
cd nostalgia
```

### 2. 启动所有服务（API + Redis + PostgreSQL + Nginx + Vue 前端）
```bash
docker compose up --build
```

#### 默认服务端口：

| 服务           | 端口   |
|--------------|------|
| Gin API      | 8080 |
| gRPC         | 9090 |
| gRPC-Gateway | 9091 |
| Nginx 前端     | 80   |
| PostgreSQL   | Docker 内部 5432；Makefile 本地映射 15432 |
| Redis        | 6379 |

#### 接口入口说明（本地docker环境）

| 功能               | 地址示例                                                                               |
|------------------|------------------------------------------------------------------------------------|
| 前台博客首页           | [http://localhost/](http://localhost/)                                             |
| 后台内容管理（Vue）      | [http://localhost/admin/](http://localhost/admin/)                                 |
| 静态资源访问           | [http://localhost/resources/{id}/xxx.jpg](http://localhost/resources/{id}/xxx.jpg) |
| RESTful API      | [http://localhost/api/](http://localhost/api/)...                                  |
| gRPC Gateway API | [http://localhost/v1/](http://localhost/v1/)...                                    |
| Swagger 文档       | [http://localhost/swagger/index.html](http://localhost/swagger/index.html)         |

### 首次初始化管理员

Nostalgia 现在使用统一的用户认证模型：公开注册用户固定为 `visitor`，后台只允许 `role = admin` 的用户访问。首次部署时通过一次性 setup 流程创建唯一管理员：

1. 复制 `.env.example` 为 `.env`，设置 `TOKEN_SYMMETRIC_KEY` 和 `SETUP_TOKEN`。
2. 启动 PostgreSQL 后运行数据库迁移。
3. 启动 API 与前端后访问 [http://localhost/setup](http://localhost/setup)。
4. 输入 `.env` 中的 `SETUP_TOKEN`，创建第一个管理员用户。
5. 初始化完成后使用该账号访问 [http://localhost/admin/login](http://localhost/admin/login)。

创建第一个管理员后，`/setup` 不再允许创建新的管理员；后续公开注册账号只能作为 `visitor` 使用评论等公开登录能力。

## 🧪 本地开发

> 确保你已安装 Docker

### 1. 克隆项目

```bash
git clone https://github.com/MonitorAllen/nostalgia.git
cd nostalgia
```

### Go API

```bash
make redis # 默认镜像为 redis:7-alpine，可在 Makefile 中自行调整
make postgres # 默认镜像为 groonga/pgroonga:3.2.3-alpine-16
make createdb # 创建数据库
make migrateup # 数据库迁移
make server # 启动 API 服务
or
make server_docker_up # 参考 Makefile
```

### 前端

```bash
cd web/frontend
bun install
bun run dev
bun run type-check
bun run build
```

## 📮 联系与支持

如需反馈，可联系作者邮箱 monitorallen.pro@gmail.com 或提交 Issue。
