# 初芽记（Chuyaji）

从孕期到育儿的全周期记录工具。微信小程序端面向宝宝和宝妈双域，Web 管理后台提供数据总览与分析。

## 仓库结构

```text
chuyaji/
├── apps/
│   ├── miniapp/             # 微信小程序（uni-app、Vue 3、Pinia）
│   └── admin/               # 管理后台（Vue 3、Element Plus、ECharts）
├── services/api/            # Go API（Gin、GORM、SQLite）
├── deploy/                  # systemd 单元文件
├── Makefile                 # 常用命令
└── README.md
```

---

## 功能概览

### 宝宝

- **工作台**：档案摘要、最近记录、快捷录入入口
- **档案管理**：昵称、孕期信息（末次月经、预产期）、出生资料（日期、体重、身长、医院）、喂养方式
- **记录**：孕产期（产检、超声、筛查、NT、糖耐）+ 出生后（喂养、睡眠、便便、生长、发育、体检、疫苗、生病）
- **生长趋势**：体重、身高、头围，叠加 WHO 2006 P3/P50/P97 参考曲线（0–60 月龄，区分性别）
- **疫苗计划**：中国免疫规划程序，出生至 12 月龄共 13 剂次

### 宝妈

- **工作台**：档案摘要、最近记录、快捷录入入口
- **档案管理**：姓名、生日、身高、孕前体重、血型、过敏史、病史、分娩日期
- **记录**：产检、复查、营养、用药、症状、体重、血压、血糖、情绪、产后恢复

### 家庭共享

- 多用户家庭组，角色分 `owner`（管理员）、`write`（可编辑）、`read`（只读）
- 邀请链接加入，支持角色分配、冲突检测、管理员转让、成员管理

### 提醒

- 根据记录自动生成（产检、疫苗、复查等类别）
- 支持推迟、完成、忽略

### 附件

- 文件上传（32MB），自动生成缩略图
- 公开分享链接（`/p/:token`）

### 管理后台

- **总览**：6 项统计卡片 + 最近用户、最近记录
- **用户管理**：分页列表、昵称搜索、用户详情（关联宝宝/宝妈档案）
- **记录浏览**：宝宝记录和宝妈记录，支持按类型、日期等筛选，抽屉详情含附件
- **提醒管理**：统计卡片 + 按状态/分类/归属筛选
- **数据分析**：记录类型分布（饼图）、每日记录趋势（折线图）、日活用户（柱状图），支持 7/30/90 天切换

---

## 技术栈

| 层     | 技术                                                        |
| ------ | ----------------------------------------------------------- |
| 小程序 | uni-app 3.x、Vue 3.4、Pinia、TypeScript、Vite 5.2、Sass    |
| 管理后台 | Vue 3.4、Vue Router 4、Pinia、Element Plus 2.9、ECharts 5.5、Axios、TypeScript、Vite 5.2 |
| 后端   | Go、Gin、GORM、SQLite、JWT（golang-jwt/v5）                   |

---

## 快速开始

### 后端

```bash
cp services/api/.env.example services/api/.env
# 编辑 .env，至少填写 CHUYAJI_JWT_SECRET
go run ./services/api/cmd/server
```

健康检查：`curl -sS http://127.0.0.1:8282/healthz`

### 小程序前端

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

用微信开发者工具打开构建输出中的小程序工程。生产构建：`npm run build:mp-weixin`，并在 `apps/miniapp/.env.production` 中配置 `VITE_CHUYAJI_API_BASE` 为线上 HTTPS API 根地址。

### 管理后台

```bash
cd apps/admin
npm install
npm run dev
```

浏览器访问 `http://localhost:5174/admin/`，使用 `CHUYAJI_ADMIN_PASSWORD` 配置的密码登录。生产构建：`npm run build`，构建产物在 `apps/admin/dist/`。

或使用 Makefile：

```bash
make admin-dev     # 开发
make admin-build   # 生产构建
```

---

## 环境变量

### 后端（`services/api`）

| 变量                                                    | 默认值         | 说明                                    |
| ------------------------------------------------------- | -------------- | --------------------------------------- |
| `CHUYAJI_HTTP_ADDR`                                     | `:8282`        | 监听地址                                |
| `CHUYAJI_JWT_SECRET`                                    | （必填）       | JWT 签名密钥                            |
| `CHUYAJI_DB_PATH`                                       | `./data/chuyaji.db` | SQLite 文件路径                    |
| `CHUYAJI_WECHAT_APP_ID`                                 |                | 微信小程序 AppID                        |
| `CHUYAJI_WECHAT_APP_SECRET`                             |                | 微信小程序 AppSecret                    |
| `CHUYAJI_DEV_MODE`                                      | `false`        | `true` 时允许 `code=dev` 登录及开发签发接口 |
| `CHUYAJI_UPLOAD_DIR`                                    | `./data/uploads` | 附件存储目录                          |
| `CHUYAJI_PUBLIC_BASE_URL`                               |                | 对外生成附件 URL 的 HTTPS 根地址        |
| `CHUYAJI_ADMIN_PASSWORD`                                |                | 管理后台登录密码；为空时禁用管理登录    |

### 小程序前端（`apps/miniapp`）

| 变量                       | 说明                                                                          |
| -------------------------- | ----------------------------------------------------------------------------- |
| `VITE_CHUYAJI_API_BASE`    | 后端 API 根地址，如 `http://127.0.0.1:8282`                                    |
| `VITE_DEV_WECHAT_LOGIN`    | 设为 `true` 时发 `code=dev`（须后端 `CHUYAJI_DEV_MODE=true`）                  |
| `VITE_CHUYAJI_DEV_TOKEN`   | 仅 Vite 开发模式：填入 JWT 可跳过微信登录；建议写在 `.env.local`、勿提交        |

开发时若后端开启 `CHUYAJI_DEV_MODE=true`，可签发 JWT 再粘贴到 `VITE_CHUYAJI_DEV_TOKEN`：

```bash
curl -sS -X POST "http://127.0.0.1:8282/api/v1/auth/dev/token" \
  -H "Content-Type: application/json" \
  -d '{"user_id":1}' | jq -r .token
```

---

## 小程序前端

围绕 **宝宝、宝妈** 双 Tab，默认进入宝宝工作台。

### 页面结构

| 页面             | 路径                          | 说明                                 |
| ---------------- | ----------------------------- | ------------------------------------ |
| 宝宝工作台       | `pages/baby/home`             | 默认首页，档案摘要、最近记录、生长入口 |
| 宝宝档案         | `pages/baby/profile-edit`     | 昵称、孕期、出生资料、喂养等         |
| 宝宝记录编辑     | `pages/baby/record-edit`      | 新增/编辑记录与附件                  |
| 宝宝记录列表     | `pages/baby/record-list`      | 完整记录时间线                       |
| 生长趋势         | `pages/baby/growth`           | 体重/身高/头围趋势图与 WHO 参考曲线  |
| 疫苗计划         | `pages/baby/vaccine-plan`     | 疫苗接种时间表                       |
| 宝妈工作台       | `pages/mom/home`              | 宝妈侧工作台                        |
| 宝妈档案         | `pages/mom/profile-edit`      | 宝妈档案编辑                         |
| 宝妈记录编辑     | `pages/mom/record-edit`       | 新增/编辑宝妈记录与附件              |
| 宝妈记录列表     | `pages/mom/record-list`       | 完整记录时间线                       |
| 提醒列表         | `pages/reminders/list`        | 提醒分类、推迟、完成                  |
| 家庭管理         | `pages/family/index`          | 家庭共享管理                         |
| 加入家庭         | `pages/family/join`           | 通过邀请链接加入家庭                  |
| 隐私说明         | `pages/privacy/privacy`       | 隐私同意                             |

### 状态管理

- `store/auth.ts`：JWT、用户 ID、微信登录流程
- `store/session.ts`：`babyId`、`motherId`、`familyId`、`familyRole`、隐私状态

### 界面风格

`src/uni.scss`「暖春手帐」主题：米杏底（`#fffdf9`）、陶土红主色（`#c96b5c`）、薄荷绿点缀（`#7dab98`）、卡片化圆角轻阴影。

---

## 管理后台

Web 端管理面板，提供数据总览、用户管理、记录查看与数据分析能力。

### 页面结构

| 页面     | 路径               | 说明                                           |
| -------- | ------------------ | ---------------------------------------------- |
| 登录     | `/admin/login`     | 密码登录（独立于小程序微信认证）               |
| 总览     | `/admin/dashboard` | 6 项统计卡片 + 最近用户、最近记录表格          |
| 用户列表 | `/admin/users`     | 分页、昵称搜索                                 |
| 用户详情 | `/admin/users/:id` | 用户信息 + 关联宝宝/宝妈档案                   |
| 宝宝记录 | `/admin/records`   | 按宝宝、孕期阶段、类型、日期筛选；抽屉详情含附件 |
| 宝妈记录 | `/admin/mother-records` | 按宝妈、类型、日期筛选；抽屉详情含附件     |
| 提醒管理 | `/admin/reminders` | 统计卡片 + 按状态/分类/归属筛选                |
| 数据分析 | `/admin/analytics` | 记录类型分布、每日趋势、日活用户；7/30/90 天切换 |

### 认证

管理后台使用独立的密码认证，与小程序微信登录互不影响：

- 通过 `CHUYAJI_ADMIN_PASSWORD` 环境变量设置管理密码
- 登录后签发带 `role: "admin"` 的 JWT（8 小时有效期）
- 未配置密码时管理登录接口返回 403

---

## 后端 API

### API 文档

服务内置 OpenAPI 3.0.3 文档，启动后可访问：

- `GET /openapi.yaml` — 原始 OpenAPI 规范文件
- `GET /docs` — 基于 [Stoplight Elements](https://stoplight.io/open-source/elements) 的交互式 API 文档界面

源文件位于 `services/api/internal/apidocs/openapi.yaml`，通过 Go `embed` 编译进二进制。

### 认证

| 接口                                      | 说明                                 |
| ----------------------------------------- | ------------------------------------ |
| `POST /api/v1/auth/wechat`                | 微信 code 登录                       |
| `POST /api/v1/auth/dev/token`             | 仅 `CHUYAJI_DEV_MODE=true` 时可用   |
| `GET /api/v1/me`                          | 当前用户信息（需 Bearer）            |
| `PATCH /api/v1/me`                        | 更新当前用户信息                     |

### 宝宝与记录

| 接口                                      | 说明           |
| ----------------------------------------- | -------------- |
| `GET/POST /api/v1/babies`                 | 宝宝列表/创建  |
| `GET/PATCH/DELETE /api/v1/babies/:id`     | 宝宝详情/更新/删除 |
| `GET /api/v1/dashboard/baby`              | 宝宝工作台数据 |
| `GET/POST /api/v1/babies/:id/records`     | 宝宝记录列表/创建 |
| `GET /api/v1/babies/:id/records/latest`   | 最近一条记录   |
| `GET /api/v1/babies/:id/growth-series`    | 生长数据序列   |
| `GET/PATCH/DELETE /api/v1/records/:id`    | 记录详情/更新/删除 |

### 宝妈与记录

| 接口                                      | 说明           |
| ----------------------------------------- | -------------- |
| `GET/POST /api/v1/mothers`                | 宝妈列表/创建  |
| `GET/PATCH /api/v1/mothers/:id`           | 宝妈详情/更新  |
| `GET /api/v1/dashboard/mother`            | 宝妈工作台数据 |
| `GET/POST /api/v1/mothers/:id/records`    | 宝妈记录列表/创建 |
| `GET /api/v1/mothers/:id/records/latest`  | 最近一条记录   |
| `GET/PATCH/DELETE /api/v1/mother-records/:id` | 记录详情/更新/删除 |

### 附件

| 接口                                                    | 说明                 |
| ------------------------------------------------------- | -------------------- |
| `GET/POST /api/v1/records/:id/attachments`              | 宝宝记录附件列表/关联 |
| `POST /api/v1/records/:id/attachments/upload`           | 宝宝记录附件上传     |
| `GET/POST /api/v1/mother-records/:id/attachments`       | 宝妈记录附件列表/关联 |
| `POST /api/v1/mother-records/:id/attachments/upload`    | 宝妈记录附件上传     |
| `DELETE /api/v1/attachments/:id`                        | 删除附件             |
| `GET /api/v1/p/:token`                                  | 公开附件访问         |

### 提醒

| 接口                                      | 说明           |
| ----------------------------------------- | -------------- |
| `GET /api/v1/reminders`                   | 提醒列表       |
| `PATCH /api/v1/reminders/:id`             | 更新提醒状态   |
| `DELETE /api/v1/reminders/:id`            | 删除提醒       |

### 家庭

| 接口                                                        | 说明               |
| ----------------------------------------------------------- | ------------------ |
| `GET/POST/PATCH /api/v1/families/current`                   | 当前家庭查询/创建/更新 |
| `POST /api/v1/families/current/leave`                       | 退出家庭           |
| `POST /api/v1/families/current/invites`                     | 创建邀请           |
| `GET /api/v1/invites/:token/preview`                        | 邀请预览           |
| `POST /api/v1/invites/:token/accept`                        | 接受邀请           |
| `PATCH/DELETE /api/v1/families/current/members/:userId`     | 更新/移除成员      |
| `PATCH /api/v1/families/current/members/:userId/nickname`   | 设置成员备注名     |
| `POST /api/v1/families/current/members/:userId/transfer-owner` | 转让管理员     |

### 管理后台 API

管理接口统一前缀 `/api/v1/admin`，需管理员认证（`role: "admin"` JWT）。

| 接口                                                 | 说明               |
| ---------------------------------------------------- | ------------------ |
| `POST /api/v1/admin/login`                           | 管理员密码登录（公开） |
| `GET /api/v1/admin/overview`                         | 总览统计           |
| `GET /api/v1/admin/users`                            | 用户列表（分页、搜索） |
| `GET /api/v1/admin/users/:id`                        | 用户详情           |
| `DELETE /api/v1/admin/users/:id`                     | 删除用户           |
| `GET /api/v1/admin/babies`                           | 宝宝列表           |
| `GET /api/v1/admin/mothers`                          | 宝妈列表           |
| `GET /api/v1/admin/records`                          | 宝宝记录列表       |
| `GET /api/v1/admin/records/:id`                      | 宝宝记录详情       |
| `GET /api/v1/admin/records/:id/attachments`          | 宝宝记录附件       |
| `DELETE /api/v1/admin/records/:id`                   | 删除宝宝记录       |
| `GET /api/v1/admin/mother-records`                   | 宝妈记录列表       |
| `GET /api/v1/admin/mother-records/:id`               | 宝妈记录详情       |
| `GET /api/v1/admin/mother-records/:id/attachments`   | 宝妈记录附件       |
| `DELETE /api/v1/admin/mother-records/:id`            | 删除宝妈记录       |
| `GET /api/v1/admin/reminders`                        | 提醒列表           |
| `GET /api/v1/admin/reminders/stats`                  | 提醒统计           |
| `GET /api/v1/admin/analytics/records`                | 记录类型分析       |
| `GET /api/v1/admin/analytics/activity`               | 用户活跃分析       |
| `GET /api/v1/admin/analytics/growth`                 | 生长数据分析       |

### 数据模型

| 模型             | 说明                                                         |
| ---------------- | ------------------------------------------------------------ |
| `User`           | 微信用户，OpenID 认证                                        |
| `Family`         | 家庭组                                                       |
| `FamilyMember`   | 家庭成员，角色：`owner`、`write`、`read`                     |
| `FamilyInvite`   | 邀请令牌，含过期时间                                         |
| `Baby`           | 宝宝档案，含孕期和出生资料                                   |
| `Mother`         | 宝妈档案                                                     |
| `Record`         | 宝宝记录，`phase` 区分孕产期/出生后，`payload` 为 JSON       |
| `MotherRecord`   | 宝妈记录，`payload` 为 JSON                                  |
| `Attachment`     | 附件，`owner_type` + `owner_id` 多态关联                     |
| `Reminder`       | 提醒，支持推迟                                               |

---

## 部署

### 构建

```bash
# 后端
make api
# 或
go build -o bin/chuyaji-api ./services/api/cmd/server

# 管理后台
make admin-build
```

### 配置

1. 复制 `services/api/.env.example` 为 `.env`
2. 必填：`CHUYAJI_JWT_SECRET`
3. 生产必填：`CHUYAJI_WECHAT_APP_ID`、`CHUYAJI_WECHAT_APP_SECRET`
4. 生产环境关闭 `CHUYAJI_DEV_MODE`
5. 设置 `CHUYAJI_ADMIN_PASSWORD` 以启用管理后台

### 进程与端口

默认监听 `:8282`。生产建议 Nginx 反向代理到 `127.0.0.1:8282`，做 HTTPS 终止。反代需放大上传体积（如 Nginx `client_max_body_size 32m`）。

### 管理后台部署

构建产物在 `apps/admin/dist/`，推荐 Nginx 以 `/admin/` 路径代理：

```nginx
location /admin/ {
    alias /path/to/apps/admin/dist/;
    try_files $uri $uri/ /admin/index.html;
}
```

### systemd

参考 `deploy/chuyaji-api.service`，按实际路径修改 `ExecStart`、`WorkingDirectory`、`EnvironmentFile` 与运行用户。

### 上线检查清单

- [ ] API 域名 TLS 有效，防火墙与反代放行
- [ ] 环境变量：JWT、微信凭证、关闭 `CHUYAJI_DEV_MODE`
- [ ] 管理后台密码 `CHUYAJI_ADMIN_PASSWORD` 已配置
- [ ] 小程序合法域名与 API 一致
- [ ] SQLite 定期备份（`CHUYAJI_DB_PATH` 所在文件）
