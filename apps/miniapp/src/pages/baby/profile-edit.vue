<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ babyId ? "编辑宝宝档案" : "创建宝宝档案" }}</text>
      <text class="head-desc">把基础信息、孕期关键日期和出生资料放在一起，后续记录会自动关联当前宝宝。</text>
    </view>

    <view class="section">
      <text class="section-label">基础信息</text>
      <view class="field">
        <text class="lab">昵称</text>
        <input v-model="nickname" class="input" placeholder="例如 小初芽" />
      </view>
      <view class="field">
        <text class="lab">性别</text>
        <input v-model="gender" class="input" placeholder="可选" />
      </view>
    </view>

    <view class="section">
      <text class="section-label">怀孕期资料</text>
      <view class="field">
        <text class="lab">末次月经</text>
        <picker mode="date" :value="lmp" @change="onDateChange('lmp', $event)">
          <view :class="['picker-input', !lmp && 'is-placeholder']">{{ lmp || "请选择日期" }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">预产期</text>
        <picker mode="date" :value="edd" @change="onDateChange('edd', $event)">
          <view :class="['picker-input', !edd && 'is-placeholder']">{{ edd || "请选择日期" }}</view>
        </picker>
        <text class="field-hint">填写末次月经后，将按孕 40 周（+280 天）自动推算；也可再手动改。</text>
      </view>
    </view>

    <view class="section">
      <text class="section-label">出生与喂养</text>
      <view class="field">
        <text class="lab">出生日期</text>
        <picker mode="date" :value="birth" @change="onDateChange('birth', $event)">
          <view :class="['picker-input', !birth && 'is-placeholder']">{{ birth || "请选择日期" }}</view>
        </picker>
      </view>
      <view class="field">
        <text class="lab">出生体重 g</text>
        <input v-model="birthWeightG" class="input" placeholder="例如 3500" />
      </view>
      <view class="field">
        <text class="lab">出生身长 cm</text>
        <input v-model="birthHeightCM" class="input" placeholder="例如 50" />
      </view>
      <view class="field">
        <text class="lab">出生医院</text>
        <input v-model="birthHospital" class="input" placeholder="可选" />
      </view>
      <view class="field">
        <text class="lab">喂养方式</text>
        <input v-model="feedingType" class="input" placeholder="例如 母乳 / 配方奶 / 混合" />
      </view>
      <view class="field">
        <text class="lab">备注</text>
        <textarea v-model="note" class="area" placeholder="可补充特殊情况或提醒" />
      </view>
    </view>

    <button class="main-btn" :loading="loading" @click="save">保存档案</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiCreateBaby, apiGetBaby, apiPatchBaby, apiGetMother } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";

const session = useSessionStore();
const babyId = ref(0);
const loading = ref(false);
const nickname = ref("");
const gender = ref("");
const lmp = ref("");
const edd = ref("");
const birth = ref("");
const birthWeightG = ref("");
const birthHeightCM = ref("");
const birthHospital = ref("");
const feedingType = ref("");
const note = ref("");

onLoad((query: Record<string, string | undefined>) => {
  session.load();
  babyId.value = Number(query.id || session.babyId || 0);
  if (babyId.value) {
    loadExisting();
  } else {
    // 新建宝宝档案：若已有宝妈档案（postpartum/parenting 且有 delivery_date），自动预填出生日期与医院
    prefillFromMother();
  }
});

async function prefillFromMother() {
  if (!session.motherId) return;
  try {
    const mother = await apiGetMother(session.motherId);
    // 仅在宝妈已分娩（postpartum/parenting）且有分娩日期时才预填
    if (
      (mother.status === "postpartum" || mother.status === "parenting") &&
      mother.delivery_date
    ) {
      birth.value = mother.delivery_date.slice(0, 10);
    }
    if (mother.note && !birthHospital.value) {
      // 医院信息在宝妈档案里没有独立字段，此处不强制预填，保留空值
    }
  } catch {
    // 无宝妈档案或加载失败时静默忽略，不影响建档流程
  }
}

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

/** 末次月经首日 +280 天（孕 40 周），用于推算预产期 */
function eddFromLmp(lmpYmd: string): string {
  const t = (lmpYmd || "").trim();
  if (!t) return "";
  return addCalendarDaysYmd(t, 280);
}

function addCalendarDaysYmd(ymd: string, deltaDays: number): string {
  const parts = ymd.split("-").map((p) => Number(p));
  if (parts.length !== 3 || parts.some((n) => !Number.isFinite(n))) return "";
  const [y, m, d] = parts;
  const dt = new Date(y, m - 1, d);
  if (Number.isNaN(dt.getTime())) return "";
  dt.setDate(dt.getDate() + deltaDays);
  const yy = dt.getFullYear();
  const mm = String(dt.getMonth() + 1).padStart(2, "0");
  const dd = String(dt.getDate()).padStart(2, "0");
  return `${yy}-${mm}-${dd}`;
}

function onDateChange(field: "lmp" | "edd" | "birth", event: { detail: { value: string } }) {
  const value = event.detail.value || "";
  if (field === "lmp") {
    lmp.value = value;
    edd.value = value ? eddFromLmp(value) : "";
  }
  if (field === "edd") edd.value = value;
  if (field === "birth") birth.value = value;
}

async function loadExisting() {
  try {
    const baby = await apiGetBaby(babyId.value);
    nickname.value = baby.nickname || "";
    gender.value = baby.gender || "";
    lmp.value = baby.lmp_date?.slice(0, 10) || "";
    const eddStored = baby.edd_date?.slice(0, 10) || "";
    edd.value = eddStored || (lmp.value ? eddFromLmp(lmp.value) : "");
    birth.value = baby.birth_date?.slice(0, 10) || "";
    birthWeightG.value = baby.birth_weight_g ? String(baby.birth_weight_g) : "";
    birthHeightCM.value = baby.birth_height_cm ? String(baby.birth_height_cm) : "";
    birthHospital.value = baby.birth_hospital || "";
    feedingType.value = baby.feeding_type || "";
    note.value = baby.note || "";
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

async function save() {
  if (!nickname.value.trim()) {
    uni.showToast({ title: "请输入昵称", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    const body = {
      nickname: nickname.value.trim(),
      gender: gender.value.trim() || undefined,
      lmp_date: toISO(lmp.value),
      edd_date: toISO(edd.value),
      birth_date: toISO(birth.value),
      birth_weight_g: toNumber(birthWeightG.value),
      birth_height_cm: toNumber(birthHeightCM.value),
      birth_hospital: birthHospital.value.trim() || undefined,
      feeding_type: feedingType.value.trim() || undefined,
      note: note.value.trim() || undefined,
    };
    const result = babyId.value ? await apiPatchBaby(babyId.value, body) : await apiCreateBaby(body);
    session.setBaby(result.id);
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => {
      const pages = getCurrentPages();
      if (pages.length > 1) {
        uni.navigateBack();
      } else {
        uni.switchTab({ url: "/pages/baby/home" });
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
  text-transform: uppercase;
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
  font-size: 24rpx;
  color: $cj-text-muted;
}

.field-hint {
  display: block;
  margin-top: 12rpx;
  font-size: 22rpx;
  line-height: 1.5;
  color: $cj-text-muted;
}

.input,
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
  background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
}
</style>
