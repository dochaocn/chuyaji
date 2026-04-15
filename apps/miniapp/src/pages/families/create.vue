<template>
  <view class="page">
    <view class="head">
      <text class="head-title">新建家庭</text>
      <text class="head-desc">名字可随时在家庭详情里再约定怎么称呼。</text>
    </view>
    <view class="field">
      <text class="lab">家庭名称</text>
      <input v-model="name" class="input" placeholder="例如：我们的小家" />
    </view>
    <button type="primary" class="btn" :loading="loading" @click="submit">创建</button>
  </view>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { apiCreateFamily } from "@/api/chuyaji";

const name = ref("");
const loading = ref(false);

async function submit() {
  if (!name.value.trim()) {
    uni.showToast({ title: "请输入名称", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    await apiCreateFamily(name.value.trim());
    uni.showToast({ title: "已创建", icon: "success" });
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
