import { API_BASE } from "@/api/config";

/** 将后端返回的附件地址转为小程序/H5 可直接加载的绝对地址 */
export function resolvePublicMediaUrl(url: string | undefined | null): string {
  const u = (url || "").trim();
  if (!u) return "";
  if (/^(https?:|data:|blob:|wxfile:|file:)/i.test(u)) return u;
  // 本地临时路径等（非站点根路径）原样返回，避免误拼 API_BASE
  if (!u.startsWith("/")) return u;
  const base = API_BASE.replace(/\/$/, "");
  return `${base}${u}`;
}
