<template>
  <view class="wrap">
    <view class="summary-row">
      <view class="summary-card">
        <text class="summary-label">样本数</text>
        <text class="summary-value">{{ sortedUser.length }}</text>
      </view>
      <view class="summary-card">
        <text class="summary-label">最新体重</text>
        <text class="summary-value">{{ latestLabel }}</text>
      </view>
    </view>
    <view class="caption">趋势条</view>
    <view v-if="sortedUser.length" class="bars">
      <view v-for="(p, idx) in sortedUser" :key="idx" class="bar-row">
        <text class="bar-age">{{ p.age_days }} 天</text>
        <view class="bar-track">
          <view class="bar-fill" :style="{ width: `${barWidth(p.kg)}%` }" />
        </view>
        <text class="bar-kg">{{ p.kg.toFixed(2) }} kg</text>
      </view>
    </view>
    <view v-if="sortedUser.length === 0" class="empty">暂无生长记录（record_type=growth）</view>
    <view v-if="sortedUser.length" class="table">
      <view class="row head">
        <text class="c">日龄</text>
        <text class="c">体重 (kg)</text>
      </view>
      <view v-for="(p, idx) in sortedUser" :key="`table-${idx}`" class="row">
        <text class="c">{{ p.age_days }}</text>
        <text class="c">{{ p.kg }}</text>
      </view>
    </view>
    <view v-if="refLabel" class="hint">{{ refLabel }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  userPoints: { age_days: number; kg: number }[];
  refLabel?: string;
}>();

const sortedUser = computed(() => [...props.userPoints].sort((a, b) => a.age_days - b.age_days));
const latestLabel = computed(() => {
  const latest = sortedUser.value[sortedUser.value.length - 1];
  return latest ? `${latest.kg.toFixed(2)} kg` : "暂无";
});
const maxKg = computed(() => Math.max(...sortedUser.value.map((item) => item.kg), 0));

function barWidth(kg: number) {
  if (!maxKg.value) return 0;
  return Math.max(16, Math.round((kg / maxKg.value) * 100));
}
</script>

<style lang="scss" scoped>
.wrap {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.summary-row {
  display: flex;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-md;
}

.summary-card {
  flex: 1;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  padding: $cj-gap-sm;
}

.summary-label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.summary-value {
  display: block;
  margin-top: 8rpx;
  font-size: 32rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.caption {
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 2rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
}

.bars {
  margin-bottom: $cj-gap-md;
}

.bar-row {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-sm;
}

.bar-age,
.bar-kg {
  width: 110rpx;
  font-size: 22rpx;
  color: $cj-text-secondary;
}

.bar-kg {
  text-align: right;
}

.bar-track {
  flex: 1;
  height: 18rpx;
  border-radius: $cj-radius-pill;
  background: rgba(201, 107, 92, 0.12);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, $cj-mint 0%, $cj-primary 100%);
}

.table {
  margin-top: $cj-gap-md;
}

.row {
  display: flex;
  flex-direction: row;
  padding: 10rpx 0;
  border-bottom: 1rpx solid $cj-border;
}
.head {
  font-weight: $cj-fw-title;
  color: $cj-text;
}
.c {
  flex: 1;
  font-size: 26rpx;
  color: $cj-text-secondary;
}
.empty {
  padding: $cj-gap-md 0;
  color: $cj-text-muted;
  font-size: 24rpx;
  text-align: center;
}
.hint {
  margin-top: $cj-gap-sm;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}
</style>
