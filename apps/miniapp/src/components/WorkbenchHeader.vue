<template>
  <view class="wb-header">
    <view class="wb-header__meta">
      <text class="wb-header__workspace">{{ workspaceLabel }}</text>
      <text v-if="badgeLabel" :class="['wb-header__badge', badgeClass]">{{ badgeLabel }}</text>
    </view>

    <view class="wb-header__main">
      <view v-if="name" :class="['wb-header__name-row', canSwitch && 'wb-header__name-row--switch']" @click="onNameTap">
        <text class="wb-header__name">{{ name }}</text>
        <text v-if="canSwitch" class="wb-header__switch-hint">切换 ▾</text>
      </view>
      <text class="wb-header__stage" :class="{ 'wb-header__stage--solo': !name }">{{ stageTitle }}</text>
      <text v-if="stageDesc" class="wb-header__desc">{{ stageDesc }}</text>
    </view>

    <view class="wb-header__actions">
      <view class="wb-header__action" @click="$emit('profile')">
        <text class="wb-header__action-icon">{{ profileIcon }}</text>
        <text class="wb-header__action-text">{{ profileLabel }}</text>
      </view>
      <view class="wb-header__action-divider" />
      <view class="wb-header__action" @click="$emit('family')">
        <text class="wb-header__action-icon">👨‍👩‍👧</text>
        <text class="wb-header__action-text">家人</text>
      </view>
    </view>

    <view v-if="showNextStep" class="wb-header__next" @click="$emit('next')">
      <view class="wb-header__next-body">
        <text class="wb-header__next-kicker">建议</text>
        <text class="wb-header__next-title">{{ nextStepTitle }}</text>
      </view>
      <text class="wb-header__next-btn">{{ nextStepBtnLabel }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    workspaceLabel: string;
    badgeLabel?: string;
    badgeClass?: string;
    name?: string;
    stageTitle: string;
    stageDesc?: string;
    profileLabel: string;
    profileIcon?: string;
    showNextStep?: boolean;
    nextStepTitle?: string;
    nextStepBtnLabel?: string;
    canSwitch?: boolean;
  }>(),
  {
    profileIcon: "📋",
    showNextStep: false,
    nextStepTitle: "",
    nextStepBtnLabel: "",
    canSwitch: false,
  }
);

const emit = defineEmits<{
  profile: [];
  family: [];
  next: [];
  switch: [];
}>();

function onNameTap() {
  if (props.canSwitch) emit("switch");
}
</script>

<style lang="scss" scoped>
.wb-header {
  margin-bottom: $cj-gap-md;
  padding: 28rpx $cj-gap-md 24rpx;
  background: linear-gradient(155deg, rgba(255, 253, 249, 0.98) 0%, rgba(255, 247, 240, 0.96) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
}

.wb-header__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12rpx;
  margin-bottom: 18rpx;
}

.wb-header__workspace {
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
  font-weight: 500;
}

.wb-header__badge {
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

.wb-header__main {
  margin-bottom: 22rpx;
}

.wb-header__name-row {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-bottom: 8rpx;
}

.wb-header__name-row--switch:active {
  opacity: 0.75;
}

.wb-header__name {
  font-size: 40rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  letter-spacing: 0.5rpx;
  line-height: 1.25;
}

.wb-header__switch-hint {
  font-size: 22rpx;
  color: $cj-primary;
  font-weight: 500;
  flex-shrink: 0;
}

.wb-header__stage {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
  line-height: 1.35;
}

.wb-header__stage--solo {
  font-size: 36rpx;
  font-weight: $cj-fw-display;
}

.wb-header__desc {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

.wb-header__actions {
  display: flex;
  align-items: stretch;
  background: rgba(255, 255, 255, 0.72);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
  overflow: hidden;
}

.wb-header__action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  padding: 20rpx 12rpx;

  &:active {
    background: rgba(201, 107, 92, 0.06);
  }
}

.wb-header__action-divider {
  width: 1rpx;
  background: $cj-border-faint;
  margin: 12rpx 0;
}

.wb-header__action-icon {
  font-size: 28rpx;
  line-height: 1;
}

.wb-header__action-text {
  font-size: 26rpx;
  color: $cj-primary-dark;
  font-weight: 500;
}

.wb-header__next {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  margin-top: 18rpx;
  padding: 18rpx 20rpx;
  background: linear-gradient(135deg, $cj-warn-bg 0%, #fff9f2 100%);
  border: 1rpx solid rgba(212, 165, 116, 0.25);
  border-radius: $cj-radius-lg;

  &:active {
    opacity: 0.9;
  }
}

.wb-header__next-body {
  flex: 1;
  min-width: 0;
}

.wb-header__next-kicker {
  display: block;
  font-size: 20rpx;
  color: $cj-text-muted;
  margin-bottom: 4rpx;
}

.wb-header__next-title {
  display: block;
  font-size: 26rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.wb-header__next-btn {
  flex-shrink: 0;
  font-size: 24rpx;
  color: #fffefb;
  font-weight: 500;
  padding: 12rpx 24rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
  box-shadow: $cj-shadow-xs;
}
</style>
