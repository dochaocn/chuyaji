import type { BodyMetricType } from "./motherRecordTypes";

/** 与宝妈首页头部时期徽章一致：用于最近记录等处的时期标签 */
export function motherStageLabel(status?: "pregnant" | "postpartum" | "parenting"): string {
  if (status === "postpartum") return "产后恢复";
  if (status === "parenting") return "育儿期";
  return "怀孕中";
}

export function motherStageBadgeClass(status?: "pregnant" | "postpartum" | "parenting"): "badge--pre" | "badge--post" {
  return status === "pregnant" ? "badge--pre" : "badge--post";
}

export type NextStepAction =
  | { type: "none" }
  | { type: "create_profile" }
  | { type: "add_record"; daysSinceLast: number }
  | { type: "check_reminder" };

const STALE_RECORD_DAYS = 7;

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

function daysSinceLastRecord(lastRecordAt: string | null | undefined): number | null {
  if (!lastRecordAt) return null;
  const last = Date.parse(lastRecordAt);
  if (!Number.isFinite(last)) return null;
  const now = Date.now();
  return Math.floor((now - last) / (24 * 60 * 60 * 1000));
}
