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
            v-for="opt in phaseFilterOptions"
            :key="`p-${opt.value}`"
            :class="['filter-chip', phaseChipActive(opt.value) && 'active']"
            @click="togglePhaseFilter(opt.value)"
          >
            <text class="filter-chip-text">{{ opt.label }}</text>
          </view>
        </view>
      </scroll-view>
      <scroll-view scroll-x class="filter-scroll filter-scroll--types" :show-scrollbar="false">
        <view class="filter-inner">
          <view
            v-for="opt in typeFilterOptions"
            :key="`t-${opt.value}`"
            :class="['filter-chip', typeChipActive(opt.value) && 'active']"
            @click="toggleTypeFilter(opt.value)"
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
                  <text :class="['tag', entry.item.phase === 'prenatal' ? 'pre' : 'post']">
                    {{ entry.item.phase === "prenatal" ? "怀孕期" : "成长期" }}
                  </text>
                  <text class="record-type">{{ labelForType(entry.item.phase, entry.item.record_type) }}</text>
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
import { apiGetBaby, apiListRecords, type RecordItem } from "@/api/chuyaji";
import { isPostnatal } from "@/utils/gestation";
import {
  BABY_PHASE_FILTER_OPTIONS,
  BABY_RECORD_TYPE_FILTER_OPTIONS,
  RECORD_TYPES,
  getBabyRecordPreviewRows,
  labelForType,
  templateForType,
} from "@/utils/recordTypes";

/** 首屏与每次触底加载条数 */
const PAGE_SIZE = 30;

const babyId = ref(0);
const records = ref<RecordItem[]>([]);
const nextCursor = ref<string | undefined>(undefined);
const hasMore = ref(true);
const loading = ref(false);
const loadingMore = ref(false);
const refreshing = ref(false);

const filtersPrimed = ref(false);

/** 未出生的宝宝：筛选条优先展示怀孕期及孕期类型；已出生则成长期优先 */
const preferPostnatalFirst = ref(false);

type BabyPhase = "prenatal" | "postnatal";
type PhaseChipValue = "all" | BabyPhase;

/** 阶段单选：全部 / 怀孕期 / 成长期（chip 顺序随宝宝阶段调整） */
const selectedPhase = ref<PhaseChipValue>("all");
const selectedTypes = ref<Set<string>>(new Set());
const searchText = ref("");
const fromDate = ref("");
const toDate = ref("");

const phaseFilterOptions = computed(() => {
  const opts = BABY_PHASE_FILTER_OPTIONS;
  if (preferPostnatalFirst.value) {
    return [opts[0], opts[2], opts[1]];
  }
  return [...opts];
});

/** 二级联动：「全部」下展示所有类型（顺序随宝宝阶段）；选定某一时期后仅展示该期类型（类型可多选） */
const typeFilterOptions = computed(() => {
  if (selectedPhase.value === "all") {
    if (preferPostnatalFirst.value) {
      return [...RECORD_TYPES.postnatal, ...RECORD_TYPES.prenatal];
    }
    return BABY_RECORD_TYPE_FILTER_OPTIONS;
  }
  if (selectedPhase.value === "prenatal") return [...RECORD_TYPES.prenatal];
  return [...RECORD_TYPES.postnatal];
});

function pruneSelectedTypesToPhase() {
  const allowed = new Set(typeFilterOptions.value.map((o) => o.value));
  const next = new Set<string>();
  for (const t of selectedTypes.value) {
    if (allowed.has(t)) next.add(t);
  }
  selectedTypes.value = next;
}

const filteredRecords = computed(() => {
  return records.value.filter((r) => {
    const phaseOk = selectedPhase.value === "all" || r.phase === selectedPhase.value;
    const typeOk = selectedTypes.value.size === 0 || selectedTypes.value.has(r.record_type);
    return phaseOk && typeOk;
  });
});

type BabyListEntry = {
  item: RecordItem;
  keyRows: ReturnType<typeof getBabyRecordPreviewRows>;
  showSummaryNote: boolean;
};

function phaseChipActive(val: PhaseChipValue) {
  return selectedPhase.value === val;
}

function togglePhaseFilter(val: PhaseChipValue) {
  if (val === "all") {
    selectedPhase.value = "all";
    selectedTypes.value = new Set();
    void loadRecords(true);
    return;
  }
  if (selectedPhase.value === val) return;
  selectedPhase.value = val;
  pruneSelectedTypesToPhase();
  void loadRecords(true);
}

function typeChipActive(val: string) {
  return selectedTypes.value.has(val);
}

function toggleTypeFilter(val: string) {
  const next = new Set(selectedTypes.value);
  if (next.has(val)) next.delete(val);
  else next.add(val);
  selectedTypes.value = next;
  void loadRecords(true);
}

const emptyText = computed(() => {
  if (searchText.value.trim()) return "没有找到匹配记录。";
  if (selectedPhase.value !== "all" || selectedTypes.value.size > 0 || fromDate.value || toDate.value) return "当前筛选下暂无记录。";
  return "还没有记录，回首页记一条吧。";
});

async function primeDefaultFilters() {
  if (!babyId.value || filtersPrimed.value) return;
  try {
    const baby = await apiGetBaby(babyId.value);
    const post = isPostnatal(baby.birth_date);
    preferPostnatalFirst.value = post;
    selectedPhase.value = post ? "postnatal" : "prenatal";
  } catch (e) {
    console.error(e);
    preferPostnatalFirst.value = false;
    selectedPhase.value = "all";
    selectedTypes.value = new Set();
  } finally {
    filtersPrimed.value = true;
  }
}

const groupedRecords = computed(() => {
  const map = new Map<string, { label: string; items: BabyListEntry[] }>();
  for (const r of filteredRecords.value) {
    const key = r.occurred_at.slice(0, 7);
    if (!map.has(key)) {
      const [y, m] = key.split("-");
      map.set(key, { label: `${y} 年 ${Number(m)} 月`, items: [] });
    }
    const keyRows = getBabyRecordPreviewRows({
      phase: r.phase,
      record_type: r.record_type,
      payload: r.payload as Record<string, unknown>,
    });
    const tmpl = templateForType(r.phase, r.record_type);
    const sum = r.summary?.trim() || "";
    const showSummaryNote = tmpl?.mode === "standard" && !!sum && keyRows.length > 0;
    map.get(key)!.items.push({ item: r, keyRows, showSummaryNote });
  }
  return [...map.entries()].sort((a, b) => b[0].localeCompare(a[0])).map(([, v]) => v);
});

onLoad((query: Record<string, string | undefined>) => {
  babyId.value = Number(query.baby_id || 0);
});

onShow(async () => {
  if (!babyId.value) return;
  await primeDefaultFilters();
  await loadRecords(true);
});

async function loadRecords(reset = false) {
  if (!babyId.value) return;
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
    const res = await apiListRecords(babyId.value, {
      limit: PAGE_SIZE,
      cursor: reset ? undefined : nextCursor.value,
      phase: selectedPhase.value === "all" ? undefined : selectedPhase.value,
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
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${id}&baby_id=${babyId.value}` });
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

.search-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 12rpx;
}

.search-input {
  flex: 1;
  min-width: 0;
  box-sizing: border-box;
  min-height: 68rpx;
  padding: 0 24rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  color: $cj-text;
  font-size: 24rpx;
}

.date-chip,
.date-clear {
  flex-shrink: 0;
  padding: 10rpx 18rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface;
  border: 1rpx solid $cj-border-light;
  color: $cj-text-secondary;
  font-size: 22rpx;
  white-space: nowrap;
}

.date-chip--empty {
  color: $cj-text-muted;
}

.date-clear {
  color: $cj-primary;
}

.filter-scroll {
  width: 100%;
}

.filter-scroll--types {
  margin-top: 4rpx;
}

.filter-scroll--types .filter-inner {
  padding-bottom: 16rpx;
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

.tag {
  flex-shrink: 0;
  padding: 4rpx 14rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  font-weight: $cj-fw-title;

  &.pre {
    background: $cj-tag-prenatal-bg;
    color: $cj-tag-prenatal-text;
  }

  &.post {
    background: $cj-tag-postnatal-bg;
    color: $cj-tag-postnatal-text;
  }
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
