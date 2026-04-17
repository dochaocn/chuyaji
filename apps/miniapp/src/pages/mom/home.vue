<template>
  <view class="page">
    <view class="hero">
      <text class="hero-kicker">宝妈工作台</text>
      <text class="hero-title">身体、产检与恢复，随记。</text>
    </view>

    <view v-if="!session.privacyOk" class="notice">
      <text class="notice-title">先阅读隐私说明</text>
      <text class="notice-desc">记录只保存在你的账号范围内，进入正式记录前先确认使用说明。</text>
      <button size="mini" class="notice-btn" @click="goPrivacy">去阅读</button>
    </view>

    <view v-if="!auth.token" class="empty-card">
      <text class="empty-title">登录后查看宝妈档案</text>
      <text class="empty-desc">请先在宝宝页完成登录，再回来管理宝妈信息与记录。</text>
    </view>

    <template v-else>
      <view v-if="dashboard?.profile" class="profile-card">
        <view class="profile-header">
          <text class="profile-name">{{ dashboard.profile.name || "宝妈档案" }}</text>
          <text :class="['profile-badge', motherBadgeClass]">{{ statusLabel }}</text>
        </view>
        <view class="pill-row">
          <text class="pill">血型：{{ dashboard.profile.blood_type || "未填写" }}</text>
          <text class="pill">分娩：{{ dashboard.profile.delivery_date ? dashboard.profile.delivery_date.slice(0, 10) : "未填写" }}</text>
        </view>
        <view class="summary-grid">
          <view class="summary-card">
            <text class="summary-label">记录数</text>
            <text class="summary-value">{{ dashboard.health_summary?.record_count ?? 0 }}</text>
          </view>
          <view class="summary-card">
            <text class="summary-label">最近更新</text>
            <text class="summary-value">{{ latestUpdateLabel }}</text>
          </view>
        </view>
      </view>

      <view v-else class="empty-card">
        <text class="empty-title">还没有宝妈档案</text>
        <text class="empty-desc">先补充基础信息，后续所有产检、症状、心情和恢复记录都从这里展开。</text>
        <button class="main-btn" @click="goProfileEdit">创建宝妈档案</button>
      </view>

      <view class="action-grid">
        <view class="action-tile warm" @click="goProfileEdit">
          <text class="action-title">档案</text>
          <text class="action-desc">基础信息、孕期与健康状况</text>
        </view>
        <view class="action-tile mint" @click="goNewRecord">
          <text class="action-title">录入记录</text>
          <text class="action-desc">产检、症状、心情、恢复</text>
        </view>
        <view class="action-tile accent" @click="goLatestOrInsight">
          <text class="action-title">最近一条</text>
          <text class="action-desc">查看详情或新建记录</text>
        </view>
      </view>

      <view class="section">
        <view class="section-head">
          <text class="section-title">最近记录</text>
          <view v-if="(dashboard?.latest_records?.length || 0) > 0" class="section-meta">
            <text class="section-meta-count">共 {{ (dashboard?.latest_records || []).length }} 条</text>
            <text class="section-meta-hint">点卡片查看详情</text>
          </view>
          <view v-else-if="dashboard?.profile && !loading" class="section-meta">
            <text class="section-meta-hint">在「录入记录」里新增一条</text>
          </view>
        </view>
        <view v-if="loading" class="placeholder">加载中…</view>
        <view v-else-if="!dashboard?.latest_records?.length" class="placeholder">还没有记录，可以从产检或身体状态开始。</view>
        <view
          v-for="item in dashboard?.latest_records || []"
          :key="item.id"
          class="record-card"
          @click="openRecord(item.id)"
        >
          <view class="record-top">
            <text class="record-tag record-tag--mom">宝妈</text>
            <text class="record-type">{{ labelForMotherType(item.record_type) }}</text>
          </view>
          <text class="record-date">{{ item.occurred_at.slice(0, 10) }}</text>
          <text class="record-summary">{{ item.summary || "未填写摘要" }}</text>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { apiMotherDashboard } from "@/api/chuyaji";
import type { MotherRecordItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import { labelForMotherType } from "@/utils/motherRecordTypes";

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

const statusLabel = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (status === "postpartum") return "产后恢复";
  if (status === "parenting") return "育儿期";
  return "怀孕中";
});

const motherBadgeClass = computed(() => {
  const status = dashboard.value?.profile?.status;
  return status === "pregnant" ? "profile-badge--pre" : "profile-badge--post";
});

const latestUpdateLabel = computed(() => {
  const at = dashboard.value?.latest_records?.[0]?.occurred_at;
  if (!at) return "暂无";
  return at.slice(0, 10);
});

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
    const data = await apiMotherDashboard(session.motherId || undefined);
    dashboard.value = data;
    if (data.profile?.id) {
      session.setMother(data.profile.id);
    }
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
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

function goLatestOrInsight() {
  const first = dashboard.value?.latest_records?.[0];
  if (first) {
    openRecord(first.id);
    return;
  }
  goNewRecord();
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/mom/record-detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 72rpx;
}

.hero {
  margin-bottom: $cj-gap-lg;
}

.hero-kicker {
  display: block;
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
  margin-bottom: 12rpx;
}

.hero-title {
  display: block;
  font-size: 44rpx;
  line-height: 1.35;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.notice,
.empty-card,
.profile-card,
.section {
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
.empty-title,
.profile-name,
.section-title {
  display: block;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.notice-title,
.section-title {
  font-size: 30rpx;
}

.empty-title,
.profile-name {
  font-size: 34rpx;
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

.profile-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12rpx 16rpx;
}

.profile-name {
  flex: 1;
  min-width: 200rpx;
}

.profile-badge {
  flex-shrink: 0;
  padding: 8rpx 20rpx;
  border-radius: $cj-radius-pill;
  font-size: 22rpx;
  font-weight: 500;
}

.profile-badge--pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.profile-badge--post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.section-head,
.record-top,
.summary-grid,
.pill-row,
.action-grid {
  display: flex;
}

.section-head {
  justify-content: space-between;
  align-items: center;
  gap: $cj-gap-sm;
}

.pill-row,
.summary-grid,
.action-grid {
  flex-wrap: wrap;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.pill,
.summary-card,
.action-tile,
.record-card {
  border-radius: $cj-radius-md;
}

.pill {
  padding: 12rpx 22rpx;
  background: $cj-accent-soft;
  color: $cj-text-secondary;
  font-size: 22rpx;
}

.summary-card {
  flex: 1;
  min-width: 220rpx;
  padding: $cj-gap-md;
  background: linear-gradient(165deg, $cj-surface-2 0%, $cj-surface 100%);
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
  font-size: 34rpx;
  font-weight: $cj-fw-display;
}

.action-grid {
  margin-bottom: $cj-gap-md;
}

.action-tile {
  flex: 1;
  min-width: 200rpx;
  padding: $cj-gap-md;
  color: $cj-ink;
}

.action-tile.warm {
  background: linear-gradient(145deg, $cj-primary-soft 0%, $cj-surface 100%);
}

.action-tile.mint {
  background: linear-gradient(145deg, $cj-mint-soft 0%, $cj-surface 100%);
}

.action-tile.accent {
  background: linear-gradient(145deg, $cj-accent-soft 0%, $cj-surface 100%);
}

.action-title {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
}

.action-desc {
  display: block;
  margin-top: 10rpx;
  font-size: 23rpx;
  line-height: 1.6;
  color: $cj-text-secondary;
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
  margin-top: $cj-gap-sm;
}

.record-top {
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
  display: block;
  margin-top: 10rpx;
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

.section-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4rpx;
  max-width: 62%;
}

.section-meta-count {
  font-size: 24rpx;
  color: $cj-ink;
  font-weight: $cj-fw-title;
}

.section-meta-hint {
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.4;
  text-align: right;
}
</style>
