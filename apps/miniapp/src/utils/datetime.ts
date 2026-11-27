/** 统一时间展示：yyyy-MM-dd HH:mm:ss（本地时区） */
export function formatDateTime(value: string | number | Date | undefined | null): string {
  if (value === undefined || value === null || value === "") return "";
  const d = value instanceof Date ? value : new Date(value as string | number);
  if (Number.isNaN(d.getTime())) {
    return String(value).replace("T", " ").slice(0, 19);
  }
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(
    d.getSeconds()
  )}`;
}
