<template>
  <view class="wb-header">
    <view class="wb-row">
      <view class="wb-info">
        <view class="wb-name-line">
          <text :class="['wb-name', !name && 'wb-name--solo']" @click="onNameTap">{{ name || stageTitle }}</text>
          <text v-if="canSwitch" class="wb-switch" @click="onNameTap">切换 ▾</text>
          <text v-if="badgeLabel" :class="['wb-badge', badgeClass]">{{ badgeLabel }}</text>
          <text v-if="identityLabel" class="wb-ident">· {{ identityLabel }}</text>
          <text v-else class="wb-ident wb-ident--empty" @click="$emit('identity')">设置身份 ›</text>
        </view>
        <text v-if="name" class="wb-stage">{{ stageTitle }}</text>
      </view>

      <view class="wb-actions">
        <view class="wb-btn" @click="$emit('profile')">
          <text class="wb-btn-icon">{{ profileIcon }}</text>
          <text class="wb-btn-text">{{ shortProfileLabel }}</text>
        </view>
        <view class="wb-btn" @click="$emit('family')">
          <text class="wb-btn-icon">👨‍👩‍👧</text>
          <text class="wb-btn-text">家人</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    workspaceLabel?: string;
    badgeLabel?: string;
    badgeClass?: string;
    /** 当前人的家庭身份（家人页设置的昵称，如「爸爸」） */
    identityLabel?: string;
    name?: string;
    stageTitle: string;
    profileLabel: string;
    profileIcon?: string;
    canSwitch?: boolean;
  }>(),
  {
    workspaceLabel: "",
    profileIcon: "📋",
    canSwitch: false,
  }
);

const emit = defineEmits<{
  profile: [];
  family: [];
  identity: [];
  switch: [];
}>();

/** 「编辑档案 / 查看档案 / 创建档案」→「编辑 / 查看 / 建档」，横向缩短 */
const shortProfileLabel = computed(() => {
  const raw = props.profileLabel || "档案";
  if (raw.startsWith("编辑")) return "编辑";
  if (raw.startsWith("查看")) return "查看";
  if (raw.startsWith("创建")) return "建档";
  return raw;
});

function onNameTap() {
  if (props.canSwitch) emit("switch");
}
</script>

<style lang="scss" scoped>
.wb-header {
  margin-bottom: $cj-gap-md;
  padding: 22rpx $cj-gap-md;
  background: linear-gradient(155deg, rgba(255, 253, 249, 0.98) 0%, rgba(255, 247, 240, 0.96) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-xs;
}

.wb-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
}

.wb-info {
  flex: 1;
  min-width: 0;
}

.wb-name-line {
  display: flex;
  align-items: center;
  gap: 10rpx;
  min-width: 0;
}

.wb-name {
  flex-shrink: 1;
  min-width: 0;
  font-size: 34rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  line-height: 1.25;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-name--solo {
  font-size: 30rpx;
  color: $cj-ink;
}

.wb-switch {
  flex-shrink: 0;
  font-size: 21rpx;
  color: $cj-primary;
}

.wb-badge {
  flex-shrink: 0;
  padding: 4rpx 14rpx;
  border-radius: $cj-radius-pill;
  font-size: 19rpx;
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

.wb-ident {
  flex-shrink: 0;
  max-width: 220rpx;
  font-size: 22rpx;
  font-weight: 500;
  color: $cj-primary-dark;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-ident--empty {
  color: $cj-primary;
  font-weight: 400;
}

.wb-stage {
  display: block;
  margin-top: 4rpx;
  font-size: 23rpx;
  color: $cj-text-secondary;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wb-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10rpx;
}

.wb-btn {
  display: flex;
  align-items: center;
  gap: 6rpx;
  padding: 10rpx 16rpx;
  background: rgba(255, 255, 255, 0.8);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-pill;

  &:active {
    background: rgba(201, 107, 92, 0.08);
  }
}

.wb-btn-icon {
  font-size: 24rpx;
  line-height: 1;
}

.wb-btn-text {
  font-size: 22rpx;
  color: $cj-primary-dark;
  font-weight: 500;
}
</style>
