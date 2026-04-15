# 初芽记 · 微信小程序（前端）

`apps/miniapp` 现在是一个围绕 `宝宝`、`宝妈` 双 Tab 的微信小程序。默认进入 `宝宝` tab，家庭入口已移除，核心体验改为“工作台 + 档案 + 记录”。

## 页面结构

| 页面 | 路径 | 说明 |
|------|------|------|
| 宝宝工作台 | `pages/baby/home` | 默认首页，展示宝宝档案摘要、最近记录、生长入口 |
| 宝宝档案 | `pages/baby/profile-edit` | 编辑昵称、孕期日期、出生资料、喂养等 |
| 宝宝记录编辑 | `pages/baby/record-edit` | 新增或编辑宝宝记录 |
| 宝宝记录详情 | `pages/baby/record-detail` | 查看记录详情、附件上传/删除 |
| 生长趋势 | `pages/baby/growth` | 查看体重趋势摘要与数据点 |
| 宝妈工作台 | `pages/mom/home` | 展示宝妈档案摘要与最近记录 |
| 宝妈档案 | `pages/mom/profile-edit` | 编辑基础信息、健康资料、状态 |
| 宝妈记录编辑 | `pages/mom/record-edit` | 新增或编辑宝妈记录 |
| 宝妈记录详情 | `pages/mom/record-detail` | 查看详情、附件上传/删除 |
| 隐私说明 | `pages/privacy/privacy` | 本地记录隐私同意状态 |

## 前端 API 封装

- `src/api/chuyaji.ts`
  - 宝宝：`apiListBabies`、`apiCreateBaby`、`apiGetBaby`、`apiPatchBaby`
  - 宝宝记录：`apiListRecords`、`apiCreateRecord`、`apiGetRecord`、`apiPatchRecord`
  - 宝妈：`apiListMothers`、`apiCreateMother`、`apiGetMother`、`apiPatchMother`
  - 宝妈记录：`apiListMotherRecords`、`apiCreateMotherRecord`、`apiGetMotherRecord`
  - 聚合接口：`apiBabyDashboard`、`apiMotherDashboard`
- `src/api/upload.ts`
  - 宝宝记录图片上传：`uploadRecordAttachment`
  - 宝妈记录图片上传：`uploadMotherRecordAttachment`

## 状态管理

- `src/store/auth.ts`
  - 保存 JWT 与当前用户 ID
- `src/store/session.ts`
  - 保存 `babyId`、`motherId` 与隐私同意状态

## 本地启动

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

类型检查：

```bash
npm run type-check
```

## 环境变量

| 变量 | 说明 |
|------|------|
| `VITE_CHUYAJI_API_BASE` | 后端 API 根地址，如 `http://127.0.0.1:8282` |
| `VITE_DEV_WECHAT_LOGIN` | 设为 `true` 时使用 `code=dev`（须后端 `CHUYAJI_DEV_MODE=true`）；默认不设置则用真实 `wx.login` code |

## 界面风格

继续沿用 `src/uni.scss` 中的「暖春手帐」视觉语言：

- 米杏底色
- 陶土红主色
- 薄荷绿点缀
- 卡片化布局
- 圆角与轻阴影

## 当前边界

- 生长页当前为“趋势摘要 + 数据点”版本，后续可继续升级为真实折线图
- 登录以微信 `code` 换票为主；可选 `VITE_DEV_WECHAT_LOGIN=true` + 后端开发模式支持 `code=dev`
