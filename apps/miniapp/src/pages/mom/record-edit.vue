<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ pageTitle }}</text>
      <text v-if="currentTemplate?.summaryPlaceholder" class="head-desc">{{ currentTemplate.summaryPlaceholder }}</text>
    </view>

    
    <view class="section">
      <text class="section-label">类型</text>
      <picker :range="typeLabels" :value="typeIndex" @change="onType">
        <view class="picker-line">
          <text class="picker-v">{{ currentTypeLabel }}</text>
        </view>
      </picker>
    </view>

    
    <view class="section">
      <text class="section-label">日期</text>
      <view class="field">
        <text class="lab">发生日期</text>
        <picker mode="date" :value="occurredAt" @change="onOccurredAtChange">
          <view :class="['picker-input', !occurredAt && 'is-placeholder']">{{ occurredAt || "请选择日期" }}</view>
        </picker>
      </view>
    </view>

    
    <view v-if="recommendedFields.length" class="section">
      <text class="section-label">关键信息</text>
      <view v-for="field in recommendedFields" :key="field.key" class="field">
        <text class="lab">{{ field.label }}<text v-if="field.unit" class="unit-text">（{{ field.unit }}）</text></text>
        <template v-if="field.type === 'select'">
          <picker :range="(field.options || []).map(o => o.label)" :value="selectIndex(field)" @change="(e) => onSelectChange(field, e)">
            <view class="picker-input">{{ selectLabel(field) || "请选择" }}</view>
          </picker>
        </template>
        <template v-else-if="field.type === 'date'">
          <picker mode="date" :value="extras[field.key] || ''" @change="(e) => extras[field.key] = e.detail.value">
            <view :class="['picker-input', !extras[field.key] && 'is-placeholder']">{{ extras[field.key] || "请选择日期" }}</view>
          </picker>
        </template>
        <template v-else>
          <input
            v-model="extras[field.key]"
            class="input"
            :type="field.type === 'number' ? 'digit' : 'text'"
            :placeholder="field.placeholder || ''"
          />
        </template>
      </view>
    </view>

    
    <view v-if="currentTemplate?.mode === 'standard'" class="section">
      <text class="section-label">摘要</text>
      <view class="field">
        <textarea
          v-model="summary"
          class="area"
          :placeholder="currentTemplate?.summaryPlaceholder || '一句话记录当前变化'"
        />
      </view>
    </view>

    
    <view v-if="optionalFields.length" class="section">
      <view class="optional-toggle" @click="showOptional = !showOptional">
        <text class="optional-toggle-text">{{ showOptional ? "收起更多信息" : "展开更多信息" }}</text>
      </view>
      <template v-if="showOptional">
        <view v-for="field in optionalFields" :key="field.key" class="field" style="margin-top: 24rpx;">
          <text class="lab">{{ field.label }}<text v-if="field.unit" class="unit-text">（{{ field.unit }}）</text></text>
          <template v-if="field.type === 'select'">
            <picker :range="(field.options || []).map(o => o.label)" :value="selectIndex(field)" @change="(e) => onSelectChange(field, e)">
              <view class="picker-input">{{ selectLabel(field) || "请选择" }}</view>
            </picker>
          </template>
          <template v-else-if="field.type === 'date'">
            <picker mode="date" :value="extras[field.key] || ''" @change="(e) => extras[field.key] = e.detail.value">
              <view :class="['picker-input', !extras[field.key] && 'is-placeholder']">{{ extras[field.key] || "请选择日期" }}</view>
            </picker>
          </template>
          <template v-else>
            <input
              v-model="extras[field.key]"
              class="input"
              :type="field.type === 'number' ? 'digit' : 'text'"
              :placeholder="field.placeholder || ''"
            />
          </template>
        </view>
      </template>
    </view>

    
    <view class="section">
      <text class="section-label">图片补充</text>
      <text v-if="currentTemplate?.attachmentHint" class="section-hint">{{ currentTemplate.attachmentHint }}</text>
      <text v-else class="section-hint">可选，点击图片可全屏预览</text>
      <view v-if="!displayPhotos.length" class="photo-placeholder">暂无图片，可点击下方添加。</view>
      <view v-else class="photo-grid">
        <view v-for="(item, idx) in displayPhotos" :key="item.key" class="photo-cell">
          <image class="photo-img" :src="item.src" mode="aspectFill" @click="previewPhoto(idx)" />
          <view class="photo-remove" @click.stop="removePhoto(item)">
            <text class="photo-remove-x">×</text>
          </view>
        </view>
      </view>
      <button class="add-photo-btn" @click="pickImages">添加图片</button>
    </view>

    <button class="main-btn" :loading="loading" @click="save">保存记录</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, reactive, ref, watch } from "vue";
import {
  apiCreateMotherRecord, apiDeleteAttachment, apiGetMotherRecord,
  apiListMotherAttachments, apiPatchMotherRecord, type AttachmentItem,
} from "@/api/chuyaji";
import { uploadMotherRecordAttachment } from "@/api/upload";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import { MOTHER_TEMPLATES, labelForMotherType, type RecordTemplate, type TemplateField } from "@/utils/motherRecordTypes";

const motherId = ref(0);
const recordId = ref(0);
const recordType = ref("checkup");
const occurredAt = ref("");
const summary = ref("");
const loading = ref(false);
const showOptional = ref(false);
const extras = reactive<Record<string, string>>({});
const serverPayload = ref<Record<string, unknown>>({});
const serverAttachments = ref<AttachmentItem[]>([]);
const pendingPaths = ref<{ key: string; path: string }[]>([]);

type DisplayPhoto =
  | { key: string; kind: "server"; id: number; src: string }
  | { key: string; kind: "pending"; src: string; path: string };

const displayPhotos = computed<DisplayPhoto[]>(() => {
  const server: DisplayPhoto[] = serverAttachments.value.map((a) => ({
    key: `s-${a.id}`, kind: "server" as const, id: a.id,
    src: resolvePublicMediaUrl((a.thumb_url?.trim()) || a.url),
  }));
  const local: DisplayPhoto[] = pendingPaths.value.map((p) => ({
    key: p.key, kind: "pending" as const, src: p.path, path: p.path,
  }));
  return [...server, ...local];
});

const typeLabels = computed(() => MOTHER_TEMPLATES.map((t) => t.label));
const typeIndex = computed(() => {
  const idx = MOTHER_TEMPLATES.findIndex((t) => t.value === recordType.value);
  return idx >= 0 ? idx : 0;
});
const currentTypeLabel = computed(() => labelForMotherType(recordType.value));

const currentTemplate = computed<RecordTemplate | undefined>(() =>
  MOTHER_TEMPLATES.find((t) => t.value === recordType.value)
);

const pageTitle = computed(() => {
  if (recordId.value) return `编辑${currentTemplate.value?.label || "记录"}`;
  return currentTemplate.value?.label ? `新增${currentTemplate.value.label}` : "新增宝妈记录";
});

const recommendedFields = computed(() => currentTemplate.value?.recommendedFields ?? []);
const optionalFields = computed(() => currentTemplate.value?.optionalFields ?? []);

function selectIndex(field: TemplateField): number {
  const opts = field.options || [];
  const idx = opts.findIndex((o) => o.value === extras[field.key]);
  return idx >= 0 ? idx : 0;
}

function selectLabel(field: TemplateField): string {
  const hit = (field.options || []).find((o) => o.value === extras[field.key]);
  return hit ? hit.label : "";
}

function onSelectChange(field: TemplateField, event: { detail: { value: string } }) {
  const opt = (field.options || [])[Number(event.detail.value)];
  if (opt) extras[field.key] = opt.value;
}

watch(
  () => recordType.value,
  () => {
    for (const key of Object.keys(extras)) delete extras[key];
    showOptional.value = false;
  },
  { flush: "sync" }
);

onLoad((query: Record<string, string | undefined>) => {
  motherId.value = Number(query.mother_id || 0);
  recordId.value = Number(query.id || 0);
  if (query.type) recordType.value = query.type;
  const now = new Date();
  occurredAt.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  if (recordId.value) loadExisting();
});

function onType(event: { detail: { value: string } }) {
  const next = MOTHER_TEMPLATES[Number(event.detail.value)]?.value;
  if (next) recordType.value = next;
}

function onOccurredAtChange(event: { detail: { value: string } }) {
  occurredAt.value = event.detail.value || "";
}

async function loadExisting() {
  try {
    const item = await apiGetMotherRecord(recordId.value);
    recordType.value = item.record_type;
    occurredAt.value = item.occurred_at.slice(0, 10);
    summary.value = item.summary || "";
    const pl = item.payload || {};
    serverPayload.value = { ...pl };
    for (const key of Object.keys(extras)) delete extras[key];
    for (const [key, val] of Object.entries(pl)) {
      if (val != null && val !== "") extras[key] = String(val);
    }
    pendingPaths.value = [];
    const att = await apiListMotherAttachments(recordId.value);
    serverAttachments.value = att.items || [];
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

function buildPayload(): Record<string, unknown> {
  const base: Record<string, unknown> = {};
  const allFields = [...recommendedFields.value, ...optionalFields.value];
  for (const field of allFields) {
    const raw = extras[field.key]?.trim();
    if (!raw) continue;
    const maybeNumber = Number(raw);
    base[field.key] = field.type === "number" && Number.isFinite(maybeNumber) ? maybeNumber : raw;
  }
  return base;
}

function buildSummary(payload: Record<string, unknown>): string {
  if (summary.value.trim()) return summary.value.trim();
  if (currentTemplate.value?.autoSummary) return currentTemplate.value.autoSummary(payload);
  return currentTemplate.value?.label + "已记录";
}

function previewPhoto(index: number) {
  const urls = displayPhotos.value.map((p) =>
    p.kind === "server" ? resolvePublicMediaUrl(serverAttachments.value.find((a) => a.id === p.id)?.url || "") : p.path
  );
  uni.previewImage({ current: urls[index], urls });
}

function pickImages() {
  const remain = 27 - displayPhotos.value.length;
  if (remain <= 0) { uni.showToast({ title: "图片数量已达上限", icon: "none" }); return; }
  uni.chooseImage({
    count: Math.min(9, remain), sizeType: ["compressed"],
    success: (res) => { for (const path of res.tempFilePaths) pendingPaths.value.push({ key: `p-${Date.now()}-${Math.random().toString(36).slice(2)}`, path }); },
  });
}

async function removePhoto(item: DisplayPhoto) {
  if (item.kind === "pending") { pendingPaths.value = pendingPaths.value.filter((p) => p.key !== item.key); return; }
  try {
    await apiDeleteAttachment(item.id);
    serverAttachments.value = serverAttachments.value.filter((a) => a.id !== item.id);
  } catch (error) { console.error(error); uni.showToast({ title: "移除失败", icon: "none" }); }
}

async function uploadPendingForRecord(id: number) {
  let failed = 0;
  for (const item of [...pendingPaths.value]) {
    try { await uploadMotherRecordAttachment(id, item.path); pendingPaths.value = pendingPaths.value.filter((p) => p.key !== item.key); }
    catch (e) { console.error(e); failed++; }
  }
  if (failed > 0) uni.showToast({ title: `${failed} 张图片上传失败`, icon: "none" });
}

async function save() {
  if (!motherId.value) { uni.showToast({ title: "缺少宝妈档案", icon: "none" }); return; }
  if (!occurredAt.value.trim()) { uni.showToast({ title: "请选择发生日期", icon: "none" }); return; }
  loading.value = true;
  try {
    const freshPayload = buildPayload();
    const payload = recordId.value ? { ...serverPayload.value, ...freshPayload } : freshPayload;
    const finalSummary = buildSummary(payload);
    const body = { record_type: recordType.value, occurred_at: new Date(`${occurredAt.value}T12:00:00`).toISOString(), summary: finalSummary, payload };
    if (recordId.value) {
      await apiPatchMotherRecord(recordId.value, body);
      await uploadPendingForRecord(recordId.value);
    } else {
      const created = await apiCreateMotherRecord(motherId.value, body);
      await uploadPendingForRecord(created.id);
    }
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => uni.navigateBack(), 250);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "保存失败", icon: "none" });
  } finally { loading.value = false; }
}
</script>

<style lang="scss" scoped>
.page { padding: $cj-page-pad-y $cj-page-pad-x 56rpx; }

.head { margin-bottom: $cj-gap-lg; }
.head-title { display: block; font-size: 42rpx; font-weight: $cj-fw-display; color: $cj-ink; }
.head-desc { display: block; margin-top: $cj-gap-sm; font-size: 26rpx; color: $cj-text-secondary; line-height: 1.6; }

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
}

.optional-toggle { padding: 4rpx 0; }
.optional-toggle-text { font-size: 24rpx; color: $cj-primary; }

.picker-line, .field { margin-bottom: $cj-gap-md; }
.picker-line:last-child, .field:last-child { margin-bottom: 0; }

.lab { display: block; margin-bottom: 10rpx; font-size: 24rpx; color: $cj-text-muted; }
.unit-text { font-size: 22rpx; }

.picker-v, .input, .picker-input, .area {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  color: $cj-text;
  font-size: 28rpx;
}

.input, .picker-v, .picker-input { min-height: 108rpx; line-height: 1.55; padding-top: 26rpx; padding-bottom: 26rpx; }
.picker-v, .picker-input { display: flex; align-items: center; }
.picker-input.is-placeholder { color: $cj-text-muted; }
.area { min-height: 200rpx; line-height: 1.55; padding-top: 28rpx; padding-bottom: 28rpx; }

.section-hint { display: block; margin-top: 8rpx; font-size: 22rpx; color: $cj-text-muted; line-height: 1.5; }

.photo-placeholder { margin-top: $cj-gap-md; padding: 28rpx 0; text-align: center; font-size: 24rpx; color: $cj-text-muted; }
.photo-grid { display: flex; flex-wrap: wrap; gap: 16rpx; margin-top: $cj-gap-md; }
.photo-cell { position: relative; width: calc((100% - 32rpx) / 3); box-sizing: border-box; }
.photo-img { display: block; width: 100%; height: 200rpx; border-radius: $cj-radius-md; background: $cj-surface-2; }
.photo-remove { position: absolute; top: 8rpx; right: 8rpx; width: 44rpx; height: 44rpx; border-radius: 50%; background: rgba(0,0,0,.45); display: flex; align-items: center; justify-content: center; }
.photo-remove-x { color: #fff; font-size: 32rpx; line-height: 1; }
.add-photo-btn { margin-top: $cj-gap-md; border-radius: $cj-radius-pill !important; background: $cj-surface !important; color: $cj-text !important; border: 1rpx solid $cj-border-light !important; }

.main-btn, .ghost-btn { border-radius: $cj-radius-pill !important; }
.main-btn { margin-top: $cj-gap-sm; background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important; color: #fffefb !important; border: none !important; }
.ghost-btn { background: $cj-surface !important; color: $cj-text !important; border: 1rpx solid $cj-border-light !important; }
</style>
