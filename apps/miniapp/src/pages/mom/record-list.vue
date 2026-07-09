<template>
  <view class="page">

    
    <view class="filter-bar">
      <view class="search-row">
        <input v-model="searchText" class="search-input" placeholder="搜索关键词" confirm-type="search" />
        <picker mode="date" :value="fromDate" @change="(e) => onDateChange('from', e.detail.value)">
          <view :class="['date-chip', !fromDate && 'date-chip--empty']">{{ fromDate || "开始日期" }}</view>
        </picker>
        <picker mode="date" :value="toDate" @change="(e) => onDateChange('to', e.detail.value)">
          <view :class="['date-chip', !toDate && 'date-chip--empty']">{{ toDate || "结束日期" }}</view>
        </picker>
        <view v-if="fromDate || toDate || searchText" class="date-clear" @click="clearSearch">清除</view>
      </view>
      <scroll-view scroll-x class="filter-scroll" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="opt in filterOptions"
            :key="opt.value"
            :class="['filter-chip', chipActive(opt.value) && 'active']"
            @click="toggleFilter(opt.value)"
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
          {{ emptyText }}
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
import { computed, ref, watch } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import { apiGetMother, apiListMotherRecords, type MotherRecordItem } from "@/api/chuyaji";
import { motherStageBadgeClass, motherStageLabel } from "@/utils/homeRules";
import {
  MOTHER_TIMELINE_FILTER_OPTIONS,
  getMotherRecordPreviewRows,
  labelForMotherType,
  templateForMotherType,
} from "@/utils/motherRecordTypes";

/** 首屏与每次触底加载条数 */
const PAGE_SIZE = 30;

const motherId = ref(0);
const motherStatus = ref<"pregnant" | "postpartum" | "parenting" | undefined>(undefined);
const records = ref<MotherRecordItem[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const hasMore = ref(true);
const loading = ref(false);
const loadingMore = ref(false);
const refreshing = ref(false);

const filterOptions = MOTHER_TIMELINE_FILTER_OPTIONS;

const selectedTypes = ref<Set<string>>(new Set());
const searchText = ref("");
const fromDate = ref("");
const toDate = ref("");

const periodLabel = computed(() => motherStageLabel(motherStatus.value));
const periodBadgeClass = computed(() => motherStageBadgeClass(motherStatus.value));

const filteredRecords = computed(() => {
  if (selectedTypes.value.size === 0) return records.value;
  return records.value.filter((r) => selectedTypes.value.has(r.record_type));
});

type MomListEntry = {
  item: MotherRecordItem;
  keyRows: ReturnType<typeof getMotherRecordPreviewRows>;
  showSummaryNote: boolean;
};

function chipActive(val: string) {
  if (val === "all") return selectedTypes.value.size === 0;
  return selectedTypes.value.has(val);
}

function toggleFilter(val: string) {
  if (val === "all") {
    selectedTypes.value = new Set();
    void loadRecords(true);
    return;
  }
  const next = new Set(selectedTypes.value);
  if (next.has(val)) next.delete(val);
  else next.add(val);
  selectedTypes.value = next;
  void loadRecords(true);
}

const emptyText = computed(() => {
  if (searchText.value.trim()) return "没有找到匹配记录。";
  if (selectedTypes.value.size > 0 || fromDate.value || toDate.value) return "当前筛选下暂无记录。";
  return "还没有记录，回首页记一条吧。";
});

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
});

onShow(() => {
  if (!motherId.value) return;
  void loadMotherMeta();
  void loadRecords(true);
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
    const singleType = selectedTypes.value.size === 1 ? [...selectedTypes.value][0] : undefined;
    const res = await apiListMotherRecords(motherId.value, {
      limit: PAGE_SIZE,
      cursor: reset ? undefined : nextCursor.value,
      record_type: singleType,
      q: searchText.value.trim() || undefined,
      from: fromDate.value || undefined,
      to: toDate.value || undefined,
    });
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

let searchTimer: ReturnType<typeof setTimeout> | null = null;
watch(searchText, () => {
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => void loadRecords(true), 400);
});

function onDateChange(which: "from" | "to", value: string) {
  if (which === "from") fromDate.value = value;
  else toDate.value = value;
  void loadRecords(true);
}

function clearSearch() {
  searchText.value = "";
  fromDate.value = "";
  toDate.value = "";
  void loadRecords(true);
}

function formatDate(iso: string) {
  const d = iso.slice(0, 10);
  const [, m, day] = d.split("-");
  return `${Number(m)} 月 ${Number(day)} 日`;
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/mom/record-edit?id=${id}&mother_id=${motherId.value}` });
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
  padding: 12rpx $cj-page-pad-x 0;
  background: $cj-page-bg;
}

.search-row {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 10rpx;
}

.search-input {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  min-height: 68rpx;
  padding: 0 24rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  color: $cj-text;
  font-size: 24rpx;
}

.date-chip,
.date-clear {
  flex-shrink: 0;
  padding: 9rpx 16rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  color: $cj-text-secondary;
  font-size: 21rpx;
  white-space: nowrap;
}

.date-chip--empty {
  color: $cj-text-muted;
}

.date-clear {
  color: $cj-primary;
  font-weight: 500;
}

.filter-scroll {
  width: 100%;
}

.filter-inner {
  display: flex;
  gap: 12rpx;
  padding-bottom: 14rpx;
  white-space: nowrap;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 9rpx 24rpx;
  border-radius: $cj-radius-pill;
  border: 1rpx solid $cj-border-faint;
  background: $cj-surface;
  flex-shrink: 0;
  transition: all 0.15s;

  &.active {
    background: $cj-primary;
    border-color: $cj-primary;
    box-shadow: $cj-shadow-soft;

    .filter-chip-text {
      color: #fffefb;
    }
  }
}

.filter-chip-text {
  font-size: 23rpx;
  color: $cj-text-secondary;
  letter-spacing: 0.3rpx;
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
  font-size: 25rpx;
}

.group-header {
  display: flex;
  align-items: baseline;
  gap: 10rpx;
  padding: $cj-gap-md 0 $cj-gap-sm;
}

.group-title {
  font-size: 25rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 0.3rpx;
}

.group-count {
  font-size: 21rpx;
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
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 26rpx;
  margin-bottom: $cj-gap-sm;

  &:active {
    opacity: 0.92;
  }
}

.record-head {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
  min-width: 0;
}

.stage-badge {
  flex-shrink: 0;
  padding: 5rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 19rpx;
  font-weight: 500;
  letter-spacing: 0.5rpx;
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
  letter-spacing: 0.2rpx;
}

.record-date {
  flex-shrink: 0;
  font-size: 21rpx;
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
  gap: 6rpx;
}

.record-key-line {
  display: flex;
  align-items: flex-start;
  gap: 14rpx;
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
  border-top: 1rpx solid $cj-border-faint;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}

.load-more {
  padding: 32rpx 0 16rpx;
  text-align: center;
}

.load-more-text {
  font-size: 23rpx;
  color: $cj-text-muted;
}

@import "@/styles/cj-record-kv-fields.scss";
</style>
