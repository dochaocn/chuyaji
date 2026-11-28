<template>
  <view class="page">
    <view class="page-head">
      <text class="head-kicker">生长</text>
      <text class="head-title">生长趋势</text>
    </view>

    <!-- 近况快照（原首页「成长快照」迁移至此） -->
    <view class="panel">
      <view class="panel-head">
        <text class="panel-title">近况</text>
        <text class="panel-link" v-if="babyId" @click="goNewGrowth">记一次生长</text>
      </view>
      <view class="snap-grid">
        <view class="snap-card">
          <text class="snap-label">最近体重</text>
          <text class="snap-value">{{ latestWeightLabel || "—" }}</text>
          <text class="snap-hint">{{ weightAgeHint }}</text>
        </view>
        <view class="snap-card">
          <text class="snap-label">本月记录</text>
          <text class="snap-value">{{ monthlyCount }}</text>
          <text class="snap-hint">条</text>
        </view>
        <view class="snap-card">
          <text class="snap-label">最近里程碑</text>
          <text class="snap-value snap-value--small">{{ latestMilestone || "—" }}</text>
        </view>
      </view>
    </view>

    <GrowthChart :series="series" :reference="reference" :ref-label="refLabel" @open-record="openRecord" />

    <view class="note-card">
      <text class="note-label">说明</text>
      <text class="note-body">{{ hintText }}</text>
    </view>

    <button class="ghost-btn" @click="exportJson">导出生长记录 JSON</button>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import GrowthChart from "@/components/GrowthChart.vue";
import {
  apiBabyDashboard,
  apiGrowthSeries,
  type GrowthMetricKey,
  type GrowthRefPoint,
  type GrowthSeriesPoint,
  type RecordItem,
} from "@/api/chuyaji";
import { copyToClipboard, exportAllRecordsJson } from "@/utils/export";
import { shanghaiMonthKey } from "@/utils/datetime";
import { useSessionStore } from "@/store/session";

const session = useSessionStore();
const babyId = ref(0);
const hint = ref("");
const series = ref<Record<GrowthMetricKey, GrowthSeriesPoint[]>>({
  weight: [],
  height: [],
  head: [],
});
const reference = ref<Record<GrowthMetricKey, GrowthRefPoint[]>>({
  weight: [],
  height: [],
  head: [],
});

const latestWeightG = ref<number | null>(null);
const latestWeightAt = ref<string>("");
const latestRecords = ref<RecordItem[]>([]);

const hintText = computed(
  () => hint.value || "参考曲线为 WHO 2006 近似百分位带（P3/P50/P97），仅供对照，不替代医生评估。"
);
const refLabel = "当前叠加 WHO 2006 参考曲线（P3 / P50 / P97）。";

const latestWeightLabel = computed(() => {
  if (latestWeightG.value === null) return "";
  return `${(latestWeightG.value / 1000).toFixed(2)} kg`;
});

const weightAgeHint = computed(() => {
  if (!latestWeightAt.value) return "暂无记录";
  const diff = Math.floor((Date.now() - Date.parse(latestWeightAt.value)) / (24 * 60 * 60 * 1000));
  return diff >= 0 ? `${diff} 天前` : "";
});

const monthlyCount = computed(() => {
  const month = shanghaiMonthKey(new Date());
  return latestRecords.value.filter((r) => shanghaiMonthKey(r.occurred_at) === month).length;
});

const latestMilestone = computed(() => {
  const hit = latestRecords.value.find((r) => r.record_type === "development");
  if (!hit) return "";
  return String((hit.payload as Record<string, unknown>)?.milestone || hit.summary || "").trim();
});

onLoad(async (query: Record<string, string | undefined>) => {
  session.load();
  babyId.value = Number(query.baby_id || session.babyId || 0);
  await Promise.all([load(), loadSnapshot()]);
});

async function load() {
  if (!babyId.value) return;
  try {
    const data = await apiGrowthSeries(babyId.value);
    series.value = data.series;
    reference.value = data.reference || { weight: [], height: [], head: [] };
    hint.value = data.hint || "";
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

async function loadSnapshot() {
  try {
    const data = await apiBabyDashboard(babyId.value || undefined, 50);
    const weight = data.growth_summary?.latest_weight_g;
    latestWeightG.value = typeof weight === "number" ? weight : null;
    latestWeightAt.value = String(data.growth_summary?.latest_weight_at || "");
    latestRecords.value = data.latest_records || [];
  } catch (error) {
    console.error(error);
  }
}

function goNewGrowth() {
  if (!babyId.value) return;
  uni.navigateTo({
    url: `/pages/baby/record-edit?baby_id=${babyId.value}&phase=postnatal&type=growth`,
  });
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
  padding: $cj-page-pad-y $cj-page-pad-x 64rpx;
}

.page-head {
  margin-bottom: $cj-gap-md;
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
  font-size: 40rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  letter-spacing: -0.5rpx;
}

.panel {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 26rpx $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.panel-title {
  font-size: 30rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.panel-link {
  font-size: 24rpx;
  color: $cj-primary;
}

.snap-grid {
  display: flex;
  gap: 14rpx;
}

.snap-card {
  flex: 1;
  min-width: 0;
  padding: 18rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-lg;
}

.snap-label {
  display: block;
  font-size: 21rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
}

.snap-value {
  display: block;
  font-size: 34rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  line-height: 1.2;
}

.snap-value--small {
  font-size: 26rpx;
}

.snap-hint {
  display: block;
  margin-top: 6rpx;
  font-size: 20rpx;
  color: $cj-text-secondary;
}

.note-card {
  margin-bottom: $cj-gap-md;
  padding: 28rpx;
  background: $cj-surface;
  border-radius: $cj-radius-xl;
  border: 1rpx solid $cj-border-faint;
  box-shadow: $cj-shadow-card;
}

.note-label {
  display: block;
  font-size: 21rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
  font-weight: 500;
}

.note-body {
  display: block;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.ghost-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-faint !important;
}
</style>
