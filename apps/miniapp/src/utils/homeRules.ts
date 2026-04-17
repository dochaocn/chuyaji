/**
 * 首页"下一步卡片"触发规则 —— 纯函数，方便测试与后续扩展。
 *
 * 规则（三选一，最多触发一条）：
 * 1. 未建档          -> action: "create_profile"
 * 2. 很久没记（>7天） -> action: "add_record"
 * 3. 有到期事项      -> action: "check_reminder"（一期暂留位，提醒功能后续实现）
 */

import type { BODY_METRIC_TYPES, BodyMetricType } from "./motherRecordTypes";

/** 下一步卡片触发结果 */
export type NextStepAction =
  | { type: "none" }
  | { type: "create_profile" }
  | { type: "add_record"; daysSinceLast: number }
  | { type: "check_reminder" };

/** 距上次记录超过几天算"很久没记" */
const STALE_RECORD_DAYS = 7;

/**
 * 计算宝宝首页的下一步卡片触发状态。
 * @param hasProfile     是否已有宝宝档案
 * @param lastRecordAt   最近一条记录的日期字符串（ISO 或 yyyy-MM-dd），没有记录传 null
 */
export function getBabyNextStep(
  hasProfile: boolean,
  lastRecordAt: string | null | undefined
): NextStepAction {
  if (!hasProfile) return { type: "create_profile" };
  const days = daysSinceLastRecord(lastRecordAt);
  if (days !== null && days > STALE_RECORD_DAYS) {
    return { type: "add_record", daysSinceLast: days };
  }
  return { type: "none" };
}

/**
 * 计算宝妈首页的下一步卡片触发状态。
 */
export function getMotherNextStep(
  hasProfile: boolean,
  lastRecordAt: string | null | undefined
): NextStepAction {
  if (!hasProfile) return { type: "create_profile" };
  const days = daysSinceLastRecord(lastRecordAt);
  if (days !== null && days > STALE_RECORD_DAYS) {
    return { type: "add_record", daysSinceLast: days };
  }
  return { type: "none" };
}

/**
 * 宝宝成长期快捷入口按月龄选择。
 * - 0-6 月：喂养、睡眠、排便、生长
 * - 6 月后：喂养、睡眠、生长、疫苗
 */
export function getBabyPostnatalShortcuts(ageMonths: number): string[] {
  if (ageMonths < 6) {
    return ["feeding", "sleep", "diaper", "growth"];
  }
  return ["feeding", "sleep", "growth", "vaccine"];
}

/**
 * 宝妈"身体指标"入口：优先进入最近常用项，没有历史则默认 weight。
 * @param recentMotherRecords 按时间倒序的最近宝妈记录列表（只需要 record_type）
 */
export function getPreferredBodyMetricType(
  recentMotherRecords: { record_type: string }[]
): BodyMetricType {
  const metricTypes: BodyMetricType[] = ["weight", "blood_pressure", "blood_sugar"];
  for (const record of recentMotherRecords) {
    const t = record.record_type as BodyMetricType;
    if (metricTypes.includes(t)) return t;
  }
  return "weight";
}

// ---------------------------------------------------------------------------
// 内部工具
// ---------------------------------------------------------------------------

function daysSinceLastRecord(lastRecordAt: string | null | undefined): number | null {
  if (!lastRecordAt) return null;
  const last = Date.parse(lastRecordAt);
  if (!Number.isFinite(last)) return null;
  const now = Date.now();
  return Math.floor((now - last) / (24 * 60 * 60 * 1000));
}
