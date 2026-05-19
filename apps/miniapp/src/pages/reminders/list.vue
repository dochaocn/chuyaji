<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">提醒</text>
      <text class="head-title">待办事项</text>
      <text class="head-desc">来自产检和复诊记录里的下次时间。</text>
    </view>

    <view v-if="loading" class="empty-card">
      <text class="empty-title">加载中...</text>
      <text class="empty-desc">正在整理最近的待办。</text>
    </view>

    <view v-else-if="!items.length" class="empty-card">
      <text class="empty-title">暂无待办</text>
      <text class="empty-desc">记录产检或复诊时填写下次时间，这里会自动出现提醒。</text>
    </view>

    <view v-else class="list">
      <view v-for="item in items" :key="item.id" class="reminder-card">
        <view class="card-main">
          <text class="card-date">{{ item.due_at.slice(0, 10) }}</text>
          <text class="card-title">{{ item.title }}</text>
          <text class="card-meta">{{ sourceLabel(item) }}</text>
        </view>
        <view class="card-actions">
          <button size="mini" class="action-btn action-btn--primary" @click="markDone(item.id)">完成</button>
          <button size="mini" class="action-btn" @click="markIgnored(item.id)">忽略</button>
          <button size="mini" class="action-btn" @click="openSource(item)">查看来源</button>
          <button size="mini" class="action-btn action-btn--danger" @click="remove(item.id)">删除</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiDeleteReminder, apiListReminders, apiPatchReminder } from "@/api/chuyaji";
import type { ReminderItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";

const auth = useAuthStore();
const loading = ref(false);
const items = ref<ReminderItem[]>([]);

onShow(async () => {
  auth.loadToken();
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
  if (auth.token) {
    await load();
  }
});

async function load() {
  loading.value = true;
  try {
    const data = await apiListReminders("pending", 50);
    items.value = data.items;
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

async function markDone(id: number) {
  await updateStatus(id, "done");
}

async function markIgnored(id: number) {
  await updateStatus(id, "ignored");
}

async function updateStatus(id: number, status: ReminderItem["status"]) {
  try {
    await apiPatchReminder(id, { status });
    items.value = items.value.filter((item) => item.id !== id);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "操作失败", icon: "none" });
  }
}

async function remove(id: number) {
  try {
    await apiDeleteReminder(id);
    items.value = items.value.filter((item) => item.id !== id);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "删除失败", icon: "none" });
  }
}

function openSource(item: ReminderItem) {
  const url =
    item.source_type === "baby_record"
      ? `/pages/baby/record-detail?id=${item.source_id}`
      : `/pages/mom/record-detail?id=${item.source_id}`;
  uni.navigateTo({ url });
}

function sourceLabel(item: ReminderItem) {
  const owner = item.owner_type === "baby" ? "宝宝" : "宝妈";
  return `${owner}记录中的下次时间`;
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 72rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
}

.head-kicker {
  display: block;
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
  margin-bottom: 8rpx;
}

.head-title {
  display: block;
  font-size: 42rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

.empty-card,
.reminder-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
}

.empty-title {
  display: block;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  font-size: 30rpx;
}

.empty-desc {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-secondary;
  font-size: 25rpx;
  line-height: 1.6;
}

.list {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-md;
}

.card-main {
  padding-bottom: $cj-gap-md;
}

.card-date {
  display: block;
  font-size: 24rpx;
  color: $cj-primary;
  margin-bottom: 8rpx;
}

.card-title {
  display: block;
  font-size: 32rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.card-meta {
  display: block;
  margin-top: 8rpx;
  font-size: 23rpx;
  color: $cj-text-muted;
}

.card-actions {
  display: flex;
  gap: $cj-gap-sm;
}

.action-btn {
  flex: 1;
  min-width: 0;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
  font-size: 22rpx !important;
}

.action-btn--primary {
  background: $cj-primary !important;
  color: #fffefb !important;
}

.action-btn--danger {
  color: $cj-primary-dark !important;
}
</style>
