/** 统计与录入统一按 Asia/Shanghai（UTC+8）日历，不跟设备时区走。 */
const SHANGHAI_OFFSET_MS = 8 * 60 * 60 * 1000;

export type ShanghaiParts = {
  y: number;
  m: number;
  d: number;
  hh: number;
  mm: number;
  ss: number;
};

export function shanghaiParts(value: string | number | Date | undefined | null): ShanghaiParts | null {
  if (value === undefined || value === null || value === "") return null;
  const date = value instanceof Date ? value : new Date(value as string | number);
  if (Number.isNaN(date.getTime())) return null;
  const shifted = new Date(date.getTime() + SHANGHAI_OFFSET_MS);
  return {
    y: shifted.getUTCFullYear(),
    m: shifted.getUTCMonth() + 1,
    d: shifted.getUTCDate(),
    hh: shifted.getUTCHours(),
    mm: shifted.getUTCMinutes(),
    ss: shifted.getUTCSeconds(),
  };
}

function pad(n: number) {
  return String(n).padStart(2, "0");
}

/** 统一时间展示：yyyy-MM-dd HH:mm:ss（Asia/Shanghai） */
export function formatDateTime(value: string | number | Date | undefined | null): string {
  const parts = shanghaiParts(value);
  if (!parts) {
    if (value === undefined || value === null || value === "") return "";
    return String(value).replace("T", " ").slice(0, 19);
  }
  return `${parts.y}-${pad(parts.m)}-${pad(parts.d)} ${pad(parts.hh)}:${pad(parts.mm)}:${pad(parts.ss)}`;
}

/** 表单上的年月日、时分是上海墙钟，转成绝对瞬间。不用无时区 ISO 字符串，避免部分 WebView 把它当成 UTC。 */
export function shanghaiWallClockToISO(ymd: string, hm: string): string {
  const [y, m, d] = ymd.split("-").map(Number);
  const [hh, mm] = (hm || "00:00").split(":").map(Number);
  const utc = Date.UTC(y, (m || 1) - 1, d || 1, (hh || 0) - 8, mm || 0, 0);
  return new Date(utc).toISOString();
}

export function shanghaiMonthKey(value: string | number | Date | undefined | null): string {
  const parts = shanghaiParts(value);
  if (!parts) return String(value ?? "").slice(0, 7);
  return `${parts.y}-${pad(parts.m)}`;
}
