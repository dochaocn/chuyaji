<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">生长趋势</text>
      <text class="head-title">体重变化与记录点</text>
      <text class="head-desc">基于已录入记录生成趋势，不作为医学判断。</text>
    </view>

    <view class="note-card">
      <text class="note-label">说明</text>
      <text class="note-body">{{ sampleNote }}</text>
    </view>

    <GrowthChart :series="series" :ref-label="refLabel" @open-record="openRecord" />

    <button class="ghost-btn" @click="exportJson">导出生长记录 JSON</button>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import GrowthChart from "@/components/GrowthChart.vue";
import sample from "@/assets/who-weight-sample.json";
import { apiGrowthSeries } from "@/api/chuyaji";
import { copyToClipboard, exportAllRecordsJson } from "@/utils/export";
import { useSessionStore } from "@/store/session";

const session = useSessionStore();
const babyId = ref(0);
type GrowthMetricKey = "weight" | "height" | "head";
type GrowthPoint = { record_id: number; age_days: number; value: number; occurred_at: string };
const series = ref<Record<GrowthMetricKey, GrowthPoint[]>>({
  weight: [],
  height: [],
  head: [],
});

const sampleNote = computed(() => (sample as { note?: string }).note || "从宝宝记录中提取体重、身长和头围数据。");
const refLabel = computed(() => "曲线基于已录入的生长记录生成，暂不包含 WHO 参考曲线。");

onLoad(async (query: Record<string, string | undefined>) => {
  session.load();
  babyId.value = Number(query.baby_id || session.babyId || 0);
  await load();
});

async function load() {
  if (!babyId.value) return;
  try {
    const data = await apiGrowthSeries(babyId.value);
    series.value = data.series;
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${id}&baby_id=${babyId.value}` });
}

async function exportJson() {
  if (!babyId.value) return;
  try {
    const text = await exportAllRecordsJson(babyId.value);
    await copyToClipboard(text);
    uni.showToast({ title: "已复制", icon: "none" });
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "导出失败", icon: "none" });
  }
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
}

.head-kicker {
  display: block;
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-mint;
  margin-bottom: 8rpx;
}

.head-title {
  display: block;
  font-size: 42rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

.note-card {
  margin-bottom: $cj-gap-md;
  padding: $cj-gap-md;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.note-label {
  display: block;
  font-size: 22rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
}

.note-body {
  display: block;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

.ghost-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
}
</style>
