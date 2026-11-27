<template>
  <view v-if="visible" class="sheet-mask" @tap.self="onClose">
    <view class="sheet" :style="sheetLiftStyle" @tap.stop>
      <view class="sheet-header">
        <text class="sheet-title">{{ template.label }}</text>
        <view class="sheet-close" @tap.stop="onClose">
          <text class="sheet-close-x">×</text>
        </view>
      </view>

      <scroll-view class="sheet-body-scroll" scroll-y :show-scrollbar="false" :enable-flex="true">
        <view class="sheet-body">
          <view v-if="todayCountHint" class="count-hint">
            <text class="count-hint-text">{{ todayCountHint }}</text>
          </view>

          <view v-for="field in allFields" :key="field.key" class="field">
            <text class="field-label">
              {{ field.label }}<text v-if="field.unit" class="field-unit">（{{ field.unit }}）</text><text v-if="isOptional(field.key)" class="field-optional">选填</text>
            </text>

            <template v-if="field.type === 'duration'">
              <view class="duration-row">
                <view class="duration-box">
                  <input
                    :value="fieldValues[durationKey(field.key, 'h')] ?? ''"
                    class="duration-field"
                    type="number"
                    placeholder="0"
                    confirm-type="done"
                    @input="onFieldInput(durationKey(field.key, 'h'), $event)"
                  />
                  <text class="duration-unit">h</text>
                </view>
                <view class="duration-box">
                  <input
                    :value="fieldValues[durationKey(field.key, 'm')] ?? ''"
                    class="duration-field"
                    type="number"
                    placeholder="0"
                    confirm-type="done"
                    @input="onFieldInput(durationKey(field.key, 'm'), $event)"
                  />
                  <text class="duration-unit">m</text>
                </view>
              </view>
            </template>

            <template v-else-if="field.type === 'select'">
              <view class="select-row">
                <view
                  v-for="opt in field.options || []"
                  :key="opt.value"
                  :class="['select-chip', fieldValues[field.key] === opt.value && 'selected']"
                  @tap.stop="setSelect(field.key, opt.value)"
                >
                  <text class="select-chip-text">{{ opt.label }}</text>
                </view>
              </view>
            </template>

            <template v-else-if="field.type === 'date'">
              <picker mode="date" :value="fieldValues[field.key] || ''" @change="(e) => setSelect(field.key, e.detail.value)">
                <view :class="['field-input', !fieldValues[field.key] && 'is-placeholder']">{{ fieldValues[field.key] || "请选择日期" }}</view>
              </picker>
            </template>

            <template v-else>
              <view v-if="quickValues(field.key).length" class="quick-row">
                <view
                  v-for="v in quickValues(field.key)"
                  :key="v"
                  :class="['quick-chip', fieldValues[field.key] === String(v) && 'selected']"
                  @tap.stop="setSelect(field.key, String(v))"
                >
                  <text class="quick-chip-text">{{ v }}</text>
                </view>
              </view>
              <input
                :value="fieldValues[field.key] ?? ''"
                class="field-input"
                :type="field.type === 'number' ? 'digit' : 'text'"
                :placeholder="field.placeholder || ''"
                confirm-type="done"
                :adjust-position="true"
                :cursor-spacing="inputCursorSpacing"
                @input="onFieldInput(field.key, $event)"
              />
            </template>
          </view>

          <view class="field">
            <text class="field-label">图片<text class="field-optional">选填</text></text>
            <view class="quick-photos">
              <view v-for="(item, idx) in photos" :key="item.key" class="quick-photo">
                <image class="quick-photo-img" :src="item.path" mode="aspectFill" @tap.stop="previewPhoto(idx)" />
                <view class="quick-photo-remove" @tap.stop="removePhoto(item.key)">
                  <text class="quick-photo-remove-x">×</text>
                </view>
              </view>
              <view class="quick-photo-add" @tap.stop="pickPhoto('camera')">
                <text class="quick-photo-add-text">拍照</text>
              </view>
              <view class="quick-photo-add" @tap.stop="pickPhoto('album')">
                <text class="quick-photo-add-text">相册</text>
              </view>
            </view>
          </view>

          <view class="field">
            <text class="field-label">备注<text class="field-optional">选填</text></text>
            <input
              v-model="noteText"
              class="field-input"
              placeholder="补充一句话"
              confirm-type="done"
              :adjust-position="true"
              :cursor-spacing="inputCursorSpacing"
            />
          </view>
        </view>
      </scroll-view>

      <button class="save-btn" :loading="loading" hover-class="save-btn-hover" @tap.stop="onSave">保存</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import type { RecordTemplate } from "@/utils/recordTypes";
import { ensureImageUnderMaxBytes } from "@/utils/imageCompress";

const props = defineProps<{
  visible: boolean;
  template: RecordTemplate;
  babyId?: number;
  /** 今日同类型记录条数，用于「今日第 N 次」提示 */
  todayCount?: number;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved", payload: Record<string, unknown>, summary: string, photos: string[]): void;
}>();

const fieldValues = ref<Record<string, string>>({});
const noteText = ref("");
const photos = ref<{ key: string; path: string }[]>([]);
const loading = ref(false);

const inputCursorSpacing = 120;

const allFields = computed(() => [...props.template.recommendedFields, ...props.template.optionalFields]);

/** 喂养不再填「次数」，由当日已有条数自动推导 */
const todayCountHint = computed(() => {
  if (props.template.value !== "feeding") return "";
  const next = Number(props.todayCount || 0) + 1;
  return `今日第 ${next} 次喂养`;
});

const keyboardHeightPx = ref(0);
let offKeyboardHeight: (() => void) | undefined;

const KEYBOARD_EXTRA_LIFT_PX = 32;

const sheetLiftStyle = computed(() => {
  const h = keyboardHeightPx.value;
  if (h <= 0) return {};
  return {
    transform: `translateY(-${h + KEYBOARD_EXTRA_LIFT_PX}px)`,
    transition: "transform 0.2s ease-out",
  };
});

function isOptional(key: string): boolean {
  return props.template.optionalFields.some((f) => f.key === key);
}

function bindKeyboardHeightListener() {
  offKeyboardHeight?.();
  offKeyboardHeight = undefined;
  keyboardHeightPx.value = 0;
  if (typeof uni.onKeyboardHeightChange !== "function") return;
  offKeyboardHeight = uni.onKeyboardHeightChange((res) => {
    const n = typeof res.height === "number" ? res.height : 0;
    keyboardHeightPx.value = n;
  }) as unknown as (() => void) | undefined;
}

function seedFields() {
  const m: Record<string, string> = {};
  for (const f of allFields.value) {
    m[f.key] = "";
    if (f.type === "duration") {
      m[durationKey(f.key, "h")] = "";
      m[durationKey(f.key, "m")] = "";
    }
  }
  fieldValues.value = m;
}

/** 数值字段常用值，点选即填，避免键盘输入 */
const QUICK_VALUES: Record<string, number[]> = {
  amount_ml: [60, 90, 120, 150, 180],
  times: [1, 2, 3, 4, 5],
};

function quickValues(key: string): number[] {
  return QUICK_VALUES[key] || [];
}

watch(
  () => [props.visible, props.template.value, props.babyId] as const,
  ([vis]) => {
    if (vis) {
      noteText.value = "";
      photos.value = [];
      seedFields();
      bindKeyboardHeightListener();
    } else {
      keyboardHeightPx.value = 0;
      offKeyboardHeight?.();
      offKeyboardHeight = undefined;
    }
  },
  { immediate: true }
);

onUnmounted(() => {
  offKeyboardHeight?.();
  offKeyboardHeight = undefined;
});

function setSelect(key: string, value: string) {
  fieldValues.value = { ...fieldValues.value, [key]: value };
}

function pickInputValue(e: unknown): string {
  if (e == null || typeof e !== "object") return "";
  const ex = e as Record<string, unknown>;
  const detail = ex.detail as Record<string, unknown> | undefined;
  if (detail && "value" in detail && detail.value != null) return String(detail.value);
  const mp = ex.mp as Record<string, unknown> | undefined;
  const mpDetail = mp?.detail as Record<string, unknown> | undefined;
  if (mpDetail && "value" in mpDetail && mpDetail.value != null) return String(mpDetail.value);
  return "";
}

function onFieldInput(key: string, e: unknown) {
  const v = pickInputValue(e);
  fieldValues.value = { ...fieldValues.value, [key]: v };
}

function onClose() {
  emit("close");
}

function durationKey(key: string, part: "h" | "m"): string {
  return `${key}__${part}`;
}

function buildPayload(): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  const fv = fieldValues.value;
  for (const field of allFields.value) {
    if (field.type === "duration") {
      const h = Number(fv[durationKey(field.key, "h")] || 0);
      const m = Number(fv[durationKey(field.key, "m")] || 0);
      const total = Math.max(0, Math.round((Number.isFinite(h) ? h : 0) * 60 + (Number.isFinite(m) ? m : 0)));
      if (total > 0) result[field.key] = total;
      continue;
    }
    const raw = fv[field.key]?.trim();
    if (!raw) continue;
    if (field.type === "number") {
      const n = Number(raw);
      result[field.key] = Number.isFinite(n) ? n : raw;
    } else {
      result[field.key] = raw;
    }
  }
  return result;
}

function buildSummary(payload: Record<string, unknown>): string {
  if (props.template.autoSummary) {
    return props.template.autoSummary(payload);
  }
  return props.template.label + "已记录";
}

function pickPhoto(source: "camera" | "album") {
  const remain = 9 - photos.value.length;
  if (remain <= 0) return;
  uni.chooseImage({
    count: source === "camera" ? 1 : Math.min(4, remain),
    sourceType: source === "camera" ? ["camera"] : ["album"],
    sizeType: ["original"],
    success: async (res) => {
      for (const path of res.tempFilePaths) {
        try {
          const finalPath = await ensureImageUnderMaxBytes(path);
          photos.value.push({ key: `p-${Date.now()}-${Math.random().toString(36).slice(2)}`, path: finalPath });
        } catch (e) {
          console.error(e);
        }
      }
    },
  });
}

function removePhoto(key: string) {
  photos.value = photos.value.filter((p) => p.key !== key);
}

function previewPhoto(idx: number) {
  const urls = photos.value.map((p) => p.path);
  uni.previewImage({ current: urls[idx], urls });
}

function onSave() {
  if (loading.value) return;
  loading.value = true;
  try {
    const payload = buildPayload();
    const note = noteText.value.trim();
    // 备注单独存 payload.note，摘要始终由字段推导，保证两者不互相污染
    if (note) payload.note = note;
    emit("saved", payload, buildSummary(payload), photos.value.map((p) => p.path));
  } finally {
    loading.value = false;
  }
}

</script>

<style lang="scss" scoped>
.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 999;
  display: flex;
  align-items: flex-end;
}

.sheet {
  width: 100%;
  background: $cj-surface;
  border-radius: $cj-radius-xl $cj-radius-xl 0 0;
  padding: $cj-gap-lg $cj-gap-md calc(env(safe-area-inset-bottom, 0px) + 32rpx);
  box-sizing: border-box;
}

.sheet-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $cj-gap-md;
}

.sheet-title {
  font-size: 33rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 0.2rpx;
}

.sheet-close {
  width: 52rpx;
  height: 52rpx;
  border-radius: 50%;
  background: $cj-surface-2;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sheet-close-x {
  font-size: 34rpx;
  color: $cj-text-muted;
  line-height: 1;
}

.sheet-body-scroll {
  max-height: 62vh;
  width: 100%;
  box-sizing: border-box;
}

.sheet-body {
  margin-bottom: $cj-gap-lg;
}

.count-hint {
  margin-bottom: $cj-gap-md;
  padding: 16rpx 22rpx;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-md;
}

.count-hint-text {
  font-size: 24rpx;
  color: $cj-primary-dark;
  font-weight: 500;
}

.field {
  margin-bottom: $cj-gap-md;
}

.field-label {
  display: block;
  margin-bottom: 10rpx;
  font-size: 23rpx;
  color: $cj-text-muted;
  font-weight: 500;
  letter-spacing: 0.3rpx;
}

.field-unit {
  font-size: 21rpx;
}

.field-optional {
  margin-left: 8rpx;
  font-size: 19rpx;
  color: $cj-text-muted;
}

.field-input {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  font-size: 32rpx;
  color: $cj-text;
  min-height: 96rpx;
  display: flex;
  align-items: center;
}

.field-input.is-placeholder {
  color: $cj-text-muted;
}

.duration-row { display: flex; gap: 16rpx; }
.duration-box { flex: 1; min-width: 0; display: flex; align-items: center; gap: 8rpx; background: $cj-surface-2; border: 1rpx solid $cj-border-faint; border-radius: $cj-radius-md; padding: 0 20rpx; height: 108rpx; box-sizing: border-box; }
.duration-field { flex: 1; min-width: 0; font-size: 32rpx; text-align: center; background: transparent; border: none; }
.duration-unit { flex-shrink: 0; font-size: 26rpx; color: $cj-text-secondary; }

.select-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.select-chip {
  padding: 12rpx 24rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  transition: all 0.15s;
}

.select-chip.selected {
  background: $cj-primary-soft;
  border-color: $cj-primary;
  box-shadow: $cj-shadow-xs;
}

.select-chip-text {
  font-size: 25rpx;
  color: $cj-text;
}

.quick-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 12rpx;
}

.quick-chip {
  padding: 10rpx 24rpx;
  font-size: 22rpx;
  color: $cj-text;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-pill;
}

.quick-chip.selected {
  color: $cj-primary-dark;
  border-color: $cj-primary;
  background: rgba(201, 107, 92, 0.1);
  font-weight: 500;
}

.quick-chip-text {
  font-size: 22rpx;
}

.quick-photos {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.quick-photo {
  position: relative;
  width: 132rpx;
  height: 132rpx;
}

.quick-photo-img {
  width: 132rpx;
  height: 132rpx;
  border-radius: 14rpx;
  background: $cj-surface-2;
}

.quick-photo-remove {
  position: absolute;
  top: 4rpx;
  right: 4rpx;
  width: 34rpx;
  height: 34rpx;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-photo-remove-x {
  color: #fff;
  font-size: 24rpx;
  line-height: 1;
}

.quick-photo-add {
  width: 132rpx;
  height: 132rpx;
  border-radius: 14rpx;
  border: 1rpx dashed $cj-border-light;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $cj-surface;
}

.quick-photo-add-text {
  font-size: 22rpx;
  color: $cj-text-secondary;
}

.save-btn {
  width: 100%;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  font-size: 30rpx;
  box-shadow: $cj-shadow-soft;
}

.save-btn-hover {
  opacity: 0.92;
}
</style>
