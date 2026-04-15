<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">生长</text>
      <text class="head-title">体重与时间</text>
      <text class="head-desc">从「生长」类记录里读取体重，按出生日换算日龄。</text>
    </view>

    <view class="note-card">
      <text class="note-label">说明</text>
      <text class="note-body">{{ sampleNote }}</text>
    </view>

    <view class="chart-wrap">
      <GrowthChart :user-points="userPts" :ref-label="refLabel" />
    </view>

    <button class="export-btn" @click="exportJson">导出宝宝记录 JSON 到剪贴板</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import GrowthChart from "@/components/GrowthChart.vue";
import { apiGetBaby, apiListRecords } from "@/api/chuyaji";
import sample from "@/assets/who-weight-sample.json";
import { copyToClipboard, exportAllRecordsJson } from "@/utils/export";
import { useSessionStore } from "@/store/session";

const babyId = ref(0);
const birth = ref<string | undefined>(undefined);
const userPts = ref<{ age_days: number; kg: number }[]>([]);
const session = useSessionStore();

const sampleNote = computed(() => (sample as { note?: string }).note || "");

const refLabel = computed(() => {
  const p50 = (sample as { p50?: { age_days: number; kg: number }[] }).p50 || [];
  if (!p50.length) return "";
  return "参考线为内置示例点，不等同于官方标准表。";
});

onLoad(async (q: Record<string, string | undefined>) => {
  babyId.value = Number(q.baby_id || session.babyId || 0);
  await load();
});

async function load() {
  if (!babyId.value) return;
  try {
    const b = await apiGetBaby(babyId.value);
    birth.value = b.birth_date;
    const pts: { age_days: number; kg: number }[] = [];
    let cursor: string | undefined;
    for (;;) {
      const page = await apiListRecords(babyId.value, 50, cursor);
      for (const r of page.items) {
        if (r.record_type !== "growth") continue;
        const w = (r.payload as { weight_g?: number }).weight_g;
        if (!birth.value || !w || !Number.isFinite(w)) continue;
        const bd = Date.parse(birth.value);
        const od = Date.parse(r.occurred_at);
        if (!Number.isFinite(bd) || !Number.isFinite(od)) continue;
        const age_days = Math.floor((od - bd) / (24 * 3600 * 1000));
        pts.push({ age_days, kg: w / 1000 });
      }
      if (!page.next_cursor) break;
      cursor = page.next_cursor;
    }
    userPts.value = pts;
  } catch (e) {
    console.error(e);
  }
}

async function exportJson() {
  if (!babyId.value) return;
  try {
    const text = await exportAllRecordsJson(babyId.value);
    await copyToClipboard(text);
    uni.showToast({ title: "已复制", icon: "none" });
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "失败", icon: "none" });
  }
}
</script>

<style lang="scss" scoped>
@keyframes cj-fade-up {
  from {
    opacity: 0;
    transform: translateY(12rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page {
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
  animation: cj-fade-up 0.5s ease backwards;
}

.head-kicker {
  display: block;
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 4rpx;
  text-transform: uppercase;
  color: $cj-mint;
  margin-bottom: 8rpx;
}

.head-title {
  display: block;
  font-size: 42rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.6;
}

.note-card {
  position: relative;
  padding: $cj-gap-md $cj-gap-md $cj-gap-md 28rpx;
  margin-bottom: $cj-gap-md;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
  animation: cj-fade-up 0.48s ease 0.06s backwards;
}

.note-card::before {
  content: "";
  position: absolute;
  left: 0;
  top: $cj-gap-md;
  bottom: $cj-gap-md;
  width: 6rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(180deg, $cj-mint 0%, $cj-accent 100%);
}

.note-label {
  display: block;
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
}

.note-body {
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.chart-wrap {
  margin-bottom: $cj-gap-lg;
  animation: cj-fade-up 0.48s ease 0.1s backwards;
}

.export-btn {
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
  border-radius: $cj-radius-pill !important;
  font-size: 26rpx !important;
  animation: cj-fade-up 0.45s ease 0.14s backwards;
}
</style>
