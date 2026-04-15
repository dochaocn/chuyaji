/** 与前端表单、后端 record_type 字符串对齐（自由扩展） */
export const RECORD_TYPES = {
  prenatal: [
    { value: "ultrasound", label: "B 超" },
    { value: "routine_checkup", label: "产检" },
    { value: "lab", label: "化验/筛查" },
    { value: "note", label: "备忘" },
  ],
  postnatal: [
    { value: "growth", label: "生长" },
    { value: "feeding", label: "喂养" },
    { value: "sleep", label: "睡眠" },
    { value: "temperature", label: "体温" },
    { value: "vaccine", label: "疫苗" },
    { value: "checkup", label: "体检" },
    { value: "note", label: "备忘" },
  ],
} as const;

export function labelForType(phase: "prenatal" | "postnatal", type: string): string {
  const list = RECORD_TYPES[phase];
  const hit = list.find((x) => x.value === type);
  return hit ? hit.label : type;
}
