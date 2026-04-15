/** 宝妈记录类型与后端 mother_records 对齐。 */
export const MOTHER_RECORD_TYPES = [
  { value: "checkup", label: "产检" },
  { value: "weight", label: "体重" },
  { value: "blood_pressure", label: "血压" },
  { value: "blood_sugar", label: "血糖" },
  { value: "symptom", label: "症状" },
  { value: "medication", label: "用药" },
  { value: "nutrition", label: "营养补充" },
  { value: "mood", label: "心情" },
  { value: "recovery", label: "产后恢复" },
  { value: "followup", label: "复诊" },
] as const;

export function labelForMotherType(type: string): string {
  const hit = MOTHER_RECORD_TYPES.find((item) => item.value === type);
  return hit ? hit.label : type;
}
