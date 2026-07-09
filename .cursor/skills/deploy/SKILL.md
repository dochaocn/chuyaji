---
name: deploy
description: >-
  初芽记（chuyaji）后端 API 及管理端前端的编译与远程部署。
  后端：编译 Go API 二进制、通过 SSH 上传至阿里云 /root/application/chuya 并重启服务。
  管理端：构建 Vue SPA、上传至 /root/application/chuyaji-admin/，通过 /admin/ 路径访问。
  在用户提到部署、上线、发布、同步到 aliyun 时使用。
---

# 初芽记：编译与远程部署

## 约定（本仓库）

| 项 | 值 |
|----|-----|
| SSH 主机别名 | `aliyun`（`~/.ssh/config` 中已配置） |
| 远端 API 根目录 | `/root/application/chuya` |
| 远端管理端目录 | `/root/application/chuyaji-admin/` |
| API 入口包 | `./services/api/cmd/server` |
| 构建产物名 | `bin/chuyaji-api`（与根目录 [Makefile](../../../Makefile) 一致） |
| 管理端访问地址 | `https://www.dochao.com.cn/admin/` |

详细配置、systemd、反代与小程序检查见仓库根目录 [README.md](../../../README.md) 中「部署与上线」；systemd 单元模板见 [deploy/chuyaji-api.service](../../../deploy/chuyaji-api.service)（需将路径改为本约定下的 `WorkingDirectory`、`ExecStart`、`EnvironmentFile`）。

## 操作步骤

### 1. 本地编译

在**仓库根目录**执行：

```bash
make api
```

等价于 `go build -o bin/chuyaji-api ./services/api/cmd/server`。

若在本机为 **macOS/Windows** 而服务器为 **Linux amd64**，在发布前交叉编译：

```bash
GOOS=linux GOARCH=amd64 go build -o bin/chuyaji-api ./services/api/cmd/server
```

（若目标为 ARM 云主机，将 `GOARCH` 改为 `arm64`。）

### 2. 部署目录与 `.env`

远端 API 工作目录约定为：`/root/application/chuya/services/api`（与 systemd `WorkingDirectory`、`godotenv` 加载路径一致）。

在**仓库根目录**执行：若远端**尚未存在** `.env`，则从本仓库 [services/api/.env.example](../../../services/api/.env.example) 复制一份到该目录并命名为 `.env`：

```bash
ssh aliyun 'mkdir -p /root/application/chuya/bin /root/application/chuya/services/api'
if ! ssh aliyun 'test -f /root/application/chuya/services/api/.env'; then
  scp services/api/.env.example aliyun:/root/application/chuya/services/api/.env
fi
```

复制后务必 **SSH 登录编辑** `/root/application/chuya/services/api/.env`，填写 `CHUYAJI_JWT_SECRET` 等必填项，并按生产环境调整 `CHUYAJI_DB_PATH`、`CHUYAJI_UPLOAD_DIR`（指向可写路径）。勿将含密钥的 `.env` 提交进 git。

若远端**已有** `.env`，不要覆盖，仅上传新二进制即可。

### 3. 上传二进制

当前 SSH 用户若**对** `/root/application/chuya/bin` **无直接写入权限**，直接 `scp` 到目标路径会失败（`dest open ... Failure`）。优先使用：**先上传到可写目录，再 `sudo mv` 到 `bin`**。

```bash
scp bin/chuyaji-api aliyun:/tmp/chuyaji-api
ssh aliyun 'sudo mv /tmp/chuyaji-api /root/application/chuya/bin/chuyaji-api && sudo chmod +x /root/application/chuya/bin/chuyaji-api'
```

若本机用户对 `bin` 目录可自行写入，也可一行直达（少数环境适用）：

```bash
scp bin/chuyaji-api aliyun:/root/application/chuya/bin/chuyaji-api
```

习惯用 rsync 时同理：可先同步到 `aliyun:/tmp/chuyaji-api`，再 `sudo mv` 到上述 `bin` 路径（或仅当 `rsync` 目标对用户可写时再直传 `bin`）。

### 4. 远端启动 / 重启

按服务器实际管理方式选一种：

**已配置 systemd（推荐）**

```bash
ssh aliyun 'sudo systemctl restart chuyaji-api && sudo systemctl status chuyaji-api --no-pager'
```

（上传与重启可合并为一条远程命令：在 `sudo mv` + `chmod +x` 成功后紧跟 `systemctl restart`。）

单元文件中 `ExecStart` 应指向 `/root/application/chuya/bin/chuyaji-api`，`WorkingDirectory` 与 `EnvironmentFile` 与线上 `.env` 位置一致。

**临时验证（无 systemd 时）**

```bash
ssh aliyun 'cd /root/application/chuya/services/api && export $(grep -v "^#" .env | xargs) && /root/application/chuya/bin/chuyaji-api'
```

（长期运行请用 systemd 或进程管理器，勿依赖手工前台进程。）

### 5. 验证

```bash
ssh aliyun 'curl -sS -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8282/healthz'
```

（若 `CHUYAJI_HTTP_ADDR` 非 `:8282`，改用实际监听地址。）

---

## 管理端前端部署

管理端为 Vue 3 SPA，部署在 `https://www.dochao.com.cn/admin/`。Nginx 通过 `alias /root/application/chuyaji-admin/` 托管静态文件，`/api/` 反向代理到后端。

### 1. 本地构建

在**仓库根目录**执行：

```bash
make admin-build
```

等价于 `cd apps/admin && npm run build`，产物在 `apps/admin/dist/`。

### 2. 上传到服务器

远端目录 `/root/application/chuyaji-admin/` 对当前 SSH 用户可能无写权限，先传到 `/tmp` 再 `sudo mv`：

```bash
ssh aliyun 'sudo rm -rf /root/application/chuyaji-admin/*'
scp -r apps/admin/dist/* aliyun:/tmp/chuyaji-admin/
ssh aliyun 'sudo cp -r /tmp/chuyaji-admin/* /root/application/chuyaji-admin/ && sudo rm -rf /tmp/chuyaji-admin'
```

若首次部署，先创建目录：

```bash
ssh aliyun 'sudo mkdir -p /root/application/chuyaji-admin'
```

### 3. 验证

```bash
ssh aliyun 'curl -sS -o /dev/null -w "%{http_code}\n" https://www.dochao.com.cn/admin/'
```

返回 200 即成功。浏览器访问 `https://www.dochao.com.cn/admin/` 应看到登录页。

## 常见故障

- **scp 报 `dest open ... Failure`**：多为对 `/root/.../bin` 无写权限，改用「上传到 `/tmp` 再 `sudo mv`」（见上文 §3）。
- **二进制无法执行**：确认交叉编译的 `GOOS`/`GOARCH` 与服务器一致；`chmod +x`。
- **启动即退出**：远端缺少 `CHUYAJI_JWT_SECRET` 或数据库路径不可写；查看 `journalctl -u chuyaji-api -e`。
- **附件/外链异常**：检查 `CHUYAJI_PUBLIC_BASE_URL` 与反代是否指向同一对外域名。

## 与本技能的关系

- 小程序构建、域名与合法域名校验：见根目录 [README.md](../../../README.md)「部署与上线」，不在此重复。
- 修改 API 后务必重新 `make api`（或交叉编译命令）再上传，避免运行旧二进制。
