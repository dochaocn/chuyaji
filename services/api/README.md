# 初芽记 · HTTP API（后端）

初芽记后端：**Gin** 路由与中间件、**GORM** 持久化、默认 **SQLite**（`github.com/glebarez/sqlite` 纯 Go 驱动）。对外提供 REST JSON、JWT 鉴权、可选本地上传附件与公开下载链接。

## 项目简介

职责：用户通过微信 `code` 登录、管理家庭/宝宝/成长记录与附件元数据；不负责小程序前端展示。非医疗诊断服务，数据由家庭自备份策略保障。

## 仓库结构（`services/api`）

```text
services/api/
├── cmd/server/           # 进程入口
├── internal/
│   ├── apidocs/          # /docs、/openapi.yaml 内嵌
│   ├── config/           # 环境变量
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # JWT 等
│   ├── model/            # GORM 模型
│   ├── router/           # 路由注册
│   └── ...
├── internal/apidocs/openapi.yaml   # OpenAPI 契约源文件
├── .env.example
└── README.md             # 本文件
```

## 技术架构

| 层级 | 说明 |
|------|------|
| 入口 | `cmd/server/main.go` 加载配置、迁移、注入 `handler.Handler`、注册 `router.New` |
| 路由 | `internal/router/router.go`：公开 `/healthz`、`/api/v1/p/:token`；`/api/v1/auth/wechat`；其余 `/api/v1/*` JWT |
| 数据 | GORM + SQLite 文件（路径 `CHUYAJI_DB_PATH`） |
| 文档 | `apidocs.Register`：`GET /docs`（Stoplight Elements）、`GET /openapi.yaml` |

认证：除表内「公开」路由外，`Authorization: Bearer <JWT>`。JWT 由 `POST /api/v1/auth/wechat` 签发。

## 接口概览

**在线文档**：服务启动后访问 **`/docs`**；原始规范 **`/openapi.yaml`**。修改路由或请求体时请同步更新 `internal/apidocs/openapi.yaml`。

默认 Base URL：`http://127.0.0.1:8282`（以 `CHUYAJI_HTTP_ADDR` 为准）。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查，返回 `ok` |
| GET | `/docs` | Stoplight Elements |
| GET | `/openapi.yaml` | OpenAPI 3 YAML |
| POST | `/api/v1/auth/wechat` | body `{"code"}`，微信换票或开发模式 `dev` |
| GET | `/api/v1/me` | 当前用户（需 JWT） |
| GET | `/api/v1/families` | 家庭列表 |
| POST | `/api/v1/families` | 创建家庭 |
| POST | `/api/v1/families/join` | 邀请码加入 |
| GET | `/api/v1/families/:id` | 家庭详情 |
| GET | `/api/v1/babies` | 宝宝列表，`family_id` 必填 |
| POST | `/api/v1/babies` | 创建宝宝 |
| GET | `/api/v1/babies/:id` | 宝宝详情 |
| PATCH | `/api/v1/babies/:id` | 更新宝宝 |
| DELETE | `/api/v1/babies/:id` | 删除宝宝 |
| GET | `/api/v1/babies/:id/records` | 记录分页，`limit`、`cursor` |
| POST | `/api/v1/babies/:id/records` | 新建记录 |
| GET | `/api/v1/records/:id` | 记录详情 |
| PATCH | `/api/v1/records/:id` | 更新记录 |
| DELETE | `/api/v1/records/:id` | 删除记录 |
| GET | `/api/v1/records/:id/attachments` | 附件列表 |
| POST | `/api/v1/records/:id/attachments` | 登记附件 URL（JSON） |
| POST | `/api/v1/records/:id/attachments/upload` | multipart 字段 `file`（需 `CHUYAJI_UPLOAD_DIR`） |
| DELETE | `/api/v1/attachments/:id` | 删除附件 |
| GET | `/api/v1/p/:token` | **公开**，按 share_token 读附件流（无 JWT） |

## 本地启动

在**仓库根目录**：

```bash
cp services/api/.env.example services/api/.env
# 必填：CHUYAJI_JWT_SECRET；开发可设 CHUYAJI_DEV_MODE=true
go mod tidy
go run ./services/api/cmd/server
```

或在 `services/api` 下：`go run ./cmd/server`（`.env` 放该目录即可）。

依赖下载慢时可设置 `GOPROXY`（如 `https://goproxy.cn,direct`）。

**健康检查**：`curl -sS http://127.0.0.1:8282/healthz`

## 环境变量

完整示例见 [.env.example](.env.example)。

| 变量 | 说明 |
|------|------|
| `CHUYAJI_HTTP_ADDR` | 监听地址，默认 `:8282` |
| `CHUYAJI_JWT_SECRET` | JWT 签名密钥（必填） |
| `CHUYAJI_DB_PATH` | SQLite 路径，默认 `./data/chuyaji.db` |
| `CHUYAJI_WECHAT_APP_ID` / `CHUYAJI_WECHAT_APP_SECRET` | 微信小程序凭证 |
| `CHUYAJI_DEV_MODE` | `true` 时允许 `code=dev` 登录（**禁止用于生产**） |
| `CHUYAJI_UPLOAD_DIR` | 非空则启用本地上传目录 |
| `CHUYAJI_PUBLIC_BASE_URL` | 生成附件外链时的公网根 URL；空则用请求 Host |

## 构建

仓库根目录：

```bash
make api
# 产出 bin/chuyaji-api
```

或：`go build -o bin/chuyaji-api ./services/api/cmd/server`

## 部署

生产环境 HTTPS、反代、systemd、Docker 与合法域名见仓库 **[deploy/README.md](../../deploy/README.md)**。根目录 [Dockerfile](../../Dockerfile) 仅构建 API 镜像。

## 规划与边界

- 大模型报告识图等能力未实现；记录 `payload` 可扩展预留。
- 日志与监控策略由部署方自行配置；勿在日志中输出敏感健康明细。

## 相关文档

- 单体仓库总览：[README.md](../../README.md)
- 小程序前端：[apps/miniapp/README.md](../../apps/miniapp/README.md)
