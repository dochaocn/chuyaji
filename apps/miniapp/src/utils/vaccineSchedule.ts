export type VaccineScheduleItem = {
  vaccine_key: string;
  dose_key: string;
  name: string;
  dose_label: string;
  recommended_age_days: number;
  type: "required" | "optional";
  description: string;
  /** 展示分组：月龄段 */
  group: "0_6m" | "6_12m" | "12_24m" | "2_6y" | "optional";
};

export const VACCINE_GROUP_LABELS: Record<VaccineScheduleItem["group"], string> = {
  "0_6m": "0–6 月龄",
  "6_12m": "6–12 月龄",
  "12_24m": "12–24 月龄",
  "2_6y": "2–6 岁",
  optional: "常见自费苗",
};

export const VACCINE_SCHEDULE: VaccineScheduleItem[] = [
  // —— 0–6 月龄（免疫规划）——
  { vaccine_key: "hepatitis_b", dose_key: "dose_1", name: "乙肝疫苗", dose_label: "第 1 剂", recommended_age_days: 0, type: "required", group: "0_6m", description: "出生后尽早接种" },
  { vaccine_key: "bcg", dose_key: "dose_1", name: "卡介苗", dose_label: "第 1 剂", recommended_age_days: 0, type: "required", group: "0_6m", description: "出生后接种" },
  { vaccine_key: "hepatitis_b", dose_key: "dose_2", name: "乙肝疫苗", dose_label: "第 2 剂", recommended_age_days: 30, type: "required", group: "0_6m", description: "约 1 月龄" },
  { vaccine_key: "polio", dose_key: "dose_1", name: "脊灰疫苗", dose_label: "第 1 剂", recommended_age_days: 60, type: "required", group: "0_6m", description: "约 2 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_1", name: "百白破疫苗", dose_label: "第 1 剂", recommended_age_days: 90, type: "required", group: "0_6m", description: "约 3 月龄" },
  { vaccine_key: "polio", dose_key: "dose_2", name: "脊灰疫苗", dose_label: "第 2 剂", recommended_age_days: 90, type: "required", group: "0_6m", description: "约 3 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_2", name: "百白破疫苗", dose_label: "第 2 剂", recommended_age_days: 120, type: "required", group: "0_6m", description: "约 4 月龄" },
  { vaccine_key: "polio", dose_key: "dose_3", name: "脊灰疫苗", dose_label: "第 3 剂", recommended_age_days: 120, type: "required", group: "0_6m", description: "约 4 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_3", name: "百白破疫苗", dose_label: "第 3 剂", recommended_age_days: 150, type: "required", group: "0_6m", description: "约 5 月龄" },
  { vaccine_key: "hepatitis_b", dose_key: "dose_3", name: "乙肝疫苗", dose_label: "第 3 剂", recommended_age_days: 180, type: "required", group: "0_6m", description: "约 6 月龄" },
  { vaccine_key: "meningococcal_a", dose_key: "dose_1", name: "A 群流脑疫苗", dose_label: "第 1 剂", recommended_age_days: 180, type: "required", group: "0_6m", description: "约 6 月龄" },

  // —— 6–12 月龄 ——
  { vaccine_key: "measles", dose_key: "dose_1", name: "麻腮风疫苗", dose_label: "第 1 剂", recommended_age_days: 240, type: "required", group: "6_12m", description: "约 8 月龄" },
  { vaccine_key: "je", dose_key: "dose_1", name: "乙脑疫苗", dose_label: "第 1 剂", recommended_age_days: 240, type: "required", group: "6_12m", description: "约 8 月龄（灭活程序另计）" },
  { vaccine_key: "meningococcal_a", dose_key: "dose_2", name: "A 群流脑疫苗", dose_label: "第 2 剂", recommended_age_days: 270, type: "required", group: "6_12m", description: "约 9 月龄" },

  // —— 12–24 月龄 ——
  { vaccine_key: "je", dose_key: "dose_2", name: "乙脑疫苗", dose_label: "第 2 剂", recommended_age_days: 540, type: "required", group: "12_24m", description: "约 18 月龄（减毒程序）" },
  { vaccine_key: "measles", dose_key: "dose_2", name: "麻腮风疫苗", dose_label: "第 2 剂", recommended_age_days: 540, type: "required", group: "12_24m", description: "约 18 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_4", name: "百白破疫苗", dose_label: "第 4 剂", recommended_age_days: 540, type: "required", group: "12_24m", description: "约 18 月龄" },
  { vaccine_key: "hepatitis_a", dose_key: "dose_1", name: "甲肝疫苗", dose_label: "第 1 剂", recommended_age_days: 540, type: "required", group: "12_24m", description: "约 18 月龄（减毒/灭活以当地为准）" },
  { vaccine_key: "hepatitis_a", dose_key: "dose_2", name: "甲肝疫苗", dose_label: "第 2 剂", recommended_age_days: 720, type: "required", group: "12_24m", description: "约 24 月龄（灭活程序第 2 剂）" },

  // —— 2–6 岁 ——
  { vaccine_key: "meningococcal_ac", dose_key: "dose_1", name: "A+C 群流脑疫苗", dose_label: "第 1 剂", recommended_age_days: 1095, type: "required", group: "2_6y", description: "约 3 岁" },
  { vaccine_key: "polio", dose_key: "dose_4", name: "脊灰疫苗", dose_label: "第 4 剂", recommended_age_days: 1460, type: "required", group: "2_6y", description: "约 4 岁" },
  { vaccine_key: "dtap", dose_key: "dose_5", name: "百白破疫苗", dose_label: "第 5 剂", recommended_age_days: 2190, type: "required", group: "2_6y", description: "约 6 岁" },
  { vaccine_key: "meningococcal_ac", dose_key: "dose_2", name: "A+C 群流脑疫苗", dose_label: "第 2 剂", recommended_age_days: 2190, type: "required", group: "2_6y", description: "约 6 岁加强" },

  // —— 常见自费（可选）——
  { vaccine_key: "pcv", dose_key: "dose_1", name: "肺炎球菌疫苗", dose_label: "第 1 剂", recommended_age_days: 60, type: "optional", group: "optional", description: "约 2 月龄起，程序以说明书为准" },
  { vaccine_key: "pcv", dose_key: "dose_2", name: "肺炎球菌疫苗", dose_label: "第 2 剂", recommended_age_days: 120, type: "optional", group: "optional", description: "约 4 月龄" },
  { vaccine_key: "pcv", dose_key: "dose_3", name: "肺炎球菌疫苗", dose_label: "第 3 剂", recommended_age_days: 180, type: "optional", group: "optional", description: "约 6 月龄" },
  { vaccine_key: "pcv", dose_key: "dose_4", name: "肺炎球菌疫苗", dose_label: "加强剂", recommended_age_days: 365, type: "optional", group: "optional", description: "约 12–15 月龄" },
  { vaccine_key: "rotavirus", dose_key: "dose_1", name: "轮状病毒疫苗", dose_label: "第 1 剂", recommended_age_days: 60, type: "optional", group: "optional", description: "约 2 月龄起口服" },
  { vaccine_key: "rotavirus", dose_key: "dose_2", name: "轮状病毒疫苗", dose_label: "第 2 剂", recommended_age_days: 90, type: "optional", group: "optional", description: "约 3 月龄" },
  { vaccine_key: "rotavirus", dose_key: "dose_3", name: "轮状病毒疫苗", dose_label: "第 3 剂", recommended_age_days: 120, type: "optional", group: "optional", description: "约 4 月龄（视产品）" },
  { vaccine_key: "hib", dose_key: "dose_1", name: "Hib 疫苗", dose_label: "第 1 剂", recommended_age_days: 60, type: "optional", group: "optional", description: "约 2 月龄" },
  { vaccine_key: "hib", dose_key: "dose_2", name: "Hib 疫苗", dose_label: "第 2 剂", recommended_age_days: 90, type: "optional", group: "optional", description: "约 3 月龄" },
  { vaccine_key: "hib", dose_key: "dose_3", name: "Hib 疫苗", dose_label: "第 3 剂", recommended_age_days: 120, type: "optional", group: "optional", description: "约 4 月龄" },
  { vaccine_key: "hib", dose_key: "dose_4", name: "Hib 疫苗", dose_label: "加强剂", recommended_age_days: 540, type: "optional", group: "optional", description: "约 18 月龄" },
  { vaccine_key: "varicella", dose_key: "dose_1", name: "水痘疫苗", dose_label: "第 1 剂", recommended_age_days: 365, type: "optional", group: "optional", description: "约 12 月龄起" },
  { vaccine_key: "varicella", dose_key: "dose_2", name: "水痘疫苗", dose_label: "第 2 剂", recommended_age_days: 1460, type: "optional", group: "optional", description: "约 4 岁（各地建议不一）" },
  { vaccine_key: "ev71", dose_key: "dose_1", name: "手足口病疫苗", dose_label: "第 1 剂", recommended_age_days: 180, type: "optional", group: "optional", description: "约 6 月龄起（EV71）" },
  { vaccine_key: "ev71", dose_key: "dose_2", name: "手足口病疫苗", dose_label: "第 2 剂", recommended_age_days: 210, type: "optional", group: "optional", description: "间隔约 1 个月" },
  { vaccine_key: "influenza", dose_key: "annual", name: "流感疫苗", dose_label: "当年", recommended_age_days: 180, type: "optional", group: "optional", description: "6 月龄起可每年接种，以当季建议为准" },
];

export function vaccineDoseId(item: VaccineScheduleItem) {
  return `${item.vaccine_key}:${item.dose_key}`;
}
