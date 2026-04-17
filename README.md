# 初芽记（chuyaji）

前后端同仓的微信小程序项目，已从「家庭 -> 宝宝 -> 时间线」重构为 `**宝宝`、`宝妈` 双 Tab 工作台**。

## 仓库结构

```text
chuyaji/
├── apps/miniapp/          # 微信小程序（uni-app、Vue 3、Pinia）
├── services/api/          # Go API（Gin、GORM、SQLite）
├── deploy/                # systemd 单元示例等
├── Makefile               # 常用命令
└── README.md
```

## 产品结构

- **宝宝**：工作台、档案、记录、生长趋势  
- **宝妈**：工作台、档案、记录

家庭能力已移除，不再作为前后端模型与页面的一部分。

---

## 快速开始

### 后端

```bash
cp services/api/.env.example services/api/.env
# 编辑 .env，至少填写 CHUYAJI_JWT_SECRET
go run ./services/api/cmd/server
```

健康检查：`curl -sS http://127.0.0.1:8282/healthz`

### 前端

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

用微信开发者工具打开构建输出中的小程序工程。生产构建：`npm run build:mp-weixin`，并在 `apps/miniapp/.env.production` 中配置 `VITE_CHUYAJI_API_BASE` 为线上 HTTPS API 根地址。

---

## 前端（`apps/miniapp`）

围绕 `宝宝`、`宝妈` 双 Tab；默认进入宝宝工作台，体验为「工作台 + 档案 + 记录」。

### 页面结构


| 页面              | 路径                                       | 说明                  |
| --------------- | ---------------------------------------- | ------------------- |
| 宝宝工作台           | `pages/baby/home`                        | 默认首页，档案摘要、最近记录、生长入口 |
| 宝宝档案            | `pages/baby/profile-edit`                | 昵称、孕期、出生资料、喂养等      |
| 宝宝记录编辑 / 详情     | `pages/baby/record-edit`、`record-detail` | 新增/编辑/查看记录与附件       |
| 生长趋势            | `pages/baby/growth`                      | 体重趋势摘要与数据点          |
| 宝妈工作台 / 档案 / 记录 | `pages/mom/*`                            | 宝妈侧工作台与记录           |
| 隐私说明            | `pages/privacy/privacy`                  | 隐私同意状态              |


### API 封装与状态

- `src/api/chuyaji.ts`：宝宝 / 宝妈 / 记录 / Dashboard 等 HTTP 封装  
- `src/api/upload.ts`：宝宝记录、宝妈记录附件上传  
- `src/store/auth.ts`：JWT、用户 ID  
- `src/store/session.ts`：`babyId`、`motherId`、隐私状态

### 环境变量（Vite）


| 变量                       | 说明                                                                          |
| ------------------------ | --------------------------------------------------------------------------- |
| `VITE_CHUYAJI_API_BASE`  | 后端 API 根地址，如 `http://127.0.0.1:8282`                                        |
| `VITE_DEV_WECHAT_LOGIN`  | 设为 `true` 时发 `code=dev`（须后端 `CHUYAJI_DEV_MODE=true`）；默认使用真实 `wx.login` code |
| `VITE_CHUYAJI_DEV_TOKEN` | 仅 **Vite 开发模式**：填入 JWT 可跳过微信登录；建议写在 `.env.local`、勿提交                        |


开发时若后端开启 `CHUYAJI_DEV_MODE=true`，可签发 JWT 再粘贴到 `VITE_CHUYAJI_DEV_TOKEN`：

```bash
curl -sS -X POST "http://127.0.0.1:8282/api/v1/auth/dev/token" \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}' | jq -r .token
```

### 界面风格

`src/uni.scss`「暖春手帐」：米杏底、陶土红主色、薄荷绿点缀、卡片化与圆角轻阴影。

### 其他命令

```bash
npm run type-check
```

### 当前边界

- 生长页为趋势摘要 + 数据点，后续可升级为完整折线图  
- 登录以微信 `code` 换票为主；开发可用 `VITE_DEV_WECHAT_LOGIN`、`VITE_CHUYAJI_DEV_TOKEN` 或 `POST /api/v1/auth/dev/token`

---

## 后端（`services/api`）

小程序 HTTP API；模型为「用户 → 宝宝 / 宝妈」双域结构。

**技术栈**：Go、Gin、GORM、SQLite、JWT。

### 认证

- `POST /api/v1/auth/wechat` — 微信 code 登录  
- `POST /api/v1/auth/dev/token` — 仅 `CHUYAJI_DEV_MODE=true` 时可用，按 `user_id` 签发 JWT  
- `GET /api/v1/me` — 当前用户（需 Bearer）

### 宝宝

- `GET/POST /api/v1/babies`，`GET/PATCH/DELETE /api/v1/babies/:id`  
- `GET /api/v1/dashboard/baby`

### 宝宝记录

- `GET/POST /api/v1/babies/:id/records`  
- `GET/PATCH/DELETE /api/v1/records/:id`

### 宝妈

- `GET/POST /api/v1/mothers`，`GET/PATCH /api/v1/mothers/:id`  
- `GET /api/v1/dashboard/mother`

### 宝妈记录

- `GET/POST /api/v1/mothers/:id/records`  
- `GET/PATCH/DELETE /api/v1/mother-records/:id`

### 附件

- 各记录下的 `attachments` 与 `upload` 路由；`GET /api/v1/p/:token` 公开附件

### 数据模型要点

- **保留**：`User`、`Baby`、`Record`、`Attachment`  
- **新增**：`Mother`、`MotherRecord`  
- **调整**：`Baby` 归属为 `user_id`；`Attachment` 使用 `owner_type` + `owner_id`

### 环境变量


| 变量                                                    | 说明                            |
| ----------------------------------------------------- | ----------------------------- |
| `CHUYAJI_HTTP_ADDR`                                   | 监听地址，默认 `:8282`               |
| `CHUYAJI_JWT_SECRET`                                  | JWT 签名密钥（必填）                  |
| `CHUYAJI_DB_PATH`                                     | SQLite 路径                     |
| `CHUYAJI_WECHAT_APP_ID` / `CHUYAJI_WECHAT_APP_SECRET` | 微信小程序凭证（生产必填）                 |
| `CHUYAJI_DEV_MODE`                                    | `true` 时允许 `code=dev` 及开发签发接口 |
| `CHUYAJI_UPLOAD_DIR`                                  | 附件目录                          |
| `CHUYAJI_PUBLIC_BASE_URL`                             | 对外生成附件 URL 的 HTTPS 根地址        |


### API 文档

- OpenAPI 源文件：`services/api/internal/apidocs/openapi.yaml`  
- 服务启动后可访问 `/docs`

---

## 部署与上线

### 构建二进制

在仓库根目录：

```bash
go build -o bin/chuyaji-api ./services/api/cmd/server
# 或
make api
```

### 配置

1. 复制 `services/api/.env.example` 为部署目录下的 `.env`（或 `services/api/.env`）。
2. 必填：`CHUYAJI_JWT_SECRET`；生产必填：微信 `CHUYAJI_WECHAT_APP_ID`、`CHUYAJI_WECHAT_APP_SECRET`。
3. 生产环境 **关闭** `CHUYAJI_DEV_MODE`。

### 进程与端口

- 默认监听 `CHUYAJI_HTTP_ADDR`（如 `:8282`）。生产建议在前面用 **Nginx** 等做 HTTPS 终止与反向代理到本机 `127.0.0.1:8282`。  
- 反代需放大上传体积（如 Nginx `client_max_body_size 32m`）以支持附件。

### systemd

可参考 `deploy/chuyaji-api.service`，按实际路径修改 `ExecStart`、`WorkingDirectory`、`EnvironmentFile` 与运行用户。

### 附件与公网 URL

- 配置 `CHUYAJI_UPLOAD_DIR` 后，multipart 上传接口可用。  
- `CHUYAJI_PUBLIC_BASE_URL` 设为对外 HTTPS 根地址，用于 `/api/v1/p/:token` 等外链。  
- 小程序 **request / uploadFile / downloadFile** 合法域名需包含该 API 域名。

### 微信小程序发布

1. `apps/miniapp` 执行 `npm install`，生产执行 `npm run build:mp-weixin`。
2. 在微信公众平台配置合法域名与小程序信息。

### 上线检查清单

- API 域名 TLS 有效，防火墙与反代放行  
- 环境变量：JWT、微信凭证、生产关闭 `CHUYAJI_DEV_MODE`  
- 小程序合法域名与 API 一致  
- SQLite（或后续数据库）备份策略

### 数据库备份

默认 SQLite：定期备份 `CHUYAJI_DB_PATH` 所在文件；若迁移至 PostgreSQL，使用 `pg_dump` 等工具。