/** 根据末次月经 LMP 与事件日期计算孕周（前端展示用，非医疗结论） */
export function weeksFromLMP(lmpISO: string, atISO: string): { weeks: number; days: number } | null {
  const lmp = Date.parse(lmpISO);
  const at = Date.parse(atISO);
  if (!Number.isFinite(lmp) || !Number.isFinite(at) || at < lmp) return null;
  const ms = at - lmp;
  const daysTotal = Math.floor(ms / (24 * 3600 * 1000));
  const weeks = Math.floor(daysTotal / 7);
  const days = daysTotal % 7;
  return { weeks, days };
}
