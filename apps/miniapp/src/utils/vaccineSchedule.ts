export type VaccineScheduleItem = {
  vaccine_key: string;
  dose_key: string;
  name: string;
  dose_label: string;
  recommended_age_days: number;
  type: "required" | "optional";
  description: string;
};

export const VACCINE_SCHEDULE: VaccineScheduleItem[] = [
  { vaccine_key: "hepatitis_b", dose_key: "dose_1", name: "乙肝疫苗", dose_label: "第 1 剂", recommended_age_days: 0, type: "required", description: "出生后尽早接种" },
  { vaccine_key: "bcg", dose_key: "dose_1", name: "卡介苗", dose_label: "第 1 剂", recommended_age_days: 0, type: "required", description: "出生后接种" },
  { vaccine_key: "hepatitis_b", dose_key: "dose_2", name: "乙肝疫苗", dose_label: "第 2 剂", recommended_age_days: 30, type: "required", description: "约 1 月龄" },
  { vaccine_key: "polio", dose_key: "dose_1", name: "脊灰疫苗", dose_label: "第 1 剂", recommended_age_days: 60, type: "required", description: "约 2 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_1", name: "百白破疫苗", dose_label: "第 1 剂", recommended_age_days: 90, type: "required", description: "约 3 月龄" },
  { vaccine_key: "polio", dose_key: "dose_2", name: "脊灰疫苗", dose_label: "第 2 剂", recommended_age_days: 90, type: "required", description: "约 3 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_2", name: "百白破疫苗", dose_label: "第 2 剂", recommended_age_days: 120, type: "required", description: "约 4 月龄" },
  { vaccine_key: "polio", dose_key: "dose_3", name: "脊灰疫苗", dose_label: "第 3 剂", recommended_age_days: 120, type: "required", description: "约 4 月龄" },
  { vaccine_key: "dtap", dose_key: "dose_3", name: "百白破疫苗", dose_label: "第 3 剂", recommended_age_days: 150, type: "required", description: "约 5 月龄" },
  { vaccine_key: "hepatitis_b", dose_key: "dose_3", name: "乙肝疫苗", dose_label: "第 3 剂", recommended_age_days: 180, type: "required", description: "约 6 月龄" },
  { vaccine_key: "meningococcal_a", dose_key: "dose_1", name: "A 群流脑疫苗", dose_label: "第 1 剂", recommended_age_days: 180, type: "required", description: "约 6 月龄" },
  { vaccine_key: "measles", dose_key: "dose_1", name: "麻腮风疫苗", dose_label: "第 1 剂", recommended_age_days: 240, type: "required", description: "约 8 月龄" },
  { vaccine_key: "varicella", dose_key: "dose_1", name: "水痘疫苗", dose_label: "第 1 剂", recommended_age_days: 365, type: "optional", description: "可按当地建议接种" },
];

export function vaccineDoseId(item: VaccineScheduleItem) {
  return `${item.vaccine_key}:${item.dose_key}`;
}
