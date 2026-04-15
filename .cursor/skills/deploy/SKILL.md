---
name: deploy
description: >-
  为初芽记（chuyaji）仓库编译 Go API 二进制、通过 SSH 上传至固定阿里云主机目录并重启服务；远端若无 .env 则从 services/api/.env.example 复制到部署目录后再填密钥。
  在用户提到部署、上线、发布后端、同步到 aliyun、/root/application/chuya，或需要打包 chuyaji-api 时使用。
---

# 初芽记：编译与远程部署

## 约定（本仓库）

| 项 | 值 |
|----|-----|
| SSH 主机别名 | `aliyun`（`~/.ssh/config` 中已配置） |
| 远端根目录 | `/root/application/chuya` |
| API 入口包 | `./services/api/cmd/server` |
| 构建产物名 | `bin/chuyaji-api`（与根目录 [Makefile](../../../Makefile) 一致） |

详细配置、systemd、反代与小程序检查见 [deploy/README.md](../../../deploy/README.md)；systemd 单元模板见 [deploy/chuyaji-api.service](../../../deploy/chuyaji-api.service)（需将路径改为本约定下的 `WorkingDirectory`、`ExecStart`、`EnvironmentFile`）。

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

```bash
scp bin/chuyaji-api aliyun:/root/application/chuya/bin/chuyaji-api
```

若习惯增量同步，可用：

```bash
rsync -avz bin/chuyaji-api aliyun:/root/application/chuya/bin/
```

### 4. 远端启动 / 重启

按服务器实际管理方式选一种：

**已配置 systemd（推荐）**

```bash
ssh aliyun 'sudo systemctl restart chuyaji-api && sudo systemctl status chuyaji-api --no-pager'
```

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

## 常见故障

- **二进制无法执行**：确认交叉编译的 `GOOS`/`GOARCH` 与服务器一致；`chmod +x`。
- **启动即退出**：远端缺少 `CHUYAJI_JWT_SECRET` 或数据库路径不可写；查看 `journalctl -u chuyaji-api -e`。
- **附件/外链异常**：检查 `CHUYAJI_PUBLIC_BASE_URL` 与反代是否指向同一对外域名。

## 与本技能的关系

- 小程序构建、域名与合法域名校验：见 [deploy/README.md](../../../deploy/README.md)，不在此重复。
- 修改 API 后务必重新 `make api`（或交叉编译命令）再上传，避免运行旧二进制。
