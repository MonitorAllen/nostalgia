---
name: nostalgia-project
description: Use when working in the Nostalgia repository, especially for Go/Gin/gRPC backend changes, sqlc migrations, proto generation, Vue frontend work, Docker workflows, or project Agent Context maintenance.
---

# Nostalgia Project Skill

## 触发条件

当任务涉及本仓库的后端、前端、数据库、部署、测试、文档或 Agent Context 时使用。

## 工作流程

1. 先阅读项目根目录的 `AGENTS.md`，并遵守其中的分支、提交、安全与验证约定。
2. 依据改动范围读取对应文件：
   - 后端：`README.md`、`Makefile`、相关 `api/`、`gapi/`、`db/`、`util/`、`worker/` 文件。
   - 数据库：`db/migration/`、`db/query/`、`sqlc.yaml`。
   - gRPC：`proto/`、`pb/`、`gapi/`。
   - 前端：`web/frontend/package.json`；公开页面与 `/backend` 后台都在统一前端内。
3. 修改 SQL 后运行或说明 `make sqlc`；修改 proto 后运行或说明 `make proto`。
4. 完成后运行最小相关验证；无法运行时说明缺少的服务、环境变量或依赖。

## 安全提醒

不要读取或传播 `.env` 中的真实敏感值。需要配置项名称时参考 `.env.example`。
Docker 镜像构建阶段不得解密、复制或打包生产 `.env`；生产配置通过 Compose `env_file` 或等价运行时环境变量注入。
