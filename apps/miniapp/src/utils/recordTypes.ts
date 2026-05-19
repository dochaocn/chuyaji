export type TemplateField = {
  key: string;
  label: string;
  placeholder?: string;
  type: "text" | "number" | "date" | "select";
  unit?: string;
  options?: { value: string; label: string }[];
};

export type RecordTemplate = {
  value: string;
  label: string;
  entryLabel: string;
  mode: "quick" | "standard";
  recommendedFields: TemplateField[];
  optionalFields: TemplateField[];
  summaryPlaceholder?: string;
  autoSummary?: (payload: Record<string, unknown>) => string;
  attachmentHint?: string;
  metricKeys?: string[];
};

const nextTimeField: TemplateField = { key: "next_time", label: "下次时间", type: "date" };

const prenatalTemplates: RecordTemplate[] = [
  {
    value: "prenatal_checkup",
    label: "产检",
    entryLabel: "产检",
    mode: "standard",
    recommendedFields: [
      { key: "gestational_weeks", label: "孕周", type: "number", placeholder: "例如 24", unit: "周" },
      { key: "doctor_advice", label: "医生建议", type: "text", placeholder: "医生有什么特别叮嘱？" },
      { key: "next_time", label: "下次时间", type: "date" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "items", label: "检查项目", type: "text", placeholder: "例如：血常规、B超" },
      { key: "result", label: "检查结果", type: "text", placeholder: "总体结论" },
    ],
    summaryPlaceholder: "今天产检结果如何？有没有医生特别提醒？",
    attachmentHint: "可上传检查单、报告单",
  },
  {
    value: "ultrasound",
    label: "B 超",
    entryLabel: "B 超",
    mode: "standard",
    recommendedFields: [
      { key: "gestational_weeks", label: "孕周", type: "number", placeholder: "例如 20", unit: "周" },
      { key: "estimated_weight_g", label: "胎儿估重", type: "number", placeholder: "例如 300", unit: "g" },
      { key: "doctor_conclusion", label: "医生结论", type: "text", placeholder: "总体情况如何？" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "bpd_mm", label: "BPD", type: "number", placeholder: "mm", unit: "mm" },
      { key: "hc_mm", label: "HC", type: "number", placeholder: "mm", unit: "mm" },
      { key: "ac_mm", label: "AC", type: "number", placeholder: "mm", unit: "mm" },
      { key: "fl_mm", label: "FL", type: "number", placeholder: "mm", unit: "mm" },
      { key: "fhr", label: "胎心", type: "number", placeholder: "次/分", unit: "次/分" },
      nextTimeField,
    ],
    summaryPlaceholder: "B 超整体情况如何？",
    attachmentHint: "可上传 B 超单、影像照片",
  },
  {
    value: "screening",
    label: "筛查",
    entryLabel: "筛查",
    mode: "standard",
    recommendedFields: [
      { key: "type", label: "筛查类型", type: "text", placeholder: "例如唐筛、无创DNA" },
      { key: "conclusion", label: "结论", type: "text", placeholder: "低风险 / 高风险 / 阴性" },
    ],
    optionalFields: [
      { key: "risk_hint", label: "风险提示", type: "text", placeholder: "如有特殊风险提示" },
      { key: "doctor_advice", label: "医生建议", type: "text", placeholder: "后续建议" },
      nextTimeField,
    ],
    summaryPlaceholder: "筛查结果和医生建议是什么？",
    attachmentHint: "可上传报告单",
  },
  {
    value: "ultrasound_4d",
    label: "四维",
    entryLabel: "四维",
    mode: "standard",
    recommendedFields: [
      { key: "gestational_weeks", label: "孕周", type: "number", placeholder: "例如 24", unit: "周" },
      { key: "doctor_conclusion", label: "医生结论", type: "text", placeholder: "结构筛查 / 总体情况" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "estimated_weight_g", label: "胎儿估重", type: "number", placeholder: "例如 800", unit: "g" },
      { key: "note", label: "备注", type: "text", placeholder: "例如面部成像、配合度等" },
      nextTimeField,
    ],
    summaryPlaceholder: "四维检查印象最深的是什么？",
    attachmentHint: "可上传四维报告或影像截图",
  },
  {
    value: "glucose_test",
    label: "糖耐",
    entryLabel: "糖耐",
    mode: "standard",
    recommendedFields: [
      { key: "conclusion", label: "结论", type: "text", placeholder: "正常 / 异常" },
    ],
    optionalFields: [
      { key: "fasting", label: "空腹", type: "number", placeholder: "mmol/L", unit: "mmol/L" },
      { key: "one_hour", label: "1 小时", type: "number", placeholder: "mmol/L", unit: "mmol/L" },
      { key: "two_hour", label: "2 小时", type: "number", placeholder: "mmol/L", unit: "mmol/L" },
      nextTimeField,
    ],
    summaryPlaceholder: "糖耐结果是否正常？",
    attachmentHint: "可上传化验单",
    metricKeys: ["fasting", "one_hour", "two_hour"],
  },
  {
    value: "nt",
    label: "NT",
    entryLabel: "NT",
    mode: "standard",
    recommendedFields: [
      { key: "nt_mm", label: "NT 值", type: "number", placeholder: "例如 1.2", unit: "mm" },
      { key: "conclusion", label: "结论", type: "text", placeholder: "低风险 / 高风险" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "gestational_weeks", label: "孕周", type: "number", placeholder: "例如 12", unit: "周" },
      nextTimeField,
    ],
    summaryPlaceholder: "NT 检查结果如何？",
    attachmentHint: "可上传检查报告",
  },
];

const postnatalTemplates: RecordTemplate[] = [
  {
    value: "feeding",
    label: "喂养",
    entryLabel: "喂养",
    mode: "quick",
    recommendedFields: [
      { key: "amount_ml", label: "奶量", type: "number", placeholder: "例如 120", unit: "ml" },
    ],
    optionalFields: [
      { key: "duration_min", label: "时长", type: "number", placeholder: "分钟", unit: "min" },
      {
        key: "feeding_type",
        label: "喂养方式",
        type: "select",
        options: [
          { value: "breast", label: "母乳" },
          { value: "formula", label: "配方奶" },
          { value: "mixed", label: "混合" },
        ],
      },
      {
        key: "spit_up",
        label: "是否吐奶",
        type: "select",
        options: [
          { value: "no", label: "否" },
          { value: "yes", label: "是" },
        ],
      },
      { key: "times", label: "次数", type: "number", placeholder: "例如 1" },
    ],
    autoSummary: (payload) => {
      const ml = payload.amount_ml;
      const min = payload.duration_min;
      if (typeof ml === "number" && ml > 0) return `喂养 ${ml}ml`;
      if (typeof min === "number" && min > 0) return `喂养 ${min}min`;
      return "喂养已记录";
    },
  },
  {
    value: "sleep",
    label: "睡眠",
    entryLabel: "睡眠",
    mode: "quick",
    recommendedFields: [
      { key: "duration_min", label: "总睡眠时长", type: "number", placeholder: "例如 480", unit: "min" },
    ],
    optionalFields: [
      { key: "night_wake", label: "夜醒次数", type: "number", placeholder: "例如 2" },
      { key: "fall_asleep", label: "入睡情况", type: "text", placeholder: "例如顺利、哭闹 10 分钟" },
    ],
    autoSummary: (payload) => {
      const min = payload.duration_min;
      if (typeof min === "number" && min > 0) return `睡眠 ${min}min`;
      return "睡眠已记录";
    },
    metricKeys: ["duration_min"],
  },
  {
    value: "diaper",
    label: "排便",
    entryLabel: "排便",
    mode: "quick",
    recommendedFields: [
      { key: "times", label: "次数", type: "number", placeholder: "例如 3" },
      {
        key: "consistency",
        label: "性状",
        type: "select",
        options: [
          { value: "normal", label: "正常" },
          { value: "loose", label: "偏稀" },
          { value: "hard", label: "偏硬" },
          { value: "watery", label: "水样" },
        ],
      },
    ],
    optionalFields: [
      { key: "color", label: "颜色", type: "text", placeholder: "例如黄色、绿色" },
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
      const times = payload.times;
      if (typeof times === "number" && times > 0) return `排便 ${times} 次`;
      return "排便已记录";
    },
  },
  {
    value: "growth",
    label: "生长",
    entryLabel: "生长",
    mode: "standard",
    recommendedFields: [
      { key: "weight_g", label: "体重", type: "number", placeholder: "例如 5200", unit: "g" },
      { key: "height_cm", label: "身长", type: "number", placeholder: "例如 58", unit: "cm" },
    ],
    optionalFields: [
      { key: "head_circumference_cm", label: "头围", type: "number", placeholder: "例如 38", unit: "cm" },
      { key: "method", label: "记录方式", type: "text", placeholder: "例如医院测量、家中体重秤" },
    ],
    summaryPlaceholder: "今天测得的生长数据如何？",
    metricKeys: ["weight_g", "height_cm", "head_circumference_cm"],
  },
  {
    value: "development",
    label: "发育",
    entryLabel: "发育",
    mode: "standard",
    recommendedFields: [
      {
        key: "milestone",
        label: "里程碑类型",
        type: "select",
        options: [
          { value: "smile", label: "第一次微笑" },
          { value: "head_up", label: "抬头" },
          { value: "roll", label: "翻身" },
          { value: "sit", label: "独坐" },
          { value: "crawl", label: "爬行" },
          { value: "stand", label: "扶站" },
          { value: "walk", label: "行走" },
          { value: "first_word", label: "第一个词" },
          { value: "other", label: "其他" },
        ],
      },
      { key: "milestone_date", label: "出现日期", type: "date" },
    ],
    optionalFields: [
      { key: "description", label: "表现描述", type: "text", placeholder: "详细描述这个里程碑" },
    ],
    summaryPlaceholder: "今天出现了什么新的成长表现？",
  },
  {
    value: "checkup",
    label: "体检",
    entryLabel: "体检",
    mode: "standard",
    recommendedFields: [
      { key: "age_months", label: "月龄", type: "number", placeholder: "例如 6", unit: "月" },
      { key: "conclusion", label: "结论", type: "text", placeholder: "总体正常 / 有需关注项" },
      { key: "next_time", label: "下次时间", type: "date" },
    ],
    optionalFields: [
      { key: "hospital", label: "医院", type: "text", placeholder: "可选" },
      { key: "advice", label: "建议", type: "text", placeholder: "医生建议" },
    ],
    summaryPlaceholder: "体检结果和医生建议是什么？",
    attachmentHint: "可上传体检报告",
  },
  {
    value: "vaccine",
    label: "疫苗",
    entryLabel: "疫苗",
    mode: "standard",
    recommendedFields: [
      { key: "name", label: "疫苗名称", type: "text", placeholder: "例如乙肝疫苗第二针" },
      { key: "vaccine_date", label: "接种日期", type: "date" },
      { key: "next_time", label: "下次时间", type: "date" },
    ],
    optionalFields: [
      { key: "site", label: "部位", type: "text", placeholder: "例如左大腿" },
      { key: "reaction", label: "接种反应", type: "text", placeholder: "例如无异常、低热" },
    ],
    summaryPlaceholder: "接种后状态如何？",
  },
  {
    value: "illness",
    label: "疾病",
    entryLabel: "疾病",
    mode: "standard",
    recommendedFields: [
      { key: "symptom", label: "症状", type: "text", placeholder: "例如发热、咳嗽、流涕" },
      { key: "temperature_celsius", label: "体温", type: "number", placeholder: "例如 38.2", unit: "℃" },
      { key: "diagnosis", label: "诊断", type: "text", placeholder: "例如上呼吸道感染" },
    ],
    optionalFields: [
      { key: "medication", label: "用药", type: "text", placeholder: "例如美林 5ml" },
      { key: "hospital", label: "就诊医院", type: "text", placeholder: "可选" },
    ],
    summaryPlaceholder: "这次生病主要情况是什么？",
    metricKeys: ["temperature_celsius"],
  },
];

export const BABY_TEMPLATES = {
  prenatal: prenatalTemplates,
  postnatal: postnatalTemplates,
} as const;

export const RECORD_TYPES = {
  prenatal: prenatalTemplates.map((t) => ({ value: t.value, label: t.label })),
  postnatal: postnatalTemplates.map((t) => ({ value: t.value, label: t.label })),
} as const;

export const BABY_PHASE_FILTER_OPTIONS: { value: "all" | "prenatal" | "postnatal"; label: string }[] = [
  { value: "all", label: "全部" },
  { value: "prenatal", label: "怀孕期" },
  { value: "postnatal", label: "成长期" },
];

export const BABY_RECORD_TYPE_FILTER_OPTIONS: { value: string; label: string }[] = [
  ...RECORD_TYPES.prenatal,
  ...RECORD_TYPES.postnatal,
];

export const BABY_HOME_PRENATAL_SHORTCUT_ORDER: readonly string[] = [
  "prenatal_checkup",
  "ultrasound",
  "screening",
];

export function babyHomePostnatalShortcutOrder(ageMonths: number): string[] {
  if (ageMonths < 6) return ["feeding", "sleep", "diaper", "growth"];
  return ["feeding", "sleep", "growth", "vaccine"];
}

export function labelForType(phase: "prenatal" | "postnatal", type: string): string {
  const hit = BABY_TEMPLATES[phase].find((x) => x.value === type);
  return hit ? hit.label : type;
}

export function templateForType(
  phase: "prenatal" | "postnatal",
  type: string
): RecordTemplate | undefined {
  return BABY_TEMPLATES[phase].find((t) => t.value === type);
}

export function formatFieldValueForDisplay(
  field: TemplateField,
  payload: Record<string, unknown>
): string | null {
  const raw = payload[field.key];
  if (raw === null || raw === undefined || raw === "") return null;
  if (field.type === "select") {
    const hit = field.options?.find((o) => o.value === String(raw));
    return hit?.label ?? String(raw);
  }
  return String(raw);
}

/** 列表/卡片预览：日期截断、数字带单位（与宝妈记录预览一致） */
export function previewValueForTemplateField(field: TemplateField, display: string): string {
  if (field.type === "date") {
    const s = display.trim();
    return s.length >= 10 ? s.slice(0, 10) : s;
  }
  if (field.unit && field.type === "number") {
    return `${display}${field.unit.startsWith("/") ? "" : " "}${field.unit}`;
  }
  return display;
}

const BABY_RECORD_PREVIEW_MAX_FIELDS = 5;

/** 宝宝记录首页/列表：按阶段模板展示已填关键字段 */
export function getBabyRecordPreviewRows(record: {
  phase: "prenatal" | "postnatal";
  record_type: string;
  payload: Record<string, unknown>;
}): { label: string; value: string }[] {
  const tmpl = templateForType(record.phase, record.record_type);
  if (!tmpl) return [];
  const payload = (record.payload || {}) as Record<string, unknown>;
  const rows: { label: string; value: string }[] = [];
  const fields = [...tmpl.recommendedFields, ...tmpl.optionalFields];
  for (const field of fields) {
    if (rows.length >= BABY_RECORD_PREVIEW_MAX_FIELDS) break;
    const raw = formatFieldValueForDisplay(field, payload);
    if (raw === null) continue;
    rows.push({ label: field.label, value: previewValueForTemplateField(field, raw) });
  }
  return rows;
}
