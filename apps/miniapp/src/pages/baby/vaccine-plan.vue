<template>
  <view class="page">
    <view class="head">
      <text class="head-title">疫苗计划</text>
      <text v-if="birthDate" class="head-sub">{{ ageLabel }} · 共 {{ planRows.length }} 项，已接种 {{ doneCount }}</text>
      <text v-else class="head-sub">按出生日期生成接种安排</text>
    </view>

    <view v-if="!birthDate" class="empty-card">
      <text class="empty-title">先补充出生日期</text>
      <text class="empty-desc">完善宝宝档案后即可生成疫苗计划。</text>
      <button class="main-btn" @click="goProfile">编辑档案</button>
    </view>

    <template v-else>
      <!-- 月龄筛选 -->
      <view class="filter-row">
        <view
          v-for="opt in filterOptions"
          :key="opt.key"
          :class="['filter-chip', activeFilter === opt.key && 'active']"
          @click="selectFilter(opt.key)"
        >
          <text class="filter-chip-text">{{ opt.label }}</text>
        </view>
      </view>

      <view v-if="!groupedPlan.length" class="empty-card">
        <text class="empty-title">该月龄段暂无计划</text>
      </view>

      <view v-else class="groups">
        <view v-for="group in groupedPlan" :key="group.key" class="group">
          <view class="group-head" @click="toggleGroup(group.key)">
            <view class="group-titles">
              <view class="group-title-row">
                <text class="group-title">{{ group.label }}</text>
                <text v-if="group.key === currentGroupKey" class="current-tag">当前月龄</text>
              </view>
              <text class="group-meta">{{ group.done }}/{{ group.rows.length }} 已接种</text>
            </view>
            <text class="group-chev">{{ openKey === group.key ? "▾" : "▸" }}</text>
          </view>

          <view v-if="openKey === group.key" class="timeline">
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
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
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
/** 当前展开的分组；null 表示全部折叠 */
const openKey = ref<VaccineScheduleItem["group"] | null>(null);
/** 月龄筛选；空串表示全部 */
const activeFilter = ref<VaccineScheduleItem["group"] | "">("");

type PlanStatus = "done" | "overdue" | "pending";

type PlanRow = VaccineScheduleItem & {
  id: string;
  scheduled_date: string;
  status: PlanStatus;
  record_id?: number;
};

type GroupKey = VaccineScheduleItem["group"];

/** 各月龄段的起始天数，用于定位「当前月龄」并按优先级排序 */
const GROUP_START_DAYS: Record<GroupKey, number> = {
  "0_6m": 0,
  "6_12m": 183,
  "12_24m": 366,
  "2_6y": 731,
  optional: 0,
};

const REQUIRED_GROUP_ORDER: GroupKey[] = ["0_6m", "6_12m", "12_24m", "2_6y"];

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

const ageDays = computed(() => {
  if (!birthDate.value) return 0;
  const start = Date.parse(birthDate.value);
  if (!Number.isFinite(start)) return 0;
  return Math.max(0, Math.floor((Date.now() - start) / (24 * 3600 * 1000)));
});

const ageLabel = computed(() => {
  if (!birthDate.value) return "";
  const days = ageDays.value;
  const months = Math.floor(days / 30);
  if (months < 1) return `出生第 ${days} 天`;
  const restDays = days - months * 30;
  return restDays > 0 ? `${months} 个月 ${restDays} 天` : `${months} 个月`;
});

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

const doneCount = computed(() => planRows.value.filter((r) => r.status === "done").length);

/** 当前月龄落在哪个分组 */
const currentGroupKey = computed<GroupKey>(() => {
  const days = ageDays.value;
  for (let i = REQUIRED_GROUP_ORDER.length - 1; i >= 0; i -= 1) {
    const key = REQUIRED_GROUP_ORDER[i];
    if (days >= GROUP_START_DAYS[key]) return key;
  }
  return "0_6m";
});

/** 排序：当前月龄 → 未来月龄 → 已过月龄（近的在前）→ 自费 */
const orderedGroupKeys = computed<GroupKey[]>(() => {
  const currentIndex = REQUIRED_GROUP_ORDER.indexOf(currentGroupKey.value);
  const future = REQUIRED_GROUP_ORDER.slice(currentIndex + 1);
  const past = REQUIRED_GROUP_ORDER.slice(0, currentIndex).reverse();
  return [currentGroupKey.value, ...future, ...past, "optional"];
});

const groupedPlan = computed(() => {
  return orderedGroupKeys.value
    .map((key) => {
      const rows = planRows.value.filter((r) => r.group === key);
      return {
        key,
        label: VACCINE_GROUP_LABELS[key],
        rows,
        done: rows.filter((r) => r.status === "done").length,
      };
    })
    .filter((g) => g.rows.length > 0)
    .filter((g) => !activeFilter.value || g.key === activeFilter.value);
});

const filterOptions = computed(() => {
  const opts = REQUIRED_GROUP_ORDER.map((key) => ({ key: key as GroupKey | "", label: VACCINE_GROUP_LABELS[key] }));
  const hasOptional = planRows.value.some((r) => r.group === "optional");
  if (hasOptional) opts.push({ key: "optional", label: VACCINE_GROUP_LABELS.optional });
  return [{ key: "" as GroupKey | "", label: "全部" }, ...opts];
});

function selectFilter(key: GroupKey | "") {
  if (activeFilter.value === key) {
    activeFilter.value = "";
    openKey.value = null;
    return;
  }
  activeFilter.value = key;
  // 选定月龄段时直接展开，省一次点击
  openKey.value = key === "" ? null : key;
}

function toggleGroup(key: GroupKey) {
  openKey.value = openKey.value === key ? null : key;
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

.head { margin-bottom: $cj-gap-md; padding-top: 4rpx; }
.head-title { display: block; font-size: 40rpx; color: $cj-ink; font-weight: $cj-fw-display; letter-spacing: -0.5rpx; }
.head-sub { display: block; margin-top: 8rpx; font-size: 23rpx; color: $cj-text-muted; }

.filter-row { display: flex; flex-wrap: wrap; gap: 10rpx; margin-bottom: $cj-gap-md; }
.filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 10rpx 22rpx;
  border-radius: $cj-radius-pill;
  border: 1rpx solid $cj-border-faint;
  background: $cj-surface;
}
.filter-chip.active { background: $cj-primary-soft; border-color: $cj-primary; }
.filter-chip-text { font-size: 23rpx; color: $cj-text-secondary; }
.filter-chip.active .filter-chip-text { color: $cj-primary-dark; font-weight: 500; }

.empty-card, .card { background: $cj-surface; border: 1rpx solid $cj-border-faint; border-radius: $cj-radius-xl; box-shadow: $cj-shadow-card; padding: 28rpx; }
.empty-title { display: block; color: $cj-ink; font-size: 28rpx; font-weight: $cj-fw-display; }
.empty-desc { display: block; margin-top: 10rpx; color: $cj-text-secondary; font-size: 25rpx; line-height: 1.65; }

.groups { display: flex; flex-direction: column; gap: 14rpx; }
.group {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-xs;
  overflow: hidden;
}
.group-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16rpx;
  padding: 22rpx 24rpx;
  &:active { background: rgba(201, 107, 92, 0.04); }
}
.group-titles { min-width: 0; }
.group-title-row { display: flex; align-items: center; gap: 10rpx; }
.group-title { font-size: 28rpx; font-weight: 600; color: $cj-ink; }
.current-tag {
  padding: 2rpx 12rpx;
  border-radius: $cj-radius-pill;
  font-size: 18rpx;
  color: $cj-primary-dark;
  background: $cj-primary-soft;
}
.group-meta { display: block; margin-top: 4rpx; font-size: 21rpx; color: $cj-text-muted; }
.group-chev { font-size: 26rpx; color: $cj-primary; flex-shrink: 0; }

.timeline { display: flex; flex-direction: column; gap: 10rpx; padding: 0 16rpx 16rpx; }
.timeline-row { display: flex; gap: 14rpx; }
.rail { width: 24rpx; display: flex; flex-direction: column; align-items: center; padding-top: 26rpx; flex-shrink: 0; }
.dot { width: 14rpx; height: 14rpx; border-radius: 50%; border: 3rpx solid $cj-surface; box-shadow: 0 0 0 2rpx $cj-border-faint; }
.dot--done { background: $cj-mint; }
.dot--overdue { background: $cj-primary; }
.dot--pending { background: $cj-accent; }
.dot--optional.dot--pending { background: #c4b5a5; }
.line { flex: 1; width: 2rpx; min-height: 36rpx; margin-top: 8rpx; background: linear-gradient(180deg, $cj-border-faint 0%, transparent 100%); }

.card { flex: 1; margin-bottom: $cj-gap-sm; box-shadow: none; border-radius: $cj-radius-md; padding: 22rpx; }
.card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: $cj-gap-sm; }
.title-wrap { min-width: 0; }
.vaccine-name { display: block; font-size: 27rpx; color: $cj-ink; font-weight: 600; }
.dose, .date, .desc { display: block; margin-top: 6rpx; font-size: 22rpx; color: $cj-text-secondary; line-height: 1.5; }
.chip-col { display: flex; flex-direction: column; align-items: flex-end; gap: 8rpx; flex-shrink: 0; }
.opt-chip { padding: 4rpx 12rpx; border-radius: $cj-radius-pill; font-size: 18rpx; background: $cj-surface-2; color: $cj-text-muted; }
.status-chip { flex-shrink: 0; padding: 5rpx 14rpx; border-radius: $cj-radius-pill; font-size: 19rpx; font-weight: 500; }
.status-chip--done { background: $cj-mint-soft; color: $cj-tag-postnatal-text; }
.status-chip--overdue { background: $cj-danger-bg; color: $cj-danger-text; }
.status-chip--pending { background: $cj-warn-bg; color: $cj-text-secondary; }

.actions { margin-top: $cj-gap-sm; display: flex; justify-content: flex-end; }
.main-btn, .ghost-btn { border-radius: $cj-radius-pill !important; }
.main-btn { margin-top: $cj-gap-md; background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important; color: #fffefb !important; border: none !important; box-shadow: $cj-shadow-soft; }
.main-btn--mini,
.ghost-btn {
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  width: auto !important;
  min-width: 0;
  margin: 0 !important;
  padding: 0 32rpx !important;
  height: 60rpx;
  line-height: 60rpx !important;
  font-size: 23rpx !important;
}
.main-btn--mini {
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
}
.ghost-btn {
  background: $cj-surface-2 !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border-faint !important;
}
</style>
