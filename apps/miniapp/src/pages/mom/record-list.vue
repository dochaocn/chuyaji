<template>
  <view class="page">

    
    <view class="filter-bar">
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="opt in filterOptions"
            :key="opt.value"
            :class="['filter-chip', activeFilter === opt.value && 'active']"
            @click="setFilter(opt.value)"
          >
            <text class="filter-chip-text">{{ opt.label }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    
    <scroll-view
      scroll-y
      class="list-scroll"
      :refresher-enabled="true"
      :refresher-triggered="refreshing"
      @refresherrefresh="onPullRefresh"
      @scrolltolower="onScrollToBottom"
    >
      <view class="list-scroll-inner">
        <view v-if="loading && !records.length" class="placeholder">加载中…</view>
        <view v-else-if="!records.length" class="placeholder">
          还没有记录，回首页记一条吧。
        </view>

        <template v-else>
          <template v-for="(group, gi) in groupedRecords" :key="gi">
            
            <view class="group-header">
              <text class="group-title">{{ group.label }}</text>
              <text class="group-count">{{ group.items.length }} 条</text>
            </view>

            
            <view
              v-for="entry in group.items"
              :key="entry.item.id"
              class="record-row"
              @click="openRecord(entry.item.id)"
            >
              <view class="record-body">
                <view class="record-head">
                  <text :class="['stage-badge', periodBadgeClass]">{{ periodLabel }}</text>
                  <text class="record-type">{{ labelForMotherType(entry.item.record_type) }}</text>
                  <text class="record-date">{{ formatDate(entry.item.occurred_at) }}</text>
                </view>
                <view v-if="entry.keyRows.length" class="record-key-grid">
                  <view v-for="(kv, ki) in entry.keyRows" :key="ki" class="record-key-line">
                    <text class="record-key-label">{{ kv.label }}</text>
                    <text class="record-key-value">{{ kv.value }}</text>
                  </view>
                </view>
                <text v-else class="record-summary">{{ entry.item.summary?.trim() || "未填写内容" }}</text>
                <text v-if="entry.showSummaryNote" class="record-summary-note">{{ entry.item.summary }}</text>
              </view>
            </view>
          </template>

          
          <view class="load-more">
            <text v-if="loadingMore" class="load-more-text">加载中…</text>
            <text v-else-if="!hasMore" class="load-more-text">已到底部</text>
          </view>
        </template>
      </view>
    </scroll-view>

  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import { apiGetMother, apiListMotherRecords, type MotherRecordItem } from "@/api/chuyaji";
import { motherStageBadgeClass, motherStageLabel } from "@/utils/homeRules";
import { getMotherRecordPreviewRows, labelForMotherType, templateForMotherType } from "@/utils/motherRecordTypes";

const motherId = ref(0);
const motherStatus = ref<"pregnant" | "postpartum" | "parenting" | undefined>(undefined);
const records = ref<MotherRecordItem[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const hasMore = ref(true);
const loading = ref(false);
const loadingMore = ref(false);
const refreshing = ref(false);

const filterOptions = [
  { value: "all", label: "全部" },
  { value: "checkup", label: "产检" },
  { value: "weight", label: "体重" },
  { value: "blood_pressure", label: "血压" },
  { value: "blood_sugar", label: "血糖" },
  { value: "symptom", label: "不适" },
  { value: "medication", label: "用药" },
  { value: "mood", label: "心情" },
];
const activeFilter = ref("all");

const periodLabel = computed(() => motherStageLabel(motherStatus.value));
const periodBadgeClass = computed(() => motherStageBadgeClass(motherStatus.value));

const filteredRecords = computed(() => {
  if (activeFilter.value === "all") return records.value;
  return records.value.filter((r) => r.record_type === activeFilter.value);
});

type MomListEntry = {
  item: MotherRecordItem;
  keyRows: ReturnType<typeof getMotherRecordPreviewRows>;
  showSummaryNote: boolean;
};

function setFilter(val: string) {
  activeFilter.value = val;
}

const groupedRecords = computed(() => {
  const map = new Map<string, { label: string; items: MomListEntry[] }>();
  for (const r of filteredRecords.value) {
    const key = r.occurred_at.slice(0, 7);
    if (!map.has(key)) {
      const [y, m] = key.split("-");
      map.set(key, { label: `${y} 年 ${Number(m)} 月`, items: [] });
    }
    const keyRows = getMotherRecordPreviewRows(r);
    const tmpl = templateForMotherType(r.record_type);
    const sum = r.summary?.trim() || "";
    const showSummaryNote = tmpl?.mode === "standard" && !!sum && keyRows.length > 0;
    map.get(key)!.items.push({ item: r, keyRows, showSummaryNote });
  }
  return [...map.entries()].sort((a, b) => b[0].localeCompare(a[0])).map(([, v]) => v);
});

onLoad((query: Record<string, string | undefined>) => {
  motherId.value = Number(query.mother_id || 0);
  if (motherId.value) {
    void loadMotherMeta();
    loadRecords(true);
  }
});

async function loadMotherMeta() {
  if (!motherId.value) return;
  try {
    const m = await apiGetMother(motherId.value);
    motherStatus.value = m.status;
  } catch (e) {
    console.error(e);
    motherStatus.value = undefined;
  }
}

async function loadRecords(reset = false) {
  if (!motherId.value) return;
  if (reset) {
    loading.value = true;
    records.value = [];
    nextCursor.value = undefined;
    hasMore.value = true;
  } else {
    if (!hasMore.value || loadingMore.value) return;
    loadingMore.value = true;
  }
  try {
    const res = await apiListMotherRecords(motherId.value, 30, reset ? undefined : nextCursor.value);
    const items = res.items ?? [];
    if (reset) {
      records.value = items;
    } else {
      records.value = [...records.value, ...items];
    }
    nextCursor.value = res.next_cursor || undefined;
    hasMore.value = !!res.next_cursor;
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
    loadingMore.value = false;
    refreshing.value = false;
  }
}

async function onPullRefresh() {
  refreshing.value = true;
  await loadMotherMeta();
  await loadRecords(true);
}

function onScrollToBottom() {
  loadRecords(false);
}

function formatDate(iso: string) {
  const d = iso.slice(0, 10);
  const [, m, day] = d.split("-");
  return `${Number(m)} 月 ${Number(day)} 日`;
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/mom/record-detail?id=${id}` });
}
</script>

<style lang="scss" scoped>
.page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  overflow-x: hidden;
  background: $cj-page-bg;
}

.filter-bar {
  flex-shrink: 0;
  padding: 16rpx $cj-page-pad-x 0;
  background: $cj-page-bg;
}

.filter-scroll {
  width: 100%;
}

.filter-inner {
  display: flex;
  gap: 16rpx;
  padding-bottom: 16rpx;
  white-space: nowrap;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 10rpx 28rpx;
  border-radius: $cj-radius-pill;
  border: 1rpx solid $cj-border-light;
  background: $cj-surface;
  flex-shrink: 0;

  &.active {
    background: $cj-primary;
    border-color: $cj-primary;

    .filter-chip-text {
      color: #fffefb;
    }
  }
}

.filter-chip-text {
  font-size: 24rpx;
  color: $cj-text-secondary;
}

.list-scroll {
  flex: 1;
  min-width: 0;
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
  padding-bottom: 72rpx;
}

.list-scroll-inner {
  box-sizing: border-box;
  width: 100%;
  max-width: 100%;
  padding: 0 $cj-page-pad-x;
}

.placeholder {
  padding: 80rpx 0;
  text-align: center;
  color: $cj-text-muted;
  font-size: 26rpx;
}

.group-header {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  padding: $cj-gap-md 0 $cj-gap-sm;
}

.group-title {
  font-size: 26rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.group-count {
  font-size: 22rpx;
  color: $cj-text-muted;
}

.record-row {
  margin-bottom: 0;
}

.record-body {
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-sm;
}

.record-head {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
  min-width: 0;
}

.stage-badge {
  flex-shrink: 0;
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

.record-type {
  flex: 1;
  min-width: 0;
  font-size: 28rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.record-date {
  flex-shrink: 0;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.record-summary {
  display: block;
  margin-top: 12rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
  overflow-wrap: break-word;
  word-break: break-word;
}

.record-key-grid {
  margin-top: 12rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.record-key-line {
  display: flex;
  align-items: flex-start;
  gap: 16rpx;
}

.record-key-label {
  flex-shrink: 0;
  width: 148rpx;
}

.record-key-value {
  flex: 1;
  min-width: 0;
}

.record-summary-note {
  display: block;
  margin-top: 10rpx;
  padding-top: 10rpx;
  border-top: 1rpx solid $cj-border-light;
  font-size: 24rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}

.load-more {
  padding: 32rpx 0 16rpx;
  text-align: center;
}

.load-more-text {
  font-size: 24rpx;
  color: $cj-text-muted;
}

@import "@/styles/cj-record-kv-fields.scss";
</style>
