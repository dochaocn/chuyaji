<template>
  <view class="page">
    <view class="head">
      <text class="head-kicker">疫苗计划</text>
      <text class="head-title">按出生日期生成接种安排</text>
      <text class="head-desc">覆盖约至 6 岁免疫规划节点与常见自费苗。计划仅用于记录和提醒，实际接种以当地接种门诊建议为准。</text>
    </view>

    <view v-if="!birthDate" class="empty-card">
      <text class="empty-title">先补充出生日期</text>
      <text class="empty-desc">完善宝宝档案后即可生成疫苗计划。</text>
      <button class="main-btn" @click="goProfile">编辑档案</button>
    </view>

    <view v-else class="groups">
      <view v-for="group in groupedPlan" :key="group.key" class="group">
        <view class="group-head" @click="toggleGroup(group.key)">
          <view class="group-titles">
            <text class="group-title">{{ group.label }}</text>
            <text class="group-meta">{{ group.done }}/{{ group.rows.length }} 已接种</text>
          </view>
          <text class="group-chev">{{ collapsed[group.key] ? "▸" : "▾" }}</text>
        </view>
        <view v-if="!collapsed[group.key]" class="timeline">
          <view v-for="item in group.rows" :key="item.id" class="timeline-row">
            <view class="rail">
              <view :class="['dot', `dot--${item.status}`, item.type === 'optional' && 'dot--optional']" />
              <view class="line" />
            </view>
            <view class="card">
              <view class="card-head">
                <view class="title-wrap">
                  <text class="vaccine-name">{{ item.name }}</text>
                  <text class="dose">{{ item.dose_label }}</text>
                </view>
                <view class="chip-col">
                  <text v-if="item.type === 'optional'" class="opt-chip">自费</text>
                  <text :class="['status-chip', `status-chip--${item.status}`]">{{ statusLabel(item.status) }}</text>
                </view>
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
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { onLoad, onShow } from "@dcloudio/uni-app";
import { apiGetBaby, apiListRecords, type RecordItem } from "@/api/chuyaji";
import { useSessionStore } from "@/store/session";
import {
  VACCINE_GROUP_LABELS,
  VACCINE_SCHEDULE,
  vaccineDoseId,
  type VaccineScheduleItem,
} from "@/utils/vaccineSchedule";

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

const collapsed = reactive<Record<string, boolean>>({
  "0_6m": false,
  "6_12m": false,
  "12_24m": false,
  "2_6y": false,
  optional: true,
});

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

const GROUP_ORDER: VaccineScheduleItem["group"][] = ["0_6m", "6_12m", "12_24m", "2_6y", "optional"];

const groupedPlan = computed(() => {
  return GROUP_ORDER.map((key) => {
    const rows = planRows.value.filter((r) => r.group === key);
    return {
      key,
      label: VACCINE_GROUP_LABELS[key],
      rows,
      done: rows.filter((r) => r.status === "done").length,
    };
  }).filter((g) => g.rows.length > 0);
});

function toggleGroup(key: string) {
  collapsed[key] = !collapsed[key];
}

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
.page { padding: $cj-page-pad-y $cj-page-pad-x 80rpx; }
.head { margin-bottom: $cj-gap-lg; padding-top: 4rpx; }
.head-kicker { display: block; font-size: 21rpx; letter-spacing: 5rpx; color: $cj-mint; margin-bottom: 8rpx; font-weight: 500; }
.head-title { display: block; font-size: 42rpx; color: $cj-ink; font-weight: $cj-fw-display; letter-spacing: -0.5rpx; }
.head-desc { display: block; margin-top: $cj-gap-sm; font-size: 25rpx; color: $cj-text-secondary; line-height: 1.65; }
.empty-card, .card { background: $cj-surface; border: 1rpx solid $cj-border-faint; border-radius: $cj-radius-xl; box-shadow: $cj-shadow-card; padding: 28rpx; }
.empty-title { display: block; color: $cj-ink; font-size: 30rpx; font-weight: $cj-fw-display; letter-spacing: 0.3rpx; }
.empty-desc { display: block; margin-top: 10rpx; color: $cj-text-secondary; font-size: 25rpx; line-height: 1.65; }
.groups { display: flex; flex-direction: column; gap: $cj-gap-md; }
.group {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  overflow: hidden;
}
.group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 24rpx 28rpx;
  &:active { background: rgba(201, 107, 92, 0.04); }
}
.group-titles { min-width: 0; }
.group-title { display: block; font-size: 30rpx; font-weight: $cj-fw-display; color: $cj-ink; }
.group-meta { display: block; margin-top: 6rpx; font-size: 22rpx; color: $cj-text-muted; }
.group-chev { font-size: 28rpx; color: $cj-primary; flex-shrink: 0; }
.timeline { display: flex; flex-direction: column; gap: 10rpx; padding: 0 20rpx 20rpx; }
.timeline-row { display: flex; gap: 18rpx; }
.rail { width: 28rpx; display: flex; flex-direction: column; align-items: center; padding-top: 28rpx; flex-shrink: 0; }
.dot { width: 16rpx; height: 16rpx; border-radius: 50%; border: 3rpx solid $cj-surface; box-shadow: 0 0 0 2rpx $cj-border-faint; }
.dot--done { background: $cj-mint; }
.dot--overdue { background: $cj-primary; }
.dot--pending { background: $cj-accent; }
.dot--optional.dot--pending { background: #c4b5a5; }
.line { flex: 1; width: 2rpx; min-height: 42rpx; margin-top: 8rpx; background: linear-gradient(180deg, $cj-border-faint 0%, transparent 100%); }
.card { flex: 1; margin-bottom: $cj-gap-sm; box-shadow: none; }
.card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: $cj-gap-sm; }
.title-wrap { min-width: 0; }
.vaccine-name { display: block; font-size: 29rpx; color: $cj-ink; font-weight: $cj-fw-display; letter-spacing: 0.2rpx; }
.dose, .date, .desc { display: block; margin-top: 8rpx; font-size: 23rpx; color: $cj-text-secondary; line-height: 1.55; }
.chip-col { display: flex; flex-direction: column; align-items: flex-end; gap: 8rpx; flex-shrink: 0; }
.opt-chip { padding: 4rpx 12rpx; border-radius: $cj-radius-pill; font-size: 18rpx; background: $cj-surface-2; color: $cj-text-muted; }
.status-chip { flex-shrink: 0; padding: 5rpx 16rpx; border-radius: $cj-radius-pill; font-size: 19rpx; font-weight: $cj-fw-title; letter-spacing: 0.3rpx; }
.status-chip--done { background: $cj-mint-soft; color: $cj-tag-postnatal-text; }
.status-chip--overdue { background: $cj-danger-bg; color: $cj-danger-text; }
.status-chip--pending { background: $cj-warn-bg; color: $cj-text-secondary; }
.actions { margin-top: $cj-gap-md; }
.main-btn, .ghost-btn { border-radius: $cj-radius-pill !important; }
.main-btn { margin-top: $cj-gap-md; background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important; color: #fffefb !important; border: none !important; box-shadow: $cj-shadow-soft; }
.main-btn--mini { margin-top: 0; width: 100%; }
.ghost-btn { width: 100%; background: $cj-surface-2 !important; color: $cj-text !important; border: 1rpx solid $cj-border-faint !important; }
</style>
