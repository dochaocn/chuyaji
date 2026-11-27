import { request } from "./http";

export interface Baby {
  id: number;
  user_id: number;
  family_id?: number;
  nickname: string;
  gender?: string;
  lmp_date?: string;
  edd_date?: string;
  birth_date?: string;
  birth_weight_g?: number;
  birth_height_cm?: number;
  birth_hospital?: string;
  feeding_type?: string;
  note?: string;
}

export interface Mother {
  id: number;
  user_id: number;
  family_id?: number;
  name: string;
  birthday?: string;
  height_cm?: number;
  pre_pregnancy_weight_kg?: number;
  blood_type?: string;
  allergy_history?: string;
  medical_history?: string;
  status?: "pregnant" | "postpartum" | "parenting";
  delivery_date?: string;
  note?: string;
}

export type FamilyRole = "owner" | "write" | "read";

export interface FamilyMember {
  user_id: number;
  nickname: string;
  avatar_url?: string;
  role: FamilyRole;
  created_at: string;
}

export interface Family {
  id: number;
  name: string;
  created_by: number;
  my_user_id?: number;
  my_role: FamilyRole;
  members: FamilyMember[];
}

export type FamilyInfo = Family;

export interface FamilyInviteConflict {
  has_existing_family: boolean;
  family_name?: string;
  baby_count: number;
  mother_count: number;
  has_other_members: boolean;
  my_role?: string;
}

export interface FamilyInvitePreview {
  family_name: string;
  inviter_nickname: string;
  role: "write" | "read";
  expires_at: string;
  member_count: number;
  already_member: boolean;
  conflict?: FamilyInviteConflict;
}

export interface RecordItem {
  id: number;
  baby_id: number;
  phase: "prenatal" | "postnatal";
  record_type: string;
  occurred_at: string;
  summary: string;
  payload: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface MotherRecordItem {
  id: number;
  mother_id: number;
  record_type: string;
  occurred_at: string;
  summary: string;
  payload: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface AttachmentItem {
  id: number;
  owner_type: "baby_record" | "mother_record";
  owner_id: number;
  url: string;
  thumb_url?: string;
  sort_order: number;
  size: number;
}

export interface BabyAlbumAttachment {
  id: number;
  url: string;
  thumb_url?: string;
  created_at: string;
  record_id: number;
  record_type: string;
  occurred_at: string;
  summary: string;
}

export interface ReminderItem {
  id: number;
  user_id: number;
  owner_type: "baby" | "mother";
  owner_id: number;
  source_type: "baby_record" | "mother_record";
  source_id: number;
  source_record_type: string;
  category?: "checkup" | "vaccine" | "followup" | "other" | string;
  title: string;
  note?: string;
  due_at: string;
  done_at?: string;
  snoozed_until?: string;
  status: "pending" | "done" | "ignored";
  created_at: string;
  updated_at: string;
}

export interface RecordListOptions {
  limit?: number;
  cursor?: string;
  phase?: "prenatal" | "postnatal";
  record_type?: string;
  q?: string;
  from?: string;
  to?: string;
}

export interface MotherRecordListOptions {
  limit?: number;
  cursor?: string;
  record_type?: string;
  q?: string;
  from?: string;
  to?: string;
}

export type GrowthMetricKey = "weight" | "height" | "head";

export interface GrowthSeriesPoint {
  record_id: number;
  age_days: number;
  occurred_at: string;
  value: number;
  unit: string;
  percentile?: number;
}

export interface GrowthRefPoint {
  age_days: number;
  p3: number;
  p50: number;
  p97: number;
}

export interface GrowthSeriesResponse {
  baby_id: number;
  birth_date?: string;
  gender?: string;
  standard?: "who";
  hint?: string;
  series: Record<GrowthMetricKey, GrowthSeriesPoint[]>;
  reference?: Record<GrowthMetricKey, GrowthRefPoint[]>;
}


export async function apiMe() {
  return request<{ id: number; nickname: string; avatar_url: string }>({ path: "/api/v1/me" });
}

export async function apiPatchMe(body: { nickname: string }) {
  return request<{ id: number; nickname: string; avatar_url: string }>({
    path: "/api/v1/me",
    method: "PATCH",
    data: body,
  });
}

export async function apiListBabies() {
  return request<{ items: Baby[] }>({ path: "/api/v1/babies" });
}

export async function apiCreateBaby(body: Partial<Baby> & { nickname: string }) {
  return request<Baby>({ path: "/api/v1/babies", method: "POST", data: body as Record<string, unknown> });
}

export async function apiGetBaby(id: number) {
  return request<Baby>({ path: `/api/v1/babies/${id}` });
}

export async function apiPatchBaby(id: number, body: Record<string, unknown>) {
  return request<Baby>({ path: `/api/v1/babies/${id}`, method: "PATCH", data: body });
}

export async function apiDeleteBaby(id: number) {
  return request<unknown>({ path: `/api/v1/babies/${id}`, method: "DELETE" });
}

export type DailyMetricBucket = {
  count: number;
  total_ml?: number;
  total_duration_min?: number;
  total_times?: number;
};

export type BabyDailySummary = {
  date: string;
  feeding: DailyMetricBucket;
  sleep: DailyMetricBucket;
  diaper: DailyMetricBucket;
};

export async function apiBabyDashboard(babyId?: number) {
  const path = babyId ? `/api/v1/dashboard/baby?baby_id=${babyId}` : "/api/v1/dashboard/baby";
  return request<{
    profile: Baby | null;
    phase_summary?: { stage: "prenatal" | "postnatal"; record_count: number };
    latest_records: RecordItem[];
    growth_summary?: Record<string, unknown>;
    daily_summary?: BabyDailySummary;
  }>({ path });
}

export async function apiListMothers() {
  return request<{ items: Mother[] }>({ path: "/api/v1/mothers" });
}

export async function apiCreateMother(body: Partial<Mother>) {
  return request<Mother>({ path: "/api/v1/mothers", method: "POST", data: body as Record<string, unknown> });
}

export async function apiGetMother(id: number) {
  return request<Mother>({ path: `/api/v1/mothers/${id}` });
}

export async function apiPatchMother(id: number, body: Record<string, unknown>) {
  return request<Mother>({ path: `/api/v1/mothers/${id}`, method: "PATCH", data: body });
}

export async function apiMotherDashboard(motherId?: number) {
  const path = motherId ? `/api/v1/dashboard/mother?mother_id=${motherId}` : "/api/v1/dashboard/mother";
  return request<{
    profile: Mother | null;
    health_summary?: { status?: string; record_count?: number };
    latest_records: MotherRecordItem[];
  }>({ path });
}

function queryString(params: Record<string, unknown>) {
  const pairs: string[] = [];
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    pairs.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
  }
  return pairs.join("&");
}

export async function apiListRecords(babyId: number, limitOrOptions: number | RecordListOptions = 20, cursor?: string) {
  const opts: RecordListOptions =
    typeof limitOrOptions === "number" ? { limit: limitOrOptions, cursor } : limitOrOptions;
  const qs = queryString({ limit: opts.limit ?? 20, cursor: opts.cursor, phase: opts.phase, record_type: opts.record_type, q: opts.q, from: opts.from, to: opts.to });
  const path = `/api/v1/babies/${babyId}/records?${qs}`;
  return request<{ items: RecordItem[]; next_cursor: string }>({ path });
}

export async function apiLatestRecord(
  babyId: number,
  recordType: string,
  phase?: "prenatal" | "postnatal"
) {
  const qs = queryString({ record_type: recordType, phase });
  return request<{ item: RecordItem | null }>({ path: `/api/v1/babies/${babyId}/records/latest?${qs}` });
}

export async function apiCreateRecord(babyId: number, body: Record<string, unknown>) {
  return request<RecordItem>({ path: `/api/v1/babies/${babyId}/records`, method: "POST", data: body });
}

export async function apiGetRecord(id: number) {
  return request<RecordItem>({ path: `/api/v1/records/${id}` });
}

export async function apiPatchRecord(id: number, body: Record<string, unknown>) {
  return request<RecordItem>({ path: `/api/v1/records/${id}`, method: "PATCH", data: body });
}

export async function apiDeleteRecord(id: number) {
  return request<unknown>({ path: `/api/v1/records/${id}`, method: "DELETE" });
}

export async function apiListMotherRecords(motherId: number, limitOrOptions: number | MotherRecordListOptions = 20, cursor?: string) {
  const opts: MotherRecordListOptions =
    typeof limitOrOptions === "number" ? { limit: limitOrOptions, cursor } : limitOrOptions;
  const qs = queryString({ limit: opts.limit ?? 20, cursor: opts.cursor, record_type: opts.record_type, q: opts.q, from: opts.from, to: opts.to });
  const path = `/api/v1/mothers/${motherId}/records?${qs}`;
  return request<{ items: MotherRecordItem[]; next_cursor: string }>({ path });
}

export async function apiLatestMotherRecord(motherId: number, recordType: string) {
  const qs = queryString({ record_type: recordType });
  return request<{ item: MotherRecordItem | null }>({ path: `/api/v1/mothers/${motherId}/records/latest?${qs}` });
}

export async function apiCreateMotherRecord(motherId: number, body: Record<string, unknown>) {
  return request<MotherRecordItem>({ path: `/api/v1/mothers/${motherId}/records`, method: "POST", data: body });
}

export async function apiGetMotherRecord(id: number) {
  return request<MotherRecordItem>({ path: `/api/v1/mother-records/${id}` });
}

export async function apiPatchMotherRecord(id: number, body: Record<string, unknown>) {
  return request<MotherRecordItem>({ path: `/api/v1/mother-records/${id}`, method: "PATCH", data: body });
}

export async function apiDeleteMotherRecord(id: number) {
  return request<unknown>({ path: `/api/v1/mother-records/${id}`, method: "DELETE" });
}

export async function apiListAttachments(recordId: number) {
  return request<{ items: AttachmentItem[] }>({
    path: `/api/v1/records/${recordId}/attachments`,
  });
}

export async function apiListBabyAttachments(babyId: number, limit = 40, cursor?: string) {
  const qs = queryString({ limit, cursor });
  return request<{ items: BabyAlbumAttachment[]; next_cursor: string }>({
    path: `/api/v1/babies/${babyId}/attachments?${qs}`,
  });
}

export async function apiListMotherAttachments(recordId: number) {
  return request<{ items: AttachmentItem[] }>({
    path: `/api/v1/mother-records/${recordId}/attachments`,
  });
}

export async function apiDeleteAttachment(id: number) {
  return request<unknown>({ path: `/api/v1/attachments/${id}`, method: "DELETE" });
}

export async function apiListReminders(
  status: ReminderItem["status"] = "pending",
  limit = 20,
  owner?: { owner_type: ReminderItem["owner_type"]; owner_id: number },
  extra?: { from?: string; to?: string; category?: string }
) {
  const params = [`status=${status}`, `limit=${limit}`];
  if (owner) {
    params.push(`owner_type=${owner.owner_type}`, `owner_id=${owner.owner_id}`);
  }
  if (extra?.from) params.push(`from=${encodeURIComponent(extra.from)}`);
  if (extra?.to) params.push(`to=${encodeURIComponent(extra.to)}`);
  if (extra?.category) params.push(`category=${encodeURIComponent(extra.category)}`);
  return request<{ items: ReminderItem[] }>({ path: `/api/v1/reminders?${params.join("&")}` });
}

export async function apiPatchReminder(id: number, body: { status?: ReminderItem["status"]; due_at?: string; snoozed_until?: string; note?: string }) {
  return request<ReminderItem>({ path: `/api/v1/reminders/${id}`, method: "PATCH", data: body });
}

export async function apiDeleteReminder(id: number) {
  return request<unknown>({ path: `/api/v1/reminders/${id}`, method: "DELETE" });
}

export async function apiGrowthSeries(babyId: number) {
  return request<GrowthSeriesResponse>({ path: `/api/v1/babies/${babyId}/growth-series` });
}

export async function apiGetCurrentFamily() {
  return request<FamilyInfo>({ path: "/api/v1/families/current" });
}

export async function apiCreateCurrentFamily() {
  return request<FamilyInfo>({ path: "/api/v1/families/current", method: "POST" });
}

export async function apiPatchCurrentFamily(body: { name: string }) {
  return request<FamilyInfo>({ path: "/api/v1/families/current", method: "PATCH", data: body });
}

export async function apiLeaveCurrentFamily() {
  return request<unknown>({ path: "/api/v1/families/current/leave", method: "POST" });
}

export async function apiCreateFamilyInvite(role: "write" | "read") {
  return request<{ token: string; path: string; role: string; expires_at: string }>({
    path: "/api/v1/families/current/invites",
    method: "POST",
    data: { role },
  });
}

export async function apiPreviewFamilyInvite(token: string) {
  return request<FamilyInvitePreview>({
    path: `/api/v1/invites/${encodeURIComponent(token)}/preview`,
  });
}

export async function apiAcceptFamilyInvite(token: string) {
  return request<FamilyInfo>({ path: `/api/v1/invites/${encodeURIComponent(token)}/accept`, method: "POST" });
}

export async function apiPatchFamilyMember(userId: number, role: FamilyRole) {
  return request<FamilyInfo>({
    path: `/api/v1/families/current/members/${userId}`,
    method: "PATCH",
    data: { role },
  });
}

export async function apiTransferFamilyOwner(userId: number) {
  return request<FamilyInfo>({
    path: `/api/v1/families/current/members/${userId}/transfer-owner`,
    method: "POST",
  });
}

export async function apiPatchFamilyMemberNickname(userId: number, nickname: string) {
  return request<FamilyInfo>({
    path: `/api/v1/families/current/members/${userId}/nickname`,
    method: "PATCH",
    data: { nickname },
  });
}

export async function apiRemoveFamilyMember(userId: number) {
  return request<unknown>({ path: `/api/v1/families/current/members/${userId}`, method: "DELETE" });
}

