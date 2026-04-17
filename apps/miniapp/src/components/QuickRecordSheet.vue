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
        <view v-for="field in template.recommendedFields" :key="field.key" class="field">
          <text class="field-label">{{ field.label }}<text v-if="field.unit" class="field-unit">（{{ field.unit }}）</text></text>
          <template v-if="field.type === 'select'">
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
          <template v-else>
            <!--
              安卓微信小程序：对 reactive 动态 key 的 v-model 常不生效。
              使用 ref + 整对象替换；@input/@blur 统一从 detail 取 value（兼容 mp 包装）。
            -->
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

        <view class="note-toggle" @tap.stop="showNote = !showNote">
          <text class="note-toggle-text">{{ showNote ? "收起备注" : "添加备注" }}</text>
        </view>
        <view v-if="showNote" class="field">
          <textarea
            v-model="noteText"
            class="note-area"
            placeholder="可选，补充一句话"
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

const props = defineProps<{
  visible: boolean;
  template: RecordTemplate;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved", payload: Record<string, unknown>, summary: string): void;
}>();

/** 用 ref + 替换对象，避免安卓端 reactive[key] 不触发渲染 */
const fieldValues = ref<Record<string, string>>({});
const noteText = ref("");
const showNote = ref(false);
const loading = ref(false);

/** 键盘与光标最小间距（px），微信 input 文档：cursor-spacing */
const inputCursorSpacing = 120;

/** 键盘弹起高度（px），用于整体上移底栏，避免固定底栏被数字键盘遮挡 */
const keyboardHeightPx = ref(0);
let offKeyboardHeight: (() => void) | undefined;

/** 在系统上报高度基础上再抬高一点，留足输入框与键盘之间的可视间隙 */
const KEYBOARD_EXTRA_LIFT_PX = 32;

const sheetLiftStyle = computed(() => {
  const h = keyboardHeightPx.value;
  if (h <= 0) return {};
  return {
    transform: `translateY(-${h + KEYBOARD_EXTRA_LIFT_PX}px)`,
    transition: "transform 0.2s ease-out",
  };
});

function bindKeyboardHeightListener() {
  offKeyboardHeight?.();
  offKeyboardHeight = undefined;
  keyboardHeightPx.value = 0;
  if (typeof uni.onKeyboardHeightChange !== "function") return;
  offKeyboardHeight = uni.onKeyboardHeightChange((res) => {
    const n = typeof res.height === "number" ? res.height : 0;
    keyboardHeightPx.value = n;
  });
}

function seedFields() {
  const m: Record<string, string> = {};
  for (const f of props.template.recommendedFields) {
    m[f.key] = "";
  }
  fieldValues.value = m;
}

watch(
  () => [props.visible, props.template.value] as const,
  ([vis]) => {
    if (vis) {
      noteText.value = "";
      showNote.value = false;
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

/** 兼容 uni-app / 微信小程序对 input 事件的封装 */
function pickInputValue(e: unknown): string {
  if (e == null || typeof e !== "object") return "";
  const ex = e as Record<string, unknown>;
  const detail = ex.detail as Record<string, unknown> | undefined;
  if (detail && "value" in detail && detail.value != null) {
    return String(detail.value);
  }
  const mp = ex.mp as Record<string, unknown> | undefined;
  const mpDetail = mp?.detail as Record<string, unknown> | undefined;
  if (mpDetail && "value" in mpDetail && mpDetail.value != null) {
    return String(mpDetail.value);
  }
  return "";
}

function onFieldInput(key: string, e: unknown) {
  const v = pickInputValue(e);
  fieldValues.value = { ...fieldValues.value, [key]: v };
}

function onClose() {
  emit("close");
}

function buildPayload(): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  const fv = fieldValues.value;
  for (const field of props.template.recommendedFields) {
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

async function onSave() {
  if (loading.value) return;
  loading.value = true;
  try {
    const payload = buildPayload();
    const summary = noteText.value.trim() || buildSummary(payload);
    emit("saved", payload, summary);
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  z-index: 999;
  display: flex;
  align-items: flex-end;
}

.sheet {
  width: 100%;
  background: $cj-surface;
  border-radius: $cj-radius-lg $cj-radius-lg 0 0;
  padding: $cj-gap-lg $cj-gap-md calc(env(safe-area-inset-bottom, 0px) + 32rpx);
  box-sizing: border-box;
}

.sheet-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $cj-gap-lg;
}

.sheet-title {
  font-size: 34rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.sheet-close {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: $cj-surface-2;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sheet-close-x {
  font-size: 36rpx;
  color: $cj-text-muted;
  line-height: 1;
}

.sheet-body-scroll {
  max-height: 56vh;
  width: 100%;
  box-sizing: border-box;
}

.sheet-body {
  margin-bottom: $cj-gap-lg;
}

.field {
  margin-bottom: $cj-gap-md;
}

.field-label {
  display: block;
  margin-bottom: 10rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
}

.field-unit {
  font-size: 22rpx;
}

.field-input {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  font-size: 32rpx;
  color: $cj-text;
  min-height: 96rpx;
  display: flex;
  align-items: center;
}

.select-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16rpx;
}

.select-chip {
  padding: 14rpx 28rpx;
  border-radius: $cj-radius-pill;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
}

.select-chip.selected {
  background: $cj-primary-soft;
  border-color: $cj-primary;
}

.select-chip-text {
  font-size: 26rpx;
  color: $cj-text;
}

.note-toggle {
  margin-bottom: $cj-gap-sm;
}

.note-toggle-text {
  font-size: 24rpx;
  color: $cj-primary;
}

.note-area {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  font-size: 28rpx;
  color: $cj-text;
  min-height: 140rpx;
  line-height: 1.6;
}

.save-btn {
  width: 100%;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  font-size: 30rpx;
}

.save-btn-hover {
  opacity: 0.92;
}
</style>
