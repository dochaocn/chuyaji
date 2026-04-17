<template>
  <view class="page">

    <view v-if="!session.privacyOk" class="notice">
      <text class="notice-title">先阅读隐私说明</text>
      <text class="notice-desc">记录只保存在你的账号范围内，进入正式记录前先确认使用说明。</text>
      <button size="mini" class="notice-btn" @click="goPrivacy">去阅读</button>
    </view>

    <view v-if="!auth.token" class="empty-card">
      <text class="empty-title">先登录后再创建档案</text>
      <text class="empty-desc">登录后可在此创建宝宝或宝妈档案，并与宝宝页数据同步。</text>
      <button class="main-btn" @tap="onLogin">微信登录</button>
    </view>

    <template v-else>

      <!-- ① 阶段头部 -->
      <view class="hero">
        <view class="hero-top">
          <text class="hero-kicker">宝妈工作台</text>
          <text v-if="dashboard?.profile" class="hero-name">{{ dashboard.profile.name || "宝妈" }}</text>
          <text v-if="dashboard?.profile" :class="['stage-badge', stageBadgeClass]">{{ statusLabel }}</text>
        </view>
        <text class="hero-title">{{ stageTitle }}</text>
        <text class="hero-desc">{{ stageDesc }}</text>
      </view>

      <!-- ② 下一步卡片（按需）；未建档时由下方空状态统一承接，避免与「创建档案」重复 -->
      <view
        v-if="nextStep.type !== 'none' && !(nextStep.type === 'create_profile' && !dashboard?.profile)"
        class="next-card"
      >
        <text class="next-label">现在可以做</text>
        <text class="next-title">{{ nextStepTitle }}</text>
        <button class="next-btn" @click="onNextStepAction">{{ nextStepBtnLabel }}</button>
      </view>

      <!-- 未建档：宝宝 / 宝妈二选一（加载中也保留卡片区域，避免空白） -->
      <view v-if="!dashboard?.profile" class="empty-card">
        <template v-if="loading">
          <text class="empty-title">加载中…</text>
          <text class="empty-desc">正在获取档案信息</text>
        </template>
        <template v-else>
          <text class="empty-title">还没有档案</text>
          <text class="empty-desc">可先创建宝宝或宝妈档案，便于分别记录成长与自身状态。</text>
          <view class="empty-actions">
            <button class="main-btn main-btn--flex" @tap="goBabyProfileEdit">创建宝宝档案</button>
            <button class="main-btn main-btn--flex" @tap="goProfileEdit">创建宝妈档案</button>
          </view>
        </template>
      </view>

      <!-- ③ 快捷记录区 -->
      <view v-if="dashboard?.profile" class="shortcut-section">
        <text class="section-label">快速记录</text>
        <view class="shortcut-grid">
          <view
            v-for="item in shortcuts"
            :key="item.value"
            class="shortcut-tile"
            @click="onShortcut(item)"
          >
            <text class="shortcut-title">{{ item.entryLabel }}</text>
            <text class="shortcut-mode">{{ item.mode === 'quick' ? '一键' : '标准' }}</text>
          </view>
        </view>
      </view>

      <!-- ④A 健康摘要：固定 4 格，一行 2 个 -->
      <view v-if="dashboard?.profile && hasSummaryData" class="summary-section">
        <text class="section-label">近况</text>
        <view class="summary-grid">
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastCheckupDate }"
            @click="lastCheckupRecord?.id && openSummaryRecord(lastCheckupRecord.id)"
          >
            <text class="summary-label">最近产检</text>
            <text class="summary-value">{{ lastCheckupDate || "暂无" }}</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastMetricLabel }"
            @click="lastMetricRecord?.id && openSummaryRecord(lastMetricRecord.id)"
          >
            <text class="summary-label">身体指标</text>
            <text class="summary-value">{{ lastMetricLabel || "暂无" }}</text>
            <text v-if="lastMetricLabel && lastMetricDate" class="summary-hint">{{ lastMetricDate }}</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastMoodLabel }"
            @click="lastMoodRecord?.id && openSummaryRecord(lastMoodRecord.id)"
          >
            <text class="summary-label">最近心情</text>
            <text class="summary-value">{{ lastMoodLabel || "暂无" }}</text>
            <text v-if="lastMoodLabel && lastMoodDate" class="summary-hint">{{ lastMoodDate }}</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastSymptomLabel }"
            @click="lastSymptomRecord?.id && openSummaryRecord(lastSymptomRecord.id)"
          >
            <text class="summary-label">最近症状</text>
            <text class="summary-value">{{ lastSymptomLabel || "暂无" }}</text>
            <text v-if="lastSymptomLabel && lastSymptomDate" class="summary-hint">{{ lastSymptomDate }}</text>
          </view>
        </view>
      </view>

      <!-- ④B 最近记录 -->
      <view v-if="dashboard?.profile" class="section">
        <view class="section-head">
          <text class="section-title">最近记录</text>
          <view v-if="(dashboard?.latest_records?.length || 0) > 0" class="section-meta">
            <text class="section-meta-hint" @click="goRecordList">查看全部时间线</text>
          </view>
          <view v-else-if="!loading" class="section-meta">
            <text class="section-meta-hint">在快速记录里新增一条</text>
          </view>
        </view>
        <view v-if="loading" class="placeholder">加载中…</view>
        <view v-else-if="!dashboard?.latest_records?.length" class="placeholder">
          还没有记录，可以从产检或身体状态开始。
        </view>
        <view
          v-for="item in dashboard?.latest_records?.slice(0, 3) || []"
          :key="item.id"
          class="record-card"
          @click="openRecord(item.id)"
        >
          <view class="record-top">
            <text class="record-tag record-tag--mom">宝妈</text>
            <text class="record-type">{{ labelForMotherType(item.record_type) }}</text>
            <text class="record-date">{{ item.occurred_at.slice(0, 10) }}</text>
          </view>
          <text class="record-summary">{{ item.summary || "未填写摘要" }}</text>
        </view>
      </view>

    </template>

    <!-- QuickRecordSheet -->
    <QuickRecordSheet
      v-if="quickTemplate"
      :visible="showQuickSheet"
      :template="quickTemplate"
      @close="showQuickSheet = false"
      @saved="onQuickSaved"
    />

    <!-- 身体指标选择弹层 -->
    <view v-if="showMetricPicker" class="sheet-mask" @click.self="showMetricPicker = false">
      <view class="picker-sheet">
        <text class="picker-title">选择指标类型</text>
        <view
          v-for="mt in metricOptions"
          :key="mt.value"
          class="picker-item"
          @click="onMetricPick(mt.value)"
        >
          <text class="picker-item-text">{{ mt.label }}</text>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { isDevWechatLogin } from "@/api/config";
import { apiMotherDashboard, apiCreateMotherRecord } from "@/api/chuyaji";
import type { MotherRecordItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import {
  labelForMotherType,
  MOTHER_TEMPLATES,
  BODY_METRIC_TYPES,
  type RecordTemplate,
} from "@/utils/motherRecordTypes";
import { getMotherNextStep, getPreferredBodyMetricType } from "@/utils/homeRules";
import QuickRecordSheet from "@/components/QuickRecordSheet.vue";

const auth = useAuthStore();
const session = useSessionStore();
const loading = ref(false);

const dashboard = ref<{
  profile: {
    id: number;
    name: string;
    status?: "pregnant" | "postpartum" | "parenting";
    blood_type?: string;
    delivery_date?: string;
  } | null;
  health_summary?: { status?: string; record_count?: number };
  latest_records: MotherRecordItem[];
} | null>(null);

const showQuickSheet = ref(false);
const quickTemplate = ref<RecordTemplate | null>(null);
const showMetricPicker = ref(false);

// ---------------------------------------------------------------------------
// 阶段
// ---------------------------------------------------------------------------

const statusLabel = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (status === "postpartum") return "产后恢复";
  if (status === "parenting") return "育儿期";
  return "怀孕中";
});

const stageBadgeClass = computed(() => {
  const status = dashboard.value?.profile?.status;
  return status === "pregnant" ? "badge--pre" : "badge--post";
});

const stageTitle = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (!dashboard.value?.profile) return "身体、产检与恢复，随记。";
  if (status === "postpartum") return "产后恢复阶段";
  if (status === "parenting") return "育儿期";
  return "怀孕阶段";
});

const stageDesc = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (!dashboard.value?.profile) return "可先创建宝宝或宝妈档案，再开始记录。";
  if (status === "postpartum") return "恢复节奏不必着急，把身体和心情慢慢记下来。";
  if (status === "parenting") return "育儿期的身体变化同样值得记录。";
  return "这一阶段最重要的是把检查、身体感受和情绪稳稳记下。";
});

// ---------------------------------------------------------------------------
// 固定 4 个快捷入口（不按阶段替换）
// 产检/复查（standard）、症状（standard）、心情（quick）、身体指标（quick 代理）
// ---------------------------------------------------------------------------

const SHORTCUT_KEYS = ["checkup", "symptom", "mood", "__body_metric__"];

const shortcuts = computed(() => {
  return SHORTCUT_KEYS.map((k) => {
    if (k === "__body_metric__") {
      return {
        value: "__body_metric__",
        label: "身体指标",
        entryLabel: "身体指标",
        mode: "quick" as const,
        recommendedFields: [],
        optionalFields: [],
      } satisfies RecordTemplate;
    }
    return MOTHER_TEMPLATES.find((t) => t.value === k)!;
  }).filter(Boolean);
});

const metricOptions = BODY_METRIC_TYPES.map((v) => ({
  value: v,
  label: labelForMotherType(v),
}));

// ---------------------------------------------------------------------------
// 下一步卡片
// ---------------------------------------------------------------------------

const nextStep = computed(() => {
  const hasProfile = !!dashboard.value?.profile;
  const lastAt = dashboard.value?.latest_records?.[0]?.occurred_at ?? null;
  return getMotherNextStep(hasProfile, lastAt);
});

const nextStepTitle = computed(() => {
  switch (nextStep.value.type) {
    case "create_profile": return "还没有宝妈档案";
    case "add_record": return `已经 ${nextStep.value.daysSinceLast} 天没有新记录了`;
    case "check_reminder": return "有即将到来的提醒";
    default: return "";
  }
});

const nextStepBtnLabel = computed(() => {
  switch (nextStep.value.type) {
    case "create_profile": return "去建档";
    case "add_record": return "补一条记录";
    case "check_reminder": return "查看提醒";
    default: return "";
  }
});

// ---------------------------------------------------------------------------
// 健康摘要
// ---------------------------------------------------------------------------

const lastCheckupRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "checkup") ?? null
);
const lastCheckupDate = computed(() => lastCheckupRecord.value?.occurred_at.slice(0, 10) ?? null);

const lastMetricRecord = computed(() => {
  const metricTypes = ["weight", "blood_pressure", "blood_sugar"];
  return dashboard.value?.latest_records?.find((r) => metricTypes.includes(r.record_type)) ?? null;
});

const lastMetricLabel = computed(() => {
  const rec = lastMetricRecord.value;
  if (!rec) return null;
  const p = rec.payload as Record<string, unknown>;
  if (rec.record_type === "weight" && p.kg) return `${p.kg} kg`;
  if (rec.record_type === "blood_pressure" && p.systolic && p.diastolic) return `${p.systolic}/${p.diastolic}`;
  if (rec.record_type === "blood_sugar" && p.mmol_l) return `${p.mmol_l} mmol/L`;
  return rec.summary || labelForMotherType(rec.record_type);
});

const lastMetricDate = computed(() => lastMetricRecord.value?.occurred_at.slice(0, 10) ?? null);

const lastMoodRecord = computed(() => {
  return dashboard.value?.latest_records?.find((r) => r.record_type === "mood") ?? null;
});

const lastMoodLabel = computed(() => {
  const rec = lastMoodRecord.value;
  if (!rec) return null;
  const score = (rec.payload as Record<string, unknown>)?.score;
  return typeof score === "number" ? `${score}/10` : rec.summary || "已记录";
});

const lastMoodDate = computed(() => lastMoodRecord.value?.occurred_at.slice(0, 10) ?? null);

const lastSymptomRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "symptom") ?? null
);

const lastSymptomLabel = computed(() => {
  const rec = lastSymptomRecord.value;
  if (!rec) return null;
  return rec.summary?.trim() || "已记录";
});

const lastSymptomDate = computed(() => lastSymptomRecord.value?.occurred_at.slice(0, 10) ?? null);

const hasSummaryData = computed(() =>
  !!(
    lastCheckupDate.value ||
    lastMetricLabel.value ||
    lastMoodLabel.value ||
    lastSymptomLabel.value
  )
);

// ---------------------------------------------------------------------------
// 事件
// ---------------------------------------------------------------------------

onShow(() => {
  auth.loadToken();
  session.load();
  if (auth.token) loadDashboard();
});

async function loadDashboard() {
  loading.value = true;
  try {
    const data = await apiMotherDashboard(session.motherId || undefined);
    dashboard.value = data;
    if (data.profile?.id) session.setMother(data.profile.id);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

function onNextStepAction() {
  const step = nextStep.value;
  if (step.type === "create_profile") return goProfileEdit();
  if (step.type === "add_record") return goNewRecord();
}

function onShortcut(tmpl: RecordTemplate) {
  if (tmpl.value === "__body_metric__") {
    // 直接进入最近常用指标，无历史则默认体重
    const preferred = getPreferredBodyMetricType(dashboard.value?.latest_records ?? []);
    const preferredTmpl = MOTHER_TEMPLATES.find((t) => t.value === preferred);
    if (preferredTmpl) {
      quickTemplate.value = preferredTmpl;
      showQuickSheet.value = true;
    } else {
      showMetricPicker.value = true;
    }
    return;
  }
  if (tmpl.mode === "quick") {
    quickTemplate.value = tmpl;
    showQuickSheet.value = true;
  } else {
    if (!session.motherId) {
      uni.showToast({ title: "请先创建宝妈档案", icon: "none" });
      return;
    }
    uni.navigateTo({ url: `/pages/mom/record-edit?mother_id=${session.motherId}&type=${tmpl.value}` });
  }
}

function onMetricPick(type: string) {
  showMetricPicker.value = false;
  const tmpl = MOTHER_TEMPLATES.find((t) => t.value === type);
  if (!tmpl) return;
  quickTemplate.value = tmpl;
  showQuickSheet.value = true;
}

async function onQuickSaved(payload: Record<string, unknown>, summary: string) {
  const motherId = Number(session.motherId) || Number(dashboard.value?.profile?.id) || 0;
  if (!motherId) {
    uni.showToast({ title: "缺少宝妈档案", icon: "none" });
    showQuickSheet.value = false;
    return;
  }
  if (motherId !== Number(session.motherId)) {
    session.setMother(motherId);
  }
  try {
    await apiCreateMotherRecord(motherId, {
      record_type: quickTemplate.value!.value,
      occurred_at: new Date().toISOString(),
      summary,
      payload,
    });
    showQuickSheet.value = false;
    uni.showToast({ title: "已保存", icon: "success" });
    await loadDashboard();
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "保存失败", icon: "none" });
  }
}

async function onLogin() {
  if (isDevWechatLogin()) {
    try {
      await auth.loginWithWeChatCode("dev");
      await loadDashboard();
      uni.showToast({ title: "登录成功", icon: "success" });
    } catch (error) {
      console.error(error);
      uni.showToast({ title: "登录失败", icon: "none" });
    }
    return;
  }
  uni.login({
    provider: "weixin",
    success: async (res) => {
      if (!res.code) {
        uni.showToast({ title: "缺少登录 code", icon: "none" });
        return;
      }
      try {
        await auth.loginWithWeChatCode(res.code);
        await loadDashboard();
        uni.showToast({ title: "登录成功", icon: "success" });
      } catch (error) {
        console.error(error);
        uni.showToast({ title: "登录失败", icon: "none" });
      }
    },
    fail: (error) => {
      console.error(error);
      uni.showToast({ title: "登录失败", icon: "none" });
    },
  });
}

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
}

function goBabyProfileEdit() {
  const id = session.babyId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/baby/profile-edit${suffix}` });
}

function goProfileEdit() {
  const id = dashboard.value?.profile?.id || session.motherId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/mom/profile-edit${suffix}` });
}

function goNewRecord() {
  if (!session.motherId) {
    uni.showToast({ title: "请先创建宝妈档案", icon: "none" });
    return;
  }
  uni.navigateTo({ url: `/pages/mom/record-edit?mother_id=${session.motherId}` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/mom/record-detail?id=${id}` });
}

function goRecordList() {
  if (!session.motherId) return;
  uni.navigateTo({ url: `/pages/mom/record-list?mother_id=${session.motherId}` });
}

function openSummaryRecord(id: number | undefined) {
  if (!id) return;
  uni.navigateTo({ url: `/pages/mom/record-detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 72rpx;
}

.notice,
.empty-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.notice {
  background: $cj-warn-bg;
}

.notice-title,
.empty-title {
  display: block;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  font-size: 30rpx;
}

.notice-desc,
.empty-desc {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-secondary;
  font-size: 25rpx;
  line-height: 1.6;
}

.notice-btn,
.main-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}

.empty-actions {
  display: flex;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.main-btn--flex {
  flex: 1;
  margin-top: 0 !important;
  font-size: 24rpx !important;
  padding: 0 12rpx !important;
}

/* ① 阶段头部 */
.hero {
  margin-bottom: $cj-gap-lg;
}

.hero-top {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-bottom: 12rpx;
}

.hero-kicker {
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
}

.hero-name {
  font-size: 22rpx;
  color: $cj-text-muted;
}

.stage-badge {
  padding: 6rpx 18rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  font-weight: 500;
}

.badge--pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.badge--post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.hero-title {
  display: block;
  font-size: 40rpx;
  line-height: 1.35;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.hero-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

/* ② 下一步卡片 */
.next-card {
  background: $cj-warn-bg;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.next-label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
}

.next-title {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  margin-bottom: $cj-gap-md;
}

.next-btn {
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  font-size: 26rpx;
  padding: 0 36rpx !important;
  height: 72rpx !important;
  line-height: 72rpx !important;
}

/* ③ 快捷入口 */
.shortcut-section,
.summary-section,
.section {
  margin-bottom: $cj-gap-md;
}

.section-label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  letter-spacing: 3rpx;
  margin-bottom: $cj-gap-sm;
}

.shortcut-grid {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-sm;
}

.shortcut-tile {
  flex: 1;
  min-width: 140rpx;
  padding: $cj-gap-md;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
}

.shortcut-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.shortcut-mode {
  display: block;
  margin-top: 6rpx;
  font-size: 20rpx;
  color: $cj-text-muted;
}

/* ④A 健康摘要：2 列 × 2 行 */
.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $cj-gap-sm;
}

.summary-card {
  min-width: 0;
  padding: $cj-gap-md;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;

  &:active {
    opacity: 0.75;
  }

  &--empty {
    .summary-value {
      color: $cj-text-muted;
      font-size: 26rpx;
      font-weight: 400;
    }
  }
}

.summary-label {
  display: block;
  color: $cj-text-muted;
  font-size: 22rpx;
}

.summary-value {
  display: block;
  margin-top: 10rpx;
  color: $cj-ink;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
}

.summary-hint {
  display: block;
  margin-top: 6rpx;
  font-size: 20rpx;
  color: $cj-text-muted;
}

/* ④B 最近记录 */
.section {
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $cj-gap-sm;
}

.section-title {
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.section-meta-hint {
  font-size: 22rpx;
  color: $cj-primary;
}

.placeholder {
  padding: 36rpx 0;
  text-align: center;
  color: $cj-text-muted;
  font-size: 24rpx;
}

.record-card {
  padding: $cj-gap-md;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  margin-top: $cj-gap-sm;
}

.record-top {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
}

.record-tag {
  padding: 6rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  font-weight: $cj-fw-title;
}

.record-tag--mom {
  background: $cj-mint-soft;
  color: $cj-text-secondary;
}

.record-type {
  font-size: 28rpx;
  color: $cj-ink;
  font-weight: $cj-fw-title;
}

.record-date {
  margin-left: auto;
  flex-shrink: 0;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.record-summary {
  display: block;
  margin-top: 10rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

/* 身体指标选择弹层 */
.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 999;
  display: flex;
  align-items: flex-end;
}

.picker-sheet {
  width: 100%;
  background: $cj-surface;
  border-radius: $cj-radius-lg $cj-radius-lg 0 0;
  padding: $cj-gap-lg $cj-gap-md calc(env(safe-area-inset-bottom, 0px) + 32rpx);
  box-sizing: border-box;
}

.picker-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-md;
}

.picker-item {
  padding: $cj-gap-md;
  border-radius: $cj-radius-md;
  margin-bottom: $cj-gap-sm;
  background: $cj-surface-2;
}

.picker-item-text {
  font-size: 30rpx;
  color: $cj-ink;
}
</style>
