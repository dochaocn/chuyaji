<template>
  <view v-if="record" class="page">
    <view class="hero">
      <view class="hero-top">
        <text :class="['tag', record.phase === 'prenatal' ? 'pre' : 'post']">
          {{ record.phase === "prenatal" ? "怀孕期" : "成长期" }}
        </text>
        <text class="type-name">{{ labelForType(record.phase, record.record_type) }}</text>
      </view>
      <text class="date-line">{{ record.occurred_at.slice(0, 10) }}</text>
      <text class="summary">{{ record.summary || "未填写摘要" }}</text>
    </view>

    <view class="section">
      <view class="section-head">
        <text class="section-title">附件</text>
        <text class="section-hint">点击图片可预览</text>
      </view>
      <view v-if="!attachments.length" class="placeholder">还没有图片，点下方按钮补充。</view>
      <view v-else class="gallery">
        <view v-for="(item, index) in attachments" :key="item.id" class="gallery-item">
          <image
            class="img"
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
      <button class="main-btn" @click="pickImage">添加图片</button>
    </view>

    <view class="actions">
      <button class="ghost-btn" @click="goEdit">编辑本条</button>
      <button class="danger-btn" @click="removeRecord">删除记录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
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
import { labelForType } from "@/utils/recordTypes";

const recordId = ref(0);
const record = ref<RecordItem | null>(null);
const attachments = ref<AttachmentItem[]>([]);

onLoad((query: Record<string, string | undefined>) => {
  recordId.value = Number(query.id || 0);
  if (recordId.value) {
    refresh();
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

.hero,
.section {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.hero {
  background: linear-gradient(155deg, $cj-surface 0%, $cj-accent-soft 100%);
}

.hero-top,
.section-head,
.actions {
  display: flex;
}

.hero-top,
.section-head {
  align-items: center;
  justify-content: space-between;
  gap: $cj-gap-sm;
}

.tag {
  padding: 6rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  font-weight: $cj-fw-title;
}

.tag.pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.tag.post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.type-name,
.section-title {
  font-size: 30rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
}

.date-line,
.section-hint {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-muted;
  font-size: 22rpx;
}

.summary {
  display: block;
  margin-top: $cj-gap-md;
  color: $cj-text-secondary;
  font-size: 28rpx;
  line-height: 1.7;
}

.placeholder {
  padding: 30rpx 0;
  text-align: center;
  color: $cj-text-muted;
  font-size: 24rpx;
}

.gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
  margin-top: $cj-gap-md;
}

.gallery-item {
  position: relative;
  width: calc((100% - 32rpx) / 3);
  box-sizing: border-box;
}

.img {
  display: block;
  width: 100%;
  height: 200rpx;
  border-radius: $cj-radius-md;
  background: $cj-surface-2;
  overflow: hidden;
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

.main-btn,
.ghost-btn,
.danger-btn {
  border-radius: $cj-radius-pill !important;
}

.main-btn {
  margin-top: $cj-gap-md;
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}

.ghost-btn {
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-light !important;
}

.actions {
  flex-direction: column;
  gap: $cj-gap-sm;
}

.danger-btn {
  background: $cj-danger-bg !important;
  color: $cj-danger-text !important;
  border: 1rpx solid rgba(143, 61, 54, 0.2) !important;
}
</style>
