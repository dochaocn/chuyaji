<template>
  <view class="page">
    <view class="toolbar">
      <view class="toolbar-inner">
        <button class="tb" size="mini" @click="goGrowth">生长曲线</button>
        <button class="tb primary" size="mini" type="primary" @click="goNew">＋ 新建记录</button>
      </view>
      <text class="toolbar-hint">向下滚动加载更多</text>
    </view>

    <view class="tline">
      <view class="rail" />
      <view class="items">
        <view v-for="r in items" :key="r.id" class="titem" @click="open(r.id)">
          <view class="node-wrap">
            <view class="node" />
          </view>
          <view class="bubble">
            <view class="bubble-top">
              <text class="tag" :class="r.phase === 'prenatal' ? 'pre' : 'post'">{{
                r.phase === "prenatal" ? "孕" : "婴"
              }}</text>
              <text class="type">{{ typeLabel(r) }}</text>
            </view>
            <text class="date">{{ r.occurred_at?.slice(0, 10) }}</text>
            <text class="sum">{{ r.summary || "（暂无摘要）" }}</text>
          </view>
        </view>
      </view>
    </view>

    <view v-if="loading" class="hint">加载中…</view>
    <view v-if="!loading && !hasMore && items.length === 0" class="empty">
      <text class="empty-e">🌿</text>
      <text class="empty-t">还没有记录</text>
      <text class="empty-d">点右上角「新建记录」，写下第一条心情与数据吧。</text>
    </view>
    <view v-if="!loading && !hasMore && items.length > 0" class="hint end">· 已经到底啦 ·</view>
  </view>
</template>

<script setup lang="ts">
import { onLoad, onReachBottom } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiListRecords } from "@/api/chuyaji";
import type { RecordItem } from "@/api/chuyaji";
import { labelForType } from "@/utils/recordTypes";
import { useSessionStore } from "@/store/session";

const babyId = ref(0);
const items = ref<RecordItem[]>([]);
const cursor = ref<string | undefined>(undefined);
const loading = ref(false);
const hasMore = ref(true);
const session = useSessionStore();

onLoad((q: Record<string, string | undefined>) => {
  babyId.value = Number(q.baby_id || session.babyId || 0);
  session.setBaby(babyId.value);
  reset();
});

function typeLabel(r: RecordItem) {
  return labelForType(r.phase, r.record_type);
}

function reset() {
  items.value = [];
  cursor.value = undefined;
  hasMore.value = true;
  loadMore();
}

async function loadMore() {
  if (!babyId.value || loading.value || !hasMore.value) return;
  loading.value = true;
  try {
    const page = await apiListRecords(babyId.value, 20, cursor.value);
    items.value.push(...page.items);
    if (page.next_cursor) {
      cursor.value = page.next_cursor;
      hasMore.value = true;
    } else {
      hasMore.value = false;
    }
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

onReachBottom(() => {
  loadMore();
});

function open(id: number) {
  uni.navigateTo({ url: `/pages/record/detail?id=${id}` });
}

function goNew() {
  uni.navigateTo({ url: `/pages/record/edit?baby_id=${babyId.value}` });
}

function goGrowth() {
  uni.navigateTo({ url: `/pages/growth/index?baby_id=${babyId.value}` });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-gap-md $cj-page-pad-x 56rpx;
}

.toolbar {
  margin-bottom: $cj-gap-lg;
}
.toolbar-inner {
  display: flex;
  gap: $cj-gap-sm;
  justify-content: flex-end;
  flex-wrap: wrap;
}
.tb {
  border-radius: $cj-radius-pill !important;
}
.toolbar-hint {
  display: block;
  margin-top: 12rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  text-align: right;
}

.tline {
  position: relative;
  display: flex;
  gap: 0;
  padding-left: 8rpx;
}

.rail {
  position: absolute;
  left: 22rpx;
  top: 8rpx;
  bottom: 8rpx;
  width: 4rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(180deg, $cj-track, rgba(201, 107, 92, 0.08));
}

.items {
  flex: 1;
  padding-left: 48rpx;
}

.titem {
  position: relative;
  display: flex;
  margin-bottom: $cj-gap-md;
}

.node-wrap {
  position: absolute;
  left: -48rpx;
  top: 24rpx;
  width: 32rpx;
  height: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.node {
  width: 18rpx;
  height: 18rpx;
  border-radius: 50%;
  background: $cj-surface;
  border: 4rpx solid $cj-primary;
  box-shadow: 0 0 0 4rpx rgba(201, 107, 92, 0.15);
}

.bubble {
  flex: 1;
  background: $cj-surface;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.bubble-top {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
  margin-bottom: 8rpx;
}

.tag {
  font-size: 20rpx;
  padding: 4rpx 12rpx;
  border-radius: $cj-radius-pill;
  font-weight: 600;
}
.tag.pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}
.tag.post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.type {
  flex: 1;
  font-size: 30rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.date {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  margin-bottom: 10rpx;
}

.sum {
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.55;
}

.hint {
  text-align: center;
  color: $cj-text-muted;
  font-size: 24rpx;
  padding: $cj-gap-md 0;
}
.hint.end {
  font-size: 22rpx;
  letter-spacing: 2rpx;
  opacity: 0.85;
}

.empty {
  text-align: center;
  padding: 72rpx $cj-gap-md 48rpx;
}
.empty-e {
  display: block;
  font-size: 72rpx;
  margin-bottom: $cj-gap-md;
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
  line-height: 1.6;
}
</style>
