import { request } from "./http";

export interface Family {
  id: number;
  name: string;
  invite_code: string;
}

export interface Baby {
  id: number;
  family_id: number;
  nickname: string;
  lmp_date?: string;
  edd_date?: string;
  birth_date?: string;
  gender?: string;
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

export async function apiMe() {
  return request<{ id: number; nickname: string; avatar_url: string }>({ path: "/api/v1/me" });
}

export async function apiListFamilies() {
  return request<{ items: Family[] }>({ path: "/api/v1/families" });
}

export async function apiCreateFamily(name: string) {
  return request<Family>({ path: "/api/v1/families", method: "POST", data: { name } });
}

export async function apiJoinFamily(invite_code: string) {
  return request<Family>({ path: "/api/v1/families/join", method: "POST", data: { invite_code } });
}

export async function apiListBabies(familyId: number) {
  return request<{ items: Baby[] }>({
    path: `/api/v1/babies?family_id=${encodeURIComponent(String(familyId))}`,
  });
}

export async function apiCreateBaby(body: Partial<Baby> & { family_id: number; nickname: string }) {
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

export async function apiListAttachments(recordId: number) {
  return request<{ items: { id: number; url: string; thumb_url?: string }[] }>({
    path: `/api/v1/records/${recordId}/attachments`,
  });
}

export async function apiDeleteAttachment(id: number) {
  return request<unknown>({ path: `/api/v1/attachments/${id}`, method: "DELETE" });
}
