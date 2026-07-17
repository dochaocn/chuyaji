<template>
  <view class="page">

    
    <view v-if="!session.privacyOk" class="notice">
      <text class="notice-title">先阅读隐私说明</text>
      <text class="notice-desc">记录只保存在你的账号范围内，进入正式记录前先确认使用说明。</text>
      <button size="mini" class="notice-btn" @click="goPrivacy">去阅读</button>
    </view>

    <view v-if="!auth.token && auth.sessionBooting" class="empty-card">
      <text class="empty-title">正在准备运行环境…</text>
      <text class="empty-desc">请稍候，完成环境与内容加载后即可使用。</text>
    </view>

    <view v-else-if="!auth.token && auth.sessionBootError" class="empty-card">
      <text class="empty-title">{{ auth.sessionBootError }}</text>
      <text class="empty-desc">请检查网络设置后重试。</text>
      <button class="main-btn" @click="retryWeChatSession">重试</button>
    </view>

    <template v-else-if="auth.token">

      <WorkbenchHeader
        workspace-label="宝宝工作台"
        :badge-label="dashboard?.profile ? babyStageLabel : undefined"
        :badge-class="babyStageBadgeClass"
        :name="dashboard?.profile?.nickname"
        :stage-title="stageTitle"
        :stage-desc="stageDesc"
        :profile-label="dashboard?.profile ? (session.canWrite ? '编辑档案' : '查看档案') : '创建档案'"
        profile-icon="👶"
        :show-next-step="nextStep.type !== 'none'"
        :next-step-title="nextStepTitle"
        :next-step-btn-label="nextStepBtnLabel"
        :can-switch="babies.length > 1"
        @profile="goProfileEdit"
        @family="goFamily"
        @next="onNextStepAction"
        @switch="onSwitchBaby"
      />

      <QuickRecordStrip
        v-if="dashboard?.profile && shortcuts.length"
        :shortcuts="shortcuts"
        @shortcut="onShortcut"
      />

      <view v-if="isPostnatalStage && dashboard?.profile" class="daily-section trend-panel">
        <view class="panel-head panel-head--inline">
          <view class="panel-head-text">
            <text class="panel-kicker">今日</text>
            <text class="panel-title">{{ dailyDateLabel }}</text>
          </view>
          <view class="panel-head-rule" />
        </view>
        <view class="summary-grid summary-grid--baby">
          <view class="summary-card summary-card--baby summary-card--weight" @click="goDailyRecords('feeding')">
            <view class="summary-card-accent summary-card-accent--coral" />
            <text class="summary-label">喂养</text>
            <text class="summary-value">{{ dailyFeedingLabel }}</text>
            <text class="summary-hint">{{ dailySummary?.feeding?.count ?? 0 }} 次</text>
          </view>
          <view class="summary-card summary-card--baby summary-card--count" @click="goDailyRecords('sleep')">
            <view class="summary-card-accent summary-card-accent--mint" />
            <text class="summary-label">睡眠</text>
            <text class="summary-value">{{ dailySleepLabel }}</text>
            <text class="summary-hint">{{ dailySummary?.sleep?.count ?? 0 }} 次</text>
          </view>
          <view class="summary-card summary-card--baby" @click="goDailyRecords('diaper')">
            <view class="summary-card-accent summary-card-accent--gold" />
            <text class="summary-label">排便</text>
            <text class="summary-value">{{ dailyDiaperLabel }}</text>
            <text class="summary-hint">{{ dailySummary?.diaper?.count ?? 0 }} 条</text>
          </view>
        </view>
      </view>

      <view v-if="isPostnatalStage && dashboard?.profile" class="trend-section trend-panel">
        <view class="panel-head panel-head--inline">
          <view class="panel-head-text">
            <text class="panel-kicker">成长快照</text>
            <text class="panel-title">近况</text>
          </view>
          <view class="panel-head-rule" />
        </view>
        <view v-if="hasTrendData" class="summary-grid summary-grid--baby">
          <view v-if="latestWeightLabel" class="summary-card summary-card--baby summary-card--weight">
            <view class="summary-card-accent summary-card-accent--coral" />
            <text class="summary-label">最近体重</text>
            <text class="summary-value">{{ latestWeightLabel }}</text>
            <text v-if="weightAgeDays !== null" class="summary-hint">{{ weightAgeDays }} 天前</text>
          </view>
          <view class="summary-card summary-card--baby summary-card--count">
            <view class="summary-card-accent summary-card-accent--mint" />
            <text class="summary-label">本月记录</text>
            <text class="summary-value">{{ monthlyCount }}</text>
            <text class="summary-hint">条</text>
          </view>
          <view v-if="latestMilestone" class="summary-card summary-card--baby summary-card--milestone">
            <view class="summary-card-accent summary-card-accent--gold" />
            <text class="summary-label">最近里程碑</text>
            <text class="summary-value milestone-value">{{ latestMilestone }}</text>
          </view>
        </view>
        <view class="growth-link-list">
          <view class="growth-link" @click="goGrowth">
            <text class="growth-link-text">查看生长趋势</text>
            <text class="growth-link-chev">›</text>
          </view>
          <view class="growth-link" @click="goVaccinePlan">
            <text class="growth-link-text">查看疫苗计划</text>
            <text class="growth-link-chev">›</text>
          </view>
          <view class="growth-link" @click="goAlbum">
            <text class="growth-link-text">成长相册</text>
            <text class="growth-link-chev">›</text>
          </view>
        </view>
      </view>

      
      <view v-if="dashboard?.profile" class="section records-panel">
        <view class="panel-head panel-head--inline">
          <view class="panel-head-text">
            <text class="panel-kicker">时间线</text>
            <text class="panel-title">最近记录</text>
          </view>
          <view class="panel-head-rule" />
          <view v-if="(dashboard?.latest_records?.length || 0) > 0" class="section-meta">
            <text class="section-meta-hint" @click="goTimeline">查看全部</text>
          </view>
          <view v-else-if="!loading" class="section-meta">
            <text class="section-meta-hint">在快速记录里新增一条</text>
          </view>
        </view>
        <view v-if="loading" class="placeholder">加载中…</view>
        <view v-else-if="!dashboard?.latest_records?.length" class="placeholder placeholder--soft">
          还没有记录，从推荐模板开始。
        </view>
        <view v-else class="record-timeline">
          <view
            v-for="(row, idx) in recentBabyRecordsWithPreview"
            :key="row.item.id"
            class="record-timeline-row"
            @click="openRecord(row.item.id)"
          >
            <view class="record-rail">
              <view
                :class="[
                  'record-dot',
                  row.item.phase === 'prenatal' ? 'record-dot--pre' : 'record-dot--post',
                ]"
              />
              <view
                v-if="idx < recentBabyRecordsWithPreview.length - 1"
                class="record-line"
                :class="row.item.phase === 'prenatal' ? 'record-line--pre' : 'record-line--post'"
              />
            </view>
            <view class="record-card record-card--timeline">
              <view class="record-card-head">
                <view class="record-card-head-row">
                  <view class="record-type-row">
                    <text class="record-type">{{ labelForType(row.item.phase, row.item.record_type) }}</text>
                    <text :class="['record-tag', row.item.phase === 'prenatal' ? 'pre' : 'post']">
                      {{ row.item.phase === "prenatal" ? "怀孕期" : "成长期" }}
                    </text>
                  </view>
                  <text class="record-date">{{ row.item.occurred_at.slice(0, 10) }}</text>
                </view>
              </view>
              <view v-if="row.keyRows.length" class="record-body">
                <view v-for="(kv, ki) in row.keyRows" :key="ki" class="record-kv-row">
                  <view class="record-key-label">{{ kv.label }}</view>
                  <view class="record-key-value">{{ kv.value }}</view>
                </view>
              </view>
              <text v-else class="record-summary">{{ row.item.summary?.trim() || "未填写内容" }}</text>
              <text v-if="row.showSummaryNote" class="record-summary-note">{{ row.item.summary }}</text>
            </view>
          </view>
        </view>
      </view>

    </template>

    
    <QuickRecordSheet
      v-if="quickTemplate"
      :visible="showQuickSheet"
      :template="quickTemplate"
      :baby-id="Number(session.babyId) || Number(dashboard?.profile?.id) || 0"
      :last-payload="lastQuickPayload"
      :last-summary="lastQuickSummary"
      @close="showQuickSheet = false"
      @saved="onQuickSaved"
    />

  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import {
  apiBabyDashboard,
  apiCreateRecord,
  apiLatestRecord,
  apiListBabies,
  apiListReminders,
  type Baby,
  type BabyDailySummary,
  type RecordItem,
  type ReminderItem,
} from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import {
  labelForType,
  BABY_TEMPLATES,
  BABY_HOME_PRENATAL_SHORTCUT_ORDER,
  babyHomePostnatalShortcutOrder,
  getBabyRecordPreviewRows,
  templateForType,
  type RecordTemplate,
} from "@/utils/recordTypes";
import { calcGestation, calcAgeDays, calcAgeMonths, isPostnatal } from "@/utils/gestation";
import { getBabyNextStep } from "@/utils/homeRules";
import QuickRecordSheet from "@/components/QuickRecordSheet.vue";
import WorkbenchHeader from "@/components/WorkbenchHeader.vue";
import QuickRecordStrip from "@/components/QuickRecordStrip.vue";

const auth = useAuthStore();
const session = useSessionStore();
const loading = ref(false);

const dashboard = ref<{
  profile: {
    id: number;
    nickname: string;
    birth_date?: string;
    edd_date?: string;
    lmp_date?: string;
  } | null;
  phase_summary?: { stage: "prenatal" | "postnatal"; record_count: number };
  latest_records: RecordItem[];
  growth_summary?: Record<string, unknown>;
  daily_summary?: BabyDailySummary;
} | null>(null);

const babies = ref<Baby[]>([]);

const showQuickSheet = ref(false);
const quickTemplate = ref<RecordTemplate | null>(null);
const lastQuickPayload = ref<Record<string, unknown> | null>(null);
const lastQuickSummary = ref("");
const reminders = ref<ReminderItem[]>([]);

const isPostnatalStage = computed(() => {
  return isPostnatal(dashboard.value?.profile?.birth_date);
});

/** 与「最近记录」阶段标签一致，样式对齐宝妈工作台时期徽章 */
const babyStageLabel = computed(() => (isPostnatalStage.value ? "成长期" : "怀孕期"));
const babyStageBadgeClass = computed(() => (isPostnatalStage.value ? "badge--post" : "badge--pre"));

const ageDays = computed(() => calcAgeDays(dashboard.value?.profile?.birth_date));
const ageMonths = computed(() => calcAgeMonths(dashboard.value?.profile?.birth_date));

const gestationResult = computed(() => {
  const p = dashboard.value?.profile;
  return calcGestation(p?.edd_date, p?.lmp_date);
});

const stageTitle = computed(() => {
  if (!dashboard.value?.profile) return "检查与成长，一站记下。";
  if (isPostnatalStage.value) {
    return `宝宝出生第 ${ageDays.value ?? 0} 天`;
  }
  if (gestationResult.value.available) {
    return `孕 ${gestationResult.value.weeks} 周 ${gestationResult.value.days} 天`;
  }
  return "检查与成长，一站记下。";
});

const stageDesc = computed(() => {
  if (!dashboard.value?.profile) return "先创建一个宝宝档案，后续所有孕期和成长期记录都围绕它展开。";
  if (isPostnatalStage.value) {
    return "喂养和生长变化都值得记一记。";
  }
  return "继续把每次检查安心记下。";
});

const shortcuts = computed<RecordTemplate[]>(() => {
  if (!dashboard.value?.profile) return [];
  if (isPostnatalStage.value) {
    const types = babyHomePostnatalShortcutOrder(ageMonths.value ?? 0);
    return types
      .map((t) => BABY_TEMPLATES.postnatal.find((tmpl) => tmpl.value === t))
      .filter((t): t is RecordTemplate => !!t);
  }
  return BABY_HOME_PRENATAL_SHORTCUT_ORDER.map((t) => BABY_TEMPLATES.prenatal.find((tmpl) => tmpl.value === t)).filter(
    (t): t is RecordTemplate => !!t
  );
});

const recentBabyRecordsWithPreview = computed(() => {
  const list = dashboard.value?.latest_records?.slice(0, 3) ?? [];
  return list.map((item) => {
    const keyRows = getBabyRecordPreviewRows({
      phase: item.phase,
      record_type: item.record_type,
      payload: item.payload as Record<string, unknown>,
    });
    const tmpl = templateForType(item.phase, item.record_type);
    const sum = item.summary?.trim() || "";
    const showSummaryNote = tmpl?.mode === "standard" && !!sum && keyRows.length > 0;
    return { item, keyRows, showSummaryNote };
  });
});

const activeReminder = computed(() => reminders.value[0] ?? null);

const nextStep = computed(() => {
  if (activeReminder.value) return { type: "check_reminder" as const };
  const hasProfile = !!dashboard.value?.profile;
  const lastAt = dashboard.value?.latest_records?.[0]?.occurred_at ?? null;
  return getBabyNextStep(hasProfile, lastAt);
});

const nextStepTitle = computed(() => {
  if (activeReminder.value) {
    const more = reminders.value.length > 1 ? `，还有 ${reminders.value.length - 1} 条` : "";
    return `${activeReminder.value.title}：${activeReminder.value.due_at.slice(0, 10)}${more}`;
  }
  switch (nextStep.value.type) {
    case "create_profile": return "还没有宝宝档案";
    case "add_record": return `已经 ${nextStep.value.daysSinceLast} 天没有新记录了`;
    case "check_reminder": return "有即将到来的提醒";
    default: return "";
  }
});

const nextStepBtnLabel = computed(() => {
  switch (nextStep.value.type) {
    case "create_profile": return "去建档";
    case "add_record": return "补一条记录";
    case "check_reminder": return "查看提醒";
    default: return "";
  }
});

const latestWeightLabel = computed(() => {
  const value = dashboard.value?.growth_summary?.latest_weight_g;
  if (typeof value !== "number") return null;
  return `${(value / 1000).toFixed(2)} kg`;
});

const weightAgeDays = computed(() => {
  const at = dashboard.value?.growth_summary?.latest_weight_at as string | undefined;
  if (!at) return null;
  const diff = Math.floor((Date.now() - Date.parse(at)) / (24 * 60 * 60 * 1000));
  return diff >= 0 ? diff : null;
});

const monthlyCount = computed(() => {
  const records = dashboard.value?.latest_records ?? [];
  const now = new Date();
  return records.filter((r) => {
    const d = new Date(r.occurred_at);
    return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
  }).length;
});

const latestMilestone = computed(() => {
  const records = dashboard.value?.latest_records ?? [];
  const hit = records.find((r) => r.record_type === "development");
  if (!hit) return null;
  return (hit.payload as Record<string, unknown>)?.milestone as string ?? hit.summary ?? "已记录";
});

const hasTrendData = computed(() => {
  return !!latestWeightLabel.value || monthlyCount.value > 0;
});

const dailySummary = computed(() => dashboard.value?.daily_summary);

const dailyDateLabel = computed(() => {
  const d = dailySummary.value?.date;
  if (!d) return "汇总";
  const parts = d.split("-");
  if (parts.length !== 3) return d;
  return `${Number(parts[1])} 月 ${Number(parts[2])} 日`;
});

const dailyFeedingLabel = computed(() => {
  const ml = dailySummary.value?.feeding?.total_ml ?? 0;
  return ml > 0 ? `${ml} ml` : "—";
});

const dailySleepLabel = computed(() => {
  const min = dailySummary.value?.sleep?.total_duration_min ?? 0;
  if (min <= 0) return "—";
  if (min < 60) return `${min} 分`;
  const h = Math.floor(min / 60);
  const m = min % 60;
  return m > 0 ? `${h} 时 ${m} 分` : `${h} 时`;
});

const dailyDiaperLabel = computed(() => {
  const times = dailySummary.value?.diaper?.total_times ?? 0;
  return times > 0 ? `${times} 次` : "—";
});

onShow(async () => {
  auth.loadToken();
  session.load();
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
  if (auth.token) {
    await session.refreshFamilyContext();
    loadDashboard();
  }
});

async function retryWeChatSession() {
  await auth.ensureWeChatSession();
  if (auth.token) {
    await loadDashboard();
  }
}

async function loadDashboard() {
  loading.value = true;
  try {
    const [data, babyPage] = await Promise.all([
      apiBabyDashboard(session.babyId || undefined),
      apiListBabies().catch(() => ({ items: [] as Baby[] })),
    ]);
    dashboard.value = data;
    babies.value = babyPage.items || [];
    if (data.profile?.id) {
      session.setBaby(data.profile.id);
      try {
        const reminderPage = await apiListReminders("pending", 20, { owner_type: "baby", owner_id: data.profile.id });
        reminders.value = reminderPage.items;
      } catch (error) {
        console.error(error);
        reminders.value = [];
      }
    } else {
      reminders.value = [];
    }
  } catch (error) {
    console.error(error);
    uni.showToast({ title: "加载失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

function discardTimerIfBabyMismatch(nextBabyId: number) {
  try {
    const raw = uni.getStorageSync("chuyaji_timer");
    if (!raw) return;
    const stored = typeof raw === "string" ? JSON.parse(raw) : raw;
    if (stored?.babyId && Number(stored.babyId) !== nextBabyId) {
      uni.removeStorageSync("chuyaji_timer");
    }
  } catch {}
}

function onSwitchBaby() {
  if (babies.value.length <= 1) return;
  const itemList = babies.value.map((b) =>
    b.id === session.babyId ? `${b.nickname}（当前）` : b.nickname
  );
  uni.showActionSheet({
    itemList,
    success: (res) => {
      const picked = babies.value[res.tapIndex];
      if (!picked || picked.id === session.babyId) return;
      discardTimerIfBabyMismatch(picked.id);
      session.setBaby(picked.id);
      void loadDashboard();
    },
  });
}

function goDailyRecords(recordType: "feeding" | "sleep" | "diaper") {
  if (!session.babyId) return;
  const date = dailySummary.value?.date || formatLocalDate(new Date());
  // ListRecords 的 to=YYYY-MM-DD 会按「次日零点」排他，故 from/to 传同一天即可筛当日
  uni.navigateTo({
    url: `/pages/baby/record-list?baby_id=${session.babyId}&phase=postnatal&record_type=${recordType}&from=${date}&to=${date}`,
  });
}

function formatLocalDate(d: Date) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

function onNextStepAction() {
  if (activeReminder.value) return goReminders();
  const step = nextStep.value;
  if (step.type === "create_profile") return goProfileEdit();
  if (step.type === "add_record") return goNewRecord();
}

async function onShortcut(tmpl: RecordTemplate) {
  if (tmpl.mode === "quick") {
    quickTemplate.value = tmpl;
    lastQuickPayload.value = null;
    lastQuickSummary.value = "";
    const bid = Number(session.babyId) || Number(dashboard.value?.profile?.id) || 0;
    const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
    if (bid) {
      try {
        const latest = await apiLatestRecord(bid, tmpl.value, phase);
        lastQuickPayload.value = (latest.item?.payload || null) as Record<string, unknown> | null;
        lastQuickSummary.value = latest.item?.summary || "";
      } catch (e) {
        console.error(e);
      }
    }
    showQuickSheet.value = true;
  } else {
    const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
    if (!session.babyId) {
      uni.showToast({ title: "请先创建宝宝档案", icon: "none" });
      return;
    }
    uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${session.babyId}&phase=${phase}&type=${tmpl.value}` });
  }
}

async function onQuickSaved(payload: Record<string, unknown>, summary: string) {
  const bid = Number(session.babyId) || Number(dashboard.value?.profile?.id) || 0;
  if (!bid) {
    uni.showToast({ title: "缺少宝宝档案", icon: "none" });
    showQuickSheet.value = false;
    return;
  }
  if (bid !== Number(session.babyId)) {
    session.setBaby(bid);
  }
  const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
  try {
    await apiCreateRecord(bid, {
      phase,
      record_type: quickTemplate.value!.value,
      occurred_at: new Date().toISOString(),
      summary,
      payload,
    });
    showQuickSheet.value = false;
    uni.showToast({ title: "已保存", icon: "success" });
    await loadDashboard();
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "保存失败", icon: "none" });
  }
}

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
}

function goFamily() {
  uni.navigateTo({ url: "/pages/family/index" });
}

function goProfileEdit() {
  const id = dashboard.value?.profile?.id || session.babyId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/baby/profile-edit${suffix}` });
}

function goNewRecord() {
  if (!session.babyId) {
    uni.showToast({ title: "请先创建宝宝档案", icon: "none" });
    return;
  }
  const phase = isPostnatalStage.value ? "postnatal" : "prenatal";
  uni.navigateTo({ url: `/pages/baby/record-edit?baby_id=${session.babyId}&phase=${phase}` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/baby/record-edit?id=${id}&baby_id=${session.babyId}` });
}

function goGrowth() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/growth?baby_id=${session.babyId}` });
}

function goVaccinePlan() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/vaccine-plan?baby_id=${session.babyId}` });
}

function goAlbum() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/album?baby_id=${session.babyId}` });
}

function goTimeline() {
  if (!session.babyId) return;
  uni.navigateTo({ url: `/pages/baby/record-list?baby_id=${session.babyId}` });
}

function goReminders() {
  uni.navigateTo({ url: "/pages/reminders/list" });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 80rpx;
}

.notice,
.empty-card {
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.notice {
  background: linear-gradient(135deg, $cj-warn-bg 0%, #fff8f0 100%);
  border-color: $cj-warn-border;
}

.notice-title,
.empty-title {
  display: block;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  font-size: 30rpx;
  letter-spacing: 0.5rpx;
}

.notice-desc,
.empty-desc {
  display: block;
  margin-top: 10rpx;
  color: $cj-text-secondary;
  font-size: 25rpx;
  line-height: 1.65;
}

.notice-btn,
.main-btn {
  margin-top: $cj-gap-md;
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  box-shadow: $cj-shadow-soft;
}

.trend-section,
.daily-section,
.section {
  margin-bottom: $cj-gap-md;
}

.daily-section .summary-card:active {
  opacity: 0.88;
}

.trend-panel {
  position: relative;
  padding: 28rpx $cj-gap-md calc($cj-gap-md + 4rpx);
  background: linear-gradient(155deg, rgba(255, 253, 249, 0.97) 0%, rgba(255, 247, 240, 0.95) 60%, rgba(230, 241, 236, 0.3) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  overflow: hidden;

  &::before {
    content: "";
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
      radial-gradient(ellipse 110% 70% at 100% 0%, rgba(232, 184, 150, 0.18) 0%, transparent 50%),
      radial-gradient(ellipse 80% 50% at 0% 100%, rgba(125, 171, 152, 0.12) 0%, transparent 45%);
    opacity: 0.85;
  }
}

.panel-head {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: flex-end;
  gap: $cj-gap-sm;
  margin-bottom: $cj-gap-md;
}

.panel-head--inline {
  align-items: center;
}

.panel-head-text {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.panel-kicker {
  font-size: 19rpx;
  letter-spacing: 6rpx;
  text-transform: uppercase;
  color: $cj-text-muted;
  font-weight: 500;
}

.panel-title {
  font-size: 32rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 0.5rpx;
}

.panel-head-rule {
  flex: 1;
  height: 1rpx;
  background: linear-gradient(90deg, $cj-border-faint 0%, transparent 100%);
  margin-bottom: 10rpx;
}

.panel-head--inline .panel-head-rule {
  margin-bottom: 0;
}

.summary-grid {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-sm;
}

.summary-grid--baby {
  position: relative;
  z-index: 1;
}

.summary-card {
  flex: 1;
  min-width: 180rpx;
  padding: 28rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
}

.summary-card--baby {
  position: relative;
  overflow: hidden;
  border: 1rpx solid $cj-border-faint;
  background: rgba(255, 253, 249, 0.94);
}

.summary-card-accent {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5rpx;
  border-radius: $cj-radius-xl 0 0 $cj-radius-xl;
}

.summary-card-accent--coral {
  background: linear-gradient(180deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
}

.summary-card-accent--mint {
  background: linear-gradient(180deg, $cj-mint 0%, #5d9482 100%);
}

.summary-card-accent--gold {
  background: linear-gradient(180deg, $cj-accent-warm 0%, $cj-accent 100%);
}

.summary-card--milestone {
  flex: 1 1 100%;
  min-width: 100%;
}

.summary-label {
  display: block;
  color: $cj-text-muted;
  font-size: 21rpx;
  font-weight: 500;
  letter-spacing: 0.5rpx;
}

.summary-value {
  display: block;
  margin-top: 10rpx;
  color: $cj-ink;
  font-size: 32rpx;
  font-weight: $cj-fw-display;
  letter-spacing: -0.3rpx;
}

.milestone-value {
  font-size: 24rpx;
  letter-spacing: 0;
}

.summary-hint {
  display: block;
  margin-top: 6rpx;
  font-size: 19rpx;
  color: $cj-text-muted;
}

.growth-link-list {
  position: relative;
  z-index: 1;
  margin-top: $cj-gap-md;
  display: flex;
  flex-direction: column;
  gap: 6rpx;
}

.growth-link {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6rpx;
  padding: 10rpx 0;
}

.growth-link-text {
  font-size: 23rpx;
  color: $cj-primary-dark;
  font-weight: 500;
  letter-spacing: 0.3rpx;
}

.growth-link-chev {
  font-size: 28rpx;
  color: $cj-primary;
  line-height: 1;
  opacity: 0.75;
}

.records-panel {
  position: relative;
  background: linear-gradient(180deg, $cj-surface 0%, rgba(255, 253, 249, 0.98) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx $cj-gap-md $cj-gap-lg;
  overflow: hidden;

  &::after {
    content: "";
    position: absolute;
    right: -40rpx;
    top: -48rpx;
    width: 180rpx;
    height: 180rpx;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(201, 107, 92, 0.06) 0%, transparent 70%);
    pointer-events: none;
  }
}

.records-panel .panel-head {
  margin-bottom: $cj-gap-md;
}

.records-panel .section-meta {
  flex-shrink: 0;
}

.section-meta-hint {
  font-size: 21rpx;
  color: $cj-primary;
  padding: 7rpx 18rpx;
  border-radius: $cj-radius-pill;
  background: rgba(201, 107, 92, 0.07);
  font-weight: 500;
  letter-spacing: 0.3rpx;
}

.placeholder {
  position: relative;
  z-index: 1;
  padding: 40rpx 0;
  text-align: center;
  color: $cj-text-muted;
  font-size: 24rpx;
}

.placeholder--soft {
  padding: 52rpx 24rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-lg;
  border: 1rpx dashed $cj-border-faint;
}

.record-timeline {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 10rpx;
}

.record-timeline-row {
  display: flex;
  flex-direction: row;
  align-items: stretch;
  gap: 18rpx;
}

.record-rail {
  flex-shrink: 0;
  width: 28rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 28rpx;
}

.record-dot {
  width: 16rpx;
  height: 16rpx;
  border-radius: 50%;
  border: 3rpx solid $cj-surface;
  box-shadow: 0 0 0 2rpx $cj-border-faint;
}

.record-dot--pre {
  background: linear-gradient(145deg, #e8c4d4 0%, #b87a92 100%);
}

.record-dot--post {
  background: linear-gradient(145deg, #b8e0d0 0%, #5a9d82 100%);
}

.record-line {
  flex: 1;
  width: 2rpx;
  min-height: 32rpx;
  margin-top: 8rpx;
  border-radius: 2rpx;
  background: linear-gradient(180deg, $cj-border-faint 0%, transparent 100%);
}

.record-line--pre {
  background: linear-gradient(180deg, rgba(184, 122, 146, 0.3) 0%, transparent 100%);
}

.record-line--post {
  background: linear-gradient(180deg, rgba(90, 157, 130, 0.3) 0%, transparent 100%);
}

.record-card {
  flex: 1;
  min-width: 0;
  padding: 24rpx 26rpx;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-faint;
  box-shadow: $cj-shadow-xs;
}

.record-card--timeline {
  background: linear-gradient(160deg, rgba(255, 253, 249, 0.98) 0%, rgba(255, 249, 242, 0.95) 100%);
  border: 1rpx solid $cj-border-faint;
  box-shadow: $cj-shadow-xs;

  &:active {
    opacity: 0.92;
  }
}

.record-card-head {
  margin-bottom: 2rpx;
}

.record-card-head-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16rpx;
}

.record-type-row {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: row;
  align-items: center;
  flex-wrap: nowrap;
  gap: 10rpx;
}

.record-type-row .record-type {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.record-type-row .record-tag {
  flex-shrink: 0;
}

.record-tag {
  padding: 5rpx 14rpx;
  border-radius: $cj-radius-pill;
  font-size: 19rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 0.3rpx;
}

.record-tag.pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.record-tag.post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.record-type {
  flex: 1;
  min-width: 0;
  font-size: 29rpx;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  line-height: 1.4;
  letter-spacing: 0.2rpx;
}

.record-date {
  flex-shrink: 0;
  padding-top: 4rpx;
  font-size: 21rpx;
  color: $cj-text-muted;
  letter-spacing: 0.5rpx;
}

.record-body {
  margin-top: 14rpx;
  padding: 4rpx 16rpx 2rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid $cj-border-faint;
}

.record-kv-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: flex-start;
  gap: 14rpx;
  padding: 14rpx 0;
  border-bottom: 1rpx solid $cj-border-faint;
}

.record-kv-row:last-child {
  border-bottom: none;
  padding-bottom: 10rpx;
}

.record-kv-row:first-child {
  padding-top: 10rpx;
}

.record-key-label {
  flex-shrink: 0;
  width: 148rpx;
}

.record-key-value {
  flex: 1;
  min-width: 0;
}

.record-summary {
  display: block;
  margin-top: 12rpx;
  padding: 14rpx 16rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid $cj-border-faint;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.record-summary-note {
  display: block;
  margin-top: 10rpx;
  padding-top: 12rpx;
  border-top: 1rpx solid $cj-border-faint;
  font-size: 22rpx;
  color: $cj-text-muted;
  line-height: 1.55;
}

@import "@/styles/cj-record-kv-fields.scss";
</style>
