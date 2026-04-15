# 初芽记 · HTTP API（后端）

`services/api` 是初芽记的小程序后端，当前模型已经从“家庭中心”调整为“用户 -> 宝宝 / 宝妈”双域结构。

技术栈：

- Go
- Gin
- GORM
- SQLite
- JWT

## 核心模块

### 认证

- `POST /api/v1/auth/wechat`
- `GET /api/v1/me`

### 宝宝

- `GET /api/v1/babies`
- `POST /api/v1/babies`
- `GET /api/v1/babies/:id`
- `PATCH /api/v1/babies/:id`
- `DELETE /api/v1/babies/:id`
- `GET /api/v1/dashboard/baby`

### 宝宝记录

- `GET /api/v1/babies/:id/records`
- `POST /api/v1/babies/:id/records`
- `GET /api/v1/records/:id`
- `PATCH /api/v1/records/:id`
- `DELETE /api/v1/records/:id`

### 宝妈

- `GET /api/v1/mothers`
- `POST /api/v1/mothers`
- `GET /api/v1/mothers/:id`
- `PATCH /api/v1/mothers/:id`
- `GET /api/v1/dashboard/mother`

### 宝妈记录

- `GET /api/v1/mothers/:id/records`
- `POST /api/v1/mothers/:id/records`
- `GET /api/v1/mother-records/:id`
- `PATCH /api/v1/mother-records/:id`
- `DELETE /api/v1/mother-records/:id`

### 附件

- `GET /api/v1/records/:id/attachments`
- `POST /api/v1/records/:id/attachments`
- `POST /api/v1/records/:id/attachments/upload`
- `GET /api/v1/mother-records/:id/attachments`
- `POST /api/v1/mother-records/:id/attachments`
- `POST /api/v1/mother-records/:id/attachments/upload`
- `DELETE /api/v1/attachments/:id`
- `GET /api/v1/p/:token`

## 数据模型

### 保留

- `User`
- `Baby`
- `Record`
- `Attachment`

### 新增

- `Mother`
- `MotherRecord`

### 调整

- `Baby` 归属由 `family_id` 调整为 `user_id`
- `Attachment` 归属由单一 `record_id` 调整为 `owner_type + owner_id`

## 本地启动

在仓库根目录：

```bash
cp services/api/.env.example services/api/.env
go run ./services/api/cmd/server
```

健康检查：

```bash
curl -sS http://127.0.0.1:8282/healthz
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `CHUYAJI_HTTP_ADDR` | 监听地址，默认 `:8282` |
| `CHUYAJI_JWT_SECRET` | JWT 签名密钥 |
| `CHUYAJI_DB_PATH` | SQLite 文件路径 |
| `CHUYAJI_WECHAT_APP_ID` / `CHUYAJI_WECHAT_APP_SECRET` | 微信小程序凭证 |
| `CHUYAJI_DEV_MODE` | 允许 `code=dev` 登录 |
| `CHUYAJI_UPLOAD_DIR` | 附件上传目录 |
| `CHUYAJI_PUBLIC_BASE_URL` | 对外生成附件 URL 的根地址 |

## 文档

- OpenAPI 源文件：`internal/apidocs/openapi.yaml`
- 在线文档：启动服务后访问 `/docs`
