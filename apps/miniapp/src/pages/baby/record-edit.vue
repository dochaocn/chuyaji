<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ pageTitle }}</text>
      <text v-if="currentTemplate?.mode === 'standard' && currentTemplate?.summaryPlaceholder" class="head-desc">
        {{ currentTemplate.summaryPlaceholder }}
      </text>
    </view>

    <!-- 阶段与类型：两行间距与「关键信息」中 .field 行距一致 -->
    <view class="section">
      <text class="section-label">阶段与类型</text>
      <view class="phase-type-stack">
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
    </view>

    <!-- 日期 -->
    <view class="section">
      <text class="section-label">日期</text>
      <view class="field">
        <text class="lab">发生日期</text>
        <picker mode="date" :value="occurredAt" @change="onOccurredAtChange">
          <view :class="['picker-input', !occurredAt && 'is-placeholder']">{{ occurredAt || "请选择日期" }}</view>
        </picker>
      </view>
    </view>

    <!-- 推荐字段 -->
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

    <!-- 摘要（standard 模式） -->
    <view v-if="currentTemplate?.mode === 'standard'" class="section">
      <text class="section-label">摘要</text>
      <view class="field">
        <textarea
          v-model="summary"
          class="area"
          :placeholder="currentTemplate?.summaryPlaceholder || '一句话描述今天发生了什么'"
        />
      </view>
    </view>

    <!-- 可选字段（折叠） -->
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

    <!-- 图片 -->
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

    <view v-if="!recordId" class="save-draft-row">
      <button class="main-btn row-btn" :loading="loading" @click="save">保存记录</button>
      <button class="ghost-btn row-btn" @click="tapLoadDraft">载入草稿</button>
    </view>
    <button v-else class="main-btn" :loading="loading" @click="save">保存记录</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, reactive, ref, watch } from "vue";
import {
  apiCreateRecord, apiDeleteAttachment, apiGetRecord, apiListAttachments, apiPatchRecord,
  type AttachmentItem,
} from "@/api/chuyaji";
import { uploadRecordAttachment } from "@/api/upload";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import { BABY_TEMPLATES, labelForType, type RecordTemplate, type TemplateField } from "@/utils/recordTypes";

const babyId = ref(0);
const recordId = ref(0);
const phase = ref<"prenatal" | "postnatal">("prenatal");
const recordType = ref("prenatal_checkup");
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

const phaseLabels = ["怀孕期", "成长期"];
const phaseIndex = computed(() => (phase.value === "prenatal" ? 0 : 1));
const typeLabels = computed(() => BABY_TEMPLATES[phase.value].map((t) => t.label));
const typeIndex = computed(() => {
  const idx = BABY_TEMPLATES[phase.value].findIndex((t) => t.value === recordType.value);
  return idx >= 0 ? idx : 0;
});
const currentTypeLabel = computed(() => labelForType(phase.value, recordType.value));

const currentTemplate = computed<RecordTemplate | undefined>(() =>
  BABY_TEMPLATES[phase.value].find((t) => t.value === recordType.value)
);

const pageTitle = computed(() => {
  if (recordId.value) return `编辑${currentTemplate.value?.label || "记录"}`;
  return currentTemplate.value?.label ? `新增${currentTemplate.value.label}` : "新增宝宝记录";
});

const recommendedFields = computed(() => currentTemplate.value?.recommendedFields ?? []);
const optionalFields = computed(() => currentTemplate.value?.optionalFields ?? []);

function selectIndex(field: TemplateField): number {
  const opts = field.options || [];
  const idx = opts.findIndex((o) => o.value === extras[field.key]);
  return idx >= 0 ? idx : 0;
}

function selectLabel(field: TemplateField): string {
  const opts = field.options || [];
  const hit = opts.find((o) => o.value === extras[field.key]);
  return hit ? hit.label : "";
}

function onSelectChange(field: TemplateField, event: { detail: { value: string } }) {
  const opts = field.options || [];
  const opt = opts[Number(event.detail.value)];
  if (opt) extras[field.key] = opt.value;
}

watch(
  () => phase.value,
  () => {
    recordType.value = BABY_TEMPLATES[phase.value][0].value;
    for (const key of Object.keys(extras)) delete extras[key];
    showOptional.value = false;
  }
);

watch(
  () => recordType.value,
  () => {
    for (const key of Object.keys(extras)) delete extras[key];
    showOptional.value = false;
  }
);

onLoad((query: Record<string, string | undefined>) => {
  babyId.value = Number(query.baby_id || 0);
  recordId.value = Number(query.id || 0);
  if (query.phase === "prenatal" || query.phase === "postnatal") phase.value = query.phase;
  if (query.type) recordType.value = query.type;
  const now = new Date();
  occurredAt.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
  if (recordId.value) loadExisting();
  else loadDraft({ silent: true });
});

function onPhase(event: { detail: { value: string } }) {
  phase.value = Number(event.detail.value) === 0 ? "prenatal" : "postnatal";
}

function onType(event: { detail: { value: string } }) {
  const next = BABY_TEMPLATES[phase.value][Number(event.detail.value)]?.value;
  if (next) recordType.value = next;
}

function onOccurredAtChange(event: { detail: { value: string } }) {
  occurredAt.value = event.detail.value || "";
}

function draftKey() {
  return `draft_record_${babyId.value}_${phase.value}_${recordType.value}`;
}

type DraftCache = { occurredAt?: string; summary?: string; pendingPaths?: { key: string; path: string }[] };

function persistDraft() {
  if (recordId.value) return;
  try {
    uni.setStorageSync(draftKey(), JSON.stringify({ occurredAt: occurredAt.value, summary: summary.value, pendingPaths: pendingPaths.value }));
  } catch { /* ignore */ }
}

async function loadExisting() {
  try {
    const item = await apiGetRecord(recordId.value);
    phase.value = item.phase;
    recordType.value = item.record_type;
    occurredAt.value = item.occurred_at.slice(0, 10);
    summary.value = item.summary || "";
    const pl = item.payload || {};
    serverPayload.value = { ...pl };
    for (const key of Object.keys(extras)) delete extras[key];
    // 全量合并 payload 进 extras，保证所有已存字段都能回显
    for (const [key, val] of Object.entries(pl)) {
      if (val != null && val !== "") extras[key] = String(val);
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
    if (!raw) { if (!silent) uni.showToast({ title: "暂无草稿", icon: "none" }); return; }
    const cached = JSON.parse(raw) as DraftCache;
    occurredAt.value = cached.occurredAt || occurredAt.value;
    summary.value = cached.summary || "";
    pendingPaths.value = Array.isArray(cached.pendingPaths) ? [...cached.pendingPaths] : [];
    if (!silent) uni.showToast({ title: "已载入草稿", icon: "success" });
  } catch { if (!silent) uni.showToast({ title: "载入失败", icon: "none" }); }
}

function tapLoadDraft() { loadDraft({ silent: false }); }

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
    try { await uploadRecordAttachment(id, item.path); pendingPaths.value = pendingPaths.value.filter((p) => p.key !== item.key); }
    catch (e) { console.error(e); failed++; }
  }
  if (failed > 0) uni.showToast({ title: `${failed} 张图片上传失败`, icon: "none" });
}

async function save() {
  if (!babyId.value) { uni.showToast({ title: "缺少宝宝档案", icon: "none" }); return; }
  if (!occurredAt.value.trim()) { uni.showToast({ title: "请选择发生日期", icon: "none" }); return; }
  loading.value = true;
  try {
    const freshPayload = buildPayload();
    // 编辑时与服务端原有 payload 合并，避免 PATCH 整段替换丢掉未出现在当前模板的键
    const payload = recordId.value ? { ...serverPayload.value, ...freshPayload } : freshPayload;
    const finalSummary = buildSummary(payload);
    const body = { phase: phase.value, record_type: recordType.value, occurred_at: new Date(`${occurredAt.value}T12:00:00`).toISOString(), summary: finalSummary, payload };
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
  } finally { loading.value = false; }
}

watch([occurredAt, summary, pendingPaths], () => { persistDraft(); }, { deep: true });
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
.field:last-child { margin-bottom: 0; }

/* 与 .field 行间距同为 $cj-gap-md（见上，勿用 .picker-line:last-child 处理多行 picker） */
.phase-type-stack > picker:not(:last-child) {
  display: block;
  margin-bottom: $cj-gap-md;
}

.phase-type-stack .picker-line {
  margin-bottom: 0;
}

.picker-k, .lab {
  display: block;
  margin-bottom: 10rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
}

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
.save-draft-row {
  display: flex;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-sm;
  align-items: stretch;
}
.save-draft-row .row-btn {
  flex: 1;
  margin-top: 0 !important;
}
.ghost-btn { background: $cj-surface !important; color: $cj-text !important; border: 1rpx solid $cj-border-light !important; }
</style>
