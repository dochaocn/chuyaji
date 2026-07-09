export const babyRecordTypeLabels: Record<string, string> = {
  prenatal_checkup: '产检',
  ultrasound: 'B超',
  screening: '筛查',
  ultrasound_4d: '四维',
  glucose_test: '糖耐',
  nt: 'NT',
  feeding: '喂养',
  sleep: '睡眠',
  diaper: '排便',
  growth: '生长',
  development: '发育',
  checkup: '体检',
  vaccine: '疫苗',
  illness: '疾病',
}

export const motherRecordTypeLabels: Record<string, string> = {
  checkup: '产检',
  followup: '复诊',
  nutrition: '营养补充',
  medication: '用药',
  symptom: '不适',
  weight: '体重',
  blood_pressure: '血压',
  blood_sugar: '血糖',
  mood: '心情',
  recovery: '产后恢复',
}

export const phaseLabels: Record<string, string> = {
  prenatal: '怀孕期',
  postnatal: '成长期',
}

export const babyRecordTypePhase: Record<string, string> = {
  prenatal_checkup: 'prenatal',
  ultrasound: 'prenatal',
  screening: 'prenatal',
  ultrasound_4d: 'prenatal',
  glucose_test: 'prenatal',
  nt: 'prenatal',
  feeding: 'postnatal',
  sleep: 'postnatal',
  diaper: 'postnatal',
  growth: 'postnatal',
  development: 'postnatal',
  checkup: 'postnatal',
  vaccine: 'postnatal',
  illness: 'postnatal',
}

export function babyTypeLabel(type: string): string {
  return babyRecordTypeLabels[type] || type
}

export function motherTypeLabel(type: string): string {
  return motherRecordTypeLabels[type] || type
}
