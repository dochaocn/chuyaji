<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ recordId ? "编辑宝宝记录" : "新增宝宝记录" }}</text>
      <text class="head-desc">统一记录怀孕期和成长期事件，便于后续在首页和趋势页里直接回看。</text>
    </view>

    <view class="section">
      <text class="section-label">阶段与类型</text>
      <picker :range="phaseLabels" :value="phaseIndex" @change="onPhase">
        <view class="picker-line">
          <text class="picker-k">阶段</text>
          <text class="picker-v">{{ phase === "prenatal" ? "怀孕期" : "成长期" }}</text>
        </view>
      </picker>
      <picker :range="typeLabels" :value="typeIndex" @change="onType">
        <view class="picker-line">
          <text class="picker-k">类型</text>
          <text class="picker-v">{{ currentTypeLabel }}</text>
        </view>
      </picker>
    </view>

    <view class="section">
      <text class="section-label">日期与摘要</text>
      <view class="field">
        <text class="lab">发生日期</text>
        <picker mode="date" :value="occurredAt" @change="onOccurredAtChange">
          <view :class="['picker-input', !occurredAt && 'is-placeholder']">{{ occurredAt || "请选择日期" }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">摘要</text>
        <textarea v-model="summary" class="area" placeholder="一句话描述今天发生了什么" />
      </view>
    </view>

    <view v-if="extraFields.length" class="section">
      <text class="section-label">扩展字段</text>
      <view v-for="field in extraFields" :key="field.key" class="field">
        <text class="lab">{{ field.label }}</text>
        <input v-model="extras[field.key]" class="input" :placeholder="field.placeholder" />
      </view>
    </view>

    <view class="section">
      <text class="section-label">图片补充</text>
      <text class="section-hint">可选，点击图片可全屏预览</text>
      <view v-if="!displayPhotos.length" class="photo-placeholder">暂无图片，可点击下方添加。</view>
      <view v-else class="photo-grid">
        <view v-for="(item, idx) in displayPhotos" :key="item.key" class="photo-cell">
          <image class="photo-img" :src="item.src" mode="aspectFill" :lazy-load="false" @click="previewPhoto(idx)" />
          <view class="photo-remove" @click.stop="removePhoto(item)">
            <text class="photo-remove-x">×</text>
          </view>
        </view>
      </view>
      <button class="add-photo-btn" @click="pickImages">添加图片</button>
    </view>

    <button class="main-btn" :loading="loading" @click="save">保存记录</button>
    <view v-if="!recordId" class="tool-row">
      <button class="ghost-btn" @click="tapLoadDraft">载入草稿</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, reactive, ref, watch } from "vue";
import {
  apiCreateRecord,
  apiDeleteAttachment,
  apiGetRecord,
  apiListAttachments,
  apiPatchRecord,
  type AttachmentItem,
} from "@/api/chuyaji";
import { uploadRecordAttachment } from "@/api/upload";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import { RECORD_TYPES, labelForType } from "@/utils/recordTypes";

const babyId = ref(0);
const recordId = ref(0);
const phase = ref<"prenatal" | "postnatal">("prenatal");
const recordType = ref("prenatal_checkup");
const occurredAt = ref("");
const summary = ref("");
const loading = ref(false);
const extras = reactive<Record<string, string>>({});
const serverAttachments = ref<AttachmentItem[]>([]);
const pendingPaths = ref<{ key: string; path: string }[]>([]);

type DisplayPhoto =
  | { key: string; kind: "server"; id: number; src: string }
  | { key: string; kind: "pending"; src: string; path: string };

const displayPhotos = computed<DisplayPhoto[]>(() => {
  const server: DisplayPhoto[] = serverAttachments.value.map((a) => ({
    key: `s-${a.id}`,
    kind: "server" as const,
    id: a.id,
    src: resolvePublicMediaUrl((a.thumb_url && a.thumb_url.trim()) || a.url),
  }));
  const local: DisplayPhoto[] = pendingPaths.value.map((p) => ({
    key: p.key,
    kind: "pending" as const,
    src: p.path,
    path: p.path,
  }));
  return [...server, ...local];
});

const phaseLabels = ["怀孕期", "成长期"];
const phaseIndex = computed(() => (phase.value === "prenatal" ? 0 : 1));
const typeLabels = computed(() => RECORD_TYPES[phase.value].map((item) => item.label));
const typeIndex = computed(() => {
  const idx = RECORD_TYPES[phase.value].findIndex((item) => item.value === recordType.value);
  return idx >= 0 ? idx : 0;
});
const currentTypeLabel = computed(() => labelForType(phase.value, recordType.value));

const extraFields = computed(() => {
  switch (recordType.value) {
    case "ultrasound":
      return [
        { key: "hospital", label: "医院", placeholder: "可选" },
        { key: "bpd_mm", label: "BPD mm", placeholder: "可选" },
      ];
    case "growth":
      return [
        { key: "weight_g", label: "体重 g", placeholder: "例如 3500" },
        { key: "height_cm", label: "身长 cm", placeholder: "例如 50" },
      ];
    case "temperature":
      return [{ key: "celsius", label: "体温 ℃", placeholder: "例如 36.7" }];
    case "blood_pressure":
      return [{ key: "value", label: "血压", placeholder: "例如 120/80" }];
    default:
      return [];
  }
});

watch(
  () => phase.value,
  () => {
    recordType.value = RECORD_TYPES[phase.value][0].value;
    for (const key of Object.keys(extras)) delete extras[key];
  }
);

onLoad((query: Record<string, string | undefined>) => {
  babyId.value = Number(query.baby_id || 0);
  recordId.value = Number(query.id || 0);
  if (query.phase === "prenatal" || query.phase === "postnatal") {
    phase.value = query.phase;
  }
  if (query.type) {
    recordType.value = query.type;
  }
  const now = new Date();
  occurredAt.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(
    2,
    "0"
  )}`;
  if (recordId.value) {
    loadExisting();
  } else {
    loadDraft({ silent: true });
  }
});

function onPhase(event: { detail: { value: string } }) {
  phase.value = Number(event.detail.value) === 0 ? "prenatal" : "postnatal";
}

function onType(event: { detail: { value: string } }) {
  const next = RECORD_TYPES[phase.value][Number(event.detail.value)]?.value;
  if (next) {
    recordType.value = next;
  }
}

function onOccurredAtChange(event: { detail: { value: string } }) {
  occurredAt.value = event.detail.value || "";
}

function draftKey() {
  return `draft_record_${babyId.value}_${phase.value}_${recordType.value}`;
}

type DraftCache = {
  occurredAt?: string;
  summary?: string;
  pendingPaths?: { key: string; path: string }[];
};

function persistDraft() {
  if (recordId.value) return;
  try {
    const data: DraftCache = {
      occurredAt: occurredAt.value,
      summary: summary.value,
      pendingPaths: pendingPaths.value,
    };
    uni.setStorageSync(draftKey(), JSON.stringify(data));
  } catch {
    /* ignore */
  }
}

async function loadExisting() {
  try {
    const item = await apiGetRecord(recordId.value);
    phase.value = item.phase;
    recordType.value = item.record_type;
    occurredAt.value = item.occurred_at.slice(0, 10);
    summary.value = item.summary || "";
    for (const key of Object.keys(extras)) delete extras[key];
    const pl = item.payload || {};
    for (const field of extraFields.value) {
      const v = pl[field.key];
      if (v != null && v !== "") {
        extras[field.key] = typeof v === "number" ? String(v) : String(v);
      }
    }
    pendingPaths.value = [];
    const att = await apiListAttachments(recordId.value);
    serverAttachments.value = att.items || [];
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

function loadDraft(options?: { silent?: boolean }) {
  const silent = options?.silent ?? true;
  try {
    const raw = uni.getStorageSync(draftKey()) as string;
    if (!raw) {
      if (!silent) {
        uni.showToast({ title: "暂无草稿", icon: "none" });
      }
      return;
    }
    const cached = JSON.parse(raw) as DraftCache;
    occurredAt.value = cached.occurredAt || occurredAt.value;
    summary.value = cached.summary || "";
    pendingPaths.value = Array.isArray(cached.pendingPaths) ? [...cached.pendingPaths] : [];
    if (!silent) {
      uni.showToast({ title: "已载入草稿", icon: "success" });
    }
  } catch {
    if (!silent) {
      uni.showToast({ title: "载入失败", icon: "none" });
    }
  }
}

function tapLoadDraft() {
  loadDraft({ silent: false });
}

function buildPayload() {
  const base: Record<string, unknown> = {};
  for (const field of extraFields.value) {
    const raw = extras[field.key]?.trim();
    if (!raw) continue;
    const maybeNumber = Number(raw);
    base[field.key] = Number.isFinite(maybeNumber) && raw === String(maybeNumber) ? maybeNumber : raw;
  }
  return base;
}

function previewPhoto(index: number) {
  const urls = displayPhotos.value.map((p) =>
    p.kind === "server"
      ? resolvePublicMediaUrl(serverAttachments.value.find((a) => a.id === p.id)?.url || "")
      : p.path
  );
  uni.previewImage({ current: urls[index], urls });
}

function pickImages() {
  const remain = 27 - displayPhotos.value.length;
  if (remain <= 0) {
    uni.showToast({ title: "图片数量已达上限", icon: "none" });
    return;
  }
  uni.chooseImage({
    count: Math.min(9, remain),
    sizeType: ["compressed"],
    success: (res) => {
      for (const path of res.tempFilePaths) {
        pendingPaths.value.push({ key: `p-${Date.now()}-${Math.random().toString(36).slice(2)}`, path });
      }
    },
  });
}

async function removePhoto(item: DisplayPhoto) {
  if (item.kind === "pending") {
    pendingPaths.value = pendingPaths.value.filter((p) => p.key !== item.key);
    return;
  }
  try {
    await apiDeleteAttachment(item.id);
    serverAttachments.value = serverAttachments.value.filter((a) => a.id !== item.id);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "移除失败", icon: "none" });
  }
}

async function uploadPendingForRecord(id: number) {
  let failed = 0;
  const queue = [...pendingPaths.value];
  for (const item of queue) {
    try {
      await uploadRecordAttachment(id, item.path);
      pendingPaths.value = pendingPaths.value.filter((p) => p.key !== item.key);
    } catch (e) {
      console.error(e);
      failed += 1;
    }
  }
  if (failed > 0) {
    uni.showToast({ title: `${failed} 张图片上传失败`, icon: "none" });
  }
}

async function save() {
  if (!babyId.value) {
    uni.showToast({ title: "缺少宝宝档案", icon: "none" });
    return;
  }
  if (!occurredAt.value.trim()) {
    uni.showToast({ title: "请选择发生日期", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    const payload = buildPayload();
    const body = {
      phase: phase.value,
      record_type: recordType.value,
      occurred_at: new Date(`${occurredAt.value}T12:00:00`).toISOString(),
      summary: summary.value,
      payload,
    };
    if (recordId.value) {
      await apiPatchRecord(recordId.value, body);
      await uploadPendingForRecord(recordId.value);
    } else {
      const created = await apiCreateRecord(babyId.value, body);
      await uploadPendingForRecord(created.id);
      uni.removeStorageSync(draftKey());
    }
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => uni.navigateBack(), 250);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "保存失败", icon: "none" });
    persistDraft();
  } finally {
    loading.value = false;
  }
}

watch(
  [occurredAt, summary, pendingPaths],
  () => {
    persistDraft();
  },
  { deep: true }
);
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
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

.section {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.section-label {
  display: block;
  margin-bottom: $cj-gap-md;
  font-size: 22rpx;
  color: $cj-text-muted;
  letter-spacing: 3rpx;
  text-transform: uppercase;
}

.picker-line,
.field {
  margin-bottom: $cj-gap-md;
}

.picker-line:last-child,
.field:last-child {
  margin-bottom: 0;
}

.picker-k,
.lab {
  display: block;
  margin-bottom: 10rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
}

.picker-v,
.input,
.picker-input,
.area {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  color: $cj-text;
  font-size: 28rpx;
}

.input {
  min-height: 108rpx;
  line-height: 1.55;
  padding-top: 26rpx;
  padding-bottom: 26rpx;
}

.picker-v {
  min-height: 108rpx;
  line-height: 1.55;
  padding-top: 26rpx;
  padding-bottom: 26rpx;
  display: flex;
  align-items: center;
}

.picker-input {
  min-height: 108rpx;
  display: flex;
  align-items: center;
  line-height: 1.55;
  padding-top: 26rpx;
  padding-bottom: 26rpx;
}

.picker-input.is-placeholder {
  color: $cj-text-muted;
}

.area {
  min-height: 200rpx;
  line-height: 1.55;
  padding-top: 28rpx;
  padding-bottom: 28rpx;
}

.section-hint {
  display: block;
  margin-top: 8rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.5;
}

.photo-placeholder {
  margin-top: $cj-gap-md;
  padding: 28rpx 0;
  text-align: center;
  font-size: 24rpx;
  color: $cj-text-muted;
}

.photo-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: $cj-gap-md;
}

.photo-cell {
  position: relative;
  width: calc((100% - 32rpx) / 3);
  box-sizing: border-box;
}

.photo-img {
  display: block;
  width: 100%;
  height: 200rpx;
  border-radius: $cj-radius-md;
  background: $cj-surface-2;
}

.photo-remove {
  position: absolute;
  top: 8rpx;
  right: 8rpx;
  width: 44rpx;
  height: 44rpx;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
}

.photo-remove-x {
  color: #fff;
  font-size: 32rpx;
  line-height: 1;
}

.add-photo-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
}

.main-btn,
.ghost-btn {
  border-radius: $cj-radius-pill !important;
}

.main-btn {
  margin-top: $cj-gap-sm;
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}

.tool-row {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.ghost-btn {
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
}
</style>
