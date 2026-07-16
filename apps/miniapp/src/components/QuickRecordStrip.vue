<template>
  <view class="quick-strip">
    <text class="quick-strip__label">快速记录</text>
    <scroll-view class="quick-strip__scroll" scroll-x :show-scrollbar="false" enable-flex>
      <view class="quick-strip__row">
        <view
          v-for="item in shortcuts"
          :key="item.value"
          class="quick-strip__chip"
          @click="$emit('shortcut', item)"
        >
          <text class="quick-strip__chip-title">{{ item.entryLabel }}</text>
          <text class="quick-strip__chip-mode">{{ item.mode === "quick" ? "一键" : "标准" }}</text>
        </view>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
defineProps<{
  shortcuts: Array<{ value: string; entryLabel: string; mode: "quick" | "standard" }>;
}>();

defineEmits<{
  shortcut: [item: { value: string; entryLabel: string; mode: "quick" | "standard" }];
}>();
</script>

<style lang="scss" scoped>
.quick-strip {
  margin-bottom: $cj-gap-md;
}

.quick-strip__label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  letter-spacing: 3rpx;
  margin-bottom: 14rpx;
  font-weight: 500;
  padding-left: 4rpx;
}

.quick-strip__scroll {
  width: 100%;
  white-space: nowrap;
}

.quick-strip__row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 12rpx;
  padding: 4rpx 2rpx 8rpx;
}

.quick-strip__chip {
  flex-shrink: 0;
  min-width: 148rpx;
  padding: 22rpx 24rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;

  &:active {
    opacity: 0.85;
    transform: scale(0.98);
  }
}

.quick-strip__chip-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  white-space: nowrap;
}

.quick-strip__chip-mode {
  display: block;
  margin-top: 6rpx;
  font-size: 20rpx;
  color: $cj-text-muted;
  white-space: nowrap;
}
</style>
