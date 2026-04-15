<template>
  <view class="page">
    <view class="head">
      <text class="head-title">加入家庭</text>
      <text class="head-desc">向亲友索要邀请码，输入后即可加入。</text>
    </view>
    <view class="field">
      <text class="lab">邀请码</text>
      <input v-model="code" class="input" placeholder="粘贴或输入邀请码" />
    </view>
    <button type="primary" class="btn" :loading="loading" @click="submit">加入</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { apiJoinFamily } from "@/api/chuyaji";

const code = ref("");
const loading = ref(false);

async function submit() {
  if (!code.value.trim()) {
    uni.showToast({ title: "请输入邀请码", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    await apiJoinFamily(code.value.trim());
    uni.showToast({ title: "已加入", icon: "success" });
    setTimeout(() => uni.navigateBack(), 400);
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x;
}

.head {
  margin-bottom: $cj-gap-lg;
}
.head-title {
  display: block;
  font-size: 40rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}
.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
}

.field {
  background: $cj-surface;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.lab {
  display: block;
  font-size: 24rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
}

.input {
  font-size: 30rpx;
  color: $cj-text;
}

.btn {
  border-radius: $cj-radius-pill !important;
}
</style>
