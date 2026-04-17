/**
 * 孕周计算工具
 *
 * 口径约定（与医生一致）：
 * 1. 优先按 EDD（预产期）反推当前孕周。
 * 2. 无 EDD 时，用 LMP（末次月经首日）+ 280 天推算 EDD 后再反推。
 *
 * EDD 对应孕 40 周 0 天，即 280 天。
 * 当前孕周 = (280 - (EDD - today)) / 7，取整数周 + 余数天。
 */

export type GestationResult =
  | { available: false }
  | { available: true; weeks: number; days: number; totalDays: number };

/**
 * 根据 eddDateStr（ISO 日期字符串或 yyyy-MM-dd）计算今天的孕周。
 */
export function calcGestationByEdd(eddDateStr: string | undefined | null): GestationResult {
  if (!eddDateStr) return { available: false };
  const edd = parseDateOnly(eddDateStr);
  if (!edd) return { available: false };
  const today = todayMidnight();
  const daysUntilEdd = Math.round((edd.getTime() - today.getTime()) / MS_PER_DAY);
  const totalPregnancyDays = 280;
  const daysPregnant = totalPregnancyDays - daysUntilEdd;
  if (daysPregnant < 0 || daysPregnant > 310) return { available: false };
  const weeks = Math.floor(daysPregnant / 7);
  const days = daysPregnant % 7;
  return { available: true, weeks, days, totalDays: daysPregnant };
}

/**
 * 根据 lmpDateStr（末次月经首日）推算 EDD 后再算孕周。
 * 仅在无 EDD 时作为回退使用。
 */
export function calcGestationByLmp(lmpDateStr: string | undefined | null): GestationResult {
  if (!lmpDateStr) return { available: false };
  const lmp = parseDateOnly(lmpDateStr);
  if (!lmp) return { available: false };
  const edd = new Date(lmp.getTime() + 280 * MS_PER_DAY);
  const eddStr = formatYmd(edd);
  return calcGestationByEdd(eddStr);
}

/**
 * 统一入口：优先 EDD，回退 LMP。
 */
export function calcGestation(
  eddDateStr: string | undefined | null,
  lmpDateStr?: string | undefined | null
): GestationResult {
  const byEdd = calcGestationByEdd(eddDateStr);
  if (byEdd.available) return byEdd;
  return calcGestationByLmp(lmpDateStr);
}

/**
 * 宝宝出生后的天龄。
 * @returns 天数（≥0），或 null（未出生 / 无数据）
 */
export function calcAgeDays(birthDateStr: string | undefined | null): number | null {
  if (!birthDateStr) return null;
  const birth = parseDateOnly(birthDateStr);
  if (!birth) return null;
  const today = todayMidnight();
  const diff = Math.round((today.getTime() - birth.getTime()) / MS_PER_DAY);
  return diff >= 0 ? diff : null;
}

/**
 * 是否处于成长期（已出生）。
 */
export function isPostnatal(birthDateStr: string | undefined | null): boolean {
  const days = calcAgeDays(birthDateStr);
  return days !== null && days >= 0;
}

/**
 * 月龄（整数月）。
 */
export function calcAgeMonths(birthDateStr: string | undefined | null): number | null {
  const days = calcAgeDays(birthDateStr);
  if (days === null) return null;
  return Math.floor(days / 30);
}

// ---------------------------------------------------------------------------
// 内部工具
// ---------------------------------------------------------------------------

const MS_PER_DAY = 24 * 60 * 60 * 1000;

function todayMidnight(): Date {
  const d = new Date();
  d.setHours(0, 0, 0, 0);
  return d;
}

function parseDateOnly(str: string): Date | null {
  const trimmed = str.trim().slice(0, 10);
  const parts = trimmed.split("-").map(Number);
  if (parts.length !== 3 || parts.some((n) => !Number.isFinite(n))) return null;
  const [y, m, d] = parts;
  const dt = new Date(y, m - 1, d);
  if (Number.isNaN(dt.getTime())) return null;
  return dt;
}

function formatYmd(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}
