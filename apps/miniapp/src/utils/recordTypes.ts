/** 与前端表单、后端 record_type 字符串对齐（自由扩展） */
export const RECORD_TYPES = {
  prenatal: [
    { value: "prenatal_checkup", label: "产检" },
    { value: "ultrasound", label: "B 超" },
    { value: "nt", label: "NT" },
    { value: "screening", label: "筛查" },
    { value: "glucose_test", label: "糖耐" },
    { value: "symptom", label: "不适" },
    { value: "medication", label: "用药" },
  ],
  postnatal: [
    { value: "growth", label: "生长" },
    { value: "checkup", label: "体检" },
    { value: "vaccine", label: "疫苗" },
    { value: "feeding", label: "喂养" },
    { value: "sleep", label: "睡眠" },
    { value: "diaper", label: "排便" },
    { value: "illness", label: "疾病" },
    { value: "development", label: "发育" },
  ],
} as const;

export function labelForType(phase: "prenatal" | "postnatal", type: string): string {
  const list = RECORD_TYPES[phase];
  const hit = list.find((x) => x.value === type);
  return hit ? hit.label : type;
}
