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

      
      <view class="hero">
        <view class="hero-top">
          <text class="hero-kicker">宝妈工作台</text>
          <text v-if="dashboard?.profile" :class="['stage-badge', stageBadgeClass]">{{ statusLabel }}</text>
        </view>
        <view class="hero-headline">
          <text v-if="dashboard?.profile" class="hero-name">{{ dashboard.profile.name || "宝妈" }}</text>
          <text
            class="hero-title"
            :class="{ 'hero-title--solo': !dashboard?.profile }"
          >{{ stageTitle }}</text>
        </view>
        <text class="hero-desc">{{ stageDesc }}</text>
      </view>

      
      <view
        v-if="nextStep.type !== 'none' && !(nextStep.type === 'create_profile' && !dashboard?.profile)"
        class="next-card"
      >
        <text class="next-label">现在可以做</text>
        <text class="next-title">{{ nextStepTitle }}</text>
        <button class="next-btn" @click="onNextStepAction">{{ nextStepBtnLabel }}</button>
      </view>

      
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

      
      <view v-if="dashboard?.profile && hasSummaryData" class="summary-section">
        <text class="section-label">近况</text>
        <view class="summary-grid">
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastCheckupRecord }"
            @click="lastCheckupRecord?.id && openSummaryRecord(lastCheckupRecord.id)"
          >
            <view class="summary-card-top">
              <text class="summary-label">最近产检</text>
              <text v-if="lastCheckupRecord && lastCheckupDate" class="summary-hint-inline">{{ lastCheckupDate }}</text>
            </view>
            <view v-if="lastCheckupPreview.length" class="summary-key-grid">
              <view v-for="(kv, ki) in lastCheckupPreview" :key="ki" class="summary-key-chip">
                <view class="summary-key-label">{{ kv.label }}</view>
                <view class="summary-key-value">{{ kv.value }}</view>
              </view>
            </view>
            <text v-else-if="lastCheckupFallback" class="summary-value">{{ lastCheckupFallback }}</text>
            <text v-else-if="lastCheckupRecord" class="summary-value">未填写内容</text>
            <text v-else class="summary-value">暂无</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastMetricRecord }"
            @click="lastMetricRecord?.id && openSummaryRecord(lastMetricRecord.id)"
          >
            <view class="summary-card-top">
              <text class="summary-label">身体指标</text>
              <text v-if="lastMetricRecord && lastMetricDate" class="summary-hint-inline">{{ lastMetricDate }}</text>
            </view>
            <view v-if="lastMetricPreview.length" class="summary-key-grid">
              <view v-for="(kv, ki) in lastMetricPreview" :key="ki" class="summary-key-chip">
                <view class="summary-key-label">{{ kv.label }}</view>
                <view class="summary-key-value">{{ kv.value }}</view>
              </view>
            </view>
            <text v-else-if="lastMetricFallback" class="summary-value">{{ lastMetricFallback }}</text>
            <text v-else-if="lastMetricRecord" class="summary-value">未填写内容</text>
            <text v-else class="summary-value">暂无</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastMoodRecord }"
            @click="lastMoodRecord?.id && openSummaryRecord(lastMoodRecord.id)"
          >
            <view class="summary-card-top">
              <text class="summary-label">最近心情</text>
              <text v-if="lastMoodRecord && lastMoodDate" class="summary-hint-inline">{{ lastMoodDate }}</text>
            </view>
            <view v-if="lastMoodPreview.length" class="summary-key-grid">
              <view v-for="(kv, ki) in lastMoodPreview" :key="ki" class="summary-key-chip">
                <view class="summary-key-label">{{ kv.label }}</view>
                <view class="summary-key-value">{{ kv.value }}</view>
              </view>
            </view>
            <text v-else-if="lastMoodFallback" class="summary-value">{{ lastMoodFallback }}</text>
            <text v-else-if="lastMoodRecord" class="summary-value">未填写内容</text>
            <text v-else class="summary-value">暂无</text>
          </view>
          <view
            class="summary-card"
            :class="{ 'summary-card--empty': !lastSymptomRecord }"
            @click="lastSymptomRecord?.id && openSummaryRecord(lastSymptomRecord.id)"
          >
            <view class="summary-card-top">
              <text class="summary-label">最近不适</text>
              <text v-if="lastSymptomRecord && lastSymptomDate" class="summary-hint-inline">{{ lastSymptomDate }}</text>
            </view>
            <view v-if="lastSymptomPreview.length" class="summary-key-grid">
              <view v-for="(kv, ki) in lastSymptomPreview" :key="ki" class="summary-key-chip">
                <view class="summary-key-label">{{ kv.label }}</view>
                <view class="summary-key-value">{{ kv.value }}</view>
              </view>
            </view>
            <text v-else-if="lastSymptomFallback" class="summary-value">{{ lastSymptomFallback }}</text>
            <text v-else-if="lastSymptomRecord" class="summary-value">未填写内容</text>
            <text v-else class="summary-value">暂无</text>
          </view>
        </view>
      </view>

      
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
        <view v-else class="record-list">
          <view
            v-for="row in recentMomRecordsWithPreview"
            :key="row.item.id"
            class="record-card"
            @click="openRecord(row.item.id)"
          >
            <view class="record-card-head">
              <view class="record-card-head-row">
                <text class="record-type">{{ labelForMotherType(row.item.record_type) }}</text>
                <text class="record-date">{{ row.item.occurred_at.slice(0, 10) }}</text>
              </view>
              <view class="record-card-tags">
                <text :class="['stage-badge', stageBadgeClass]">{{ statusLabel }}</text>
              </view>
            </view>
            <view v-if="row.keyRows.length" class="record-body">
              <view v-for="(kv, ki) in row.keyRows" :key="ki" class="record-kv-row">
                <view class="record-key-label">{{ kv.label }}</view>
                <view class="record-key-value">{{ kv.value }}</view>
              </view>
            </view>
            <text v-else class="record-summary">{{ row.item.summary?.trim() || "未填写内容" }}</text>
            <text v-if="row.showSummaryNote" class="record-summary-note">{{ row.item.summary }}</text>
          </view>
        </view>
      </view>

    </template>

    
    <QuickRecordSheet
      v-if="quickTemplate"
      :visible="showQuickSheet"
      :template="quickTemplate"
      @close="showQuickSheet = false"
      @saved="onQuickSaved"
    />

    
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
  getMotherRecordPreviewRows,
  templateForMotherType,
  type RecordTemplate,
} from "@/utils/motherRecordTypes";
import {
  getMotherNextStep,
  getPreferredBodyMetricType,
  motherStageBadgeClass,
  motherStageLabel,
} from "@/utils/homeRules";
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

const statusLabel = computed(() => motherStageLabel(dashboard.value?.profile?.status));

const stageBadgeClass = computed(() => motherStageBadgeClass(dashboard.value?.profile?.status));

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

const SHORTCUT_KEYS = ["checkup", "symptom", "medication", "mood", "__body_metric__"];

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

/** 始终列出三种指标；最近记录过的类型排在最前，便于一键延续，同时不会隐藏体重/血压 */
const metricOptions = computed(() => {
  const preferred = getPreferredBodyMetricType(dashboard.value?.latest_records ?? []);
  const rest = BODY_METRIC_TYPES.filter((v) => v !== preferred);
  const order = [preferred, ...rest];
  return order.map((v) => ({
    value: v,
    label: labelForMotherType(v),
  }));
});

/** 最近记录：优先展示模板关键字段行，摘要仅在无字段或标准记录补充说明时展示 */
const recentMomRecordsWithPreview = computed(() => {
  const list = dashboard.value?.latest_records?.slice(0, 3) ?? [];
  return list.map((item) => {
    const keyRows = getMotherRecordPreviewRows(item);
    const tmpl = templateForMotherType(item.record_type);
    const sum = item.summary?.trim() || "";
    const showSummaryNote = tmpl?.mode === "standard" && !!sum && keyRows.length > 0;
    return { item, keyRows, showSummaryNote };
  });
});

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

const SUMMARY_CARD_PREVIEW_MAX = 3;

function motherSummaryFallbackText(r: MotherRecordItem | null, previewLen: number): string | null {
  if (!r || previewLen > 0) return null;
  return r.summary?.trim() || null;
}

const lastCheckupRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "checkup") ?? null
);
const lastCheckupDate = computed(() => lastCheckupRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastCheckupPreview = computed(() => {
  const r = lastCheckupRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastCheckupFallback = computed(() =>
  motherSummaryFallbackText(lastCheckupRecord.value, lastCheckupPreview.value.length)
);

const lastMetricRecord = computed(() => {
  const metricTypes = ["weight", "blood_pressure", "blood_sugar"];
  return dashboard.value?.latest_records?.find((r) => metricTypes.includes(r.record_type)) ?? null;
});
const lastMetricDate = computed(() => lastMetricRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastMetricPreview = computed(() => {
  const r = lastMetricRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastMetricFallback = computed(() =>
  motherSummaryFallbackText(lastMetricRecord.value, lastMetricPreview.value.length)
);

const lastMoodRecord = computed(() => {
  return dashboard.value?.latest_records?.find((r) => r.record_type === "mood") ?? null;
});
const lastMoodDate = computed(() => lastMoodRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastMoodPreview = computed(() => {
  const r = lastMoodRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastMoodFallback = computed(() =>
  motherSummaryFallbackText(lastMoodRecord.value, lastMoodPreview.value.length)
);

const lastSymptomRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "symptom") ?? null
);
const lastSymptomDate = computed(() => lastSymptomRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastSymptomPreview = computed(() => {
  const r = lastSymptomRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastSymptomFallback = computed(() =>
  motherSummaryFallbackText(lastSymptomRecord.value, lastSymptomPreview.value.length)
);

const hasSummaryData = computed(() =>
  !!(
    lastCheckupRecord.value ||
    lastMetricRecord.value ||
    lastMoodRecord.value ||
    lastSymptomRecord.value
  )
);

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
    showMetricPicker.value = true;
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

.hero {
  margin-bottom: $cj-gap-lg;
}

.hero-top {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex-wrap: wrap;
  margin-bottom: 8rpx;
}

.hero-kicker {
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
}

.hero-headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 28rpx;
  min-width: 0;
}

.hero-name {
  flex-shrink: 0;
  max-width: 48%;
  font-size: 36rpx;
  font-weight: $cj-fw-title;
  color: $cj-primary-dark;
  letter-spacing: 1rpx;
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
  flex: 1;
  min-width: 0;
  font-size: 40rpx;
  line-height: 1.35;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  text-align: right;
}

.hero-title--solo {
  flex: none;
  width: 100%;
  text-align: left;
}

.hero-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

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

.summary-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $cj-gap-sm;
}

.summary-card {
  min-width: 0;
  box-sizing: border-box;
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

.summary-card-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16rpx;
  min-width: 0;
}

.summary-label {
  flex: 1;
  min-width: 0;
  color: $cj-text-muted;
  font-size: 22rpx;
}

.summary-hint-inline {
  flex-shrink: 0;
  font-size: 20rpx;
  color: $cj-text-muted;
}

.summary-value {
  display: block;
  margin-top: 10rpx;
  color: $cj-ink;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  line-height: 1.5;
  word-break: break-word;
}

/* 横向排列、自动换行，避免左标签右大空的纵向行布局 */
.summary-key-grid {
  margin-top: 10rpx;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 10rpx 12rpx;
}

.summary-key-chip {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: flex-start;
  max-width: 100%;
  box-sizing: border-box;
  padding: 8rpx 14rpx;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-md;
  gap: 8rpx;
}

.summary-key-label {
  flex-shrink: 0;
  max-width: 46%;
  font-size: 20rpx;
  color: $cj-text-muted;
  line-height: 1.45;
}

.summary-key-value {
  flex: 1;
  min-width: 0;
  font-size: 22rpx;
  color: $cj-ink;
  font-weight: 500;
  line-height: 1.45;
  word-break: break-word;
}

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

.record-list {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.record-card {
  padding: 24rpx 26rpx;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-soft;
}

.record-card-head {
  margin-bottom: 2rpx;
}

.record-card-head-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20rpx;
}

.record-card-tags {
  margin-top: 12rpx;
}

.record-type {
  flex: 1;
  min-width: 0;
  font-size: 30rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  line-height: 1.4;
}

.record-date {
  flex-shrink: 0;
  padding-top: 4rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  letter-spacing: 0.5rpx;
}

.record-body {
  margin-top: 16rpx;
  padding: 6rpx 18rpx 4rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid rgba(234, 217, 204, 0.55);
}

.record-kv-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: flex-start;
  gap: 16rpx;
  padding: 16rpx 0;
  border-bottom: 1rpx solid rgba(234, 217, 204, 0.45);
}

.record-kv-row:last-child {
  border-bottom: none;
  padding-bottom: 12rpx;
}

.record-kv-row:first-child {
  padding-top: 12rpx;
}

.record-key-label {
  flex-shrink: 0;
  width: 148rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.5;
}

.record-key-value {
  flex: 1;
  min-width: 0;
  font-size: 26rpx;
  color: $cj-text;
  line-height: 1.55;
  word-break: break-word;
}

.record-summary {
  display: block;
  margin-top: 14rpx;
  padding: 16rpx 18rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid rgba(234, 217, 204, 0.45);
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.record-summary-note {
  display: block;
  margin-top: 12rpx;
  padding-top: 14rpx;
  border-top: 1rpx solid rgba(234, 217, 204, 0.55);
  font-size: 23rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}

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
