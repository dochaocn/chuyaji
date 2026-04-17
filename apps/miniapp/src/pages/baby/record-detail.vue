<template>
  <view v-if="record" class="page">
    <view class="head">
      <text class="head-title">{{ detailTitle }}</text>
      <text v-if="headDesc" class="head-desc">{{ headDesc }}</text>
    </view>

    <!-- 与编辑页：阶段与类型 -->
    <view class="section">
      <text class="section-label">阶段与类型</text>
      <view class="phase-type-stack">
        <view class="readonly-line">
          <text class="picker-k">阶段</text>
          <view class="picker-input readonly">{{ record.phase === "prenatal" ? "怀孕期" : "成长期" }}</view>
        </view>
        <view class="readonly-line">
          <text class="picker-k">类型</text>
          <view class="picker-input readonly">{{ labelForType(record.phase, record.record_type) }}</view>
        </view>
      </view>
    </view>

    <view class="section">
      <text class="section-label">日期</text>
      <view class="field">
        <text class="lab">发生日期</text>
        <view class="picker-input readonly">{{ record.occurred_at.slice(0, 10) }}</view>
      </view>
    </view>

    <view v-if="recommendedFields.length" class="section">
      <text class="section-label">关键信息</text>
      <view v-for="field in recommendedFields" :key="field.key" class="field">
        <text class="lab">{{ field.label }}<text v-if="field.unit" class="unit-text">（{{ field.unit }}）</text></text>
        <view class="picker-input readonly">{{ displayFor(field) }}</view>
      </view>
    </view>

    <view v-if="currentTemplate?.mode === 'standard'" class="section">
      <text class="section-label">摘要</text>
      <view class="field">
        <view class="area readonly">{{ record.summary || "未填写" }}</view>
      </view>
    </view>

    <view v-if="optionalFields.length" class="section">
      <view class="optional-toggle" @click="showOptional = !showOptional">
        <text class="optional-toggle-text">{{ showOptional ? "收起更多信息" : "展开更多信息" }}</text>
      </view>
      <template v-if="showOptional">
        <view v-for="field in optionalFields" :key="field.key" class="field field--optional">
          <text class="lab">{{ field.label }}<text v-if="field.unit" class="unit-text">（{{ field.unit }}）</text></text>
          <view class="picker-input readonly">{{ displayFor(field) }}</view>
        </view>
      </template>
    </view>

    <view class="section">
      <text class="section-label">图片补充</text>
      <text v-if="currentTemplate?.attachmentHint" class="section-hint">{{ currentTemplate.attachmentHint }}</text>
      <text v-else class="section-hint">可选，点击图片可全屏预览</text>
      <view v-if="!attachments.length" class="photo-placeholder">还没有图片，点下方按钮补充。</view>
      <view v-else class="photo-grid">
        <view v-for="(item, index) in attachments" :key="item.id" class="photo-cell">
          <image
            class="photo-img"
            :src="attachmentThumbSrc(item)"
            mode="aspectFill"
            :lazy-load="false"
            @click="preview(index)"
          />
          <view class="photo-remove" @click.stop="removeAttachment(item.id)">
            <text class="photo-remove-x">×</text>
          </view>
        </view>
      </view>
      <button class="add-photo-btn" @click="pickImage">添加图片</button>
    </view>

    <view class="actions-row">
      <button class="ghost-btn action-btn" @click="goEdit">编辑本条</button>
      <button class="ghost-btn action-btn" @click="goNewWithDraft">载入草稿</button>
    </view>
    <button class="danger-btn danger-full" @click="removeRecord">删除记录</button>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import {
  apiDeleteAttachment,
  apiDeleteRecord,
  apiGetRecord,
  apiListAttachments,
  type AttachmentItem,
  type RecordItem,
} from "@/api/chuyaji";
import { uploadRecordAttachment } from "@/api/upload";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import {
  formatFieldValueForDisplay,
  labelForType,
  templateForType,
  type RecordTemplate,
  type TemplateField,
} from "@/utils/recordTypes";

const recordId = ref(0);
const record = ref<RecordItem | null>(null);
const attachments = ref<AttachmentItem[]>([]);
const showOptional = ref(false);

const currentTemplate = computed<RecordTemplate | undefined>(() => {
  const r = record.value;
  if (!r) return undefined;
  return templateForType(r.phase, r.record_type);
});

const payload = computed(() => (record.value?.payload || {}) as Record<string, unknown>);

const recommendedFields = computed(() => currentTemplate.value?.recommendedFields ?? []);
const optionalFields = computed(() => currentTemplate.value?.optionalFields ?? []);

const detailTitle = computed(() => {
  const label = currentTemplate.value?.label;
  return label ? `${label} · 详情` : "记录详情";
});

const headDesc = computed(() => {
  const t = currentTemplate.value;
  const r = record.value;
  if (!t || !r) return "";
  if (t.mode === "standard" && t.summaryPlaceholder) return t.summaryPlaceholder;
  if (t.mode === "quick" && r.summary) return r.summary;
  return "";
});

function displayFor(field: TemplateField): string {
  return formatFieldValueForDisplay(field, payload.value) ?? "—";
}

onLoad((query: Record<string, string | undefined>) => {
  recordId.value = Number(query.id || 0);
});

/** 从编辑页返回时页面不重建，onLoad 不会再次执行，需在每次展示时拉取最新数据 */
onShow(() => {
  if (recordId.value) {
    void refresh();
  }
});

async function refresh() {
  try {
    record.value = await apiGetRecord(recordId.value);
    const result = await apiListAttachments(recordId.value);
    attachments.value = result.items || [];
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

function attachmentThumbSrc(item: AttachmentItem) {
  const raw = (item.thumb_url && item.thumb_url.trim()) || item.url;
  return resolvePublicMediaUrl(raw);
}

function preview(index: number) {
  const urls = attachments.value.map((a) => resolvePublicMediaUrl(a.url));
  uni.previewImage({ current: urls[index], urls });
}

async function removeAttachment(id: number) {
  try {
    await apiDeleteAttachment(id);
    await refresh();
  } catch (error) {
    console.error(error);
  }
}

function pickImage() {
  uni.chooseImage({
    count: 9,
    sizeType: ["compressed"],
    success: async (result) => {
      for (const file of result.tempFilePaths) {
        try {
          await uploadRecordAttachment(recordId.value, file);
        } catch (error) {
          console.error(error);
        }
      }
      await refresh();
    },
  });
}

function goEdit() {
  if (!record.value) return;
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${recordId.value}&baby_id=${record.value.baby_id}` });
}

/** 新建同阶段/同类型记录页并载入本地草稿（与编辑页草稿键一致） */
function goNewWithDraft() {
  const r = record.value;
  if (!r) return;
  uni.navigateTo({
    url: `/pages/baby/record-edit?baby_id=${r.baby_id}&phase=${r.phase}&type=${encodeURIComponent(r.record_type)}`,
  });
}

function removeRecord() {
  uni.showModal({
    title: "确认删除？",
    success: async (result) => {
      if (!result.confirm) return;
      try {
        await apiDeleteRecord(recordId.value);
        uni.navigateBack();
      } catch (error) {
        console.error(error);
      }
    },
  });
}
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
}

.optional-toggle {
  padding: 4rpx 0;
}

.optional-toggle-text {
  font-size: 24rpx;
  color: $cj-primary;
}

.phase-type-stack > .readonly-line:not(:last-child) {
  display: block;
  margin-bottom: $cj-gap-md;
}

.field {
  margin-bottom: $cj-gap-md;
}

.field:last-child {
  margin-bottom: 0;
}

.field--optional {
  margin-top: 24rpx;
}

.picker-k,
.lab {
  display: block;
  margin-bottom: 10rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
}

.unit-text {
  font-size: 22rpx;
}

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

.picker-input {
  min-height: 108rpx;
  line-height: 1.55;
  padding-top: 26rpx;
  padding-bottom: 26rpx;
  display: flex;
  align-items: center;
}

.picker-input.readonly {
  color: $cj-text-secondary;
}

.area.readonly {
  min-height: 120rpx;
  line-height: 1.55;
  padding-top: 28rpx;
  padding-bottom: 28rpx;
  white-space: pre-wrap;
  word-break: break-word;
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

.actions-row {
  display: flex;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.action-btn {
  flex: 1;
  margin: 0 !important;
  border-radius: $cj-radius-pill !important;
}

.ghost-btn {
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
}

.danger-btn {
  background: $cj-danger-bg !important;
  color: $cj-danger-text !important;
  border: 1rpx solid rgba(143, 61, 54, 0.2) !important;
}

.danger-full {
  width: 100%;
  margin-top: $cj-gap-sm !important;
  border-radius: $cj-radius-pill !important;
}
</style>
