import {
  formatFieldValueForDisplay,
  previewValueForTemplateField,
  type RecordTemplate,
  type TemplateField,
} from "./recordTypes";

export type { RecordTemplate, TemplateField };

const motherTemplates: RecordTemplate[] = [
  {
    value: "checkup",
    label: "产检",
    entryLabel: "产检/复查",
    mode: "standard",
    recommendedFields: [
      { key: "gestational_weeks", label: "孕周", type: "number", placeholder: "例如 28", unit: "周" },
      { key: "conclusion", label: "结论", type: "text", placeholder: "总体正常 / 有需关注项" },
      { key: "next_time", label: "下次时间", type: "date" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "items", label: "检查项目", type: "text", placeholder: "例如血压、尿常规" },
      { key: "doctor_advice", label: "医生建议", type: "text", placeholder: "有无特别叮嘱" },
    ],
    summaryPlaceholder: "这次产检最重要的信息是什么？",
    attachmentHint: "可上传检查单、报告单",
  },
  {
    value: "followup",
    label: "复诊",
    entryLabel: "复诊",
    mode: "standard",
    recommendedFields: [
      { key: "reason", label: "复诊原因", type: "text", placeholder: "例如产后 42 天检查" },
      { key: "result", label: "结果", type: "text", placeholder: "总体情况" },
      { key: "next_time", label: "下次时间", type: "date" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "doctor_advice", label: "医生建议", type: "text", placeholder: "有无特别建议" },
    ],
    summaryPlaceholder: "复诊结果和后续安排是什么？",
    attachmentHint: "可上传复诊报告",
  },
  {
    value: "nutrition",
    label: "营养补充",
    entryLabel: "营养",
    mode: "standard",
    recommendedFields: [
      { key: "type", label: "补充类型", type: "text", placeholder: "例如 DHA、铁、叶酸" },
      { key: "dose", label: "剂量", type: "text", placeholder: "例如 200mg" },
    ],
    optionalFields: [
      { key: "frequency", label: "频次", type: "text", placeholder: "例如每天一次" },
      { key: "note", label: "备注", type: "text", placeholder: "可选" },
    ],
    summaryPlaceholder: "今天补充了什么？",
  },
  {
    value: "medication",
    label: "用药",
    entryLabel: "用药",
    mode: "standard",
    recommendedFields: [
      { key: "name", label: "药名", type: "text", placeholder: "例如叶酸、钙片" },
      { key: "reason", label: "原因", type: "text", placeholder: "例如补钙、止吐" },
    ],
    optionalFields: [
      { key: "dose", label: "剂量", type: "text", placeholder: "例如 600mg" },
      { key: "doctor_advice", label: "医生建议", type: "text", placeholder: "有无医嘱" },
    ],
    summaryPlaceholder: "因为什么开始或继续用药？",
  },
  {
    value: "symptom",
    label: "不适",
    entryLabel: "不适",
    mode: "standard",
    recommendedFields: [
      { key: "type", label: "症状类型", type: "text", placeholder: "例如恶心、水肿、腰酸" },
      {
        key: "severity",
        label: "程度",
        type: "select",
        options: [
          { value: "mild", label: "轻微" },
          { value: "moderate", label: "中等" },
          { value: "severe", label: "严重" },
        ],
      },
    ],
    optionalFields: [
      { key: "duration", label: "持续时间", type: "text", placeholder: "例如 2 天" },
      {
        key: "relieved",
        label: "是否缓解",
        type: "select",
        options: [
          { value: "yes", label: "已缓解" },
          { value: "no", label: "未缓解" },
        ],
      },
    ],
    summaryPlaceholder: "今天最明显的身体感受是什么？",
  },
  {
    value: "weight",
    label: "体重",
    entryLabel: "体重",
    mode: "quick",
    recommendedFields: [
      { key: "kg", label: "体重", type: "number", placeholder: "例如 62.5", unit: "kg" },
    ],
    optionalFields: [
      { key: "measure_time", label: "测量时间", type: "text", placeholder: "例如早晨空腹" },
      { key: "note", label: "备注", type: "text", placeholder: "可选" },
    ],
    autoSummary: (payload) => {
      const kg = payload.kg;
      if (typeof kg === "number" && kg > 0) return `体重 ${kg}kg`;
      return "体重已记录";
    },
    metricKeys: ["kg"],
  },
  {
    value: "blood_pressure",
    label: "血压",
    entryLabel: "血压",
    mode: "quick",
    recommendedFields: [
      { key: "systolic", label: "收缩压", type: "number", placeholder: "例如 120", unit: "mmHg" },
      { key: "diastolic", label: "舒张压", type: "number", placeholder: "例如 80", unit: "mmHg" },
    ],
    optionalFields: [
      { key: "measure_time", label: "测量时间", type: "text", placeholder: "例如早晨" },
      {
        key: "abnormal",
        label: "是否异常",
        type: "select",
        options: [
          { value: "no", label: "否" },
          { value: "yes", label: "是" },
        ],
      },
    ],
    autoSummary: (payload) => {
      const s = payload.systolic;
      const d = payload.diastolic;
      if (typeof s === "number" && typeof d === "number") return `血压 ${s}/${d}`;
      return "血压已记录";
    },
    metricKeys: ["systolic", "diastolic"],
  },
  {
    value: "blood_sugar",
    label: "血糖",
    entryLabel: "血糖",
    mode: "quick",
    recommendedFields: [
      { key: "mmol_l", label: "血糖", type: "number", placeholder: "例如 5.2", unit: "mmol/L" },
    ],
    optionalFields: [
      {
        key: "period",
        label: "测量时段",
        type: "select",
        options: [
          { value: "fasting", label: "空腹" },
          { value: "postprandial_1h", label: "餐后 1 小时" },
          { value: "postprandial_2h", label: "餐后 2 小时" },
          { value: "other", label: "其他" },
        ],
      },
      {
        key: "is_fasting",
        label: "是否空腹",
        type: "select",
        options: [
          { value: "yes", label: "是" },
          { value: "no", label: "否" },
        ],
      },
    ],
    autoSummary: (payload) => {
      const v = payload.mmol_l;
      if (typeof v === "number" && v > 0) return `血糖 ${v} mmol/L`;
      return "血糖已记录";
    },
    metricKeys: ["mmol_l"],
  },
  {
    value: "mood",
    label: "心情",
    entryLabel: "心情",
    mode: "quick",
    recommendedFields: [
      {
        key: "score",
        label: "心情分值",
        type: "number",
        placeholder: "1-10",
      },
    ],
    optionalFields: [
      { key: "tag", label: "情绪标签", type: "text", placeholder: "例如焦虑、平静、开心" },
      { key: "trigger", label: "触发原因", type: "text", placeholder: "是什么影响了今天的情绪？" },
      { key: "relief", label: "缓解方式", type: "text", placeholder: "有没有让你好受一些的事情？" },
    ],
    autoSummary: (payload) => {
      const score = payload.score;
      if (typeof score === "number" && score > 0) return `心情 ${score}/10`;
      return "心情已记录";
    },
    metricKeys: ["score"],
  },
  {
    value: "recovery",
    label: "产后恢复",
    entryLabel: "恢复",
    mode: "standard",
    recommendedFields: [
      { key: "feeling", label: "自我感受", type: "text", placeholder: "整体状态如何？" },
      { key: "pain_level", label: "疼痛程度", type: "number", placeholder: "0-10", unit: "/10" },
    ],
    optionalFields: [
      { key: "sleep_quality", label: "睡眠", type: "text", placeholder: "例如睡了 5 小时，质量一般" },
      { key: "wound_lochia", label: "伤口/恶露", type: "text", placeholder: "例如恶露减少、伤口无异常" },
      { key: "feeding_status", label: "喂养状态", type: "text", placeholder: "例如母乳顺利、涨奶" },
    ],
    summaryPlaceholder: "今天恢复状态最值得记下的是什么？",
  },
];

export const MOTHER_TEMPLATES: RecordTemplate[] = motherTemplates;

export const MOTHER_RECORD_TYPES = motherTemplates.map((t) => ({ value: t.value, label: t.label }));

export const MOTHER_TIMELINE_FILTER_OPTIONS: { value: string; label: string }[] = [
  { value: "all", label: "全部" },
  ...MOTHER_RECORD_TYPES,
];

export function labelForMotherType(type: string): string {
  const hit = motherTemplates.find((item) => item.value === type);
  return hit ? hit.label : type;
}

export function templateForMotherType(type: string): RecordTemplate | undefined {
  return motherTemplates.find((t) => t.value === type);
}

const MOTHER_RECORD_PREVIEW_MAX_FIELDS = 5;

/** 首页/列表用：按模板把已填关键字段排成行，便于扫读 */
export function getMotherRecordPreviewRows(record: {
  record_type: string;
  payload: Record<string, unknown>;
}): { label: string; value: string }[] {
  const tmpl = templateForMotherType(record.record_type);
  if (!tmpl) return [];
  const payload = (record.payload || {}) as Record<string, unknown>;
  const rows: { label: string; value: string }[] = [];
  const fields = [...tmpl.recommendedFields, ...tmpl.optionalFields];
  for (const field of fields) {
    if (rows.length >= MOTHER_RECORD_PREVIEW_MAX_FIELDS) break;
    const raw = formatFieldValueForDisplay(field, payload);
    if (raw === null) continue;
    rows.push({ label: field.label, value: previewValueForTemplateField(field, raw) });
  }
  return rows;
}

export const BODY_METRIC_TYPES = ["weight", "blood_pressure", "blood_sugar"] as const;
export type BodyMetricType = (typeof BODY_METRIC_TYPES)[number];
