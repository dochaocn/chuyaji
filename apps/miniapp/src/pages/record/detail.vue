<template>
  <view class="page" v-if="rec">
    <view class="hero">
      <view class="hero-top">
        <text :class="['tag', rec.phase === 'prenatal' ? 'tag--pre' : 'tag--post']">
          {{ rec.phase === "prenatal" ? "孕期" : "产后" }}
        </text>
        <text class="type-name">{{ typeLabel(rec) }}</text>
      </view>
      <text class="date-line">{{ formatDate(rec.occurred_at) }}</text>
      <text class="sum">{{ rec.summary || "（暂无摘要，可去编辑补充一句）" }}</text>
      <view class="hero-blob" aria-hidden="true" />
    </view>

    <view class="card att-card">
      <view class="card-head">
        <text class="card-title">照片与附件</text>
        <text class="card-hint">轻点图片可预览</text>
      </view>
      <view v-if="!atts.length" class="empty-att">还没有图片，点下方按钮添加</view>
      <view v-else class="att-grid">
        <view v-for="a in atts" :key="a.id" class="att-cell">
          <image v-if="a.url" class="img" :src="a.url" mode="aspectFill" @click="preview(a.url)" />
          <button size="mini" class="del" @click.stop="delAtt(a.id)">移除</button>
        </view>
      </view>
      <button type="primary" class="add-btn" @click="pickImage">添加图片</button>
    </view>

    <view class="actions">
      <button class="btn-edit" @click="goEdit">编辑本条</button>
      <button class="danger" @click="remove">删除记录</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiDeleteAttachment, apiDeleteRecord, apiGetRecord, apiListAttachments } from "@/api/chuyaji";
import type { RecordItem } from "@/api/chuyaji";
import { uploadRecordAttachment } from "@/api/upload";
import { labelForType } from "@/utils/recordTypes";

const id = ref(0);
const rec = ref<RecordItem | null>(null);
const atts = ref<{ id: number; url: string }[]>([]);

onLoad((q: Record<string, string | undefined>) => {
  id.value = Number(q.id || 0);
  if (id.value) refresh();
});

function typeLabel(r: RecordItem) {
  return labelForType(r.phase, r.record_type);
}

function formatDate(iso: string) {
  if (!iso) return "";
  const d = iso.slice(0, 10);
  const t = iso.slice(11, 16);
  return t ? `${d} ${t}` : d;
}

async function refresh() {
  if (!id.value) return;
  try {
    rec.value = await apiGetRecord(id.value);
    const a = await apiListAttachments(id.value);
    atts.value = a.items || [];
  } catch (e) {
    console.error(e);
  }
}

function preview(url: string) {
  uni.previewImage({ urls: [url] });
}

async function delAtt(aid: number) {
  try {
    await apiDeleteAttachment(aid);
    await refresh();
  } catch (e) {
    console.error(e);
  }
}

function pickImage() {
  if (!id.value) return;
  uni.chooseImage({
    count: 3,
    sizeType: ["compressed"],
    success: async (r) => {
      for (const p of r.tempFilePaths) {
        try {
          await uploadRecordAttachment(id.value, p);
        } catch (e) {
          console.error(e);
        }
      }
      await refresh();
      uni.showToast({ title: "已上传", icon: "none" });
    },
  });
}

function goEdit() {
  uni.navigateTo({ url: `/pages/record/edit?id=${id.value}&baby_id=${rec.value?.baby_id || 0}` });
}

async function remove() {
  uni.showModal({
    title: "确认删除？",
    success: async (r) => {
      if (!r.confirm || !id.value) return;
      try {
        await apiDeleteRecord(id.value);
        uni.navigateBack();
      } catch (e) {
        console.error(e);
      }
    },
  });
}
</script>

<style lang="scss" scoped>
@keyframes cj-fade-up {
  from {
    opacity: 0;
    transform: translateY(14rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page {
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
}

.hero {
  position: relative;
  overflow: hidden;
  background: linear-gradient(155deg, $cj-surface 0%, $cj-accent-soft 100%);
  border-radius: $cj-radius-lg;
  padding: $cj-gap-lg $cj-gap-md 40rpx;
  margin-bottom: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-lift;
  animation: cj-fade-up 0.55s cubic-bezier(0.22, 1, 0.36, 1) backwards;
}

.hero-blob {
  position: absolute;
  right: -40rpx;
  bottom: -50rpx;
  width: 200rpx;
  height: 200rpx;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(201, 107, 92, 0.12) 0%, transparent 68%);
  pointer-events: none;
}

.hero-top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-sm;
}

.tag {
  font-size: 22rpx;
  font-weight: $cj-fw-title;
  padding: 8rpx 20rpx;
  border-radius: $cj-radius-pill;
  letter-spacing: 1rpx;
}

.tag--pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.tag--post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.type-name {
  font-size: 34rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.date-line {
  display: block;
  font-size: 24rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-md;
}

.sum {
  display: block;
  font-size: 30rpx;
  color: $cj-text-secondary;
  line-height: 1.75;
  position: relative;
  z-index: 1;
}

.card {
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
  box-shadow: $cj-shadow-card;
  border: 1rpx solid $cj-border-light;
  animation: cj-fade-up 0.5s cubic-bezier(0.22, 1, 0.36, 1) 0.08s backwards;
}

.card-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: $cj-gap-md;
}

.card-title {
  font-size: 28rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.card-hint {
  font-size: 22rpx;
  color: $cj-text-muted;
}

.empty-att {
  font-size: 26rpx;
  color: $cj-text-muted;
  text-align: center;
  padding: $cj-gap-lg 0;
}

.att-grid {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.att-cell {
  width: calc(50% - 12rpx);
}

.img {
  width: 100%;
  height: 220rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid $cj-border-light;
  display: block;
}

.del {
  margin-top: 8rpx;
  width: 100% !important;
  font-size: 22rpx !important;
  background: $cj-page-bg !important;
  color: $cj-text-secondary !important;
  border: 1rpx solid $cj-border-light !important;
  border-radius: $cj-radius-pill !important;
}

.add-btn {
  border-radius: $cj-radius-pill !important;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: $cj-gap-sm;
  animation: cj-fade-up 0.48s cubic-bezier(0.22, 1, 0.36, 1) 0.14s backwards;
}

.btn-edit {
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border-radius: $cj-radius-pill !important;
  font-weight: 600;
  letter-spacing: 2rpx;
  box-shadow: $cj-shadow-soft;
  border: none !important;
}

.danger {
  background: $cj-danger-bg !important;
  color: $cj-danger-text !important;
  border-radius: $cj-radius-pill !important;
  border: 1rpx solid rgba(143, 61, 54, 0.2) !important;
}
</style>
