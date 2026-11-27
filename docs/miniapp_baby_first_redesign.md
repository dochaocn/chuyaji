# 宝宝工作台重构（宝宝优先）实施计划

> 文档版本：v2
> 适用范围：小程序 `apps/miniapp` + 后端 `services/api`（仅首页/记录相关接口）+ 生产服务器 `/root/application/chuya`
> 背景：当前小程序同时承载"宝宝"与"宝妈"两条业务线，首页信息密度高、提示类文案多、日常记录（喂养/睡眠/排便）淹没在"成长快照""建议卡片"等次要模块里。本次聚焦宝宝日常记录，砍掉宝妈线，重构首页布局与记录表单。

---

## 1. 需求

### 1.1 问题

| 现象 | 根因（已核实） |
|------|----------------|
| 首页"AI 味"重、提示文案多 | `pages/baby/home.vue` 里有 `nextStep` 建议卡（`homeRules.ts` 的 `getBabyNextStep`，7 天无记录就弹"已经 N 天没有新记录了"）、`hero-desc`、`"还没有记录，从推荐模板开始。"` 等引导语 |
| 三日常板块不突出 | 首页顺序为：头部 → 快速记录 → 今日汇总 → **成长快照（近况）** → 时间线；成长快照占据三卡 + 三个跳转链接，视觉权重高于日常 |
| 今日汇总信息太少 | 后端 `buildBabyDailySummary` 已算好 `total_ml` / `total_duration_min` / `total_times`，但前端只显示 `count`（`home.vue` 各卡片 `summary-hint`） |
| 时间线只能看 5 条且无图 | `dashboard.go:52` 硬编码 `Limit(5)`；`recordOut` 不含附件，`recordToOut` 只回 payload |
| 记录不能删 | `apiDeleteRecord`（`api/chuyaji.ts:306`）已定义却**无任何调用点**；后端 `DELETE /api/v1/records/:id` 已实现（含附件与提醒级联） |
| 表单折叠多、输入多 | `record-edit.vue` 的"展开更多信息"折叠（`showOptional`）、`QuickRecordSheet.vue` 的 `note-toggle` 折叠；select 字段走 `picker` 弹窗而非平铺 |
| 宝妈线与宝宝线并列 | `pages.json` tabBar 两项：`宝宝` / `宝妈`；`pages/mom/*` 共 2153 行 |

### 1.2 目标

1. 取消宝妈页：前端下线 tab 与 4 个页面，后端接口与历史数据保留。
2. 重构宝宝首页布局，去掉提示类文案，降低"AI 味"。
3. 三大日常（喂养/睡眠/排便）作为唯一主入口；生长降为二级入口。
4. 今日汇总展示总量：总奶量、总喂养时长、喂养次数、总睡眠时长、睡眠次数、排便次数。
5. 成长快照移出首页，作为跳转页（`growth.vue`）顶部内容。
6. 时间线紧贴今日汇总下方，首页显示 20 条，每条尽量完整展示记录内容，图片可点击放大。
7. 记录可删除、可修改。
8. 备注与"更多信息"不再折叠，选填字段平铺，尽量用选择代替输入。
9. 每种记录支持直接调起相机拍照；图片压缩到 1MB 以内。

### 1.3 范围边界

| 端 | 改动 | 不改 |
|----|------|------|
| 小程序前端 | `pages.json`、`pages/baby/*`、`pages/reminders/list.vue`、`components/*`、`utils/recordTypes.ts`、`store/session.ts`、`api/chuyaji.ts`（仅前端调用层） | `pages/family/*`、`pages/privacy/*`、登录/邀请逻辑 |
| 后端 | `handler/dashboard.go`（条数与附件）、`handler/record.go`（`recordOut` 增加附件） | `handler/mother.go`、`model` 里 `Mother/MotherRecord` 表、`router` 的 mom 路由、附件上传实现（已有压缩约定） |
| 管理后台 | 无 | — |
| 生产服务器 | 替换 `/root/application/chuya/bin/chuyaji-api` + 重启 systemd 服务 | `nginx`、`halo-mysql`、`pagespy`、systemd unit 文件、远程数据库内容 |

### 1.4 非目标

- 不删除后端宝妈接口，不迁移/删除 `mothers`、`mother_records` 表与数据（老版本小程序仍在线上，删接口会 404）。
- 不改数据库 schema、不做数据迁移。
- 不改附件上传协议（multipart + 本地落盘）。
- 不涉及发布/审核流程（发布另行处理）。
- 不重做疫苗计划、成长相册、生长曲线的业务逻辑（只调整入口位置）。
- **不修改服务器上的任何数据**：不在远端执行写库操作（`code=dev` 登录、创建/删除记录、迁移脚本一律禁止。`code=dev` 会创建用户，属于写库，必须避免）。
- **不修改 `services/api/internal/model`**：只要模型不变，进程启动时的 `AutoMigrate` 就不会变更线上表结构；任何 Phase 若需要改 model，必须先停下评审。
- 不改服务器的 `nginx` 配置、`systemd` unit 文件、其它目录（`pagespy`、`halo-mysql`）。
- 不把本地数据库上传覆盖远端。

---

## 2. 期望（验收标准）

### 2.1 页面与导航

- [ ] 小程序底部不再有 tabBar（宝妈 tab 移除后仅剩一个 tab 无意义），启动页仍为 `pages/baby/home`
- [ ] 全仓搜索 `pages/mom` 无残留跳转（`reminders/list.vue` 两处改为提示）
- [ ] 首页自上而下顺序：头部 → 三日常快捷记录 → 今日汇总 → 时间线 → 更多入口（生长/相册/疫苗/提醒）
- [ ] 首页不出现："建议""去建档""N 天没有新记录""从推荐模板开始"等提示类文案
- [ ] 成长快照内容（最近体重、本月记录数、最近里程碑）出现在 `pages/baby/growth` 页顶部，首页不再有该面板

### 2.2 今日汇总

- [ ] 喂养卡显示：总奶量 `xx ml`、总时长 `xx 分钟`、`xx 次`（无数据显示 `—`）
- [ ] 睡眠卡显示：总时长（`x 小时 y 分`）、`xx 次`
- [ ] 排便卡显示：总次数 `xx 次`、`xx 条记录`
- [ ] 数值均来自后端 `daily_summary`，不前端重复计算
- [ ] 点三张卡分别跳转到按类型筛选的时间线（保留现有行为）

### 2.3 时间线

- [ ] 首页时间线一次返回并显示 20 条（后端 `records_limit` 参数，默认 20，上限 50）
- [ ] 每条记录展示：类型名、时间（`HH:mm`，非今日显示日期）、摘要、最多 4 个关键字段、备注（若有）、缩略图（最多 3 张，超出显示 `+N`）
- [ ] 点缩略图 → `uni.previewImage` 大图，不触发进入详情
- [ ] 点卡片其他区域 → 进入记录详情页
- [ ] 「查看全部」跳 `pages/baby/record-list`

### 2.4 记录详情 / 删除 / 修改

- [ ] 新增 `pages/baby/record-detail`：展示全部字段（含选填）、全部图片、创建/更新时间
- [ ] 详情页有「修改」「删除」按钮；删除二次确认（`uni.showModal`），成功后返回上一页并刷新列表
- [ ] `familyRole === "read"`（只读成员）看不到修改/删除按钮
- [ ] 删除后该记录的附件与关联提醒一并消失（后端已实现级联，前端只需调 `DELETE`）

### 2.5 记录表单

- [ ] 新建/编辑页无"展开更多信息/收起"折叠，选填字段直接平铺并标注「选填」
- [ ] 「备注」为常显单行输入，非折叠；摘要自动生成，允许覆盖
- [ ] 选择型字段（喂养方式、是否吐奶、性状、是否异常）为平铺 chip，点击即选，不走 picker 弹窗
- [ ] 数值型字段提供常用值快捷 chip（如奶量 90/120/150/180 ml，时长 15/30/45/60 min），点选即填入
- [ ] 图片区有两个入口：「拍照」（`sourceType: ["camera"]`）与「从相册选择」
- [ ] 所有图片在提交前经 `ensureImageUnderMaxBytes` 压到 ≤ 1MB，压缩失败给明确 toast

### 2.6 记录类型收敛

- [ ] 首页快捷记录仅有：喂养、睡眠、排便
- [ ] 新建记录时可选类型仅上述三种（生长记录从 `growth` 页的「记一次生长」进入，类型固定 `growth`）
- [ ] 历史记录中的其它类型（产检/疫苗/体检/发育/疾病等）在时间线与「全部时间线」正常显示、可查看详情，只是不能新建
- [ ] 时间线类型筛选列出"当前存在历史数据的类型"，不硬编码固定列表

### 2.7 提醒

- [ ] 提醒页只列出 `owner_type === "baby"` 的提醒
- [ ] 原先跳宝妈记录的两处（`reminders/list.vue:98`、`:142`）改为 `uni.showToast` 提示并中止

### 2.8 回归安全

- [ ] 家庭主/可读写成员可新建、修改、删除记录；只读成员只能看
- [ ] 邀请加入、退出家庭、切换宝宝不受影响
- [ ] 附件上传/预览、成长相册、疫苗计划、生长曲线行为不变
- [ ] 后端老接口（`/api/v1/dashboard/mother`、`/api/v1/mothers/*`）保持可用，返回结构不变

---

## 3. 实现设计

### 3.1 宝妈页面下线（仅前端）

- `pages.json`：删除 `tabBar` 整段；从 `pages` 数组移除 `pages/mom/home`、`pages/mom/profile-edit`、`pages/mom/record-edit`、`pages/mom/record-list`。
- 删除文件：`pages/mom/*.vue`（4 个）。
- 删除不再被引用的工具：`utils/motherRecordTypes.ts`；`utils/homeRules.ts` 中的 `getMotherNextStep`、`motherStageLabel`、`motherStageBadgeClass`、`getPreferredBodyMetricType`。
- `api/chuyaji.ts`：删除 mom 相关类型与函数（`MotherItem`/`MotherRecordItem`/`apiListMothers`/`apiMotherDashboard`/`apiCreateMother*`/`apiPatchMother*`/`apiDeleteMother*` 等），保留 `ReminderItem`（含 `owner_type`）。
- `store/session.ts`：保留 `motherId` 字段（后端仍在返回 `mothers`，`refreshFamilyContext` 里可继续设置，避免后端 401/404 连锁），但删除对 `apiListMothers` 的**可选**依赖——第一阶段先保留，Phase 6 再决定是否移除，避免一次改太多。
- `pages/reminders/list.vue`：
  - 列表过滤 `owner_type === "baby"`；
  - 第 98、142 行的 mom 跳转分支改为 `uni.showToast({ title: "宝妈记录已下线", icon: "none" })`。

### 3.2 后端：记录带附件 + 时间线条数

**`recordOut` 增加附件**（`services/api/internal/handler/record.go`）：

```go
type recordAttachmentOut struct {
    ID       uint64 `json:"id"`
    URL      string `json:"url"`
    ThumbURL string `json:"thumb_url,omitempty"`
}

type recordOut struct {
    ... // 现有字段不动
    Attachments []recordAttachmentOut `json:"attachments,omitempty"`
}
```

填充策略（避免 N+1）：在需要返回列表的 handler 里，先收集所有 record id，一次 `IN` 查询 `attachments`（`owner_type='baby_record'`），按 `owner_id` 分组后挂回。抽一个 helper：

```go
func (h *Handler) attachRecordAttachments(list []recordOut) []recordOut
```

挂载点：
- `ListRecords`（`record.go:52`）
- `GetRecord`（`record.go:233`）
- `BabyDashboard` 的 `latest_records`（`dashboard.go:74` 附近）

**时间线条数**（`dashboard.go`）：

```go
limit, _ := strconv.Atoi(c.DefaultQuery("records_limit", "20"))
if limit <= 0 || limit > 50 { limit = 20 }
// 替换现有 Limit(5)
```

前端 `apiBabyDashboard` 增加 `records_limit` 查询参数（默认 20）。

**今日汇总补充字段**（`buildBabyDailySummary`，可选但建议一并做，避免前端二次请求）：

```go
"feeding": {"count", "total_ml", "total_duration_min", "last_at"},
"sleep":   {"count", "total_duration_min", "longest_min", "last_at"},
"diaper":  {"count", "total_times", "last_consistency", "last_at"},
```

- `last_at`：最后一次该类记录时间（ISO8601，无则 `null`）
- `sleep.longest_min`：当日单次最长睡眠
- `diaper.last_consistency`：当日最后一条的 `consistency`（无则空串）

### 3.3 首页布局重构（`pages/baby/home.vue`）

目标结构（按渲染顺序）：

| 区块 | 内容 | 说明 |
|------|------|------|
| 1 头部 `WorkbenchHeader` | 宝宝名 / 阶段 / `编辑档案` `家人` | 保留现有组件，去掉 `hero-desc` 类说明文字 |
| 2 今日汇总 | 喂养 / 睡眠 / 排便 三卡（两行三列或一行三列，按窄屏实测） | 每卡：主数值 + 两行明细（总量/时长/次数） |
| 3 快捷记录 | 喂养 / 睡眠 / 排便 三个按钮，打开 `QuickRecordSheet` | 原 `QuickRecordStrip` 精简为 3 项 |
| 4 时间线 | 20 条，卡片式（非时间轴竖线） | 保留时间轴视觉可选，但要压缩高度、提高信息密度 |
| 5 更多入口 | 生长趋势 / 成长相册 / 疫苗计划 / 提醒 | 单列链接列表，不再放"成长快照"三卡 |

删除项：
- `trend-section`（成长快照/近况）整块 → 迁移到 `growth.vue` 顶部
- `WorkbenchHeader` 的 `next-step` 建议卡（`showNextStep` 传 `false` 或移除该能力）
- 文案：`"还没有记录，从推荐模板开始。"` → `"还没有记录"`；`"在快速记录里新增一条"` → 删除
- `home.vue` 里 `nextStep*`、`stageDesc`（若只是引导语）等计算属性与相关样式

### 3.4 时间线卡片设计

单条卡片内容：

```
[类型名]                         [HH:mm / 11-26]
摘要（一句话）
体重 4.1 kg   |   身长 55 cm        ← 关键字段 2~4 个，两列
备注：今天吃得不错                   ← 若有
[图1][图2][图3] +2                  ← 缩略图 48x48，最多 3 张
```

- 关键字段取 `recordTypes.ts` 里 `recommendedFields` 的前 4 个有值的字段（现有 `getRecordPreviewRows` 类似能力，`utils/recordTypes.ts` 已有 `formatFieldValueForDisplay`）
- 图片：`resolvePublicMediaUrl(thumb_url || url)`（`utils/mediaUrl.ts`），点击 `uni.previewImage`（`stopPropagation` 防止冒泡到卡片跳转）
- 卡片点击 → `pages/baby/record-detail?id=...&baby_id=...`

### 3.5 记录详情页（新增）

`pages/baby/record-detail.vue`：

- 拉 `GET /api/v1/records/:id` + 附件（后端已在 Phase 2 带上）
- 展示：类型、发生时间、全部 payload 字段（按模板顺序，含选填）、备注、图片网格
- 底部操作条：`修改`（跳 `record-edit?id=...`）、`删除`（`showModal` 确认 → `DELETE` → `uni.navigateBack` + 触发上一页刷新）
- 只读成员（`session.canWrite === false`）隐藏操作条
- 图片点击 `previewImage`

刷新机制：返回上一页时 uni 不自动刷新，用 `uni.$emit("record:changed")` / `uni.$on` 或在 `onShow` 重新拉 dashboard。

### 3.6 表单平铺化（`record-edit.vue` + `QuickRecordSheet.vue`）

- 移除 `showOptional` 折叠与 `optional-toggle`，选填字段紧随必填字段平铺；字段左侧标签加「选填」小灰标
- `type: "select"` 的字段渲染为 chip 行（`flex-wrap`），选中态用主色描边 + 浅底
- `type: "number"` 且字段属于常用值集合（奶量、时长、次数）时，渲染一行快捷 chip + 一个可输入的 input
- 摘要：默认用 `autoSummary(payload)` 自动生成并实时回显；下方「备注」为常显单行输入（写入 `payload.note` 或 `summary`，需与后端约定：目前备注是 `payload.note`，摘要字段是 `summary`）

> **归属约定（必须遵守）**：`summary` 字段由 `buildSummary()` 生成；用户手填内容写入 `payload.note`，不覆盖 `summary`，避免时间线摘要被人工文字污染。

- 图片区：`拍照`（`sourceType: ["camera"]`）、`相册` 两个按钮；压缩仍走 `ensureImageUnderMaxBytes`

### 3.7 记录类型收敛

- `utils/recordTypes.ts`：新增导出 `ACTIVE_POSTNATAL_TYPES = ["feeding", "sleep", "diaper"]`，供快捷入口与新建页使用
- `growth.vue` 增加「记一次生长」按钮 → `record-edit?baby_id=x&record_type=growth&phase=postnatal`
- 时间线类型筛选：改为从已加载记录中汇总出现过的 `record_type`，动态生成筛选项（保留现有 `RECORD_TYPES` 作为标签来源）

### 3.8 生产部署与本地调试环境（已实测确认）

#### 服务器现状

| 项 | 值 |
|---|---|
| 登录方式 | `ssh aliyun`（免密，已验证可用） |
| 系统 / 架构 | Alibaba Cloud Linux 3，`x86_64`，主机名 `iZ2ze480zyde0t4pfthuv1Z` |
| 服务器 Go | `/usr/bin/go`，`go1.25.10 linux/amd64`（可直接在服务器编译） |
| 二进制 | `/root/application/chuya/bin/chuyaji-api`（39 MB，2026-07-13） |
| 工作目录 | `/root/application/chuya/services/api`（= unit 的 `WorkingDirectory`） |
| 配置 | `/root/application/chuya/services/api/.env`（`CHUYAJI_HTTP_ADDR=:8282`，`CHUYAJI_DEV_MODE=false`，`PUBLIC_BASE_URL=https://...`） |
| 数据库 | `/root/application/chuya/services/api/data/chuyaji.db`（约 164 KB） |
| 附件 | `/root/application/chuya/services/api/data/uploads/` |
| 服务 | `systemd chuyaji-api.service`（`ExecStart=…/bin/chuyaji-api`，`Restart=on-failure`，当前 active） |
| 代码检出 | `/root/application/chuya/src`（分支 `anan`，停在 `887fda0`，落后本地 `anan-grow`）→ **不作为部署源**，部署以本地构建产物为准 |

> 说明：`nginx` 在服务器上是 `inactive`，443 由外层代理终结（本机 DNS 被本地代理接管，解析到 `198.18.1.36`，不影响部署）。因此部署只需换二进制 + 重启服务，不碰反代。

#### 构建（本地交叉编译，CGO 关闭）

本项目 SQLite 驱动是 `github.com/glebarez/sqlite`（纯 Go / modernc），**不需要 CGO**，可在 macOS 上直接交叉编译：

```bash
cd /Users/Zhuanz1/chuyaji
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/chuyaji-api-linux ./services/api/cmd/server
file bin/chuyaji-api-linux   # 期望: ELF 64-bit LSB executable, x86-64
```

#### 部署步骤（每次后端有改动）

```bash
STAMP=$(date +%Y%m%d%H%M)
# 1) 备份远端旧二进制（只读操作，不动数据库）
ssh aliyun "cp /root/application/chuya/bin/chuyaji-api /root/application/chuya/bin/chuyaji-api.bak-$STAMP"
# 2) 上传到临时名
scp bin/chuyaji-api-linux aliyun:/root/application/chuya/bin/chuyaji-api.new
# 3) 替换 + 重启
ssh aliyun "mv /root/application/chuya/bin/chuyaji-api.new /root/application/chuya/bin/chuyaji-api && chmod +x /root/application/chuya/bin/chuyaji-api && systemctl restart chuyaji-api"
# 4) 验证
curl -sS https://www.dochao.com.cn/healthz          # 期望 ok
ssh aliyun "systemctl is-active chuyaji-api; journalctl -u chuyaji-api -n 30 --no-pager"
```

回滚：

```bash
ssh aliyun "cp /root/application/chuya/bin/chuyaji-api.bak-$STAMP /root/application/chuya/bin/chuyaji-api && systemctl restart chuyaji-api"
```

#### 部署红线（每次部署前确认）

- [ ] 本次改动**没有**动 `services/api/internal/model` → `AutoMigrate` 不会改表
- [ ] 没有在远端执行任何写库命令（`code=dev` 登录会建用户，禁止）
- [ ] 没有上传本地 `*.db` 覆盖远端
- [ ] 上传前已备份远端二进制

#### 本地调试环境（已搭好）

- 生产库一致性快照：`services/api/data/prod-snapshot.db`（用 `sqlite3 <db> ".backup …"` 生成，非直接 `cp`，避免半写状态）
- 快照内容：`users=10, babies=1, records=12, mothers=1, mother_records=5, families=1, attachments=24, reminders=1`
- 启动快照后端（端口分开，不干扰日常 dev 实例）：

```bash
cd services/api
CHUYAJI_HTTP_ADDR=:8283 CHUYAJI_DB_PATH=./data/prod-snapshot.db go run ./cmd/server
```

- 因线上与本地 `CHUYAJI_JWT_SECRET` 相同（**这是已知安全问题，见 §7.5**），生产 JWT 可直接打 :8283，看到「安安」真实数据
- 附件 URL 由 `CHUYAJI_PUBLIC_BASE_URL` 生成，指向线上 → 本地预览图片仍从线上加载（只读，无副作用）

### 3.9 H5 预览的 tabBar 兼容（实施中发现的坑）

删掉 `pages.json` 的 `tabBar` 后，**H5 端直接白屏**：

```
TypeError: Cannot read properties of undefined (reading 'height')
    at useShowTabBar (uni-h5.es.js)
    at setup (<Layout>)
```

`uni-h5` 的 `Layout` 组件假定一定存在 tabBar 配置。小程序端没有这个问题（原生 tabBar，不配就没有）。

处理方式（不影响小程序产物）：

```jsonc
// apps/miniapp/src/pages.json
// #ifdef H5
"tabBar": { "list": [{ "pagePath": "pages/baby/home", "text": "宝宝" }] },
// #endif
```

```vue
<!-- apps/miniapp/src/App.vue -->
/* #ifdef H5 */
uni-tabbar { display: none !important; }
/* #endif */
```

验证：小程序产物 `dist/dev/mp-weixin/app.json` 里 `tabBar` 为 `undefined`；H5 页面正常渲染且底部不显示导航条。

---

## 4. 实施步骤

> 原则：小步提交、每步可回滚。Phase 1 与 Phase 2 相互独立，可并行；Phase 3 起依赖 Phase 2 的接口字段。

### Phase 0：准备

- [ ] 在 `anan-grow` 新建工作分支（建议 `feature/baby-first-redesign`）
- [ ] 记录基线：截取当前首页/表单/时间线三张图，存入 `docs/assets/before/`（可选）
- [ ] 确认本地后端（`:8282`）+ H5（`:5173`）+ 小程序编译三套环境可用
- **验收**：三端可正常打开首页并显示真实数据
- **回滚**：无代码改动

### Phase 1：宝妈前端下线（可独立上线）

- [ ] `pages.json` 删除 tabBar 与 4 条 mom 页面注册
- [ ] 删除 `pages/mom/*.vue`
- [ ] `reminders/list.vue` 过滤 baby 提醒 + 两处跳转改 toast
- [ ] 清理 `api/chuyaji.ts` 的 mom 导出与 `utils/motherRecordTypes.ts`
- [ ] `utils/homeRules.ts` 删除 mother 相关函数，保留 baby 所需
- **验收**：小程序无 tabBar；`npm run type-check` 通过；`grep -rn "pages/mom" apps/miniapp/src` 无结果；提醒页可用
- **回滚**：revert 该提交（无后端改动，无数据影响）

### Phase 2：后端小改（附件 + 条数 + 汇总字段）

- [ ] `record.go`：`recordOut` 增加 `attachments`，新增 `attachRecordAttachments` helper
- [ ] 在 `ListRecords` / `GetRecord` / `BabyDashboard` 挂载
- [ ] `dashboard.go`：`records_limit` 参数（默认 20，上限 50）
- [ ] `buildBabyDailySummary` 增加 `last_at` / `longest_min` / `last_consistency`
- [ ] `api/chuyaji.ts`：`apiBabyDashboard` 传 `records_limit=20`；`RecordItem` 类型加 `attachments`
- **验收**：`curl` 校验三个接口返回含 `attachments`；`daily_summary` 含新字段；老版本调用（无 `records_limit`）行为不变
- **回滚**：接口新增字段是向后兼容的，revert 即可

### Phase 2.5：后端上线（可选提前发布）

后端改动全部向后兼容，可先于小程序发布，让老客户端先拿到 `attachments`（虽不使用，无副作用）。

- [ ] `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/chuyaji-api-linux ./services/api/cmd/server`
- [ ] 确认 §3.8 部署红线四条
- [ ] 备份远端二进制 → 上传 → 替换 → `systemctl restart chuyaji-api`
- [ ] 验证 `/healthz` 与线上 `dashboard/baby` 返回结构（用只读 GET + 生产 JWT，**不写库**）
- **验收**：线上 `/api/v1/dashboard/baby` 返回含 `attachments` 字段；`daily_summary.sleep.longest_min` 等新字段存在；服务 `active (running)`；无 5xx
- **回滚**：恢复 `.bak-<时间戳>` 二进制并重启（30 秒内完成）

### Phase 3：首页布局重构

- [ ] 重排区块顺序（头部 → 汇总 → 快捷记录 → 时间线 → 更多入口）
- [ ] 移除成长快照面板、建议卡、提示文案
- [ ] 今日汇总三卡改为多指标展示（直接消费 `daily_summary` 新字段）
- [ ] 快捷记录收敛为 3 项
- [ ] 时间线改卡片布局（含缩略图、关键字段、备注）
- [ ] 「成长快照」三卡迁移到 `growth.vue` 顶部
- **验收**：第 2.1、2.2、2.3 节全部勾选；首屏无提示类文案；空数据时占位合理
- **回滚**：单文件重构，revert `home.vue` 即回到旧布局（`growth.vue` 顶部新增可单独保留）

### Phase 4：记录详情 + 删除/修改

- [ ] 新增 `pages/baby/record-detail.vue` 并在 `pages.json` 注册
- [ ] 首页/全部时间线卡片点击改为进入详情页
- [ ] 详情页接入 `apiGetRecord` / `apiDeleteRecord`；修改跳 `record-edit`
- [ ] 权限控制（`session.canWrite`）与 `record:changed` 刷新事件
- **验收**：第 2.4 节全部勾选；删除后首页与全部时间线都不再出现该记录
- **回滚**：新页面独立，revert 后卡片点击回到直接编辑（旧行为）

### Phase 5：表单平铺化 + 拍照

- [ ] `record-edit.vue` 去掉折叠、select 改 chip、数字加快捷值、备注常显
- [ ] `QuickRecordSheet.vue` 同步（去掉 `note-toggle`）
- [ ] 图片区增加「拍照」按钮
- [ ] `growth.vue` 增加「记一次生长」入口
- **验收**：第 2.5、2.6 节全部勾选；1MB 压缩在真机与模拟器均生效
- **回滚**：表单改动集中在两个文件，可单独 revert

### Phase 6：清理与回归

- [ ] `grep` 检查残留：`showOptional`、`note-toggle`、`nextStep`、`motherRecordTypes`、`pages/mom`
- [ ] 删除无用样式与计算属性
- [ ] 更新 `README.md` 的页面清单（若其中有 mom 页面描述）
- [ ] 全量回归（见第 5 章）
- **验收**：`vue-tsc --noEmit` 通过、`go build ./...` 通过、第 5 章清单全绿
- **回滚**：清理项可单独 revert，不影响功能

### Phase 7：发布

- [ ] 后端最终部署（若 Phase 2.5 之后又改了后端，重新走 §3.8 流程；否则跳过）
- [ ] 小程序构建：`cd apps/miniapp && npm run build:mp-weixin`（需先确认 `manifest.json` 的 `mp-weixin.appid` 已填真实 appid）
- [ ] 开发者工具打开 `dist/build/mp-weixin` → 上传 → 公众平台提审 → 发布
- [ ] 发布后观察线上日志：`ssh aliyun "journalctl -u chuyaji-api -f"`，重点看 500/panic
- **验收**：线上版本可正常新建/编辑/删除记录，时间线带缩略图
- **回滚**：小程序在公众平台「版本回退」；后端回滚见 §3.8

---

## 5. 测试点

### 5.1 类型与构建

- [ ] `cd apps/miniapp && npx vue-tsc --noEmit` 无错误
- [ ] `cd services/api && go build ./...` 通过
- [ ] `npm run dev:mp-weixin` 编译无报错（仅允许 sass legacy 警告）

### 5.2 接口测试（本地 `:8282`）

```bash
TOKEN=$(curl -sS -X POST http://127.0.0.1:8282/api/v1/auth/wechat \
  -H 'Content-Type: application/json' -d '{"code":"dev"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')

# 1) 首页：条数 + 附件 + 汇总字段
curl -sS -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:8282/api/v1/dashboard/baby?records_limit=20" | python3 -m json.tool | head -60

# 2) 记录详情含附件
curl -sS -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:8282/api/v1/records/1" | python3 -m json.tool

# 3) 列表分页仍正常
curl -sS -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:8282/api/v1/babies/1/records?limit=5"

# 4) 删除（用测试记录 id，勿删生产数据）
curl -sS -X DELETE -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:8282/api/v1/records/<TEST_ID>" -w "%{http_code}\n"
```

### 5.3 业务回归（必测链路）

| 链路 | 用例 |
|------|------|
| 新建记录 | 喂养（chip 选奶量）→ 保存 → 首页时间线出现、今日汇总总量增加 |
| 编辑记录 | 详情 → 修改 → 改奶量 → 保存 → 汇总数值同步变化 |
| 删除记录 | 详情 → 删除 → 确认 → 返回列表消失、汇总减少、附件与提醒同步消失 |
| 只读成员 | 用 `read` 角色账号打开详情，无修改/删除按钮；表单页只读 |
| 图片 | 拍照 → 压缩 ≤1MB → 上传 → 时间线显示缩略图 → 点击预览大图 |
| 今日汇总 | 造 0 条 / 1 条 / 多条三种数据，验证 `—` 与单位展示 |
| 生长入口 | 首页更多入口 → 生长趋势 → 顶部快照 + 「记一次生长」→ 保存后曲线更新 |
| 全部时间线 | 类型筛选动态生成；翻页/日期筛选仍正常 |
| 提醒 | 只显示宝宝提醒；宝妈提醒不再出现且无跳转报错 |
| 家庭协作 | 邀请家人、退出家庭、切换宝宝 → 首页数据正确刷新 |
| 历史类型 | 打开一条产检/疫苗历史记录 → 详情正常显示（含附件） |

### 5.4 发版前检查清单

```bash
cd /Users/Zhuanz1/chuyaji
rg -n "pages/mom|motherRecordTypes|apiListMothers|apiMotherDashboard" apps/miniapp/src
rg -n "showOptional|note-toggle" apps/miniapp/src
rg -n "nextStep|getBabyNextStep" apps/miniapp/src/pages/baby/home.vue
rg -n "Limit\(5\)" services/api/internal
cd apps/miniapp && npx vue-tsc --noEmit
cd ../../services/api && go build ./...
```

- [ ] 上述 `rg` 结果为空或只剩有意保留项
- [ ] 后端接口向后兼容（老客户端不带 `records_limit` 仍得 5 条 → **注意**：默认值改 20 会让老客户端也多返回，属预期，需确认管理端/其它调用方不受影响）
- [ ] 小程序包体积增量 < 100KB（删 4 页约 -80KB，新增详情页约 +15KB）

### 5.5 灰度与回滚
- 后端改动只加字段，可先发后端再发小程序（顺序：后端 → 小程序），避免小程序读到缺失字段
- 若首页出现异常，revert Phase 3 的 `home.vue`，其它 Phase 保留
- 宝妈页面删除不可通过配置恢复，需 revert 提交（后端数据始终安全）

### 5.6 部署后端前后验证（生产，只读）

```bash
# 部署前先确认线上数据结构（用生产 JWT 只做 GET）
TOKEN=<线上 JWT>
curl -sS -H "Authorization: Bearer $TOKEN" "https://www.dochao.com.cn/api/v1/dashboard/baby" | python3 -m json.tool | head -30

# 部署后同样命令，比对：
#  - latest_records 长度 5 → 20（含 records_limit 生效）
#  - 每条记录新增 attachments 字段（无图记录不出现该 key，omitempty）
#  - daily_summary.sleep 出现 longest_min，feeding/diaper 出现 last_at

# 服务健康
curl -sS https://www.dochao.com.cn/healthz
ssh aliyun "systemctl is-active chuyaji-api; systemctl show chuyaji-api -p ActiveEnterTimestamp"
```

- [ ] 部署前后 `latest_records` 首条的 `id`/`occurred_at` 一致（说明读的还是同一份数据，未错库）
- [ ] 数据库文件大小与 mtime 未因部署发生异常变化（`ssh aliyun "ls -la /root/application/chuya/services/api/data/"`）

---

## 6. 涉及仓库与文件索引

| 类型 | 路径 | 动作 |
|------|------|------|
| 路由/配置 | `apps/miniapp/src/pages.json` | 删 tabBar、删 mom 页、加 record-detail |
| 页面 | `apps/miniapp/src/pages/baby/home.vue` | 重构主体 |
| 页面 | `apps/miniapp/src/pages/baby/growth.vue` | 顶部加快照 + 记一次生长 |
| 页面 | `apps/miniapp/src/pages/baby/record-edit.vue` | 表单平铺 + 拍照 |
| 页面 | `apps/miniapp/src/pages/baby/record-detail.vue` | 新增 |
| 页面 | `apps/miniapp/src/pages/baby/record-list.vue` | 卡片点击改详情、动态筛选 |
| 页面 | `apps/miniapp/src/pages/reminders/list.vue` | 过滤 baby、改 mom 跳转 |
| 组件 | `apps/miniapp/src/components/WorkbenchHeader.vue` | 去掉 next-step 区块 |
| 组件 | `apps/miniapp/src/components/QuickRecordStrip.vue` | 收敛为 3 项 |
| 组件 | `apps/miniapp/src/components/QuickRecordSheet.vue` | 去折叠、chip 化 |
| 工具 | `apps/miniapp/src/utils/recordTypes.ts` | 新增 `ACTIVE_POSTNATAL_TYPES`、chip 快捷值 |
| 工具 | `apps/miniapp/src/utils/homeRules.ts` | 删 mother 函数 |
| 工具 | `apps/miniapp/src/utils/motherRecordTypes.ts` | 删除 |
| 状态 | `apps/miniapp/src/store/session.ts` | 视清理结果移除 mother 依赖 |
| API | `apps/miniapp/src/api/chuyaji.ts` | 删 mom 导出、加 attachments 类型与 `records_limit` |
| 删除 | `apps/miniapp/src/pages/mom/*.vue` | 4 个文件 |
| 后端 | `services/api/internal/handler/record.go` | `recordOut.attachments` + helper |
| 后端 | `services/api/internal/handler/dashboard.go` | `records_limit`、汇总新字段、挂附件 |
| 文档 | `README.md` | 页面清单若有 mom 描述则更新 |
| 构建产物 | `bin/chuyaji-api-linux` | 交叉编译产物（gitignore 已忽略 `bin/`） |
| 部署参考 | `deploy/chuyaji-api.service` | 与线上 unit 一致，仅作参考不改远端 |

---

## 7. 附录

### 7.1 记录类型与入口映射

| 类型 key | 名称 | 首页快捷 | 其它入口 | 历史记录可见 |
|----------|------|----------|----------|--------------|
| `feeding` | 喂养 | ✅ | — | ✅ |
| `sleep` | 睡眠 | ✅ | — | ✅ |
| `diaper` | 排便 | ✅ | — | ✅ |
| `growth` | 生长 | ❌ | 生长趋势页「记一次生长」 | ✅ |
| `development` | 发育 | ❌ | — | ✅ |
| `checkup` | 体检 | ❌ | — | ✅ |
| `vaccine` | 疫苗 | ❌ | 疫苗计划页 | ✅ |
| `illness` | 疾病 | ❌ | — | ✅ |
| `prenatal_*` | 产检类 | ❌ | — | ✅ |

### 7.2 今日汇总字段来源

| 展示项 | 来源字段 |
|--------|----------|
| 喂养总奶量 | `daily_summary.feeding.total_ml` |
| 喂养总时长 | `daily_summary.feeding.total_duration_min` |
| 喂养次数 | `daily_summary.feeding.count` |
| 睡眠总时长 | `daily_summary.sleep.total_duration_min` |
| 睡眠次数 | `daily_summary.sleep.count` |
| 睡眠最长单次 | `daily_summary.sleep.longest_min` |
| 排便总次数 | `daily_summary.diaper.total_times` |
| 排便记录条数 | `daily_summary.diaper.count` |

### 7.3 归属字段传递链

- 记录：`babyId`（`session.babyId`）→ 列表/详情接口 → `record.baby_id`
- 附件：`owner_type = "baby_record"` + `owner_id = record.id` → 时间线按 `owner_id` 分组
- 提醒：`owner_type = "baby"` + `source_type = "baby_record"` → 提醒页过滤依据

### 7.4 服务器速查

| 用途 | 命令 / 路径 |
|------|-------------|
| 登录 | `ssh aliyun` |
| 服务状态 | `ssh aliyun "systemctl status chuyaji-api"` |
| 实时日志 | `ssh aliyun "journalctl -u chuyaji-api -f"` |
| 二进制目录 | `/root/application/chuya/bin/` |
| 配置 | `/root/application/chuya/services/api/.env` |
| 数据库 | `/root/application/chuya/services/api/data/chuyaji.db` |
| 附件 | `/root/application/chuya/services/api/data/uploads/` |
| 拉取数据库快照 | `ssh aliyun "sqlite3 <db> '.backup /tmp/snap.db'" && scp aliyun:/tmp/snap.db services/api/data/prod-snapshot.db` |

### 7.5 已知问题（本次不修，仅记录）

- **线上与本地共用同一 `CHUYAJI_JWT_SECRET`**：实测同一 JWT 在 `https://www.dochao.com.cn` 与本地 `:8282` 均返回 200。任何拿到仓库 `.env` 的人可伪造线上任意用户 token。建议线上换成独立随机值（会让所有用户重新登录一次），需单独评估窗口期。
- `/root/application/chuya/src` 的检出停在 `887fda0`（分支 `anan`），与线上二进制并不对应；不要以它为部署源，也不要 `git pull` 后直接编译覆盖（可能回退功能）。
- **线上后端落后本地一个提交**：实测 `GET /api/v1/dashboard/baby` 返回的 `daily_summary` 为 `null`（字段缺失），而本地代码有完整实现。说明线上跑的是「家庭协作」那次提交之前的二进制，`97a1931`（今日汇总/多宝切换/计时/相册/疫苗计划）**尚未上线**。Phase 2.5 部署后即可恢复；部署前线上首页「今日汇总」会一直显示 0。

---

## 8. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| v1 | 2026-11-27 | 初稿：确认宝妈仅前端下线、记录类型收敛为三日常、时间线 20 条、提醒过滤宝宝 |
| v2 | 2026-11-27 | 补充 §3.8 生产部署与本地调试环境（ssh aliyun、目录、systemd、交叉编译、红线）、Phase 2.5 后端上线、Phase 7 发布、§5.6 部署验证、§7.4 服务器速查、§7.5 已知问题 |

---

## 9. 实施记录（2026-11-27）

### 9.1 各阶段完成情况

| Phase | 状态 | 关键产出 |
|---|---|---|
| 1 宝妈前端下线 | ✅ | 删 `pages/mom/*` 4 页、`utils/motherRecordTypes.ts`；`pages.json` 去 tabBar 与 mom 路由；`reminders/list.vue` 只留宝宝提醒；`api/chuyaji.ts` 删 12 个 mom 导出；`session.ts` 去掉 `apiListMothers`；`profile-edit.vue` 删掉调用已删接口的 `prefillFromMother()` |
| 2 后端字段 | ✅ | `recordOut.attachments` + `attachRecordAttachments`（一次 IN 查询）；`dashboard?records_limit`（默认 20，上限 50）；`daily_summary` 增 `last_at` / `longest_min` / `last_consistency` |
| 2.5 后端上线 | ✅ | 交叉编译 `bin/chuyaji-api-linux`（37 MB，静态 ELF），备份为 `chuyaji-api.bak-202611271516`，`systemctl restart` 后 `active`，线上 `/healthz` = ok |
| 3 首页重构 | ✅ | 顺序改为 头部 → 今日汇总 → 三快捷 → 时间线（20 条卡片）→ 更多入口；删成长快照/建议卡/引导文案；快照迁到 `growth.vue` 顶部 |
| 4 详情与删除 | ✅ | 新增 `pages/baby/record-detail.vue`；首页与「全部时间线」跳详情；修改/删除按 `session.canWrite` 显隐 |
| 5 表单平铺 | ✅ | `record-edit` 去折叠、select 与数值字段 chip 化、备注常显、拍照/相册双入口；`QuickRecordSheet` 同步 + 新增图片区（保存后由首页统一上传） |
| 6 清理回归 | ✅ | 删 `QuickRecordStrip.vue`、`utils/homeRules.ts`、`WorkbenchHeader` 的 next-step 区块、无用样式；类型筛选改为按已加载数据生成；README 同步 |
| 7 构建发布 | ⏸ | `npm run build:mp-weixin` 与上传提审需要人工操作，未执行 |

### 9.2 实施中发现的两个问题（已处理）

1. **H5 白屏**：删 tabBar 后 `uni-h5` 的 `Layout#useShowTabBar` 读 `undefined.height` 抛错。用 `// #ifdef H5` 给 H5 单独保留单 tab 配置，并用 `App.vue` 的 `#ifdef H5` 样式把 `uni-tabbar` 隐藏；小程序产物 `app.json` 中 `tabBar` 为 `undefined`（见 §3.9）。
2. **部署触发启动迁移**：新二进制启动时 `migrateEnsureFamilies` 为「无家庭成员记录」的用户补建家庭。本次**创建了 8 个空家庭**（`families` id 23–30，对应用户 23–30，各 1 条 `family_members` 记录，`name='我的家庭'`）。**未改动** `babies` / `mothers` / `records` / `attachments` 任何一行；`baby.family_id`、`mother.family_id` 仍为 1。`AutoMigrate` 前后 schema 完全一致（diff 为空）。该行为是新代码的设计逻辑，回滚二进制后下次启动仍会触发，故未回退。

### 9.3 验证证据

- `npx vue-tsc --noEmit`：通过（每阶段执行）
- `go build ./...`：通过；`gofmt` 仅 `admin.go` 为历史未格式化（未改该文件）
- 线上接口：`GET /api/v1/dashboard/baby?records_limit=20` 返回 12 条、含 `attachments`、`daily_summary` 含新字段
- 本地快照库 `:8283` 全链路：创建 → 详情 → PATCH → DELETE（204）→ 再查 404
- H5 冒烟 13 个页面：无 JS 报错（`album` 仅首次 `goto` 因大图加载超时，页面内容正常）

### 9.4 遗留与后续建议

1. **缩略图成本**：附件表有 `thumb_url` 但上传时未生成缩略图，时间线目前直接用原图（单张可达 800 KB、3000×4000）。数据量上来后一屏 20 条会明显吃流量，建议后端上传时生成 300px 缩略图并回填 `thumb_url`。
2. **拍照/压缩需真机验证**：H5 无法覆盖 `uni.chooseImage(sourceType:["camera"])` 与 `uni.compressImage`，需在微信开发者工具或真机上过一遍。
3. **只读成员校验**：`session.canWrite` 依赖 `familyRole` 落库，仅在家庭页/首页刷新时写入；直接进入详情页前需保证已 `refreshFamilyContext`，否则可能短暂显示操作按钮。
4. **`motherId` 残留**：`store/session.ts` 仍保留 `motherId`/`setMother` 字段以便后端兼容，当前无引用，可在确认后端宝妈接口彻底废弃后清理。

### 9.5 第二轮微调（2026-11-27 晚）

| # | 反馈 | 处理 |
|---|---|---|
| 1 | 首页记录要与宝宝当前阶段一致，成长期不要出现怀孕期数据 | `dashboard/baby` 的 `latest_records` 增加 `phase = 当前阶段` 过滤（`dashboard.go`，阶段先算后查）；「查看全部」走 `GET /babies/:id/records` 仍返回两阶段 |
| 2 | 顶部「宝宝」黑字多余 | `pages.json` 里 `pages/baby/home` 的 `navigationBarTitleText` 置空 |
| 3 | 编辑档案/家人占空间，向上合并 | `WorkbenchHeader` 由「两行大按钮」改为「一行：昵称+阶段徽章+出生天数 + 右侧两个小胶囊（👶 编辑 / 👨‍👩‍👧 家人）」；卡片高度约减少 40% |
| 4 | 不需要开始计时 | `QuickRecordSheet` 整块计时能力（状态、storage、tick、按钮）移除；首页 `discardTimerIfBabyMismatch` 一并删除 |
| 5 | 喂养次数无意义，由系统统计 | 喂养模板删掉 `times`；快捷弹层顶部新增「今日第 N 次喂养」（N = 当日已有喂养条数 + 1，数据来自 `daily_summary.feeding.count`） |
| 6 | 夜醒次数、入睡情况改选择 | `night_wake` → 0/1/2/3/4 次及以上；`fall_asleep` → 顺利入睡/需拍哄/需抱睡/哭闹后入睡 |
| 7 | 排便颜色改选择、删是否异常 | `color` → 黄色/绿色/棕色/黑色/带血丝/其他；删除 `abnormal` 字段 |

上线记录：备份 `chuyaji-api.bak-202611271533`，重启后 `active`；线上 `dashboard/baby` 返回 5 条且 `phases=['postnatal']`，`/babies/1/records` 仍 12 条含两阶段。生产包重建为 460 KB、12 页、无 tabBar、产物内无 timer 残留。

兼容说明：历史记录里已有的 `feeding.times` / `diaper.abnormal` / 文本型 `color`、`night_wake` 仍在 payload 中，只是不再渲染对应字段（`color`/`night_wake` 旧值会以原始字符串展示）。

### 9.6 第三轮微调（2026-11-27 晚）

| # | 反馈 | 处理 |
|---|---|---|
| 1 | 首页记录内容重设计：简洁、美观、准确 | 卡片改为「类型 + 时间 → 关键字段行（label 小字 + value 大字，最多 3 项）→ 手写摘要（仅标准类型）→ 备注（左侧细线）→ 缩略图」。日常类型不再展示自动摘要，消除「喂养 100ml」与「奶量 100 ml」重复；新增 `resolveRecordSummary()`，带 `autoSummary` 的类型一律按 payload 重新推导，修正历史脏摘要（例：记录 9 存的 `排便 2 次` 与 `times=1` 矛盾，现显示「排便 1 次」） |
| 2 | 全部时间线：选成长期→选喂养后其它类型选项消失 | 上一轮「按已加载数据生成类型清单」的逻辑在筛选后把清单收缩了。`typeFilterOptions` 改回随阶段固定（`allTypeOptionsForPhase()`），筛选只影响结果不影响选项 |

附带一致性修正：快捷弹层的「备注」此前被当作 summary 存储，现改为写入 `payload.note`，summary 始终由 `autoSummary` 推导，与 `record-edit` 行为统一；`record-detail` 的摘要同样走 `resolveRecordSummary`。

验证：H5 实测卡片文案（喂养 → `奶量 100 ml`；排便 → `次数 1 / 性状 正常`；生长 → `体重 4100 g / 身长 55 cm / 记录方式 社区医院测量`）；筛选交互实测 `全部,成长期,怀孕期,喂养,睡眠,排便,生长,发育,体检,疫苗,疾病` 在选中「成长期」与「喂养」后保持不变；详情页记录 9 显示「排便 1 次」；`vue-tsc` 通过；生产包 460 KB。

### 9.7 筛选栏结构改造（2026-11-27 晚，第三轮反馈复现失败后的加固）

**现象**：用户在「全部时间线」选「成长期」→ 选「喂养」后，认为其它类型选项消失。

**排查**：
- 源码里 `typeFilterOptions = allTypeOptionsForPhase()` 已是静态列表（上一轮已改回）
- H5 逐帧实测（点击前/点成长期/点喂养/再点生长），类型行始终为 `喂养,睡眠,排便,生长,发育,体检,疫苗,疾病`，DOM 高度不变
- 小程序产物 `dist/dev|build/mp-weixin/pages/baby/record-list.wxml` 已确认无收缩逻辑

**结论**：代码逻辑没有复现该现象；最可能的原因是 `scroll-view scroll-x` 在小程序端重新渲染/筛选后滚动位置或内容宽度异常，导致用户只看到部分 chip。

**改造**：把两行筛选从 `scroll-view scroll-x` + `white-space: nowrap` 改为**普通容器 + `flex-wrap` 自动换行**（`.filter-row`）。chip 常驻文档流、不依赖横向滚动，无论筛选如何变化都完整可见；顺带去掉 `.filter-chip` 的 `flex-shrink: 0`。

验证：H5 逐帧复测，点「成长期」→「喂养」→「生长」后两行 chip 全部保持（`init/phase/type/type2` 四帧一致，高度 h=28 / h=65 不变）；`vue-tsc` 通过；dev 与 build 两份产物均已重建（build 460 KB）。

### 9.8 第四轮调整（2026-11-27 晚）

| # | 反馈 | 处理 |
|---|---|---|
| 1 | 成长相册只显示成长期图片 | `ListBabyAttachments`（`attachment.go`）先取宝宝阶段，再给关联记录加 `r.phase = 阶段` 过滤；已出生宝宝只返回成长期图片 |
| 2 | 提醒前端整体删除 | 删 `pages/reminders/list.vue`、`pages.json` 路由、首页「更多入口 → 提醒」与 `goReminders`、`api/chuyaji.ts` 里 `ReminderItem` + 3 个提醒接口；后端 `/api/v1/reminders` 与已有提醒数据保留 |

前端页面数 12 → 11；后端重新部署（备份 `chuyaji-api.bak-202611271541`）。

验证：线上 `GET /api/v1/babies/1/attachments?limit=40` 返回 **2 条**（`diaper` 2026-11-26、`growth` 2026-10-26），怀孕期检查单已不再出现；`rg reminders apps/miniapp/src` 无结果；11 个页面 H5 冒烟 0 报错；生产包 444 KB、11 页。

### 9.9 第五轮调整：疫苗计划页（2026-11-27 晚）

| 需求 | 实现 |
|---|---|
| 默认不展开 | `openKey: GroupKey \| null`，初始 `null`；所有月龄段默认折叠，点击标题展开/收起（同时只展开一段） |
| 增加月龄筛选 | 顶部 chip 行：`全部 / 0–6 月龄 / 6–12 月龄 / 12–24 月龄 / 2–6 岁 / 常见自费苗`；选中后只显示该段并自动展开，再点一次回到全部 |
| 当前月龄优先、已过月龄靠后 | `ageDays` 由出生日期算出当前月龄段；排序 `当前 → 未来段（升序）→ 已过段（近的在前）→ 自费`；当前段标题加「当前月龄」标签 |

其他：头部显示「X 个月 Y 天 · 共 N 项，已接种 M」；筛选 chip 改为 `flex-wrap` 两行常显（与记录筛选一致）；卡片操作按钮由整行宽改为右侧小胶囊。

验证（H5，宝宝 1 个月 2 天）：分组顺序 `0–6 月龄 | 6–12 月龄 | 12–24 月龄 | 2–6 岁 | 常见自费苗`；初始展开行数 **0**；点第二段展开后 **3** 行；点「12–24 月龄」筛选后仅剩该段 **5** 行；`vue-tsc` 通过；生产包 448 KB、11 页。

### 9.10 第六轮调整：时间展示与历史时间修正（2026-11-27 晚）

| 需求 | 实现 |
|---|---|
| 记录时间带年月日，格式 `yyyy-MM-dd HH:mm:ss` | 新增 `utils/datetime.ts` 的 `formatDateTime()`（本地时区、补零到秒），首页卡片、全部时间线、记录详情（含创建时间）、成长相册统一使用 |
| 历史数据改成只知道日期 | 线上 10 条 `12:00:00`（旧表单只选日期、代码写死正午）记录，时间统一置为**当地 `00:00:00`**，存库为 UTC 前一天 16:00；真实时间的两条（11、12）未动 |
| 以后能选具体时间 | `record-edit` 增加「发生时间」`mode="time"` 选择器（精确到分，秒记 0），默认当前时分；编辑历史记录时按本地时区回填 |

**线上数据操作记录（可回滚）**：
- 先 `sqlite3 .backup` 备份到 `/root/application/chuya/services/api/data/chuyaji.db.bak-202611271546`
- `UPDATE records SET occurred_at = 当地日期 00:00:00 (存为 UTC) WHERE 当地时间为 12:00:00`，`changes()=10`
- 复核：线上 12 条记录，1–10 为 `00:00:00`，11 `19:01:25`、12 `15:00:46` 不变；`/healthz` = ok

**注意**：00:00:00 只是「只知道日期」的标记，不代表真实发生时刻；SQL 用 `strftime(..., '+8 hours')` 判定本地时间，避免时区误判。

### 9.11 第七轮调整（2026-11-27 晚）

| 需求 | 实现 |
|---|---|
| 点记录直接进入修改，且与新增界面一致、所有内容可改 | 首页时间线卡片、「全部时间线」行点击均直达 `record-edit`（带 `id` 进入编辑态）；下线中间的 `pages/baby/record-detail`（文件 + 路由 + README 条目），删除能力下沉到编辑页底部「删除这条记录」（仅编辑态 + `canWrite` 显示，二次确认） |
| 记录卡片尽量展示录入信息 | 首页卡片与列表行的关键字段上限从 3 / 5 提升到 **12**（模板里所有已填字段都会展示）；仍保留类型、完整时间、备注、缩略图 |

前端页面数 11 → 10。验证：首页 5 张卡片分别展示 `奶量 100 ml`、`次数 1 / 性状 正常`、`体重 4100 g / 身长 55 cm / 记录方式 社区医院测量`；点击卡片 URL 变为 `record-edit?id=12&baby_id=1`，编辑页含「发生时间 15:00」「保存记录」「删除这条记录」；`vue-tsc` 通过；生产包 440 KB、10 页。

### 9.12 第八轮微调：编辑态隐藏「阶段与类型」

编辑既有记录（`recordId > 0`）时不再显示「阶段与类型」区块，避免误改类型/阶段；新建记录仍保留该区块。实测：`record-edit?id=12` 页面无「阶段与类型」，新建页仍有两项；`vue-tsc` 通过；生产包已重建。

### 9.13 第九轮微调：喂养去时长、睡眠改 h/m 选择

| 需求 | 实现 |
|---|---|
| 喂养取消时长录入 | 喂养模板删除 `duration_min` 字段（`recordTypes.ts`），喂养摘要只按奶量生成；首页今日汇总的喂养卡去掉「0 分钟」行，只保留总奶量与次数 |
| 睡眠用 h/m，录入给时间选择 | 睡眠 `duration_min` 改为 `select`，选项 `30m / 50m / 1h / 1h30m / 2h / 2h30m / 3h / 4h / 6h / 8h`（值为分钟数，兼容历史数字）；新增 `utils/duration.ts` 的 `formatDurationMinutes()`，睡眠摘要与今日汇总（总量、最长单次）统一输出 `2h30m` 这类格式 |

兼容性：历史睡眠记录的 `duration_min` 仍是数字（如 480），`formatFieldValueForDisplay` 用 `String(raw)` 匹配选项值，仍会显示为 `8h`；后端 `payloadFloat` 支持字符串，汇总不受影响。

验证：睡眠表单 chips 顺序 `30m…8h`；喂养表单已无「时长」；今日汇总 `喂养 100 ml / 1 次`、`睡眠 0m / 最长 0m`；`vue-tsc` 通过；生产包已重建。

### 9.14 第十轮微调：今日汇总三卡信息精简

| 需求 | 实现 |
|---|---|
| 睡眠不要「最长 0m」 | 睡眠卡删除 `sleepLongestLabel`，只保留总量 + 次数 |
| 排便去掉「无记录 / 0 条」，改正常/异常 | 移除 `diaperConsistencyLabel` 与 `diaperCount`；新增 `diaperStatusLabel`：当天有排便记录时按最后一条性状判定（`normal` 或未填 → 正常，`loose/hard/watery` → 异常），当天无记录显示 `—` |

现状：`喂养 100 ml / 1 次`、`睡眠 0m / 0 次`、`排便 0 次 / —`。
验证：本地快照库造一条 `consistency=loose` 的当日记录后，接口 `daily_summary.diaper.last_consistency='loose'`、`count=1`，对应判定为「异常」，测试记录已删除；`vue-tsc` 通过；生产包已重建。

### 9.15 睡眠时长改为「时 + 分」双输入

`TemplateField` 新增 `type: "duration"`；睡眠 `duration_min` 从选择器改为双输入框（`h` / `m`），提交时合并为分钟数。

- `record-edit.vue`：`durationKey(key,'h'|'m')` 拆分状态，保存前 `syncDurationFields()` 合并写入 `duration_min`；载入既有记录时 `splitDurationFields()` 回填（90 → 1h 30m）
- `QuickRecordSheet.vue`：同样的双输入与合并逻辑
- 样式用 `.duration-row > .duration-box`（输入框 + 单位共用一个带边框容器），实测两框等宽、不溢出卡片
- 展示仍由 `formatDurationMinutes()` 输出 `1h30m`；存储仍是分钟数，后端汇总与历史数据不受影响

验证（拦截 POST、未写库）：表单填 `1h 30m` → `payload.duration_min=90`、`summary="睡眠 1h30m"`；弹层填 `2h 15m` → `duration_min=135`、`summary="睡眠 2h15m"`。两框实测宽度 158.67px / 158.67px。

### 9.16 今日卡片直接录入

| 需求 | 实现 |
|---|---|
| 今日三张卡点击不再跳「全部时间线」 | `daily-card` 的 `@click` 从 `goDailyRecords(type)` 改为 `onDailyCardTap(type)`，直接打开对应类型的快捷录入弹层 |
| 下方三日常快捷按钮删除 | 移除 `.quick-row` 区块与 `.quick-btn` 样式；`goDailyRecords()` / `formatLocalDate()` 一并删除 |

验证（H5）：`.quick-row` 数量 0；依次点击三张卡 → 弹层标题分别为「喂养」「睡眠」「排便」，URL 保持首页不变；`vue-tsc` 通过；生产包已重建。

### 9.17 删除「沿用上次」

快捷录入弹层里的「沿用上次」按钮移除：删 `reuse-btn`、`onReuseLast()`、`hasLast`、`lastPayload` / `lastSummary` 两个 prop；首页不再为打开弹层去拉 `apiLatestRecord`，`lastQuickPayload` / `lastQuickSummary` 状态一并删除。

验证：弹层文本不含「沿用上次」，只剩「保存」；`rg "reuse|lastPayload|apiLatestRecord"` 在 home 与 QuickRecordSheet 中无结果；`vue-tsc` 通过；生产包已重建。

### 9.18 保存 / 删除按钮并排

编辑既有记录时，底部「保存记录」与「删除」改为同一行等宽两个按钮（`.bottom-actions` flex + `flex:1`）；新建记录（无 `recordId`）仍是通栏「保存记录」。删除按钮文案由「删除这条记录」简化为「删除」。

验证：编辑页两按钮 `x=17/w=173`、`x=198/w=175`，同一 y 基线；新建页只有一个 `w=357` 的保存按钮；`vue-tsc` 通过；生产包已重建。

> 9.18 补充：为兼容小程序 `button` 默认 `width:100%` 的行为，按钮改成 `flex: 1 1 0` + `width: 0 !important` + 固定 88rpx 行高，确保单行等宽（H5 实测 173/175px，同一 y 基线）。
