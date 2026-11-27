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
        <view v-if="supportsTimer" class="timer-block">
          <view class="timer-head">
            <text class="timer-label">计时</text>
            <text class="timer-display">{{ timerDisplay }}</text>
          </view>
          <view class="timer-actions">
            <button
              v-if="!timerRunning"
              size="mini"
              class="timer-btn timer-btn--start"
              hover-class="timer-btn-hover"
              @tap.stop="startTimer"
            >
              开始计时
            </button>
            <template v-else>
              <button size="mini" class="timer-btn timer-btn--stop" hover-class="timer-btn-hover" @tap.stop="stopTimer">
                结束并填入
              </button>
              <button size="mini" class="timer-btn timer-btn--reset" hover-class="timer-btn-hover" @tap.stop="resetTimer">
                重置
              </button>
            </template>
          </view>
          <text v-if="timerHint" class="timer-hint">{{ timerHint }}</text>
        </view>
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
          <template v-else-if="field.type === 'date'">
            <picker mode="date" :value="fieldValues[field.key] || ''" @change="(e) => setSelect(field.key, e.detail.value)">
              <view :class="['field-input', !fieldValues[field.key] && 'is-placeholder']">{{ fieldValues[field.key] || "请选择日期" }}</view>
            </picker>
          </template>
          <template v-else>
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

        <view v-if="template.optionalFields.length" class="note-toggle" @tap.stop="showMore = !showMore">
          <text class="note-toggle-text">{{ showMore ? "收起更多信息" : "更多信息" }}</text>
        </view>
        <template v-if="showMore">
          <view v-for="field in template.optionalFields" :key="field.key" class="field">
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
            <template v-else-if="field.type === 'date'">
              <picker mode="date" :value="fieldValues[field.key] || ''" @change="(e) => setSelect(field.key, e.detail.value)">
                <view :class="['field-input', !fieldValues[field.key] && 'is-placeholder']">{{ fieldValues[field.key] || "请选择日期" }}</view>
              </picker>
            </template>
            <template v-else>
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
        </template>

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

      <button v-if="hasLast" class="reuse-btn" hover-class="save-btn-hover" @tap.stop="onReuseLast">沿用上次</button>
      <button class="save-btn" :loading="loading" hover-class="save-btn-hover" @tap.stop="onSave">保存</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import type { RecordTemplate } from "@/utils/recordTypes";

const TIMER_STORAGE_KEY = "chuyaji_timer";
const TIMER_MAX_MS = 24 * 60 * 60 * 1000;

type StoredTimer = {
  type: string;
  babyId: number;
  startedAt: number;
};

const props = defineProps<{
  visible: boolean;
  template: RecordTemplate;
  babyId?: number;
  lastPayload?: Record<string, unknown> | null;
  lastSummary?: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved", payload: Record<string, unknown>, summary: string): void;
}>();

const fieldValues = ref<Record<string, string>>({});
const noteText = ref("");
const showNote = ref(false);
const showMore = ref(false);
const loading = ref(false);

const timerStartedAt = ref(0);
const timerNow = ref(Date.now());
const timerHint = ref("");
let timerTick: ReturnType<typeof setInterval> | null = null;

const supportsTimer = computed(
  () => props.template.value === "feeding" || props.template.value === "sleep"
);
const timerRunning = computed(() => timerStartedAt.value > 0);
const elapsedMs = computed(() =>
  timerRunning.value ? Math.max(0, timerNow.value - timerStartedAt.value) : 0
);
const timerDisplay = computed(() => formatElapsed(elapsedMs.value));

const inputCursorSpacing = 120;
const hasLast = computed(() => !!props.lastPayload && Object.keys(props.lastPayload).length > 0);

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
  for (const f of [...props.template.recommendedFields, ...props.template.optionalFields]) {
    m[f.key] = "";
  }
  fieldValues.value = m;
}

function formatElapsed(ms: number) {
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  if (h > 0) {
    return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  }
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

function clearTimerTick() {
  if (timerTick) {
    clearInterval(timerTick);
    timerTick = null;
  }
}

function ensureTimerTick() {
  clearTimerTick();
  if (!timerRunning.value) return;
  timerTick = setInterval(() => {
    timerNow.value = Date.now();
    if (elapsedMs.value > TIMER_MAX_MS) {
      timerHint.value = "已超过 24 小时，请重置后重新计时";
    }
  }, 1000);
}

function readStoredTimer(): StoredTimer | null {
  try {
    const raw = uni.getStorageSync(TIMER_STORAGE_KEY);
    if (!raw) return null;
    if (typeof raw === "string") {
      return JSON.parse(raw) as StoredTimer;
    }
    return raw as StoredTimer;
  } catch {
    return null;
  }
}

function writeStoredTimer(data: StoredTimer | null) {
  try {
    if (!data) {
      uni.removeStorageSync(TIMER_STORAGE_KEY);
      return;
    }
    uni.setStorageSync(TIMER_STORAGE_KEY, data);
  } catch {}
}

function applyDurationMinutes(ms: number) {
  const minutes = Math.max(1, Math.round(ms / 60000));
  fieldValues.value = { ...fieldValues.value, duration_min: String(minutes) };
  if (props.template.value === "feeding") {
    showMore.value = true;
  }
}

function restoreTimerFromStorage() {
  timerHint.value = "";
  if (!supportsTimer.value) {
    timerStartedAt.value = 0;
    clearTimerTick();
    return;
  }
  const stored = readStoredTimer();
  if (!stored?.startedAt) {
    timerStartedAt.value = 0;
    clearTimerTick();
    return;
  }
  const babyId = Number(props.babyId || 0);
  if (stored.babyId && babyId && stored.babyId !== babyId) {
    writeStoredTimer(null);
    timerStartedAt.value = 0;
    timerHint.value = "已切换宝宝，旧计时已清除";
    clearTimerTick();
    return;
  }
  if (stored.type !== props.template.value) {
    timerStartedAt.value = 0;
    clearTimerTick();
    return;
  }
  timerStartedAt.value = stored.startedAt;
  timerNow.value = Date.now();
  if (elapsedMs.value > TIMER_MAX_MS) {
    timerHint.value = "已超过 24 小时，请重置后重新计时";
  }
  ensureTimerTick();
}

function startTimer() {
  const babyId = Number(props.babyId || 0);
  const startedAt = Date.now();
  timerStartedAt.value = startedAt;
  timerNow.value = startedAt;
  timerHint.value = "";
  writeStoredTimer({ type: props.template.value, babyId, startedAt });
  ensureTimerTick();
}

function stopTimer() {
  if (!timerRunning.value) return;
  const ms = elapsedMs.value;
  if (ms > TIMER_MAX_MS) {
    timerHint.value = "已超过 24 小时，请重置后重新计时";
    return;
  }
  applyDurationMinutes(ms);
  timerStartedAt.value = 0;
  writeStoredTimer(null);
  clearTimerTick();
  timerHint.value = "已填入时长，可再手改";
}

function resetTimer() {
  timerStartedAt.value = 0;
  timerHint.value = "";
  writeStoredTimer(null);
  clearTimerTick();
}

function clearTimerAfterSave() {
  if (supportsTimer.value) {
    writeStoredTimer(null);
    timerStartedAt.value = 0;
    timerHint.value = "";
    clearTimerTick();
  }
}

watch(
  () => [props.visible, props.template.value, props.babyId] as const,
  ([vis]) => {
    if (vis) {
      noteText.value = "";
      showNote.value = false;
      showMore.value = false;
      seedFields();
      bindKeyboardHeightListener();
      restoreTimerFromStorage();
    } else {
      keyboardHeightPx.value = 0;
      offKeyboardHeight?.();
      offKeyboardHeight = undefined;
      clearTimerTick();
    }
  },
  { immediate: true }
);

onUnmounted(() => {
  offKeyboardHeight?.();
  offKeyboardHeight = undefined;
  clearTimerTick();
});

function setSelect(key: string, value: string) {
  fieldValues.value = { ...fieldValues.value, [key]: value };
}

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
  for (const field of [...props.template.recommendedFields, ...props.template.optionalFields]) {
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
    clearTimerAfterSave();
    emit("saved", payload, summary);
  } finally {
    loading.value = false;
  }
}

function onReuseLast() {
  if (!props.lastPayload) return;
  const payload = { ...props.lastPayload };
  const summary = props.lastSummary?.trim() || buildSummary(payload);
  clearTimerAfterSave();
  emit("saved", payload, summary);
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
  margin-bottom: $cj-gap-lg;
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
  max-height: 56vh;
  width: 100%;
  box-sizing: border-box;
}

.sheet-body {
  margin-bottom: $cj-gap-lg;
}

.timer-block {
  margin-bottom: $cj-gap-md;
  padding: 20rpx 22rpx;
  background: linear-gradient(155deg, rgba(255, 247, 240, 0.95) 0%, rgba(230, 241, 236, 0.45) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
}

.timer-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16rpx;
  margin-bottom: 14rpx;
}

.timer-label {
  font-size: 23rpx;
  color: $cj-text-muted;
  font-weight: 500;
}

.timer-display {
  font-size: 40rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 1rpx;
  font-variant-numeric: tabular-nums;
}

.timer-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.timer-btn {
  margin: 0 !important;
  border-radius: $cj-radius-pill !important;
  font-size: 24rpx !important;
  padding: 0 28rpx !important;
}

.timer-btn--start {
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}

.timer-btn--stop {
  background: $cj-mint-soft !important;
  color: $cj-tag-postnatal-text !important;
  border: 1rpx solid rgba(125, 171, 152, 0.35) !important;
}

.timer-btn--reset {
  background: $cj-surface !important;
  color: $cj-text-secondary !important;
  border: 1rpx solid $cj-border-faint !important;
}

.timer-btn-hover {
  opacity: 0.9;
}

.timer-hint {
  display: block;
  margin-top: 12rpx;
  font-size: 22rpx;
  color: $cj-text-secondary;
  line-height: 1.45;
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

.note-toggle {
  margin-bottom: $cj-gap-sm;
}

.note-toggle-text {
  font-size: 23rpx;
  color: $cj-primary;
  font-weight: 500;
}

.note-area {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
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
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  font-size: 30rpx;
  box-shadow: $cj-shadow-soft;
}

.reuse-btn {
  width: 100%;
  margin-bottom: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-faint !important;
  font-size: 28rpx;
}

.save-btn-hover {
  opacity: 0.92;
}
</style>
