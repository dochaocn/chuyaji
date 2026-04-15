# 初芽记（chuyaji）

当前仓库是一个前后端同仓的微信小程序项目，已经从原来的“家庭 -> 宝宝 -> 时间线”结构重构为 `宝宝`、`宝妈` 双 Tab 工作台。

## 仓库结构

```text
chuyaji/
├── apps/miniapp/          # 微信小程序（uni-app、Vue 3、Pinia）
├── services/api/          # Go API（Gin、GORM、SQLite）
├── deploy/                # 部署示例
├── Dockerfile             # API 镜像构建
├── Makefile               # 常用命令
└── README.md
```

## 产品结构

- `宝宝`
  - 宝宝工作台
  - 宝宝档案
  - 宝宝记录
  - 生长趋势
- `宝妈`
  - 宝妈工作台
  - 宝妈档案
  - 宝妈记录

家庭能力已经移除，不再作为前后端模型和页面的一部分。

## 快速开始

### 启动后端

```bash
cp services/api/.env.example services/api/.env
go run ./services/api/cmd/server
```

### 启动前端

```bash
cd apps/miniapp
npm install
npm run dev:mp-weixin
```

## 相关文档

- [apps/miniapp/README.md](apps/miniapp/README.md)
- [services/api/README.md](services/api/README.md)
- [deploy/README.md](deploy/README.md)
