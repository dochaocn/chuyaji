# 初芽记（chuyaji）

家庭自用：记录孕期检查与婴幼儿成长。**微信小程序**（uni-app + Vue 3）与 **Go HTTP API**（Gin + GORM + SQLite）同仓维护，前后端可分开部署。

## 文档怎么读

| 文档 | 内容 |
|------|------|
| [apps/miniapp/README.md](apps/miniapp/README.md) | **前端**：页面、路由、调用的接口封装、本地运行与小程序发布 |
| [services/api/README.md](services/api/README.md) | **后端**：目录结构、架构、REST 接口、环境变量与本地启动 |
| [deploy/README.md](deploy/README.md) | **部署**：反代、systemd、Docker、合法域名与备份 |

根目录本文只做**仓库总览**与**跨端约定**；细节以前后端各自 README 为准。

## 远端仓库

[https://github.com/dochaocn/chuyaji.git](https://github.com/dochaocn/chuyaji.git)

## 仓库结构

```text
chuyaji/
├── apps/miniapp/          # 微信小程序（uni-app、Vue 3、Pinia）
├── services/api/          # Go API（Gin、GORM、SQLite）
├── deploy/                # Caddy / systemd 等部署示例
├── Dockerfile             # 仅构建 API 二进制镜像
├── Makefile               # make api / make run-api 等
├── go.mod                 # module: github.com/dochaocn/chuyaji
└── README.md              # 本文件
```

## 快速开始

### 1. 启动 API（后端）

在仓库根目录：

```bash
cp services/api/.env.example services/api/.env
# 至少填写 CHUYAJI_JWT_SECRET；本地联调可设 CHUYAJI_DEV_MODE=true
go mod tidy
go run ./services/api/cmd/server
```

默认监听 **`http://127.0.0.1:8282`**，健康检查：`GET /healthz`。在线文档：`/docs`，OpenAPI：`/openapi.yaml`。

也可用：`make run-api`（见 [Makefile](Makefile)）。

### 2. 启动小程序（前端）

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

用微信开发者工具打开构建产物目录（一般为 `dist/dev/mp-weixin`，以本地 CLI 输出为准），开发设置中关闭「合法域名校验」。API 根地址在 `apps/miniapp/.env.development` 的 **`VITE_CHUYAJI_API_BASE`**。

更多联调说明（真机 IP、`code=dev`、40029 等）见 [apps/miniapp/README.md](apps/miniapp/README.md)。

### 3. 可选：Docker 只跑 API

在仓库根目录：

```bash
docker build -t chuyaji-api .
docker run --rm -p 8282:8282 -e CHUYAJI_JWT_SECRET=dev-secret -v chuyaji-data:/data chuyaji-api
```

详见 [deploy/README.md](deploy/README.md)。

## 跨端约定（摘要）

- **认证**：除公开路由外，`/api/v1/*` 需请求头 `Authorization: Bearer <JWT>`；令牌由 `POST /api/v1/auth/wechat` 返回。
- **开发登录**：后端 `CHUYAJI_DEV_MODE=true` 时，请求体可用 `{"code":"dev"}`；前端开发构建默认倾向走 dev 登录，详见前端 README。
- **接口契约**：以 `services/api/internal/apidocs/openapi.yaml` 为源，服务内嵌 `/docs` 与 `/openapi.yaml`。

## 与 GitHub 同步（示例）

```bash
cd chuyaji
git init
git remote add origin https://github.com/dochaocn/chuyaji.git
git add -A && git commit -m "chore: initial monorepo skeleton"
git branch -M main
git push -u origin main
```

## 规划与已知边界

- **F-LLM-001**：报告识图与大模型摘要（`payload` 预留扩展，未实现）。

## License

Private / 自行约定。
