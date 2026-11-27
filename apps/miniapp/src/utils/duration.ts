/** 把分钟数格式化成 h/m 组合：50 → 50m、120 → 2h、150 → 2h30m */
export function formatDurationMinutes(minutes: number | undefined | null): string {
  const total = Math.max(0, Math.round(Number(minutes) || 0));
  if (total <= 0) return "0m";
  const h = Math.floor(total / 60);
  const m = total % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h${m}m`;
}
