export const API_BASE = import.meta.env.VITE_CHUYAJI_API_BASE || "http://127.0.0.1:8282";

export function isDevWechatLogin(): boolean {
  const v = import.meta.env.VITE_DEV_WECHAT_LOGIN;
  return v === "true" || v === "1" || v === "yes";
}
