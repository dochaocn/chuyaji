# 初芽记 · 微信小程序（前端）

uni-app（Vue 3 + TypeScript + Pinia）编译为**微信小程序**。界面为「暖春手帐」亲子向视觉；全局样式与 `$cj-*` 变量集中在 `src/uni.scss` 顶部（受 uni-app 注入顺序限制，勿在业务样式里 `@import` token 文件）。

## 项目简介

供家长在小程序内登录、管理家庭与宝宝、浏览成长时间线、编辑记录与附件、查看生长数据汇总等。数据来自同仓库 [services/api](../../services/api/README.md) 暴露的 REST API。

## 页面与功能

| 页面/模块 | 路径（`pages.json`） | 功能说明 |
|-----------|----------------------|----------|
| 首页 | `pages/index/index` | 隐私入口、微信登录/退出、跳转家庭与时间线 |
| 隐私说明 | `pages/privacy/privacy` | 阅读后本地标记同意（Pinia） |
| 我的家庭 | `pages/families/index` | 列出家庭、进入宝宝列表、创建/加入家庭 |
| 创建家庭 | `pages/families/create` | 输入名称创建家庭 |
| 加入家庭 | `pages/families/join` | 邀请码加入 |
| 宝宝列表 | `pages/baby/list` | 按家庭列出宝宝、进入时间线或新增宝宝 |
| 编辑宝宝 | `pages/baby/edit` | 新建/编辑昵称与日期字段 |
| 成长时间线 | `pages/timeline/index` | 分页拉取记录、进详情/新建记录/生长页 |
| 记录详情 | `pages/record/detail` | 展示记录、附件列表、上传图、编辑/删除 |
| 编辑记录 | `pages/record/edit` | 阶段/类型、日期、摘要、扩展字段、草稿 |
| 生长曲线 | `pages/growth/index` | 从生长类记录汇总体重与日龄、导出 JSON |

## 路由（页面路径）

小程序使用 **配置式路由**，与 `pages.json` 一致：

```text
/pages/index/index
/pages/privacy/privacy
/pages/families/index
/pages/families/create
/pages/families/join
/pages/baby/list
/pages/baby/edit
/pages/timeline/index
/pages/record/detail
/pages/record/edit
/pages/growth/index
```

页面间通过 `uni.navigateTo` 传 query（如 `baby_id`、`family_id`、`id`）。

## 每页接口大纲

封装均在 `src/api/chuyaji.ts` 与 `src/api/http.ts`；上传在 `src/api/upload.ts`。

**登录 / 会话**

- `POST /api/v1/auth/wechat` — `auth.loginWithWeChatCode`：body `{ code }`，换 JWT（开发可用 `code=dev`，须后端 `CHUYAJI_DEV_MODE=true`）。

**首页**

- 无强制接口；登录成功后仅本地存 token。

**家庭**

- `GET /api/v1/families` — `apiListFamilies`
- `POST /api/v1/families` — `apiCreateFamily`
- `POST /api/v1/families/join` — `apiJoinFamily`

**宝宝**

- `GET /api/v1/babies?family_id=` — `apiListBabies`
- `POST /api/v1/babies` — `apiCreateBaby`
- `GET /api/v1/babies/:id` — `apiGetBaby`
- `PATCH /api/v1/babies/:id` — `apiPatchBaby`
- `DELETE /api/v1/babies/:id` — `apiDeleteBaby`

**记录与时间线**

- `GET /api/v1/babies/:id/records` — `apiListRecords`（`limit`、`cursor`）
- `POST /api/v1/babies/:id/records` — `apiCreateRecord`
- `GET /api/v1/records/:id` — `apiGetRecord`
- `PATCH /api/v1/records/:id` — `apiPatchRecord`
- `DELETE /api/v1/records/:id` — `apiDeleteRecord`

**附件**

- `GET /api/v1/records/:id/attachments` — `apiListAttachments`
- `DELETE /api/v1/attachments/:id` — `apiDeleteAttachment`
- `POST /api/v1/records/:id/attachments/upload` — `uploadRecordAttachment`（`uni.uploadFile`，multipart 字段 `file`）

**生长页导出**

- `utils/export.ts` 内分页调用 `apiListRecords` 拼 JSON，无单独后端导出接口。

## 技术栈

Vue 3、TypeScript、Pinia、uni-app、Vite（`@dcloudio/vite-plugin-uni`）、Sass。

## 本地启动

**环境**：Node.js（建议 LTS），包管理器使用 **npm**（仓库脚本以 npm 为准）。

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

若 CLI 提示找不到 `uni`，可尝试：`npx uni -p mp-weixin`。

### 环境变量（`apps/miniapp`）

| 变量 | 含义 |
|------|------|
| `VITE_CHUYAJI_API_BASE` | API 根 URL，如 `http://127.0.0.1:8282` 或局域网 `http://192.168.x.x:8282`（真机必填本机 IP） |
| `VITE_DEV_WECHAT_LOGIN` | 显式 `false` 时强制走真实 `wx.login` code；未设置时开发构建默认倾向 dev 登录（见 `src/api/config.ts`） |

开发环境可编辑 `.env.development`；修改后需重新执行 `npm run dev:mp-weixin`。

### 构建

```bash
npm run build:mp-weixin
```

产物目录以 CLI 输出为准，用于上传微信后台或真机预览。

### 类型检查

```bash
npm run type-check
```

（若 `tsconfig` 与本地 TypeScript 版本不兼容报错，需单独调整配置，与业务代码无关。）

## 部署（小程序侧）

1. 生产构建后，用微信开发者工具上传；在微信公众平台配置 **request / uploadFile / downloadFile** 合法域名为你的 HTTPS API 域名。
2. 生产包勿依赖 `code=dev`；后端关闭 `CHUYAJI_DEV_MODE`，并配置与小程序一致的 AppID/Secret。

服务器与反代见仓库 [deploy/README.md](../../deploy/README.md)。

## 联调常见问题（摘要）

1. **请求超时**：确认本机 `curl` 能访问 `VITE_CHUYAJI_API_BASE` 对应 `/healthz`；真机须用电脑局域网 IP，不可用 `127.0.0.1`。
2. **40029 invalid code**：小程序 AppID 与后端 `CHUYAJI_WECHAT_*` 不一致或 Secret 错误；本地可临时用 dev 登录。
3. **主题修改**：改 `src/uni.scss` 顶部 `$cj-*` 变量后重新编译。

更完整的后端环境变量与接口表见 [services/api/README.md](../../services/api/README.md)。
