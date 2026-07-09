<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ motherId ? "编辑宝妈档案" : "创建宝妈档案" }}</text>
      <text class="head-desc">把宝妈的基础信息、孕前状态和产后阶段记录下来，方便后续持续补充。</text>
    </view>

    <view class="section">
      <text class="section-label">基础信息</text>
      <view class="field">
        <text class="lab">称呼</text>
        <input v-model="name" class="input" placeholder="例如 妈妈 / 小琪" />
      </view>
      <view class="field">
        <text class="lab">状态</text>
        <picker :range="statusLabels" :value="statusIndex" @change="onStatusChange">
          <view class="picker-input">{{ statusLabel }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">生日</text>
        <picker mode="date" :value="birthday" @change="onDateChange('birthday', $event)">
          <view :class="['picker-input', !birthday && 'is-placeholder']">{{ birthday || "请选择日期" }}</view>
        </picker>
      </view>
    </view>

    <view class="section">
      <text class="section-label">健康资料</text>
      <view class="field">
        <text class="lab">身高 cm</text>
        <input v-model="heightCM" class="input" placeholder="例如 165" />
      </view>
      <view class="field">
        <text class="lab">孕前体重 kg</text>
        <input v-model="preWeightKG" class="input" placeholder="例如 52" />
      </view>
      <view class="field">
        <text class="lab">血型</text>
        <picker :range="bloodLabels" :value="bloodIndex" @change="onBloodChange">
          <view :class="['picker-input', !bloodType && 'is-placeholder']">{{ bloodDisplayLabel }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">过敏史</text>
        <textarea v-model="allergyHistory" class="area" placeholder="可选" />
      </view>
      <view class="field">
        <text class="lab">病史</text>
        <textarea v-model="medicalHistory" class="area" placeholder="可选" />
      </view>
      <view class="field">
        <text class="lab">分娩日期</text>
        <picker mode="date" :value="deliveryDate" @change="onDateChange('deliveryDate', $event)">
          <view :class="['picker-input', !deliveryDate && 'is-placeholder']">
            {{ deliveryDate || "请选择日期" }}
          </view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">备注</text>
        <textarea v-model="note" class="area" placeholder="可补充重点提醒" />
      </view>
    </view>

    <button class="main-btn" :loading="loading" @click="save">保存档案</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { apiCreateMother, apiGetMother, apiPatchMother } from "@/api/chuyaji";
import type { Mother } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";

const session = useSessionStore();
const motherId = ref(0);
const loading = ref(false);
const name = ref("");
const status = ref("pregnant");
const birthday = ref("");
const heightCM = ref("");
const preWeightKG = ref("");
const bloodType = ref("");
const allergyHistory = ref("");
const medicalHistory = ref("");
const deliveryDate = ref("");
const note = ref("");

const STATUS_ITEMS: { value: Mother["status"]; label: string }[] = [
  { value: "pregnant", label: "怀孕中" },
  { value: "postpartum", label: "产后恢复" },
  { value: "parenting", label: "育儿期" },
];

const statusLabels = STATUS_ITEMS.map((i) => i.label);

const statusIndex = computed(() => {
  const idx = STATUS_ITEMS.findIndex((i) => i.value === status.value);
  return idx >= 0 ? idx : 0;
});

const statusLabel = computed(() => {
  const hit = STATUS_ITEMS.find((i) => i.value === status.value);
  return hit?.label ?? "怀孕中";
});

const BLOOD_ITEMS: { value: string; label: string }[] = [
  { value: "", label: "未填写" },
  { value: "A 型", label: "A 型" },
  { value: "B 型", label: "B 型" },
  { value: "AB 型", label: "AB 型" },
  { value: "O 型", label: "O 型" },
  { value: "A+", label: "A（Rh+）" },
  { value: "A-", label: "A（Rh-）" },
  { value: "B+", label: "B（Rh+）" },
  { value: "B-", label: "B（Rh-）" },
  { value: "AB+", label: "AB（Rh+）" },
  { value: "AB-", label: "AB（Rh-）" },
  { value: "O+", label: "O（Rh+）" },
  { value: "O-", label: "O（Rh-）" },
  { value: "未知", label: "未知" },
];

const bloodLabels = BLOOD_ITEMS.map((i) => i.label);

const bloodIndex = computed(() => {
  const idx = BLOOD_ITEMS.findIndex((i) => i.value === bloodType.value);
  return idx >= 0 ? idx : 0;
});

const bloodDisplayLabel = computed(() => {
  const hit = BLOOD_ITEMS.find((i) => i.value === bloodType.value);
  if (hit) return hit.label;
  if (bloodType.value.trim()) return bloodType.value;
  return "未填写";
});

onLoad((query: Record<string, string | undefined>) => {
  session.load();
  motherId.value = Number(query.id || session.motherId || 0);
  if (motherId.value) {
    loadExisting();
  }
});

function toISO(value: string): string | undefined {
  const time = Date.parse(value);
  if (!Number.isFinite(time)) return undefined;
  return new Date(time).toISOString();
}

function toNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function coerceBloodTypeFromServer(raw: string): string {
  const t = (raw || "").trim();
  if (!t) return "";
  if (BLOOD_ITEMS.some((i) => i.value === t)) return t;
  const aliases: Record<string, string> = {
    O型: "O 型",
    A型: "A 型",
    B型: "B 型",
    AB型: "AB 型",
    o型: "O 型",
    a型: "A 型",
    b型: "B 型",
    ab型: "AB 型",
  };
  return aliases[t] ?? t;
}

function onDateChange(field: "birthday" | "deliveryDate", event: { detail: { value: string } }) {
  const value = event.detail.value || "";
  if (field === "birthday") birthday.value = value;
  if (field === "deliveryDate") deliveryDate.value = value;
}

function onStatusChange(event: { detail: { value: string } }) {
  const item = STATUS_ITEMS[Number(event.detail.value)];
  if (item) {
    status.value = item.value;
  }
}

function onBloodChange(event: { detail: { value: string } }) {
  const item = BLOOD_ITEMS[Number(event.detail.value)];
  if (item) {
    bloodType.value = item.value;
  }
}

async function loadExisting() {
  try {
    const mother = await apiGetMother(motherId.value);
    name.value = mother.name || "";
    status.value = mother.status || "pregnant";
    birthday.value = mother.birthday?.slice(0, 10) || "";
    heightCM.value = mother.height_cm ? String(mother.height_cm) : "";
    preWeightKG.value = mother.pre_pregnancy_weight_kg ? String(mother.pre_pregnancy_weight_kg) : "";
    bloodType.value = coerceBloodTypeFromServer(mother.blood_type || "");
    allergyHistory.value = mother.allergy_history || "";
    medicalHistory.value = mother.medical_history || "";
    deliveryDate.value = mother.delivery_date?.slice(0, 10) || "";
    note.value = mother.note || "";
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

async function save() {
  loading.value = true;
  try {
    const nextStatus = (status.value.trim() || "pregnant") as Mother["status"];
    const body = {
      name: name.value.trim() || undefined,
      status: nextStatus,
      birthday: toISO(birthday.value),
      height_cm: toNumber(heightCM.value),
      pre_pregnancy_weight_kg: toNumber(preWeightKG.value),
      blood_type: bloodType.value.trim() || undefined,
      allergy_history: allergyHistory.value.trim() || undefined,
      medical_history: medicalHistory.value.trim() || undefined,
      delivery_date: toISO(deliveryDate.value),
      note: note.value.trim() || undefined,
    };
    const result = motherId.value ? await apiPatchMother(motherId.value, body) : await apiCreateMother(body);
    session.setMother(result.id);
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => {
      const pages = getCurrentPages();
      if (pages.length > 1) {
        uni.navigateBack();
      } else {
        uni.switchTab({ url: "/pages/mom/home" });
      }
    }, 250);
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "保存失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 64rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
  padding-top: 4rpx;
}

.head-title {
  display: block;
  font-size: 42rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: -0.5rpx;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.section {
  background: $cj-surface;
  border-radius: $cj-radius-xl;
  border: 1rpx solid $cj-border-faint;
  box-shadow: $cj-shadow-card;
  padding: 28rpx;
  margin-bottom: $cj-gap-md;
}

.section-label {
  display: block;
  margin-bottom: $cj-gap-md;
  font-size: 21rpx;
  color: $cj-text-muted;
  letter-spacing: 3rpx;
  font-weight: 500;
}

.field {
  margin-bottom: $cj-gap-md;
}

.field:last-child {
  margin-bottom: 0;
}

.lab {
  display: block;
  margin-bottom: 10rpx;
  font-size: 23rpx;
  color: $cj-text-muted;
  font-weight: 500;
  letter-spacing: 0.3rpx;
}

.input,
.picker-input,
.area {
  width: 100%;
  box-sizing: border-box;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
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

.main-btn {
  margin-top: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  box-shadow: $cj-shadow-soft;
}
</style>
