<template>
  <view v-if="record" class="page">
    <view class="head">
      <text class="head-title">{{ detailTitle }}</text>
      <text v-if="headDesc" class="head-desc">{{ headDesc }}</text>
    </view>

    <view class="section">
      <text class="section-label">类型</text>
      <view class="readonly-line">
        <view class="picker-input readonly">{{ labelForMotherType(record.record_type) }}</view>
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
      <view v-if="!attachments.length" class="photo-placeholder">还没有图片，点下方按钮添加。</view>
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
      <button class="danger-btn action-btn" @click="removeRecord">删除记录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import {
  apiDeleteAttachment,
  apiDeleteMotherRecord,
  apiGetMotherRecord,
  apiListMotherAttachments,
  type AttachmentItem,
  type MotherRecordItem,
} from "@/api/chuyaji";
import { uploadMotherRecordAttachment } from "@/api/upload";
import { ensureImageUnderMaxBytes } from "@/utils/imageCompress";
import { resolvePublicMediaUrl } from "@/utils/mediaUrl";
import { formatFieldValueForDisplay, type TemplateField } from "@/utils/recordTypes";
import { labelForMotherType, MOTHER_TEMPLATES, type RecordTemplate } from "@/utils/motherRecordTypes";

const recordId = ref(0);
const record = ref<MotherRecordItem | null>(null);
const attachments = ref<AttachmentItem[]>([]);
const showOptional = ref(false);

const currentTemplate = computed<RecordTemplate | undefined>(() => {
  const r = record.value;
  if (!r) return undefined;
  return MOTHER_TEMPLATES.find((t) => t.value === r.record_type);
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

onShow(() => {
  if (recordId.value) {
    void refresh();
  }
});

async function refresh() {
  try {
    record.value = await apiGetMotherRecord(recordId.value);
    const result = await apiListMotherAttachments(recordId.value);
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
    sizeType: ["original"],
    success: async (result) => {
      for (const file of result.tempFilePaths) {
        try {
          const path = await ensureImageUnderMaxBytes(file);
          await uploadMotherRecordAttachment(recordId.value, path);
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
  uni.navigateTo({ url: `/pages/mom/record-edit?id=${recordId.value}&mother_id=${record.value.mother_id}` });
}

function removeRecord() {
  uni.showModal({
    title: "确认删除？",
    success: async (result) => {
      if (!result.confirm) return;
      try {
        await apiDeleteMotherRecord(recordId.value);
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

</style>
