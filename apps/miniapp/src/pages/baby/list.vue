<template>
  <view class="page">
    <view v-if="!familyId" class="empty">缺少 family_id</view>
    <view v-else class="intro">
      <text class="intro-t">选择宝宝</text>
      <text class="intro-d">进入「成长时间线」查看与记录。</text>
    </view>

    <view v-for="b in items" :key="b.id" class="card" @click="pick(b)">
      <view class="avatar">{{ (b.nickname && b.nickname[0]) || "宝" }}</view>
      <view class="meta">
        <text class="name">{{ b.nickname }}</text>
        <text class="sub">ID {{ b.id }}</text>
      </view>
      <text class="chev">›</text>
    </view>

    <button type="primary" class="btn-add" @click="add">＋ 添加宝宝</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onShow } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiListBabies } from "@/api/chuyaji";
import type { Baby } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";

const familyId = ref(0);
const items = ref<Baby[]>([]);
const session = useSessionStore();

onLoad((q: Record<string, string | undefined>) => {
  familyId.value = Number(q.family_id || session.familyId || 0);
});

onShow(async () => {
  if (!familyId.value) return;
  try {
    const r = await apiListBabies(familyId.value);
    items.value = r.items || [];
  } catch (e) {
    console.error(e);
  }
});

function pick(b: Baby) {
  session.setBaby(b.id);
  uni.navigateTo({ url: `/pages/timeline/index?baby_id=${b.id}` });
}

function add() {
  uni.navigateTo({ url: `/pages/baby/edit?family_id=${familyId.value}` });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 48rpx;
}

.intro {
  margin-bottom: $cj-gap-lg;
}
.intro-t {
  display: block;
  font-size: 36rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}
.intro-d {
  display: block;
  margin-top: 8rpx;
  font-size: 26rpx;
  color: $cj-text-secondary;
}

.card {
  display: flex;
  align-items: center;
  gap: $cj-gap-md;
  background: $cj-surface;
  padding: $cj-gap-md;
  border-radius: $cj-radius-md;
  margin-bottom: $cj-gap-md;
  box-shadow: $cj-shadow-card;
  border: 1rpx solid $cj-border-light;
}

.avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background: linear-gradient(145deg, $cj-mint-soft, $cj-accent-soft);
  color: $cj-ink;
  font-size: 36rpx;
  font-weight: $cj-fw-title;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 2rpx solid $cj-border-light;
}

.meta {
  flex: 1;
  min-width: 0;
}
.name {
  font-size: 32rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}
.sub {
  display: block;
  margin-top: 6rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
}
.chev {
  font-size: 32rpx;
  color: $cj-text-muted;
}

.empty {
  color: $cj-text-secondary;
  text-align: center;
  padding: $cj-gap-lg 0;
  font-size: 28rpx;
}

.btn-add {
  margin-top: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
}
</style>
