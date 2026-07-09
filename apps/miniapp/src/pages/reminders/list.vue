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
      <template v-for="group in groups" :key="group.key">
        <view v-if="group.items.length" class="group">
          <text class="group-title">{{ group.label }}</text>
          <view v-for="item in group.items" :key="item.id" class="reminder-card">
            <view class="card-main">
              <view class="card-topline">
                <text :class="['status-chip', `status-chip--${group.key}`]">{{ group.label }}</text>
                <text class="card-date">{{ item.due_at.slice(0, 10) }}</text>
              </view>
              <text class="card-title">{{ item.title }}</text>
              <text class="card-meta">{{ item.note || sourceLabel(item) }}</text>
            </view>
            <view class="card-actions">
              <button size="mini" class="action-btn action-btn--primary" @click="markDone(item)">完成</button>
              <button size="mini" class="action-btn" @click="snooze(item, 1)">延后 1 天</button>
              <button size="mini" class="action-btn" @click="snooze(item, 7)">延后 1 周</button>
            </view>
            <view class="card-actions card-actions--secondary">
              <button size="mini" class="action-btn" @click="markIgnored(item.id)">忽略</button>
              <button size="mini" class="action-btn" @click="openSource(item)">查看来源</button>
              <button size="mini" class="action-btn action-btn--danger" @click="remove(item.id)">删除</button>
            </view>
          </view>
        </view>
      </template>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { apiDeleteReminder, apiListReminders, apiPatchReminder } from "@/api/chuyaji";
import type { ReminderItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";

const auth = useAuthStore();
const loading = ref(false);
const items = ref<ReminderItem[]>([]);

const groups = computed(() => {
  const today = new Date();
  const todayText = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;
  return [
    { key: "overdue", label: "已逾期", items: items.value.filter((item) => item.due_at.slice(0, 10) < todayText) },
    { key: "today", label: "今天", items: items.value.filter((item) => item.due_at.slice(0, 10) === todayText) },
    { key: "future", label: "未来", items: items.value.filter((item) => item.due_at.slice(0, 10) > todayText) },
  ];
});

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

async function markDone(item: ReminderItem) {
  await updateStatus(item.id, "done");
  if (item.owner_type === "baby") {
    uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${item.owner_id}` });
  } else {
    uni.navigateTo({ url: `/pages/mom/record-edit?mother_id=${item.owner_id}` });
  }
}

async function markIgnored(id: number) {
  await updateStatus(id, "ignored");
}

async function snooze(item: ReminderItem, days: number) {
  const base = new Date(item.due_at);
  base.setDate(base.getDate() + days);
  try {
    await apiPatchReminder(item.id, { snoozed_until: base.toISOString() });
    await load();
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "操作失败", icon: "none" });
  }
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
      ? `/pages/baby/record-edit?id=${item.source_id}&baby_id=${item.owner_id}`
      : `/pages/mom/record-edit?id=${item.source_id}&mother_id=${item.owner_id}`;
  uni.navigateTo({ url });
}

function sourceLabel(item: ReminderItem) {
  const owner = item.owner_type === "baby" ? "宝宝" : "宝妈";
  return `${owner}记录中的下次时间`;
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 80rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
  padding-top: 4rpx;
}

.head-kicker {
  display: block;
  font-size: 21rpx;
  letter-spacing: 5rpx;
  color: $cj-primary;
  margin-bottom: 8rpx;
  font-weight: 500;
}

.head-title {
  display: block;
  font-size: 42rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  letter-spacing: -0.5rpx;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.empty-card,
.reminder-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx;
}

.empty-title {
  display: block;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  font-size: 30rpx;
  letter-spacing: 0.3rpx;
}

.empty-desc {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-secondary;
  font-size: 25rpx;
  line-height: 1.65;
}

.list {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-md;
}

.group {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-sm;
}

.group-title {
  font-size: 25rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  letter-spacing: 0.3rpx;
}

.card-main {
  padding-bottom: $cj-gap-md;
}

.card-topline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $cj-gap-sm;
  margin-bottom: 8rpx;
}

.status-chip {
  padding: 5rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 19rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 0.3rpx;
}

.status-chip--overdue {
  background: $cj-danger-bg;
  color: $cj-danger-text;
}

.status-chip--today {
  background: $cj-warn-bg;
  color: $cj-text-secondary;
}

.status-chip--future {
  background: $cj-mint-soft;
  color: $cj-tag-postnatal-text;
}

.card-date {
  display: block;
  font-size: 23rpx;
  color: $cj-primary;
  margin-bottom: 8rpx;
  font-weight: 500;
}

.card-title {
  display: block;
  font-size: 31rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 0.2rpx;
}

.card-meta {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.card-actions {
  display: flex;
  gap: $cj-gap-sm;
}

.card-actions--secondary {
  margin-top: $cj-gap-sm;
}

.action-btn {
  flex: 1;
  min-width: 0;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-faint !important;
  font-size: 21rpx !important;
}

.action-btn--primary {
  background: $cj-primary !important;
  color: #fffefb !important;
  box-shadow: $cj-shadow-xs;
}

.action-btn--danger {
  color: $cj-primary-dark !important;
}
</style>
