<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">成长相册</text>
      <text class="head-title">记录里的照片与报告</text>
      <text class="head-desc">按上传时间倒序浏览，点开可预览或跳回来源记录。</text>
    </view>

    <view v-if="loading && !items.length" class="empty-card">
      <text class="empty-title">加载中…</text>
    </view>
    <view v-else-if="!items.length" class="empty-card">
      <text class="empty-title">还没有照片</text>
      <text class="empty-desc">在宝宝记录里上传检查单、日常照片后，会自动出现在这里。</text>
      <button class="main-btn" @click="goTimeline">去看时间线</button>
    </view>
    <view v-else class="grid">
      <view v-for="(item, idx) in items" :key="item.id" class="cell">
        <image
          class="thumb"
          :src="item.thumb_url || item.url"
          mode="aspectFill"
          @click="previewAt(idx)"
        />
        <view class="cell-meta">
          <text class="cell-type">{{ typeLabel(item.record_type) }}</text>
          <text class="cell-date">{{ item.occurred_at.slice(0, 10) }}</text>
        </view>
        <text class="cell-link" @click="openRecord(item.record_id)">查看来源 ›</text>
      </view>
    </view>

    <view v-if="hasMore" class="more-wrap">
      <button class="ghost-btn" :loading="loadingMore" @click="loadMore">加载更多</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import { apiListBabyAttachments, type BabyAlbumAttachment } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";
import { BABY_TEMPLATES } from "@/utils/recordTypes";

const session = useSessionStore();
const babyId = ref(0);
const items = ref<BabyAlbumAttachment[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const hasMore = ref(true);
const loading = ref(false);
const loadingMore = ref(false);

onLoad((query: Record<string, string | undefined>) => {
  session.load();
  babyId.value = Number(query.baby_id || session.babyId || 0);
});

onShow(() => {
  if (!babyId.value) return;
  void load(true);
});

function typeLabel(recordType: string) {
  const hit =
    BABY_TEMPLATES.postnatal.find((t) => t.value === recordType) ||
    BABY_TEMPLATES.prenatal.find((t) => t.value === recordType);
  return hit?.label || recordType;
}

async function load(reset: boolean) {
  if (!babyId.value) return;
  if (reset) {
    loading.value = true;
    items.value = [];
    nextCursor.value = undefined;
    hasMore.value = true;
  } else {
    if (!hasMore.value || loadingMore.value) return;
    loadingMore.value = true;
  }
  try {
    const res = await apiListBabyAttachments(babyId.value, 40, reset ? undefined : nextCursor.value);
    const page = res.items || [];
    items.value = reset ? page : [...items.value, ...page];
    nextCursor.value = res.next_cursor || undefined;
    hasMore.value = !!res.next_cursor;
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
}

function loadMore() {
  void load(false);
}

function previewAt(index: number) {
  const urls = items.value.map((i) => i.url);
  uni.previewImage({ urls, current: urls[index] });
}

function openRecord(recordId: number) {
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${recordId}&baby_id=${babyId.value}` });
}

function goTimeline() {
  if (!babyId.value) return;
  uni.navigateTo({ url: `/pages/baby/record-list?baby_id=${babyId.value}` });
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
  color: $cj-mint;
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

.empty-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx;
}

.empty-title {
  display: block;
  color: $cj-ink;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
}

.empty-desc {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-secondary;
  font-size: 25rpx;
  line-height: 1.65;
}

.main-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}

.grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.cell {
  width: calc(50% - 8rpx);
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-xs;
  overflow: hidden;
  padding-bottom: 14rpx;
}

.thumb {
  width: 100%;
  height: 220rpx;
  background: $cj-surface-2;
  display: block;
}

.cell-meta {
  display: flex;
  justify-content: space-between;
  gap: 8rpx;
  padding: 12rpx 14rpx 0;
}

.cell-type {
  font-size: 22rpx;
  color: $cj-ink;
  font-weight: 500;
}

.cell-date {
  font-size: 20rpx;
  color: $cj-text-muted;
}

.cell-link {
  display: block;
  padding: 8rpx 14rpx 0;
  font-size: 22rpx;
  color: $cj-primary-dark;
  font-weight: 500;
}

.more-wrap {
  margin-top: $cj-gap-lg;
}

.ghost-btn {
  width: 100%;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-faint !important;
}
</style>
