<template>
  <view class="wrap">
    <view class="tabs">
      <view
        v-for="metric in metrics"
        :key="metric.key"
        :class="['tab', selectedMetric === metric.key ? 'tab--active' : '']"
        @click="selectedMetric = metric.key"
      >
        <text class="tab-text">{{ metric.label }}</text>
      </view>
    </view>

    <view class="range-tabs">
      <view
        v-for="range in ranges"
        :key="range.key"
        :class="['range-tab', selectedRange === range.key ? 'range-tab--active' : '']"
        @click="selectedRange = range.key"
      >
        <text class="range-tab-text">{{ range.label }}</text>
      </view>
    </view>

    <view class="summary-row">
      <view class="summary-card">
        <text class="summary-label">样本数</text>
        <text class="summary-value">{{ activePoints.length }}</text>
      </view>
      <view class="summary-card">
        <text class="summary-label">最新{{ activeMetric.label }}</text>
        <text class="summary-value">{{ latestLabel }}</text>
      </view>
      <view class="summary-card">
        <text class="summary-label">较上次</text>
        <text class="summary-value">{{ deltaLabel }}</text>
      </view>
    </view>

    <view v-if="activePoints.length >= 2" class="chart">
      <view class="axis-line axis-line--x" />
      <view class="axis-line axis-line--y" />
      <view
        v-for="(segment, idx) in segments"
        :key="`seg-${idx}`"
        class="line-segment"
        :style="{
          left: `${segment.left}%`,
          bottom: `${segment.bottom}%`,
          width: `${segment.width}%`,
          transform: `rotate(${segment.angle}deg)`,
        }"
      />
      <view
        v-for="(point, idx) in normalizedPoints"
        :key="`point-${idx}`"
        class="point"
        :style="{ left: `${point.x}%`, bottom: `${point.y}%` }"
        @click="openPoint(point.record_id)"
      >
        <text class="point-label">{{ formatValue(point.value) }}</text>
      </view>
    </view>

    <view v-else class="empty">
      <text class="empty-title">{{ activePoints.length === 1 ? "已有 1 个记录点" : "暂无生长记录" }}</text>
      <text class="empty-desc">{{ activePoints.length === 1 ? "再记录一次可查看趋势。" : "新增生长记录后会在这里生成曲线。" }}</text>
    </view>

    <view v-if="activePoints.length" class="table">
      <view class="row head">
        <text class="c">日龄</text>
        <text class="c">{{ activeMetric.label }} ({{ activeMetric.unit }})</text>
      </view>
      <view v-for="(p, idx) in activePoints" :key="`table-${idx}`" class="row">
        <text class="c">{{ p.age_days }}</text>
        <text class="c">{{ formatValue(p.value) }}</text>
      </view>
    </view>

    <view v-if="refLabel" class="hint">{{ refLabel }}</view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

export type GrowthMetricKey = "weight" | "height" | "head";

export type GrowthPoint = {
  record_id: number;
  age_days: number;
  value: number;
  occurred_at: string;
};

const props = defineProps<{
  series: Record<GrowthMetricKey, GrowthPoint[]>;
  refLabel?: string;
}>();

const emit = defineEmits<{
  (e: "open-record", id: number): void;
}>();

const metrics: { key: GrowthMetricKey; label: string; unit: string }[] = [
  { key: "weight", label: "体重", unit: "kg" },
  { key: "height", label: "身长", unit: "cm" },
  { key: "head", label: "头围", unit: "cm" },
];

const selectedMetric = ref<GrowthMetricKey>("weight");
const selectedRange = ref<"all" | "3m" | "6m">("all");

const ranges: { key: "all" | "3m" | "6m"; label: string; days: number | null }[] = [
  { key: "all", label: "全部", days: null },
  { key: "3m", label: "近 3 月", days: 92 },
  { key: "6m", label: "近 6 月", days: 184 },
];

const firstAvailableMetric = computed<GrowthMetricKey>(() => {
  return metrics.find((metric) => (props.series[metric.key] ?? []).length > 0)?.key ?? "weight";
});

watch(
  () => firstAvailableMetric.value,
  (metric) => {
    if (!(props.series[selectedMetric.value] ?? []).length) {
      selectedMetric.value = metric;
    }
  },
  { immediate: true }
);

const activeMetric = computed(() => metrics.find((metric) => metric.key === selectedMetric.value) ?? metrics[0]);
const activePoints = computed(() => {
  const sorted = [...(props.series[selectedMetric.value] ?? [])].sort((a, b) => a.age_days - b.age_days);
  const range = ranges.find((item) => item.key === selectedRange.value);
  if (!range?.days || !sorted.length) return sorted;
  const maxAge = sorted[sorted.length - 1].age_days;
  return sorted.filter((point) => point.age_days >= maxAge - range.days!);
});

const latestLabel = computed(() => {
  const latest = activePoints.value[activePoints.value.length - 1];
  return latest ? formatValue(latest.value) : "暂无";
});

const deltaLabel = computed(() => {
  const points = activePoints.value;
  if (points.length < 2) return "暂无";
  const delta = points[points.length - 1].value - points[points.length - 2].value;
  const sign = delta > 0 ? "+" : "";
  return `${sign}${formatValue(delta)}`;
});

const normalizedPoints = computed(() => {
  const points = activePoints.value;
  if (!points.length) return [];
  const minAge = Math.min(...points.map((point) => point.age_days));
  const maxAge = Math.max(...points.map((point) => point.age_days));
  const minValue = Math.min(...points.map((point) => point.value));
  const maxValue = Math.max(...points.map((point) => point.value));
  const ageRange = Math.max(maxAge - minAge, 1);
  const valueRange = Math.max(maxValue - minValue, 1);
  return points.map((point) => ({
    ...point,
    x: 8 + ((point.age_days - minAge) / ageRange) * 84,
    y: 10 + ((point.value - minValue) / valueRange) * 78,
  }));
});

const segments = computed(() => {
  const points = normalizedPoints.value;
  return points.slice(1).map((point, idx) => {
    const prev = points[idx];
    const dx = point.x - prev.x;
    const dy = point.y - prev.y;
    return {
      left: prev.x,
      bottom: prev.y,
      width: Math.sqrt(dx * dx + dy * dy),
      angle: -Math.atan2(dy, dx) * (180 / Math.PI),
    };
  });
});

function formatValue(value: number) {
  return activeMetric.value.key === "weight" ? value.toFixed(2) : value.toFixed(1);
}

function openPoint(id: number) {
  if (id) emit("open-record", id);
}
</script>

<style lang="scss" scoped>
.wrap {
  background: $cj-surface;
  border-radius: $cj-radius-xl;
  padding: 28rpx;
  border: 1rpx solid $cj-border-faint;
  box-shadow: $cj-shadow-card;
}

.tabs {
  display: flex;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-md;
}

.range-tabs {
  display: flex;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-md;
}

.tab {
  flex: 1;
  text-align: center;
  border-radius: $cj-radius-pill;
  padding: 13rpx 0;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  transition: all 0.15s;
}

.tab--active {
  background: $cj-primary;
  border-color: $cj-primary;
  box-shadow: $cj-shadow-soft;
}

.range-tab {
  flex: 1;
  text-align: center;
  border-radius: $cj-radius-pill;
  padding: 9rpx 0;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  transition: all 0.15s;
}

.range-tab--active {
  background: $cj-mint-soft;
  border-color: $cj-mint;
}

.tab-text {
  font-size: 23rpx;
  color: $cj-text;
  letter-spacing: 0.3rpx;
}

.range-tab-text {
  font-size: 21rpx;
  color: $cj-text-secondary;
}

.tab--active .tab-text {
  color: #fffefb;
}

.summary-row {
  display: flex;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-md;
}

.summary-card {
  flex: 1;
  background: $cj-surface-2;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-sm;
}

.summary-label {
  display: block;
  font-size: 21rpx;
  color: $cj-text-muted;
  font-weight: 500;
  letter-spacing: 0.5rpx;
}

.summary-value {
  display: block;
  margin-top: 8rpx;
  font-size: 32rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  letter-spacing: -0.3rpx;
}

.chart {
  position: relative;
  height: 360rpx;
  margin: $cj-gap-md 0;
  border-radius: $cj-radius-lg;
  background: linear-gradient(180deg, rgba(230, 241, 236, 0.4) 0%, rgba(255, 247, 240, 0.4) 100%);
  overflow: hidden;
}

.axis-line {
  position: absolute;
  background: rgba(139, 128, 119, 0.18);
}

.axis-line--x {
  left: 6%;
  right: 6%;
  bottom: 9%;
  height: 1rpx;
}

.axis-line--y {
  left: 7%;
  top: 8%;
  bottom: 9%;
  width: 1rpx;
}

.line-segment {
  position: absolute;
  height: 4rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(90deg, $cj-mint 0%, $cj-primary 100%);
  transform-origin: left center;
}

.point {
  position: absolute;
  width: 16rpx;
  height: 16rpx;
  margin-left: -8rpx;
  margin-bottom: -8rpx;
  border-radius: 50%;
  background: $cj-primary;
  border: 3rpx solid #fffefb;
  box-shadow: 0 3rpx 10rpx rgba(120, 72, 60, 0.14);
}

.point-label {
  position: absolute;
  left: 50%;
  bottom: 22rpx;
  transform: translateX(-50%);
  white-space: nowrap;
  font-size: 19rpx;
  color: $cj-text-secondary;
}

.table {
  margin-top: $cj-gap-md;
}

.row {
  display: flex;
  flex-direction: row;
  padding: 10rpx 0;
  border-bottom: 1rpx solid $cj-border-faint;
}

.head {
  font-weight: $cj-fw-title;
  color: $cj-text;
}

.c {
  flex: 1;
  font-size: 25rpx;
  color: $cj-text-secondary;
}

.empty {
  padding: $cj-gap-md 0;
  color: $cj-text-muted;
  font-size: 24rpx;
  text-align: center;
}

.empty-title,
.empty-desc {
  display: block;
}

.empty-title {
  color: $cj-ink;
  font-size: 28rpx;
  font-weight: $cj-fw-title;
}

.empty-desc {
  margin-top: 8rpx;
  color: $cj-text-secondary;
  font-size: 24rpx;
}

.hint {
  margin-top: $cj-gap-sm;
  font-size: 21rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}
</style>
