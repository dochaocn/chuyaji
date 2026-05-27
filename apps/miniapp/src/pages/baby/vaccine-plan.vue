<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">疫苗计划</text>
      <text class="head-title">按出生日期生成接种安排</text>
      <text class="head-desc">计划仅用于记录和提醒，实际接种以当地接种门诊建议为准。</text>
    </view>

    <view v-if="!birthDate" class="empty-card">
      <text class="empty-title">先补充出生日期</text>
      <text class="empty-desc">完善宝宝档案后即可生成疫苗计划。</text>
      <button class="main-btn" @click="goProfile">编辑档案</button>
    </view>

    <view v-else class="timeline">
      <view v-for="item in planRows" :key="item.id" class="timeline-row">
        <view class="rail">
          <view :class="['dot', `dot--${item.status}`]" />
          <view class="line" />
        </view>
        <view class="card">
          <view class="card-head">
            <view class="title-wrap">
              <text class="vaccine-name">{{ item.name }}</text>
              <text class="dose">{{ item.dose_label }}</text>
            </view>
            <text :class="['status-chip', `status-chip--${item.status}`]">{{ statusLabel(item.status) }}</text>
          </view>
          <text class="date">{{ item.scheduled_date }}</text>
          <text class="desc">{{ item.description }}</text>
          <view class="actions">
            <button v-if="item.record_id" size="mini" class="ghost-btn" @click="openRecord(item.record_id)">查看记录</button>
            <button v-else size="mini" class="main-btn main-btn--mini" @click="createRecord(item)">创建接种记录</button>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import { apiGetBaby, apiListRecords, type RecordItem } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";
import { VACCINE_SCHEDULE, vaccineDoseId, type VaccineScheduleItem } from "@/utils/vaccineSchedule";

const session = useSessionStore();
const babyId = ref(0);
const birthDate = ref("");
const records = ref<RecordItem[]>([]);

type PlanStatus = "done" | "overdue" | "pending";

type PlanRow = VaccineScheduleItem & {
  id: string;
  scheduled_date: string;
  status: PlanStatus;
  record_id?: number;
};

onLoad((query: Record<string, string | undefined>) => {
  session.load();
  babyId.value = Number(query.baby_id || session.babyId || 0);
});

onShow(load);

async function load() {
  if (!babyId.value) return;
  try {
    const baby = await apiGetBaby(babyId.value);
    birthDate.value = baby.birth_date || "";
    if (!birthDate.value) return;
    const page = await apiListRecords(babyId.value, { limit: 100, phase: "postnatal", record_type: "vaccine" });
    records.value = page.items || [];
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  }
}

const planRows = computed<PlanRow[]>(() => {
  if (!birthDate.value) return [];
  const start = Date.parse(birthDate.value);
  const today = new Date();
  return VACCINE_SCHEDULE.map((item) => {
    const scheduled = new Date(start + item.recommended_age_days * 24 * 3600 * 1000);
    const id = vaccineDoseId(item);
    const record = records.value.find((r) => {
      const payload = r.payload as Record<string, unknown>;
      return `${payload.vaccine_key}:${payload.dose_key}` === id || (payload.name === item.name && payload.dose_label === item.dose_label);
    });
    const status: PlanStatus = record ? "done" : scheduled < today ? "overdue" : "pending";
    return { ...item, id, scheduled_date: formatDateValue(scheduled), status, record_id: record?.id };
  });
});

function formatDateValue(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function statusLabel(status: PlanStatus) {
  if (status === "done") return "已接种";
  if (status === "overdue") return "已逾期";
  return "待接种";
}

function createRecord(item: PlanRow) {
  const payload = {
    vaccine_key: item.vaccine_key,
    dose_key: item.dose_key,
    name: item.name,
    dose_label: item.dose_label,
    scheduled_date: item.scheduled_date,
    vaccine_date: formatDateValue(new Date()),
    next_time: item.scheduled_date,
  };
  uni.setStorageSync("chuyaji_baby_record_prefill", { payload, summary: `${item.name}${item.dose_label}接种记录` });
  uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${babyId.value}&phase=postnatal&type=vaccine` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${id}&baby_id=${babyId.value}` });
}

function goProfile() {
  uni.navigateTo({ url: `/pages/baby/profile-edit?baby_id=${babyId.value}` });
}
</script>

<style lang="scss" scoped>
.page { padding: $cj-page-pad-y $cj-page-pad-x 72rpx; }
.head { margin-bottom: $cj-gap-lg; }
.head-kicker { display: block; font-size: 22rpx; letter-spacing: 4rpx; color: $cj-mint; margin-bottom: 8rpx; }
.head-title { display: block; font-size: 42rpx; color: $cj-ink; font-weight: $cj-fw-display; }
.head-desc { display: block; margin-top: $cj-gap-sm; font-size: 26rpx; color: $cj-text-secondary; line-height: 1.6; }
.empty-card, .card { background: $cj-surface; border: 1rpx solid $cj-border-light; border-radius: $cj-radius-lg; box-shadow: $cj-shadow-card; padding: $cj-gap-md; }
.empty-title { display: block; color: $cj-ink; font-size: 30rpx; font-weight: $cj-fw-display; }
.empty-desc { display: block; margin-top: 10rpx; color: $cj-text-secondary; font-size: 25rpx; line-height: 1.6; }
.timeline { display: flex; flex-direction: column; gap: 8rpx; }
.timeline-row { display: flex; gap: 20rpx; }
.rail { width: 28rpx; display: flex; flex-direction: column; align-items: center; padding-top: 28rpx; flex-shrink: 0; }
.dot { width: 18rpx; height: 18rpx; border-radius: 50%; border: 3rpx solid $cj-surface; box-shadow: 0 0 0 2rpx rgba(234,217,204,.9); }
.dot--done { background: $cj-mint; }
.dot--overdue { background: $cj-primary; }
.dot--pending { background: $cj-accent; }
.line { flex: 1; width: 2rpx; min-height: 42rpx; margin-top: 8rpx; background: $cj-border; }
.card { flex: 1; margin-bottom: $cj-gap-sm; }
.card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: $cj-gap-sm; }
.title-wrap { min-width: 0; }
.vaccine-name { display: block; font-size: 30rpx; color: $cj-ink; font-weight: $cj-fw-display; }
.dose, .date, .desc { display: block; margin-top: 8rpx; font-size: 24rpx; color: $cj-text-secondary; line-height: 1.55; }
.status-chip { flex-shrink: 0; padding: 6rpx 18rpx; border-radius: $cj-radius-pill; font-size: 20rpx; font-weight: $cj-fw-title; }
.status-chip--done { background: $cj-mint-soft; color: $cj-tag-postnatal-text; }
.status-chip--overdue { background: $cj-danger-bg; color: $cj-danger-text; }
.status-chip--pending { background: $cj-warn-bg; color: $cj-text-secondary; }
.actions { margin-top: $cj-gap-md; }
.main-btn, .ghost-btn { border-radius: $cj-radius-pill !important; }
.main-btn { margin-top: $cj-gap-md; background: linear-gradient(165deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important; color: #fffefb !important; border: none !important; }
.main-btn--mini { margin-top: 0; width: 100%; }
.ghost-btn { width: 100%; background: $cj-surface-2 !important; color: $cj-text !important; border: 1rpx solid $cj-border-light !important; }
</style>
