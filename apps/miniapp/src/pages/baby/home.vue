<template>
  <view class="page">

    <!-- 隐私检查 -->
    <view v-if="!session.privacyOk" class="notice">
      <text class="notice-title">先阅读隐私说明</text>
      <text class="notice-desc">记录只保存在你的账号范围内，进入正式记录前先确认使用说明。</text>
      <button size="mini" class="notice-btn" @click="goPrivacy">去阅读</button>
    </view>

    <!-- 未登录 -->
    <view v-if="!auth.token" class="empty-card">
      <text class="empty-title">先登录后再开始记录</text>
      <text class="empty-desc">登录后会自动加载当前宝宝档案与最近记录。</text>
      <button class="main-btn" @click="onLogin">微信登录</button>
    </view>

    <template v-else>

      <!-- ① 阶段头部 -->
      <view class="hero">
        <view class="hero-top">
          <text class="hero-kicker">宝宝工作台</text>
          <text v-if="dashboard?.profile" class="hero-name">{{ dashboard.profile.nickname || "未命名宝宝" }}</text>
        </view>
        <text class="hero-title">{{ stageTitle }}</text>
        <text class="hero-desc">{{ stageDesc }}</text>
      </view>

      <!-- ② 下一步卡片（按需出现，三种情况触发） -->
      <view v-if="nextStep.type !== 'none'" class="next-card">
        <text class="next-label">现在可以做</text>
        <text class="next-title">{{ nextStepTitle }}</text>
        <button class="next-btn" @click="onNextStepAction">{{ nextStepBtnLabel }}</button>
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

      <!-- ④A 趋势摘要（成长期） -->
      <view v-if="isPostnatalStage && hasTrendData" class="trend-section">
        <text class="section-label">近况</text>
        <view class="summary-grid">
          <view v-if="latestWeightLabel" class="summary-card">
            <text class="summary-label">最近体重</text>
            <text class="summary-value">{{ latestWeightLabel }}</text>
            <text v-if="weightAgeDays !== null" class="summary-hint">{{ weightAgeDays }} 天前</text>
          </view>
          <view class="summary-card">
            <text class="summary-label">本月记录</text>
            <text class="summary-value">{{ monthlyCount }}</text>
            <text class="summary-hint">条</text>
          </view>
          <view v-if="latestMilestone" class="summary-card">
            <text class="summary-label">最近里程碑</text>
            <text class="summary-value milestone-value">{{ latestMilestone }}</text>
          </view>
        </view>
        <view class="growth-link" @click="goGrowth">
          <text class="growth-link-text">查看生长趋势</text>
        </view>
      </view>

      <!-- ④B 最近记录 -->
      <view v-if="dashboard?.profile" class="section">
        <view class="section-head">
          <text class="section-title">最近记录</text>
          <view v-if="(dashboard?.latest_records?.length || 0) > 0" class="section-meta">
            <text class="section-meta-hint" @click="goTimeline">查看全部时间线</text>
          </view>
          <view v-else-if="!loading" class="section-meta">
            <text class="section-meta-hint">在快速记录里新增一条</text>
          </view>
        </view>
        <view v-if="loading" class="placeholder">加载中…</view>
        <view v-else-if="!dashboard?.latest_records?.length" class="placeholder">
          还没有记录，从推荐模板开始。
        </view>
        <view
          v-for="item in dashboard?.latest_records?.slice(0, 4) || []"
          :key="item.id"
          class="record-card"
          @click="openRecord(item.id)"
        >
          <view class="record-top">
            <text :class="['record-tag', item.phase === 'prenatal' ? 'pre' : 'post']">
              {{ item.phase === "prenatal" ? "怀孕期" : "成长期" }}
            </text>
            <text class="record-type">{{ labelForType(item.phase, item.record_type) }}</text>
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

  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { isDevWechatLogin } from "@/api/config";
import { apiBabyDashboard, apiCreateRecord } from "@/api/chuyaji";
import type { RecordItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import { labelForType, BABY_TEMPLATES, type RecordTemplate } from "@/utils/recordTypes";
import { calcGestation, calcAgeDays, calcAgeMonths, isPostnatal } from "@/utils/gestation";
import { getBabyNextStep, getBabyPostnatalShortcuts } from "@/utils/homeRules";
import QuickRecordSheet from "@/components/QuickRecordSheet.vue";

const auth = useAuthStore();
const session = useSessionStore();
const loading = ref(false);

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
} | null>(null);

// quick sheet 状态
const showQuickSheet = ref(false);
const quickTemplate = ref<RecordTemplate | null>(null);

// ---------------------------------------------------------------------------
// 阶段判断
// ---------------------------------------------------------------------------

const isPostnatalStage = computed(() => {
  return isPostnatal(dashboard.value?.profile?.birth_date);
});

const ageDays = computed(() => calcAgeDays(dashboard.value?.profile?.birth_date));
const ageMonths = computed(() => calcAgeMonths(dashboard.value?.profile?.birth_date));

const gestationResult = computed(() => {
  const p = dashboard.value?.profile;
  return calcGestation(p?.edd_date, p?.lmp_date);
});

const stageTitle = computed(() => {
  if (!dashboard.value?.profile) return "检查与成长，一站记下。";
  if (isPostnatalStage.value) {
    return `宝宝出生第 ${ageDays.value ?? 0} 天`;
  }
  if (gestationResult.value.available) {
    return `孕 ${gestationResult.value.weeks} 周 ${gestationResult.value.days} 天`;
  }
  return "检查与成长，一站记下。";
});

const stageDesc = computed(() => {
  if (!dashboard.value?.profile) return "先创建一个宝宝档案，后续所有孕期和成长期记录都围绕它展开。";
  if (isPostnatalStage.value) {
    return "喂养和生长变化都值得记一记。";
  }
  return "继续把每次检查安心记下。";
});

// ---------------------------------------------------------------------------
// 快捷入口
// ---------------------------------------------------------------------------

const shortcuts = computed<RecordTemplate[]>(() => {
  if (!dashboard.value?.profile) return [];
  if (isPostnatalStage.value) {
    const types = getBabyPostnatalShortcuts(ageMonths.value ?? 0);
    return types
      .map((t) => BABY_TEMPLATES.postnatal.find((tmpl) => tmpl.value === t))
      .filter((t): t is RecordTemplate => !!t);
  }
  // 怀孕期：产检、B 超、症状、用药
  const prenatalKeys = ["prenatal_checkup", "ultrasound", "symptom", "medication"];
  return prenatalKeys
    .map((t) => BABY_TEMPLATES.prenatal.find((tmpl) => tmpl.value === t))
    .filter((t): t is RecordTemplate => !!t);
});

// ---------------------------------------------------------------------------
// 下一步卡片
// ---------------------------------------------------------------------------

const nextStep = computed(() => {
  const hasProfile = !!dashboard.value?.profile;
  const lastAt = dashboard.value?.latest_records?.[0]?.occurred_at ?? null;
  return getBabyNextStep(hasProfile, lastAt);
});

const nextStepTitle = computed(() => {
  switch (nextStep.value.type) {
    case "create_profile": return "还没有宝宝档案";
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
// 趋势摘要
// ---------------------------------------------------------------------------

const latestWeightLabel = computed(() => {
  const value = dashboard.value?.growth_summary?.latest_weight_g;
  if (typeof value !== "number") return null;
  return `${(value / 1000).toFixed(2)} kg`;
});

const weightAgeDays = computed(() => {
  const at = dashboard.value?.growth_summary?.latest_weight_at as string | undefined;
  if (!at) return null;
  const diff = Math.floor((Date.now() - Date.parse(at)) / (24 * 60 * 60 * 1000));
  return diff >= 0 ? diff : null;
});

const monthlyCount = computed(() => {
  const records = dashboard.value?.latest_records ?? [];
  const now = new Date();
  return records.filter((r) => {
    const d = new Date(r.occurred_at);
    return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
  }).length;
});

const latestMilestone = computed(() => {
  const records = dashboard.value?.latest_records ?? [];
  const hit = records.find((r) => r.record_type === "development");
  if (!hit) return null;
  return (hit.payload as Record<string, unknown>)?.milestone as string ?? hit.summary ?? "已记录";
});

const hasTrendData = computed(() => {
  return !!latestWeightLabel.value || monthlyCount.value > 0;
});

// ---------------------------------------------------------------------------
// 事件
// ---------------------------------------------------------------------------

onShow(() => {
  auth.loadToken();
  session.load();
  if (auth.token) {
    loadDashboard();
  }
});

async function loadDashboard() {
  loading.value = true;
  try {
    const data = await apiBabyDashboard(session.babyId || undefined);
    dashboard.value = data;
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

function onNextStepAction() {
  const step = nextStep.value;
  if (step.type === "create_profile") return goProfileEdit();
  if (step.type === "add_record") return goNewRecord();
}

function onShortcut(tmpl: RecordTemplate) {
  if (tmpl.mode === "quick") {
    quickTemplate.value = tmpl;
    showQuickSheet.value = true;
  } else {
    const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
    if (!session.babyId) {
      uni.showToast({ title: "请先创建宝宝档案", icon: "none" });
      return;
    }
    uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${session.babyId}&phase=${phase}&type=${tmpl.value}` });
  }
}

async function onQuickSaved(payload: Record<string, unknown>, summary: string) {
  const bid = Number(session.babyId) || Number(dashboard.value?.profile?.id) || 0;
  if (!bid) {
    uni.showToast({ title: "缺少宝宝档案", icon: "none" });
    showQuickSheet.value = false;
    return;
  }
  if (bid !== Number(session.babyId)) {
    session.setBaby(bid);
  }
  const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
  try {
    await apiCreateRecord(bid, {
      phase,
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

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
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

function goProfileEdit() {
  const id = dashboard.value?.profile?.id || session.babyId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/baby/profile-edit${suffix}` });
}

function goNewRecord() {
  if (!session.babyId) {
    uni.showToast({ title: "请先创建宝宝档案", icon: "none" });
    return;
  }
  const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
  uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${session.babyId}&phase=${phase}` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/baby/record-detail?id=${id}` });
}

function goGrowth() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/growth?baby_id=${session.babyId}` });
}

function goTimeline() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/record-list?baby_id=${session.babyId}` });
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

/* ① 阶段头部 */
.hero {
  margin-bottom: $cj-gap-lg;
}

.hero-top {
  display: flex;
  align-items: center;
  gap: 16rpx;
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

/* ③ 快捷记录区 */
.shortcut-section,
.trend-section,
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

/* ④A 趋势摘要 */
.summary-grid {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-sm;
}

.summary-card {
  flex: 1;
  min-width: 180rpx;
  padding: $cj-gap-md;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
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
  font-size: 32rpx;
  font-weight: $cj-fw-display;
}

.milestone-value {
  font-size: 24rpx;
}

.summary-hint {
  display: block;
  margin-top: 6rpx;
  font-size: 20rpx;
  color: $cj-text-muted;
}

.growth-link {
  margin-top: $cj-gap-sm;
  text-align: right;
}

.growth-link-text {
  font-size: 24rpx;
  color: $cj-primary;
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

.record-tag.pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.record-tag.post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
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
</style>
