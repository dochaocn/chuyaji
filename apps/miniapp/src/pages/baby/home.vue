<template>
  <view class="page">
    <view v-if="!session.privacyOk" class="notice">
      <text class="notice-title">先阅读隐私说明</text>
      <text class="notice-desc">记录只保存在你的账号范围内，进入正式记录前先确认使用说明。</text>
      <button size="mini" class="notice-btn" @click="goPrivacy">去阅读</button>
    </view>

    <view v-if="!auth.token && auth.sessionBooting" class="empty-card">
      <text class="empty-title">正在准备运行环境…</text>
    </view>

    <view v-else-if="!auth.token && auth.sessionBootError" class="empty-card">
      <text class="empty-title">{{ auth.sessionBootError }}</text>
      <button class="main-btn" @click="retryWeChatSession">重试</button>
    </view>

    <view v-else-if="!auth.token" class="empty-card">
      <text class="empty-title">未登录</text>
      <button class="main-btn" @click="retryWeChatSession">重新登录</button>
    </view>

    <template v-else>
      <WorkbenchHeader
        :badge-label="dashboard?.profile ? babyStageLabel : undefined"
        :badge-class="babyStageBadgeClass"
        :identity-label="resolvedIdentity"
        :name="dashboard?.profile?.nickname"
        :stage-title="stageTitle"
        :profile-label="dashboard?.profile ? (session.canWrite ? '编辑档案' : '查看档案') : '创建档案'"
        profile-icon="👶"
        :can-switch="babies.length > 1"
        @profile="goProfileEdit"
        @family="goFamily"
        @identity="goFamily"
        @switch="onSwitchBaby"
      />

      <view v-if="!dashboard?.profile && !loading" class="empty-card">
        <view class="empty-actions">
          <button class="main-btn" @click="goProfileEdit">创建宝宝档案</button>
          <button v-if="showFamilyFeatures" class="family-btn" @click="goFamily">家人</button>
        </view>
      </view>

      <template v-else-if="dashboard?.profile">
        <!-- 今日汇总：总量优先 -->
        <view class="panel">
          <view class="panel-head">
            <text class="panel-title">今日</text>
            <text class="panel-date">{{ dailyDateLabel }}</text>
          </view>
          <view class="daily-grid">
            <view class="daily-card" @click="onDailyCardTap('feeding')">
              <text class="daily-label">喂养</text>
              <text class="daily-main">{{ feedingTotalMl }}</text>
              <view class="daily-meta">
                <text>{{ feedingCount }} 次</text>
              </view>
            </view>
            <view class="daily-card" @click="onDailyCardTap('diaper')">
              <text class="daily-label">排便</text>
              <text class="daily-main">{{ diaperTotalLabel }}</text>
              <view class="daily-meta">
                <text>{{ diaperStatusLabel }}</text>
              </view>
            </view>
            <view class="daily-card" @click="onDailyCardTap('sleep')">
              <text class="daily-label">睡眠</text>
              <text class="daily-main">{{ sleepTotalLabel }}</text>
              <view class="daily-meta">
                <text>{{ sleepCount }} 次</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 时间线 -->
        <view class="panel">
          <view class="panel-head">
            <text class="panel-title">记录</text>
            <text class="panel-link" @click="goTimeline">查看全部</text>
          </view>

          <view v-if="loading && !timelineRows.length" class="tl-empty">加载中…</view>
          <view v-else-if="!timelineRows.length" class="tl-empty">还没有记录</view>

          <view v-else class="tl-list">
            <view v-for="row in timelineRows" :key="row.item.id" class="tl-card" @click="openRecord(row.item.id)">
              <view class="tl-head">
                <text class="tl-type">{{ row.typeLabel }}</text>
                <text class="tl-time">{{ row.timeLabel }}</text>
              </view>

              <view v-if="row.metrics.length" class="tl-metrics">
                <view v-for="m in row.metrics" :key="m.label" class="tl-metric">
                  <text class="tl-metric-label">{{ m.label }}</text>
                  <text class="tl-metric-value">{{ m.value }}</text>
                </view>
              </view>

              <text v-if="row.summary" class="tl-summary">{{ row.summary }}</text>
              <text v-if="row.note" class="tl-note">备注 · {{ row.note }}</text>

              <view v-if="row.thumbs.length" class="tl-thumbs">
                <image
                  v-for="(thumb, idx) in row.thumbs"
                  :key="thumb.id"
                  class="tl-thumb"
                  :src="thumb.src"
                  mode="aspectFill"
                  @click.stop="previewRow(row, idx)"
                />
                <text v-if="row.moreCount" class="tl-thumb-more">+{{ row.moreCount }}</text>
              </view>
            </view>
          </view>
        </view>

        <!-- 更多入口 -->
        <view class="more-list">
          <view class="more-item" @click="goGrowth">
            <text class="more-text">生长趋势</text>
            <text class="more-chev">›</text>
          </view>
          <view class="more-item" @click="goAlbum">
            <text class="more-text">成长相册</text>
            <text class="more-chev">›</text>
          </view>
          <view class="more-item" @click="goVaccinePlan">
            <text class="more-text">疫苗计划</text>
            <text class="more-chev">›</text>
          </view>
        </view>
      </template>
    </template>

    <QuickRecordSheet
      v-if="quickTemplate"
      :visible="showQuickSheet"
      :template="quickTemplate"
      :baby-id="Number(session.babyId) || Number(dashboard?.profile?.id) || 0"
      :today-count="feedingTodayCount"
      @close="showQuickSheet = false"
      @saved="onQuickSaved"
    />
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { SHOW_FAMILY_FEATURES } from "@/api/config";
import {
  apiBabyDashboard,
  apiCreateRecord,
  apiListBabies,
  apiMe,
  type Baby,
  type BabyDailySummary,
  type RecordItem,
} from "@/api/chuyaji";
import { uploadRecordAttachment } from "@/api/upload";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import {
  ACTIVE_POSTNATAL_TYPES,
  BABY_TEMPLATES,
  labelForType,
  getBabyRecordPreviewRows,
  resolveRecordSummary,
  templateForType,
  type RecordTemplate,
} from "@/utils/recordTypes";
import { calcGestation, calcAgeDays, isPostnatal } from "@/utils/gestation";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import { formatDateTime } from "@/utils/datetime";
import { formatDurationMinutes } from "@/utils/duration";
import QuickRecordSheet from "@/components/QuickRecordSheet.vue";
import WorkbenchHeader from "@/components/WorkbenchHeader.vue";

const auth = useAuthStore();
const session = useSessionStore();
const showFamilyFeatures = SHOW_FAMILY_FEATURES;
const loading = ref(false);
/** 首页展示用的身份；家庭接口同步失败时用 /me 的昵称兜底 */
const identityLabel = ref("");
const resolvedIdentity = computed(() => identityLabel.value || session.myNickname);
let identityChecked = false;

const dashboard = ref<{
  profile: {
    id: number;
    nickname: string;
    birth_date?: string;
    edd_date?: string;
    lmp_date?: string;
  } | null;
  phase_summary?: { stage: "prenatal" | "postnatal"; record_count: number };
  latest_records: RecordItem[];
  growth_summary?: Record<string, unknown>;
  daily_summary?: BabyDailySummary;
} | null>(null);

const babies = ref<Baby[]>([]);
const showQuickSheet = ref(false);
const quickTemplate = ref<RecordTemplate | null>(null);

const isPostnatalStage = computed(() => isPostnatal(dashboard.value?.profile?.birth_date));

const babyStageLabel = computed(() => (isPostnatalStage.value ? "成长期" : "怀孕期"));
const babyStageBadgeClass = computed(() => (isPostnatalStage.value ? "badge--post" : "badge--pre"));

const ageDays = computed(() => calcAgeDays(dashboard.value?.profile?.birth_date));
const gestationResult = computed(() => {
  const p = dashboard.value?.profile;
  return calcGestation(p?.edd_date, p?.lmp_date);
});

const stageTitle = computed(() => {
  if (!dashboard.value?.profile) return "还没有宝宝档案";
  if (isPostnatalStage.value) return `出生第 ${ageDays.value ?? 0} 天`;
  if (gestationResult.value.available) return `孕 ${gestationResult.value.weeks} 周 ${gestationResult.value.days} 天`;
  return "宝宝档案";
});

/** 三个日常记录模板（顺序固定：喂养 / 睡眠 / 排便） */
const dailyTemplates = computed<RecordTemplate[]>(() =>
  ACTIVE_POSTNATAL_TYPES.map((t) => BABY_TEMPLATES.postnatal.find((tmpl) => tmpl.value === t)).filter(
    (t): t is RecordTemplate => !!t
  )
);

const dailySummary = computed(() => dashboard.value?.daily_summary);

const dailyDateLabel = computed(() => {
  const d = dailySummary.value?.date;
  if (!d) return "";
  const parts = d.split("-");
  if (parts.length !== 3) return d;
  return `${Number(parts[1])} 月 ${Number(parts[2])} 日`;
});

const feedingTotalMl = computed(() => {
  const ml = dailySummary.value?.feeding?.total_ml ?? 0;
  return `${ml} ml`;
});
const feedingCount = computed(() => dailySummary.value?.feeding?.count ?? 0);

const sleepTotalLabel = computed(() => formatDurationMinutes(dailySummary.value?.sleep?.total_duration_min ?? 0));
const sleepCount = computed(() => dailySummary.value?.sleep?.count ?? 0);

const diaperTotalLabel = computed(() => {
  const times = dailySummary.value?.diaper?.total_times ?? 0;
  return `${times} 次`;
});

/** 当日排便状态：有异常性状记「异常」，其余（含未填性状）记「正常」，当天无记录显示 — */
const diaperStatusLabel = computed(() => {
  const summary = dailySummary.value?.diaper;
  if (!summary?.count) return "—";
  const raw = summary.last_consistency || "";
  return raw && raw !== "normal" ? "异常" : "正常";
});
const feedingTodayCount = computed(() => dailySummary.value?.feeding?.count ?? 0);

type TimelineRow = {
  item: RecordItem;
  typeLabel: string;
  timeLabel: string;
  /** 关键字段（来自 payload，最多 3 条），保证展示与记录内容一致 */
  metrics: { label: string; value: string }[];
  /** 仅标准类型展示手写摘要，日常类型不再重复显示自动摘要 */
  summary: string;
  note: string;
  thumbs: { id: number; src: string; full: string }[];
  moreCount: number;
};

const MAX_THUMBS = 3;

const timelineRows = computed<TimelineRow[]>(() => {
  const list = dashboard.value?.latest_records ?? [];
  return list.map((item) => {
    const tmpl = templateForType(item.phase, item.record_type);
    // 尽量把录入过的字段都摊在卡片上，一眼能看清这条记录的内容
    const metrics = getBabyRecordPreviewRows(
      {
        phase: item.phase,
        record_type: item.record_type,
        payload: item.payload as Record<string, unknown>,
      },
      12
    );
    const payload = (item.payload || {}) as Record<string, unknown>;
    const note = String(payload.note || "").trim();
    // 日常类型（有 autoSummary）不展示摘要，避免与关键字段重复或历史脏摘要冲突
    const summary = tmpl?.mode === "standard" ? resolveRecordSummary(item) : "";
    const atts = item.attachments ?? [];
    const thumbs = atts.slice(0, MAX_THUMBS).map((a) => ({
      id: a.id,
      src: resolvePublicMediaUrl(a.thumb_url || a.url),
      full: resolvePublicMediaUrl(a.url),
    }));
    return {
      item,
      typeLabel: labelForType(item.phase, item.record_type),
      timeLabel: formatDateTime(item.occurred_at),
      metrics,
      summary,
      note,
      thumbs,
      moreCount: Math.max(0, atts.length - MAX_THUMBS),
    };
  });
});

function previewRow(row: TimelineRow, idx: number) {
  const urls = row.thumbs.map((t) => t.full);
  uni.previewImage({ current: urls[idx], urls });
}

onShow(async () => {
  auth.loadToken();
  session.load();
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
  if (auth.token) {
    await session.refreshFamilyContext();
    await syncIdentity();
    loadDashboard();
  }
});

/**
 * 身份以家庭接口同步到 session 为准；万一家庭成员里取不到，
 * 再用当前用户昵称兜底，避免家人页已有身份、首页仍显示未设置。
 */
async function syncIdentity() {
  const sessionName = session.myNickname;
  if (sessionName) {
    identityLabel.value = sessionName;
    return;
  }
  if (identityLabel.value || identityChecked) return;
  identityChecked = true;
  try {
    const me = await apiMe();
    const name = String(me.nickname || "").trim();
    identityLabel.value = name;
    if (name) session.setMyNickname(name);
  } catch (error) {
    console.error(error);
  }
}

async function retryWeChatSession() {
  await auth.ensureWeChatSession();
  if (auth.token) {
    await loadDashboard();
  }
}

async function loadDashboard() {
  loading.value = true;
  try {
    const [data, babyPage] = await Promise.all([
      apiBabyDashboard(session.babyId || undefined),
      apiListBabies().catch(() => ({ items: [] as Baby[] })),
    ]);
    dashboard.value = data;
    babies.value = babyPage.items || [];
    if (data.profile?.id) {
      session.setBaby(data.profile.id);
    }
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

function onSwitchBaby() {
  if (babies.value.length <= 1) return;
  const itemList = babies.value.map((b) => (b.id === session.babyId ? `${b.nickname}（当前）` : b.nickname));
  uni.showActionSheet({
    itemList,
    success: (res) => {
      const picked = babies.value[res.tapIndex];
      if (!picked || picked.id === session.babyId) return;
      session.setBaby(picked.id);
      void loadDashboard();
    },
  });
}

function onDailyCardTap(value: "feeding" | "sleep" | "diaper") {
  const tmpl = dailyTemplates.value.find((t) => t.value === value);
  if (tmpl) onDailyRecord(tmpl);
}

function onDailyRecord(tmpl: RecordTemplate) {
  const bid = Number(session.babyId) || Number(dashboard.value?.profile?.id) || 0;
  if (!bid) {
    uni.showToast({ title: "请先创建宝宝档案", icon: "none" });
    return;
  }
  quickTemplate.value = tmpl;
  showQuickSheet.value = true;
}

async function onQuickSaved(payload: Record<string, unknown>, summary: string, photos: string[] = []) {
  const bid = Number(session.babyId) || Number(dashboard.value?.profile?.id) || 0;
  if (!bid) {
    uni.showToast({ title: "缺少宝宝档案", icon: "none" });
    showQuickSheet.value = false;
    return;
  }
  if (bid !== Number(session.babyId)) session.setBaby(bid);
  try {
    const created = await apiCreateRecord(bid, {
      phase: "postnatal",
      record_type: quickTemplate.value!.value,
      occurred_at: new Date().toISOString(),
      summary,
      payload,
    });
    let failed = 0;
    for (const path of photos) {
      try {
        await uploadRecordAttachment(created.id, path);
      } catch (e) {
        console.error(e);
        failed++;
      }
    }
    showQuickSheet.value = false;
    uni.showToast({ title: failed ? `${failed} 张图片上传失败` : "已保存", icon: failed ? "none" : "success" });
    await loadDashboard();
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "保存失败", icon: "none" });
  }
}

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
}

function goFamily() {
  uni.navigateTo({ url: "/pages/family/index" });
}

function goProfileEdit() {
  const id = dashboard.value?.profile?.id || session.babyId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/baby/profile-edit${suffix}` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${id}&baby_id=${session.babyId}` });
}

function goGrowth() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/growth?baby_id=${session.babyId}` });
}

function goVaccinePlan() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/vaccine-plan?baby_id=${session.babyId}` });
}

function goAlbum() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/album?baby_id=${session.babyId}` });
}

function goTimeline() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/record-list?baby_id=${session.babyId}` });
}

</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 80rpx;
}

.notice,
.empty-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.notice {
  background: linear-gradient(135deg, $cj-warn-bg 0%, #fff8f0 100%);
  border-color: $cj-warn-border;
}

.notice-title,
.empty-title {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  margin-bottom: 12rpx;
}

.notice-desc {
  display: block;
  font-size: 24rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
  margin-bottom: 18rpx;
}

.notice-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8rpx 28rpx;
  font-size: 22rpx;
  color: #fffefb;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
  border-radius: $cj-radius-pill;
}

.main-btn {
  margin-top: 8rpx;
  padding: 20rpx 0;
  font-size: 28rpx;
  color: #fffefb;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
  border-radius: $cj-radius-pill;
  box-shadow: $cj-shadow-soft;
}

.empty-actions {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.empty-actions .main-btn {
  flex: 1;
  min-width: 0;
}

.family-btn {
  flex-shrink: 0;
  margin-top: 8rpx;
  padding: 20rpx 34rpx;
  font-size: 28rpx;
  color: $cj-primary-dark;
  background: $cj-surface;
  border: 1rpx solid $cj-border;
  border-radius: $cj-radius-pill;
}

.panel {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 22rpx;
}

.panel-title {
  font-size: 32rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 1rpx;
}

.panel-date {
  font-size: 22rpx;
  color: $cj-text-muted;
}

.panel-link {
  font-size: 24rpx;
  color: $cj-primary;
}

/* 今日汇总 */
.daily-grid {
  display: flex;
  gap: 14rpx;
}

.daily-card {
  flex: 1;
  min-width: 0;
  padding: 20rpx 18rpx;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;

  &:active {
    opacity: 0.85;
  }
}

.daily-label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  margin-bottom: 10rpx;
}

.daily-main {
  display: block;
  font-size: 36rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  line-height: 1.2;
  margin-bottom: 10rpx;
}

.daily-meta {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
  font-size: 20rpx;
  color: $cj-text-secondary;
}

/* 时间线 */
.tl-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.tl-card {
  padding: 20rpx 22rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;

  &:active {
    background: $cj-surface-2;
  }
}

.tl-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12rpx;
}

.tl-type {
  font-size: 27rpx;
  font-weight: 600;
  color: $cj-ink;
  letter-spacing: 0.3rpx;
}

.tl-time {
  flex-shrink: 0;
  font-size: 21rpx;
  color: $cj-text-muted;
  font-variant-numeric: tabular-nums;
}

.tl-metrics {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 6rpx 22rpx;
  margin-top: 10rpx;
}

.tl-metric {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
  max-width: 100%;
}

.tl-metric-label {
  font-size: 20rpx;
  color: $cj-text-muted;
}

.tl-metric-value {
  font-size: 26rpx;
  color: $cj-text;
  font-weight: 500;
}

.tl-summary {
  display: block;
  margin-top: 10rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tl-note {
  display: block;
  margin-top: 10rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.45;
  padding-left: 12rpx;
  border-left: 3rpx solid rgba(201, 107, 92, 0.35);
}

.tl-thumbs {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-top: 14rpx;
}

.tl-thumb {
  width: 112rpx;
  height: 112rpx;
  border-radius: $cj-radius-sm;
  background: $cj-surface-2;
}

.tl-thumb-more {
  font-size: 21rpx;
  color: $cj-text-muted;
  padding-left: 4rpx;
}

.tl-empty {
  padding: 40rpx 0;
  text-align: center;
  font-size: 24rpx;
  color: $cj-text-muted;
}

/* 更多入口 */
.more-list {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  overflow: hidden;
}

.more-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 26rpx $cj-gap-md;

  & + & {
    border-top: 1rpx solid $cj-border-faint;
  }

  &:active {
    background: $cj-surface-2;
  }
}

.more-text {
  font-size: 27rpx;
  color: $cj-text;
}

.more-chev {
  font-size: 30rpx;
  color: $cj-text-muted;
}
</style>
