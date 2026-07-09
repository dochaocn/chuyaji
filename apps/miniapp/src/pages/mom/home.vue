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
      <button class="main-btn" @tap="retryWeChatSession">重试</button>
    </view>

    <template v-else-if="auth.token">

      
      <view class="hero">
        <view class="hero-top">
          <text class="hero-kicker">宝妈工作台</text>
          <text v-if="dashboard?.profile" :class="['stage-badge', stageBadgeClass]">{{ statusLabel }}</text>
          <text v-if="dashboard?.profile" class="hero-edit-link" @click="goProfileEdit">编辑档案</text>
        </view>
        <view class="hero-headline">
          <text v-if="dashboard?.profile" class="hero-name">{{ dashboard.profile.name || "宝妈" }}</text>
          <text
            class="hero-title"
            :class="{ 'hero-title--solo': !dashboard?.profile }"
          >{{ stageTitle }}</text>
        </view>
        <text class="hero-desc">{{ stageDesc }}</text>
      </view>

      
      <view
        v-if="nextStep.type !== 'none' && !(nextStep.type === 'create_profile' && !dashboard?.profile)"
        class="next-card"
      >
        <text class="next-label">现在可以做</text>
        <text class="next-title">{{ nextStepTitle }}</text>
        <button class="next-btn" @click="onNextStepAction">{{ nextStepBtnLabel }}</button>
      </view>

      <view v-if="!dashboard?.profile" class="empty-card">
        <template v-if="loading">
          <text class="empty-title">加载中…</text>
          <text class="empty-desc">正在获取档案信息</text>
        </template>
        <template v-else>
          <text class="empty-title">还没有档案</text>
          <text class="empty-desc">可先创建宝宝或宝妈档案，便于分别记录成长与自身状态。</text>
          <view class="empty-actions">
            <button class="main-btn main-btn--flex" @tap="goBabyProfileEdit">创建宝宝档案</button>
            <button class="main-btn main-btn--flex" @tap="goProfileEdit">创建宝妈档案</button>
          </view>
        </template>
      </view>

      
      <view v-if="dashboard?.profile" class="shortcut-section">
        <text class="section-label">快速记录</text>
        <view class="shortcut-grid">
          <view
            v-for="item in shortcuts"
            :key="item.value"
            class="shortcut-tile"
            @click="onShortcut(item)"
          >
            <text class="shortcut-title">{{ item.entryLabel }}</text>
            <text class="shortcut-mode">{{ item.mode === 'quick' ? '一键' : '标准' }}</text>
          </view>
        </view>
      </view>

      
      <view v-if="dashboard?.profile && hasSummaryData" class="summary-section mom-snapshot-panel">
        <view class="panel-head panel-head--inline">
          <view class="panel-head-text">
            <text class="panel-kicker">身体与情绪</text>
            <text class="panel-title">近况</text>
          </view>
          <view class="panel-head-rule" />
        </view>
        <view class="summary-grid summary-grid--stack">
          <view
            class="summary-card summary-card--mom summary-card--metric"
            :class="{ 'summary-card--empty': !lastMetricRecord }"
            @click="lastMetricRecord?.id && openSummaryRecord(lastMetricRecord.id)"
          >
            <view class="summary-card-glow summary-card-glow--sage" />
            <view class="summary-card-top">
              <text class="summary-label">身体指标</text>
              <text v-if="lastMetricRecord && lastMetricDate" class="summary-hint-inline">{{ lastMetricDate }}</text>
            </view>
            <scroll-view
              v-if="lastMetricPreview.length"
              class="summary-key-scroll"
              scroll-x
              :show-scrollbar="false"
              enable-flex
            >
              <view class="summary-key-row">
                <view v-for="(kv, ki) in lastMetricPreview" :key="ki" class="summary-key-chip">
                  <view class="summary-key-label">{{ kv.label }}</view>
                  <view class="summary-key-value">{{ kv.value }}</view>
                </view>
              </view>
            </scroll-view>
            <text v-else-if="lastMetricFallback" class="summary-value summary-value--single">{{ lastMetricFallback }}</text>
            <text v-else-if="lastMetricRecord" class="summary-value summary-value--single">未填写内容</text>
            <text v-else class="summary-value summary-value--single">暂无</text>
          </view>
          <view
            class="summary-card summary-card--mom summary-card--checkup"
            :class="{ 'summary-card--empty': !lastCheckupRecord }"
            @click="lastCheckupRecord?.id && openSummaryRecord(lastCheckupRecord.id)"
          >
            <view class="summary-card-glow summary-card-glow--rose" />
            <view class="summary-card-top">
              <text class="summary-label">最近产检</text>
              <text v-if="lastCheckupRecord && lastCheckupDate" class="summary-hint-inline">{{ lastCheckupDate }}</text>
            </view>
            <scroll-view
              v-if="lastCheckupPreview.length"
              class="summary-key-scroll"
              scroll-x
              :show-scrollbar="false"
              enable-flex
            >
              <view class="summary-key-row">
                <view v-for="(kv, ki) in lastCheckupPreview" :key="ki" class="summary-key-chip">
                  <view class="summary-key-label">{{ kv.label }}</view>
                  <view class="summary-key-value">{{ kv.value }}</view>
                </view>
              </view>
            </scroll-view>
            <text v-else-if="lastCheckupFallback" class="summary-value summary-value--single">{{ lastCheckupFallback }}</text>
            <text v-else-if="lastCheckupRecord" class="summary-value summary-value--single">未填写内容</text>
            <text v-else class="summary-value summary-value--single">暂无</text>
          </view>
          <view
            class="summary-card summary-card--mom summary-card--symptom"
            :class="{ 'summary-card--empty': !lastSymptomRecord }"
            @click="lastSymptomRecord?.id && openSummaryRecord(lastSymptomRecord.id)"
          >
            <view class="summary-card-glow summary-card-glow--clay" />
            <view class="summary-card-top">
              <text class="summary-label">最近不适</text>
              <text v-if="lastSymptomRecord && lastSymptomDate" class="summary-hint-inline">{{ lastSymptomDate }}</text>
            </view>
            <scroll-view
              v-if="lastSymptomPreview.length"
              class="summary-key-scroll"
              scroll-x
              :show-scrollbar="false"
              enable-flex
            >
              <view class="summary-key-row">
                <view v-for="(kv, ki) in lastSymptomPreview" :key="ki" class="summary-key-chip">
                  <view class="summary-key-label">{{ kv.label }}</view>
                  <view class="summary-key-value">{{ kv.value }}</view>
                </view>
              </view>
            </scroll-view>
            <text v-else-if="lastSymptomFallback" class="summary-value summary-value--single">{{ lastSymptomFallback }}</text>
            <text v-else-if="lastSymptomRecord" class="summary-value summary-value--single">未填写内容</text>
            <text v-else class="summary-value summary-value--single">暂无</text>
          </view>
          <view
            class="summary-card summary-card--mom summary-card--mood"
            :class="{ 'summary-card--empty': !lastMoodRecord }"
            @click="lastMoodRecord?.id && openSummaryRecord(lastMoodRecord.id)"
          >
            <view class="summary-card-glow summary-card-glow--sun" />
            <view class="summary-card-top">
              <text class="summary-label">最近心情</text>
              <text v-if="lastMoodRecord && lastMoodDate" class="summary-hint-inline">{{ lastMoodDate }}</text>
            </view>
            <scroll-view
              v-if="lastMoodPreview.length"
              class="summary-key-scroll"
              scroll-x
              :show-scrollbar="false"
              enable-flex
            >
              <view class="summary-key-row">
                <view v-for="(kv, ki) in lastMoodPreview" :key="ki" class="summary-key-chip">
                  <view class="summary-key-label">{{ kv.label }}</view>
                  <view class="summary-key-value">{{ kv.value }}</view>
                </view>
              </view>
            </scroll-view>
            <text v-else-if="lastMoodFallback" class="summary-value summary-value--single">{{ lastMoodFallback }}</text>
            <text v-else-if="lastMoodRecord" class="summary-value summary-value--single">未填写内容</text>
            <text v-else class="summary-value summary-value--single">暂无</text>
          </view>
        </view>
      </view>

      
      <view v-if="dashboard?.profile" class="section records-panel">
        <view class="panel-head panel-head--inline">
          <view class="panel-head-text">
            <text class="panel-kicker">随记</text>
            <text class="panel-title">最近记录</text>
          </view>
          <view class="panel-head-rule" />
          <view v-if="(dashboard?.latest_records?.length || 0) > 0" class="section-meta">
            <text class="section-meta-hint" @click="goRecordList">查看全部</text>
          </view>
          <view v-else-if="!loading" class="section-meta">
            <text class="section-meta-hint">在快速记录里新增一条</text>
          </view>
        </view>
        <view v-if="loading" class="placeholder">加载中…</view>
        <view v-else-if="!dashboard?.latest_records?.length" class="placeholder placeholder--soft">
          还没有记录，可以从产检或身体状态开始。
        </view>
        <view v-else class="record-timeline">
          <view
            v-for="(row, idx) in recentMomRecordsWithPreview"
            :key="row.item.id"
            class="record-timeline-row"
            @click="openRecord(row.item.id)"
          >
            <view class="record-rail">
              <view class="record-dot record-dot--mom" />
              <view
                v-if="idx < recentMomRecordsWithPreview.length - 1"
                class="record-line record-line--mom"
              />
            </view>
            <view class="record-card record-card--timeline">
              <view class="record-card-head">
                <view class="record-card-head-row">
                  <view class="record-type-row">
                    <text class="record-type">{{ labelForMotherType(row.item.record_type) }}</text>
                    <text :class="['stage-badge', 'stage-badge--inline', stageBadgeClass]">{{ statusLabel }}</text>
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
      :last-payload="lastQuickPayload"
      :last-summary="lastQuickSummary"
      @close="showQuickSheet = false"
      @saved="onQuickSaved"
    />

    
    <view v-if="showMetricPicker" class="sheet-mask" @click.self="showMetricPicker = false">
      <view class="picker-sheet">
        <text class="picker-title">选择指标类型</text>
        <view
          v-for="mt in metricOptions"
          :key="mt.value"
          class="picker-item"
          @click="onMetricPick(mt.value)"
        >
          <text class="picker-item-text">{{ mt.label }}</text>
        </view>
      </view>
    </view>

  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed, ref } from "vue";
import { apiMotherDashboard, apiCreateMotherRecord, apiLatestMotherRecord, apiListReminders } from "@/api/chuyaji";
import type { MotherRecordItem, ReminderItem } from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import {
  labelForMotherType,
  MOTHER_TEMPLATES,
  BODY_METRIC_TYPES,
  getMotherRecordPreviewRows,
  templateForMotherType,
  type RecordTemplate,
} from "@/utils/motherRecordTypes";
import {
  getMotherNextStep,
  getPreferredBodyMetricType,
  motherStageBadgeClass,
  motherStageLabel,
} from "@/utils/homeRules";
import QuickRecordSheet from "@/components/QuickRecordSheet.vue";

const auth = useAuthStore();
const session = useSessionStore();
const loading = ref(false);

const dashboard = ref<{
  profile: {
    id: number;
    name: string;
    status?: "pregnant" | "postpartum" | "parenting";
    blood_type?: string;
    delivery_date?: string;
  } | null;
  health_summary?: { status?: string; record_count?: number };
  latest_records: MotherRecordItem[];
} | null>(null);

const showQuickSheet = ref(false);
const quickTemplate = ref<RecordTemplate | null>(null);
const lastQuickPayload = ref<Record<string, unknown> | null>(null);
const lastQuickSummary = ref("");
const showMetricPicker = ref(false);
const reminders = ref<ReminderItem[]>([]);

const statusLabel = computed(() => motherStageLabel(dashboard.value?.profile?.status));

const stageBadgeClass = computed(() => motherStageBadgeClass(dashboard.value?.profile?.status));

const stageTitle = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (!dashboard.value?.profile) return "身体、产检与恢复，随记。";
  if (status === "postpartum") return "产后恢复阶段";
  if (status === "parenting") return "育儿期";
  return "怀孕阶段";
});

const stageDesc = computed(() => {
  const status = dashboard.value?.profile?.status;
  if (!dashboard.value?.profile) return "可先创建宝宝或宝妈档案，再开始记录。";
  if (status === "postpartum") return "恢复节奏不必着急，把身体和心情慢慢记下来。";
  if (status === "parenting") return "育儿期的身体变化同样值得记录。";
  return "这一阶段最重要的是把检查、身体感受和情绪稳稳记下。";
});

const SHORTCUT_KEYS = ["checkup", "symptom", "medication", "mood", "__body_metric__"];

const shortcuts = computed(() => {
  return SHORTCUT_KEYS.map((k) => {
    if (k === "__body_metric__") {
      return {
        value: "__body_metric__",
        label: "身体指标",
        entryLabel: "身体指标",
        mode: "quick" as const,
        recommendedFields: [],
        optionalFields: [],
      } satisfies RecordTemplate;
    }
    return MOTHER_TEMPLATES.find((t) => t.value === k)!;
  }).filter(Boolean);
});

/** 始终列出三种指标；最近记录过的类型排在最前，便于一键延续，同时不会隐藏体重/血压 */
const metricOptions = computed(() => {
  const preferred = getPreferredBodyMetricType(dashboard.value?.latest_records ?? []);
  const rest = BODY_METRIC_TYPES.filter((v) => v !== preferred);
  const order = [preferred, ...rest];
  return order.map((v) => ({
    value: v,
    label: labelForMotherType(v),
  }));
});

/** 最近记录：优先展示模板关键字段行，摘要仅在无字段或标准记录补充说明时展示 */
const recentMomRecordsWithPreview = computed(() => {
  const list = dashboard.value?.latest_records?.slice(0, 3) ?? [];
  return list.map((item) => {
    const keyRows = getMotherRecordPreviewRows(item);
    const tmpl = templateForMotherType(item.record_type);
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
  return getMotherNextStep(hasProfile, lastAt);
});

const nextStepTitle = computed(() => {
  if (activeReminder.value) {
    const more = reminders.value.length > 1 ? `，还有 ${reminders.value.length - 1} 条` : "";
    return `${activeReminder.value.title}：${activeReminder.value.due_at.slice(0, 10)}${more}`;
  }
  switch (nextStep.value.type) {
    case "create_profile": return "还没有宝妈档案";
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

const SUMMARY_CARD_PREVIEW_MAX = 3;

function motherSummaryFallbackText(r: MotherRecordItem | null, previewLen: number): string | null {
  if (!r || previewLen > 0) return null;
  return r.summary?.trim() || null;
}

const lastCheckupRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "checkup") ?? null
);
const lastCheckupDate = computed(() => lastCheckupRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastCheckupPreview = computed(() => {
  const r = lastCheckupRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastCheckupFallback = computed(() =>
  motherSummaryFallbackText(lastCheckupRecord.value, lastCheckupPreview.value.length)
);

const lastMetricRecord = computed(() => {
  const metricTypes = ["weight", "blood_pressure", "blood_sugar"];
  return dashboard.value?.latest_records?.find((r) => metricTypes.includes(r.record_type)) ?? null;
});
const lastMetricDate = computed(() => lastMetricRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastMetricPreview = computed(() => {
  const r = lastMetricRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastMetricFallback = computed(() =>
  motherSummaryFallbackText(lastMetricRecord.value, lastMetricPreview.value.length)
);

const lastMoodRecord = computed(() => {
  return dashboard.value?.latest_records?.find((r) => r.record_type === "mood") ?? null;
});
const lastMoodDate = computed(() => lastMoodRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastMoodPreview = computed(() => {
  const r = lastMoodRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastMoodFallback = computed(() =>
  motherSummaryFallbackText(lastMoodRecord.value, lastMoodPreview.value.length)
);

const lastSymptomRecord = computed(() =>
  dashboard.value?.latest_records?.find((r) => r.record_type === "symptom") ?? null
);
const lastSymptomDate = computed(() => lastSymptomRecord.value?.occurred_at.slice(0, 10) ?? null);
const lastSymptomPreview = computed(() => {
  const r = lastSymptomRecord.value;
  if (!r) return [];
  return getMotherRecordPreviewRows(r).slice(0, SUMMARY_CARD_PREVIEW_MAX);
});
const lastSymptomFallback = computed(() =>
  motherSummaryFallbackText(lastSymptomRecord.value, lastSymptomPreview.value.length)
);

const hasSummaryData = computed(() =>
  !!(
    lastCheckupRecord.value ||
    lastMetricRecord.value ||
    lastMoodRecord.value ||
    lastSymptomRecord.value
  )
);

onShow(async () => {
  auth.loadToken();
  session.load();
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
  if (auth.token) {
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
    const data = await apiMotherDashboard(session.motherId || undefined);
    dashboard.value = data;
    if (data.profile?.id) {
      session.setMother(data.profile.id);
      try {
        const reminderPage = await apiListReminders("pending", 20, { owner_type: "mother", owner_id: data.profile.id });
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

function onNextStepAction() {
  if (activeReminder.value) return goReminders();
  const step = nextStep.value;
  if (step.type === "create_profile") return goProfileEdit();
  if (step.type === "add_record") return goNewRecord();
}

async function onShortcut(tmpl: RecordTemplate) {
  if (tmpl.value === "__body_metric__") {
    showMetricPicker.value = true;
    return;
  }
  if (tmpl.mode === "quick") {
    quickTemplate.value = tmpl;
    await loadLatestQuick(tmpl.value);
    showQuickSheet.value = true;
  } else {
    if (!session.motherId) {
      uni.showToast({ title: "请先创建宝妈档案", icon: "none" });
      return;
    }
    uni.navigateTo({ url: `/pages/mom/record-edit?mother_id=${session.motherId}&type=${tmpl.value}` });
  }
}

async function onMetricPick(type: string) {
  showMetricPicker.value = false;
  const tmpl = MOTHER_TEMPLATES.find((t) => t.value === type);
  if (!tmpl) return;
  quickTemplate.value = tmpl;
  await loadLatestQuick(tmpl.value);
  showQuickSheet.value = true;
}

async function loadLatestQuick(type: string) {
  lastQuickPayload.value = null;
  lastQuickSummary.value = "";
  const motherId = Number(session.motherId) || Number(dashboard.value?.profile?.id) || 0;
  if (!motherId) return;
  try {
    const latest = await apiLatestMotherRecord(motherId, type);
    lastQuickPayload.value = (latest.item?.payload || null) as Record<string, unknown> | null;
    lastQuickSummary.value = latest.item?.summary || "";
  } catch (e) {
    console.error(e);
  }
}

async function onQuickSaved(payload: Record<string, unknown>, summary: string) {
  const motherId = Number(session.motherId) || Number(dashboard.value?.profile?.id) || 0;
  if (!motherId) {
    uni.showToast({ title: "缺少宝妈档案", icon: "none" });
    showQuickSheet.value = false;
    return;
  }
  if (motherId !== Number(session.motherId)) {
    session.setMother(motherId);
  }
  try {
    await apiCreateMotherRecord(motherId, {
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

function goBabyProfileEdit() {
  const id = session.babyId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/baby/profile-edit${suffix}` });
}

function goProfileEdit() {
  const id = dashboard.value?.profile?.id || session.motherId || 0;
  const suffix = id ? `?id=${id}` : "";
  uni.navigateTo({ url: `/pages/mom/profile-edit${suffix}` });
}

function goNewRecord() {
  if (!session.motherId) {
    uni.showToast({ title: "请先创建宝妈档案", icon: "none" });
    return;
  }
  uni.navigateTo({ url: `/pages/mom/record-edit?mother_id=${session.motherId}` });
}

function openRecord(id: number) {
  uni.navigateTo({ url: `/pages/mom/record-edit?id=${id}&mother_id=${session.motherId}` });
}

function goRecordList() {
  if (!session.motherId) return;
  uni.navigateTo({ url: `/pages/mom/record-list?mother_id=${session.motherId}` });
}


function openSummaryRecord(id: number | undefined) {
  if (!id) return;
  uni.navigateTo({ url: `/pages/mom/record-edit?id=${id}&mother_id=${session.motherId}` });
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

.empty-actions {
  display: flex;
  gap: $cj-gap-sm;
  margin-top: $cj-gap-md;
}

.main-btn--flex {
  flex: 1;
  margin-top: 0 !important;
  font-size: 24rpx !important;
  padding: 0 12rpx !important;
}

.hero {
  margin-bottom: $cj-gap-xl;
  padding-top: 8rpx;
}

.hero-top {
  display: flex;
  align-items: center;
  gap: 14rpx;
  flex-wrap: wrap;
  margin-bottom: 10rpx;
}

.hero-kicker {
  font-size: 21rpx;
  letter-spacing: 5rpx;
  color: $cj-primary;
  font-weight: 500;
}

.hero-headline {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 24rpx;
  min-width: 0;
}

.hero-name {
  flex-shrink: 0;
  max-width: 48%;
  font-size: 34rpx;
  font-weight: $cj-fw-title;
  color: $cj-primary-dark;
  letter-spacing: 1rpx;
}

.stage-badge {
  padding: 5rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 19rpx;
  font-weight: 500;
  letter-spacing: 0.5rpx;
}

.stage-badge--inline {
  flex-shrink: 0;
  align-self: center;
}

.badge--pre {
  background: $cj-tag-prenatal-bg;
  color: $cj-tag-prenatal-text;
}

.badge--post {
  background: $cj-tag-postnatal-bg;
  color: $cj-tag-postnatal-text;
}

.hero-edit-link {
  margin-left: auto;
  font-size: 23rpx;
  color: $cj-primary;
  padding: 8rpx 0;
  font-weight: 500;
}

.hero-title {
  flex: 1;
  min-width: 0;
  font-size: 40rpx;
  line-height: 1.3;
  color: $cj-ink;
  font-weight: $cj-fw-display;
  text-align: right;
  letter-spacing: -0.5rpx;
}

.hero-title--solo {
  flex: none;
  width: 100%;
  text-align: left;
}

.hero-desc {
  display: block;
  margin-top: 12rpx;
  font-size: 25rpx;
  color: $cj-text-secondary;
  line-height: 1.65;
}

.next-card {
  background: linear-gradient(135deg, $cj-warn-bg 0%, #fff9f2 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  padding: 28rpx $cj-gap-md;
  margin-bottom: $cj-gap-md;
}

.next-label {
  display: block;
  font-size: 21rpx;
  color: $cj-text-muted;
  margin-bottom: 8rpx;
  letter-spacing: 1rpx;
}

.next-title {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  margin-bottom: $cj-gap-md;
  letter-spacing: 0.3rpx;
}

.next-btn {
  border-radius: $cj-radius-pill !important;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%) !important;
  color: #fffefb !important;
  border: none !important;
  font-size: 26rpx;
  padding: 0 36rpx !important;
  height: 72rpx !important;
  line-height: 72rpx !important;
  box-shadow: $cj-shadow-soft;
  letter-spacing: 1rpx;
}

.shortcut-section,
.summary-section,
.section {
  margin-bottom: $cj-gap-md;
}

.mom-snapshot-panel {
  position: relative;
  padding: 28rpx $cj-gap-md calc($cj-gap-md + 4rpx);
  background: linear-gradient(150deg, rgba(255, 253, 249, 0.98) 0%, rgba(255, 247, 240, 0.96) 50%, rgba(230, 241, 236, 0.3) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  overflow: hidden;

  &::before {
    content: "";
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.8;
    background:
      radial-gradient(ellipse 100% 70% at 0% 0%, rgba(201, 107, 92, 0.07) 0%, transparent 50%),
      radial-gradient(ellipse 80% 60% at 100% 100%, rgba(125, 171, 152, 0.1) 0%, transparent 45%);
  }

  .summary-card-top {
    flex-wrap: nowrap;
    min-width: 0;
  }

  .summary-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
  }

  .summary-hint-inline {
    white-space: nowrap;
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

.section-label {
  display: block;
  font-size: 21rpx;
  color: $cj-text-muted;
  letter-spacing: 3rpx;
  margin-bottom: $cj-gap-sm;
  font-weight: 500;
}

.shortcut-grid {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-sm;
}

.shortcut-tile {
  flex: 1;
  min-width: 140rpx;
  padding: 28rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
  transition: opacity 0.15s;

  &:active {
    opacity: 0.8;
    transform: scale(0.98);
  }
}

.shortcut-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 0.3rpx;
}

.shortcut-mode {
  display: block;
  margin-top: 6rpx;
  font-size: 19rpx;
  color: $cj-text-muted;
  letter-spacing: 0.5rpx;
}

.summary-grid {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $cj-gap-sm;
}

.summary-grid--stack {
  grid-template-columns: 1fr;
}

.summary-link {
  position: relative;
  z-index: 1;
  margin-top: $cj-gap-md;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6rpx;
  padding: 10rpx 0;
}

.summary-link-text {
  font-size: 23rpx;
  color: $cj-primary-dark;
  font-weight: 500;
}

.summary-link-chev {
  font-size: 28rpx;
  color: $cj-primary;
  line-height: 1;
  opacity: 0.75;
}

.summary-card {
  position: relative;
  min-width: 0;
  box-sizing: border-box;
  padding: 28rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;

  &:active {
    opacity: 0.8;
    transform: scale(0.99);
  }

  &--empty {
    .summary-value {
      color: $cj-text-muted;
      font-size: 26rpx;
      font-weight: 400;
    }
  }
}

.summary-card--mom {
  overflow: hidden;
  border: 1rpx solid $cj-border-faint;
  background: rgba(255, 253, 249, 0.95);
}

.summary-card-glow {
  position: absolute;
  width: 160rpx;
  height: 160rpx;
  border-radius: 50%;
  pointer-events: none;
  opacity: 0.45;
  top: -64rpx;
  right: -48rpx;
}

.summary-card-glow--rose {
  background: radial-gradient(circle, rgba(201, 107, 92, 0.18) 0%, transparent 65%);
}

.summary-card-glow--sage {
  background: radial-gradient(circle, rgba(125, 171, 152, 0.22) 0%, transparent 65%);
}

.summary-card-glow--sun {
  background: radial-gradient(circle, rgba(212, 165, 116, 0.2) 0%, transparent 65%);
}

.summary-card-glow--clay {
  background: radial-gradient(circle, rgba(168, 79, 66, 0.12) 0%, transparent 68%);
}

.summary-card--checkup .summary-label {
  color: $cj-primary-dark;
  font-weight: $cj-fw-title;
}

.summary-card--metric .summary-label {
  color: #2a5644;
  font-weight: $cj-fw-title;
}

.summary-card--mood .summary-label {
  color: #76583a;
  font-weight: $cj-fw-title;
}

.summary-card--symptom .summary-label {
  color: #684a42;
  font-weight: $cj-fw-title;
}

.summary-card-top {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 14rpx;
  min-width: 0;
}

.summary-label {
  flex: 1;
  min-width: 0;
  color: $cj-text-muted;
  font-size: 21rpx;
  font-weight: 500;
  letter-spacing: 0.5rpx;
}

.summary-hint-inline {
  flex-shrink: 0;
  font-size: 19rpx;
  color: $cj-text-muted;
}

.summary-value {
  position: relative;
  z-index: 1;
  display: block;
  margin-top: 10rpx;
  color: $cj-ink;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  line-height: 1.5;
  word-break: break-word;
  letter-spacing: -0.2rpx;
}

.summary-value--single {
  word-break: keep-all;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.summary-key-scroll {
  position: relative;
  z-index: 1;
  width: 100%;
  margin-top: 10rpx;
  white-space: nowrap;
}

.summary-key-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10rpx;
  padding: 4rpx 2rpx 8rpx;
  box-sizing: border-box;
}

.summary-key-chip {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  flex-shrink: 0;
  box-sizing: border-box;
  padding: 7rpx 12rpx;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-md;
  gap: 8rpx;
  max-width: 85vw;
}

.summary-key-label {
  flex-shrink: 0;
  max-width: none;
  white-space: nowrap;
}

.summary-key-value {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
    left: -56rpx;
    bottom: -64rpx;
    width: 200rpx;
    height: 200rpx;
    border-radius: 50%;
    background: radial-gradient(circle, rgba(125, 171, 152, 0.08) 0%, transparent 70%);
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

.record-dot--mom {
  background: linear-gradient(145deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
}

.record-line {
  flex: 1;
  width: 2rpx;
  min-height: 32rpx;
  margin-top: 8rpx;
  border-radius: 2rpx;
  background: linear-gradient(180deg, $cj-border-faint 0%, transparent 100%);
}

.record-line--mom {
  background: linear-gradient(180deg, rgba(201, 107, 92, 0.28) 0%, transparent 100%);
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

.sheet-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 999;
  display: flex;
  align-items: flex-end;
}

.picker-sheet {
  width: 100%;
  background: $cj-surface;
  border-radius: $cj-radius-xl $cj-radius-xl 0 0;
  padding: $cj-gap-lg $cj-gap-md calc(env(safe-area-inset-bottom, 0px) + 32rpx);
  box-sizing: border-box;
}

.picker-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-display;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-md;
  letter-spacing: 0.3rpx;
}

.picker-item {
  padding: $cj-gap-md;
  border-radius: $cj-radius-lg;
  margin-bottom: $cj-gap-sm;
  background: $cj-surface-2;

  &:active {
    opacity: 0.85;
  }
}

.picker-item-text {
  font-size: 30rpx;
  color: $cj-ink;
}

@import "@/styles/cj-record-kv-fields.scss";

.mom-snapshot-panel .summary-key-label,
.mom-snapshot-panel .summary-key-value {
  word-break: normal;
}

.mom-snapshot-panel .summary-key-label {
  white-space: nowrap;
}

.mom-snapshot-panel .summary-key-value {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
