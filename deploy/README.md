# 部署说明（初芽记）

本文面向 **运维与发布**：API 服务、反向代理、小程序后台配置。产品功能与接口细节见仓库根 [README.md](../README.md)、前端 [apps/miniapp/README.md](../apps/miniapp/README.md)、后端 [services/api/README.md](../services/api/README.md)。

## 后端 API

### 构建二进制

在仓库根目录：

```bash
go build -o bin/chuyaji-api ./services/api/cmd/server
```

或使用 `make api`（见根目录 [Makefile](../Makefile)）。

### 配置

1. 复制 `services/api/.env.example` 为 `services/api/.env`（或部署目录下的 `.env`）。
2. 必填：`CHUYAJI_JWT_SECRET`；生产必填：微信小程序 `CHUYAJI_WECHAT_APP_ID`、`CHUYAJI_WECHAT_APP_SECRET`。
3. 生产 **关闭** `CHUYAJI_DEV_MODE`（勿允许 `code=dev`）。

### 进程与端口

- 默认监听 `CHUYAJI_HTTP_ADDR`（如 `:8282`）。前面放置 **Caddy / Nginx** 做 HTTPS 终止与反代到 `127.0.0.1:8282`。
- 可参考本目录下的 `Caddyfile`（将域名改为实际值）。
- Nginx 需设置足够大的 `client_max_body_size`（如 `32m`）以支持附件上传。

### systemd

可参考同目录 `chuyaji-api.service`，按实际安装路径与运行用户修改 `ExecStart`、`WorkingDirectory`、环境文件路径。

### Docker

在**仓库根目录**构建（见根 [Dockerfile](../Dockerfile)）：

```bash
docker build -t chuyaji-api .
docker run --rm -p 8282:8282 \
  -e CHUYAJI_JWT_SECRET=... \
  -v chuyaji-data:/data \
  chuyaji-api
```

镜像内默认 `CHUYAJI_DB_PATH=/data/chuyaji.db`、`CHUYAJI_UPLOAD_DIR=/data/uploads`；请持久化 `/data` 卷。

同目录 `docker-compose.yml` 仅为占位示例，生产请改用自行构建的镜像并妥善管理密钥。

### 附件与公网 URL

- 设置 `CHUYAJI_UPLOAD_DIR` 后，`POST /api/v1/records/:id/attachments/upload` 可用。
- 设置 `CHUYAJI_PUBLIC_BASE_URL` 为对外 HTTPS 根地址，以便生成 `/api/v1/p/:token` 外链。
- 小程序 **downloadFile / 图片** 合法域名需包含该 API 域名。

## 微信小程序

1. 在 `apps/miniapp` 执行 `npm install`（或 pnpm，与仓库脚本一致即可）。
2. 开发：`npm run dev:mp-weixin`，用微信开发者工具打开构建输出中的微信小程序工程。
3. 生产：`npm run build:mp-weixin`，上传审核前在[微信公众平台](https://mp.weixin.qq.com/)配置 **request / uploadFile / downloadFile** 合法域名为你的 **HTTPS API 域名**。

## 上线检查清单

- [ ] API 域名 TLS 有效，反代与防火墙放行
- [ ] 环境变量：`CHUYAJI_JWT_SECRET`、微信凭证、生产关闭 `CHUYAJI_DEV_MODE`
- [ ] 小程序合法域名与服务器域名一致
- [ ] SQLite 文件或后续若改用 PostgreSQL 的备份策略
- [ ] 日志不落敏感健康明细（自行配置脱敏与保留周期）

## 数据库备份

当前默认 **SQLite**：定期备份 `services/api/data/`（或 `CHUYAJI_DB_PATH` 所在目录）下数据库文件。若迁移至 PostgreSQL，使用 `pg_dump` 等工具做定时备份。
