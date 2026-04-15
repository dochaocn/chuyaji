import { request } from "./http";

export interface Baby {
  id: number;
  user_id: number;
  nickname: string;
  gender?: string;
  avatar_url?: string;
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

export interface RecordItem {
  id: number;
  baby_id: number;
  phase: "prenatal" | "postnatal";
  record_type: string;
  occurred_at: string;
  gestational_weeks?: number;
  gestational_days?: number;
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

export async function apiMe() {
  return request<{ id: number; nickname: string; avatar_url: string }>({ path: "/api/v1/me" });
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

export async function apiBabyDashboard(babyId?: number) {
  const path = babyId ? `/api/v1/dashboard/baby?baby_id=${babyId}` : "/api/v1/dashboard/baby";
  return request<{
    profile: Baby | null;
    phase_summary?: { stage: "prenatal" | "postnatal"; record_count: number };
    latest_records: RecordItem[];
    growth_summary?: Record<string, unknown>;
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
    health_summary?: Record<string, unknown>;
    latest_records: MotherRecordItem[];
  }>({ path });
}

export async function apiListRecords(babyId: number, limit = 20, cursor?: string) {
  let path = `/api/v1/babies/${babyId}/records?limit=${limit}`;
  if (cursor) path += `&cursor=${encodeURIComponent(cursor)}`;
  return request<{ items: RecordItem[]; next_cursor: string }>({ path });
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

export async function apiListMotherRecords(motherId: number, limit = 20, cursor?: string) {
  let path = `/api/v1/mothers/${motherId}/records?limit=${limit}`;
  if (cursor) path += `&cursor=${encodeURIComponent(cursor)}`;
  return request<{ items: MotherRecordItem[]; next_cursor: string }>({ path });
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

export async function apiListMotherAttachments(recordId: number) {
  return request<{ items: AttachmentItem[] }>({
    path: `/api/v1/mother-records/${recordId}/attachments`,
  });
}

export async function apiDeleteAttachment(id: number) {
  return request<unknown>({ path: `/api/v1/attachments/${id}`, method: "DELETE" });
}
