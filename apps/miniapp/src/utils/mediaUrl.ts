import { API_BASE } from "@/api/config";

export function resolvePublicMediaUrl(url: string | undefined | null): string {
  const u = (url || "").trim();
  if (!u) return "";
  if (/^(https?:|data:|blob:|wxfile:|file:)/i.test(u)) return u;
  if (!u.startsWith("/")) return u;
  const base = API_BASE.replace(/\/$/, "");
  return `${base}${u}`;
}
