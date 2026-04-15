<template>
  <view class="page">
    <view class="head">
      <view class="head-deco" aria-hidden="true" />
      <text class="head-title">{{ recordId ? "编辑记录" : "写一条记录" }}</text>
      <text class="head-desc">选阶段与类型，填日期和一句话摘要；详细数据可写在下方扩展字段里。</text>
    </view>

    <view class="section">
      <text class="section-label">阶段与类型</text>
      <picker :range="phaseLabels" :value="phaseIndex" @change="onPhase">
        <view class="picker-line">
          <text class="picker-k">阶段</text>
          <text class="picker-v">{{ phase === "prenatal" ? "孕期" : "产后" }}</text>
          <text class="picker-chev">›</text>
        </view>
      </picker>
      <picker :range="typeLabels" :value="typeIndex" @change="onType">
        <view class="picker-line">
          <text class="picker-k">类型</text>
          <text class="picker-v">{{ currentTypeLabel }}</text>
          <text class="picker-chev">›</text>
        </view>
      </picker>
    </view>

    <view class="section">
      <text class="section-label">日期与摘要</text>
      <view class="field-inner">
        <text class="lab">发生日期</text>
        <input v-model="occurredAt" class="input bare" placeholder="yyyy-mm-dd" />
      </view>
      <view class="field-inner">
        <text class="lab">摘要</text>
        <textarea v-model="summary" class="area bare" placeholder="一句话，方便在时间线里辨认" />
      </view>
    </view>

    <view v-if="extraFields.length" class="section extra-block">
      <text class="section-label">扩展字段</text>
      <view v-for="f in extraFields" :key="f.key" class="row">
        <text class="lk">{{ f.label }}</text>
        <input v-model="extras[f.key]" class="mini" :placeholder="f.ph" />
      </view>
    </view>

    <view class="section adv">
      <text class="section-label muted">高级（可选）</text>
      <textarea v-model="payloadJson" class="area small" placeholder="payload JSON，一般可留空" />
    </view>

    <button type="primary" class="save-btn" :loading="loading" @click="save">保存</button>
    <view v-if="!recordId" class="tools">
      <button class="btn2" @click="loadDraft">载入草稿</button>
      <button class="btn2" @click="copyLast">复制上一条同类型</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, reactive, ref, watch } from "vue";
import { apiCreateRecord, apiGetRecord, apiListRecords, apiPatchRecord } from "@/api/chuyaji";
import { RECORD_TYPES, labelForType } from "@/utils/recordTypes";

const babyId = ref(0);
const recordId = ref(0);
const phase = ref<"prenatal" | "postnatal">("prenatal");
const recordType = ref("ultrasound");
const occurredAt = ref("");
const summary = ref("");
const payloadJson = ref("");
const loading = ref(false);

const extras = reactive<Record<string, string>>({});

const phaseLabels = ["prenatal:孕期", "postnatal:产后"];
const phaseIndex = computed(() => (phase.value === "prenatal" ? 0 : 1));

const typeLabels = computed(() => RECORD_TYPES[phase.value].map((x) => `${x.value}:${x.label}`));
const typeIndex = computed(() => {
  const list = RECORD_TYPES[phase.value];
  const idx = list.findIndex((x) => x.value === recordType.value);
  return idx >= 0 ? idx : 0;
});

const currentTypeLabel = computed(() => labelForType(phase.value, recordType.value));

const extraFields = computed(() => {
  const t = recordType.value;
  if (t === "ultrasound") {
    return [
      { key: "hospital", label: "医院", ph: "可选" },
      { key: "bpd_mm", label: "BPD mm", ph: "可选" },
    ];
  }
  if (t === "growth") {
    return [{ key: "weight_g", label: "体重 g", ph: "例如 3500" }];
  }
  if (t === "temperature") {
    return [{ key: "celsius", label: "体温 ℃", ph: "例如 36.7" }];
  }
  return [];
});

watch(
  () => phase.value,
  () => {
    recordType.value = RECORD_TYPES[phase.value][0].value;
    for (const k of Object.keys(extras)) delete extras[k];
  }
);

onLoad((q: Record<string, string | undefined>) => {
  babyId.value = Number(q.baby_id || 0);
  recordId.value = Number(q.id || 0);
  if (q.phase === "prenatal" || q.phase === "postnatal") phase.value = q.phase;
  if (q.type) recordType.value = q.type;
  const d = new Date();
  occurredAt.value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
  if (recordId.value) {
    loadExisting();
  } else {
    loadDraft();
  }
});

function onPhase(e: { detail: { value: string } }) {
  phase.value = Number(e.detail.value) === 0 ? "prenatal" : "postnatal";
  recordType.value = RECORD_TYPES[phase.value][0].value;
}

function onType(e: { detail: { value: string } }) {
  const i = Number(e.detail.value);
  const v = RECORD_TYPES[phase.value][i]?.value;
  if (v) recordType.value = v;
}

function draftKey() {
  return `draft_record_${babyId.value}_${phase.value}_${recordType.value}`;
}

async function loadExisting() {
  if (!recordId.value) return;
  const r = await apiGetRecord(recordId.value);
  phase.value = r.phase;
  recordType.value = r.record_type;
  occurredAt.value = (r.occurred_at || "").slice(0, 10);
  summary.value = r.summary || "";
  payloadJson.value = JSON.stringify(r.payload || {}, null, 2);
}

function loadDraft() {
  try {
    const raw = uni.getStorageSync(draftKey()) as string;
    if (!raw) return;
    const o = JSON.parse(raw) as Record<string, string>;
    occurredAt.value = o.occurredAt || occurredAt.value;
    summary.value = o.summary || "";
    payloadJson.value = o.payloadJson || "";
    uni.showToast({ title: "已载入草稿", icon: "none" });
  } catch {
    /* ignore */
  }
}

async function copyLast() {
  try {
    const page = await apiListRecords(babyId.value, 50);
    const hit = page.items.find((x) => x.phase === phase.value && x.record_type === recordType.value);
    if (!hit) {
      uni.showToast({ title: "没有同类型记录", icon: "none" });
      return;
    }
    summary.value = hit.summary || "";
    payloadJson.value = JSON.stringify(hit.payload || {}, null, 2);
    occurredAt.value = (hit.occurred_at || "").slice(0, 10);
    uni.showToast({ title: "已复制", icon: "none" });
  } catch (e) {
    console.error(e);
  }
}

function buildPayload(): Record<string, unknown> {
  let base: Record<string, unknown> = {};
  if (payloadJson.value.trim()) {
    try {
      base = JSON.parse(payloadJson.value) as Record<string, unknown>;
    } catch {
      throw new Error("payload JSON 无效");
    }
  }
  for (const f of extraFields.value) {
    const v = extras[f.key]?.trim();
    if (!v) continue;
    if (f.key === "weight_g" || f.key === "bpd_mm") {
      const n = Number(v);
      if (Number.isFinite(n)) base[f.key] = n;
    } else if (f.key === "celsius") {
      const n = Number(v);
      if (Number.isFinite(n)) base[f.key] = n;
    } else {
      base[f.key] = v;
    }
  }
  base.ai_summary = null;
  return base;
}

async function save() {
  if (!babyId.value) {
    uni.showToast({ title: "缺少 baby_id", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    const at = new Date(occurredAt.value + "T12:00:00").toISOString();
    const payload = buildPayload();
    if (recordId.value) {
      await apiPatchRecord(recordId.value, {
        phase: phase.value,
        record_type: recordType.value,
        occurred_at: at,
        summary: summary.value,
        payload,
      });
    } else {
      await apiCreateRecord(babyId.value, {
        phase: phase.value,
        record_type: recordType.value,
        occurred_at: at,
        summary: summary.value,
        payload,
      });
      try {
        uni.removeStorageSync(draftKey());
      } catch {
        /* ignore */
      }
    }
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => uni.navigateBack(), 400);
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "保存失败", icon: "none" });
    try {
      uni.setStorageSync(
        draftKey(),
        JSON.stringify({ occurredAt: occurredAt.value, summary: summary.value, payloadJson: payloadJson.value })
      );
    } catch {
      /* ignore */
    }
  } finally {
    loading.value = false;
  }
}

watch(
  () => [occurredAt.value, summary.value, payloadJson.value],
  () => {
    if (recordId.value) return;
    try {
      uni.setStorageSync(
        draftKey(),
        JSON.stringify({ occurredAt: occurredAt.value, summary: summary.value, payloadJson: payloadJson.value })
      );
    } catch {
      /* ignore */
    }
  }
);
</script>

<style lang="scss" scoped>
@keyframes cj-fade-up {
  from {
    opacity: 0;
    transform: translateY(16rpx);
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
  position: relative;
  margin-bottom: $cj-gap-lg;
  padding-left: 8rpx;
  animation: cj-fade-up 0.52s cubic-bezier(0.22, 1, 0.36, 1) backwards;
}

.head-deco {
  position: absolute;
  left: 0;
  top: -8rpx;
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  background: radial-gradient(circle at 30% 30%, $cj-primary-soft 0%, transparent 70%);
  opacity: 0.9;
  z-index: 0;
}

.head-title {
  position: relative;
  z-index: 1;
  display: block;
  font-size: 44rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 1rpx;
}

.head-desc {
  position: relative;
  z-index: 1;
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.55;
  max-width: 100%;
}

.section {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-md $cj-gap-md $cj-gap-lg;
  margin-bottom: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}

.section-label {
  display: block;
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 3rpx;
  text-transform: uppercase;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-md;
}

.section-label.muted {
  opacity: 0.85;
}

.page > .section {
  animation: cj-fade-up 0.48s cubic-bezier(0.22, 1, 0.36, 1) backwards;
}

.page > .section:nth-of-type(2) {
  animation-delay: 0.06s;
}
.page > .section:nth-of-type(3) {
  animation-delay: 0.12s;
}
.page > .section:nth-of-type(4) {
  animation-delay: 0.18s;
}
.page > .section:nth-of-type(5) {
  animation-delay: 0.22s;
}

.tools {
  animation: cj-fade-up 0.45s cubic-bezier(0.22, 1, 0.36, 1) 0.3s backwards;
}

.picker-line {
  display: flex;
  align-items: center;
  padding: 22rpx 0;
  border-bottom: 1rpx solid $cj-border-light;
  font-size: 28rpx;
}

.picker-line:last-of-type {
  border-bottom: none;
  padding-bottom: 0;
}

.picker-k {
  width: 120rpx;
  color: $cj-text-secondary;
}

.picker-v {
  flex: 1;
  color: $cj-ink;
  font-weight: $cj-fw-title;
}

.picker-chev {
  color: $cj-accent;
  font-size: 36rpx;
  line-height: 1;
  opacity: 0.7;
}

.field-inner {
  margin-bottom: $cj-gap-md;
}

.field-inner:last-child {
  margin-bottom: 0;
}

.lab {
  display: block;
  font-size: 24rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
}

.input.bare,
.area.bare {
  background: $cj-accent-soft;
  border: 1rpx dashed rgba(212, 165, 116, 0.45);
}

.input,
.area {
  background: $cj-surface;
  padding: $cj-gap-md;
  border-radius: $cj-radius-md;
  width: 100%;
  box-sizing: border-box;
  border: 1rpx solid $cj-border-light;
  font-size: 28rpx;
  color: $cj-text;
}

.area {
  min-height: 160rpx;
}

.small {
  min-height: 120rpx;
  font-size: 22rpx;
  font-family: ui-monospace, monospace;
  background: $cj-surface-2;
  border-style: solid;
  border-color: $cj-border-light;
}

.extra-block {
  background: linear-gradient(165deg, $cj-mint-soft 0%, $cj-surface 48%);
  border-color: rgba(143, 184, 168, 0.35);
}

.extra-block .row:last-child {
  margin-bottom: 0;
}

.adv {
  background: $cj-surface-2;
  border-style: dashed;
  border-color: rgba(234, 217, 204, 0.9);
}

.row {
  display: flex;
  align-items: center;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-sm;
}

.lk {
  width: 160rpx;
  font-size: 26rpx;
  color: $cj-text-secondary;
  flex-shrink: 0;
}

.mini {
  flex: 1;
  background: rgba(255, 253, 249, 0.95);
  padding: 16rpx 18rpx;
  border-radius: $cj-radius-sm;
  border: 1rpx solid $cj-border-light;
}

.save-btn {
  margin-top: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
  animation: cj-fade-up 0.45s cubic-bezier(0.22, 1, 0.36, 1) 0.26s backwards;
}

.tools {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.btn2 {
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
  border-radius: $cj-radius-pill !important;
  font-size: 26rpx !important;
}
</style>
