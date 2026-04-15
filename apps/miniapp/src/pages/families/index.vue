<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">与家人一起</text>
      <text class="head-title">我的家庭</text>
      <text class="head-desc">点选家庭进入宝宝列表；邀请码仅家庭成员可见。</text>
    </view>

    <view v-if="items.length === 0" class="empty">
      <text class="empty-e">🏡</text>
      <text class="empty-t">还没有家庭</text>
      <text class="empty-d">创建一个新家，或用邀请码加入亲友。</text>
    </view>

    <view v-for="f in items" :key="f.id" class="card" @click="openFamily(f.id)">
      <view class="stripe" />
      <view class="card-main">
        <text class="name">{{ f.name }}</text>
        <text class="sub">邀请码 · {{ f.invite_code }}</text>
      </view>
      <text class="chev">›</text>
    </view>

    <view class="actions">
      <button class="btn primary" type="primary" @click="goCreate">创建家庭</button>
      <button class="btn ghost" @click="goJoin">加入家庭</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiListFamilies } from "@/api/chuyaji";
import type { Family } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";

const items = ref<Family[]>([]);
const session = useSessionStore();

async function refresh() {
  try {
    const r = await apiListFamilies();
    items.value = r.items || [];
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

onShow(() => {
  refresh();
});

function openFamily(id: number) {
  session.setFamily(id);
  uni.navigateTo({ url: `/pages/baby/list?family_id=${id}` });
}

function goCreate() {
  uni.navigateTo({ url: "/pages/families/create" });
}

function goJoin() {
  uni.navigateTo({ url: "/pages/families/join" });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 48rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
}
.head-kicker {
  display: block;
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
}
.head-title {
  display: block;
  font-size: 44rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 2rpx;
}
.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
}

.empty {
  text-align: center;
  padding: 48rpx 0 32rpx;
}
.empty-e {
  font-size: 64rpx;
  display: block;
  margin-bottom: $cj-gap-sm;
}
.empty-t {
  display: block;
  font-size: 32rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
  margin-bottom: $cj-gap-sm;
}
.empty-d {
  display: block;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.55;
}

.card {
  position: relative;
  display: flex;
  align-items: stretch;
  background: $cj-surface;
  border-radius: $cj-radius-md;
  margin-bottom: $cj-gap-md;
  overflow: hidden;
  box-shadow: $cj-shadow-card;
  border: 1rpx solid $cj-border-light;
}
.stripe {
  width: 10rpx;
  background: linear-gradient(180deg, $cj-primary, $cj-accent);
  flex-shrink: 0;
}
.card-main {
  flex: 1;
  padding: $cj-gap-md $cj-gap-sm $cj-gap-md $cj-gap-md;
  min-width: 0;
}
.name {
  display: block;
  font-size: 32rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}
.sub {
  display: block;
  margin-top: 10rpx;
  font-size: 24rpx;
  color: $cj-text-secondary;
}
.chev {
  align-self: center;
  padding-right: $cj-gap-md;
  font-size: 36rpx;
  color: $cj-text-muted;
}

.actions {
  margin-top: $cj-gap-md;
}
.btn {
  margin-bottom: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
}
.btn.ghost {
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border !important;
}
</style>
