<template>
  <view class="wrap">
    <view class="caption">数据点</view>
    <view class="row head">
      <text class="c">日龄</text>
      <text class="c">体重 (kg)</text>
    </view>
    <view v-for="(p, idx) in sortedUser" :key="idx" class="row">
      <text class="c">{{ p.age_days }}</text>
      <text class="c">{{ p.kg }}</text>
    </view>
    <view v-if="sortedUser.length === 0" class="empty">暂无生长记录（record_type=growth）</view>
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
</script>

<style lang="scss" scoped>
.wrap {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.caption {
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 2rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
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
