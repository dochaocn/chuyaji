export const API_BASE = import.meta.env.VITE_CHUYAJI_API_BASE || "http://127.0.0.1:8282";

export function isDevWechatLogin(): boolean {
  const v = import.meta.env.VITE_DEV_WECHAT_LOGIN;
  return v === "true" || v === "1" || v === "yes";
}

/**
 * 仅开发模式生效：用 VITE_CHUYAJI_DEV_TOKEN 注入一个已有 JWT，跳过微信登录。
 * 用于 H5 / 浏览器里预览真实数据（微信开发者工具之外的场景）。
 * 生产构建中 import.meta.env.DEV 为 false，此函数恒返回空串。
 */
export function devToken(): string {
  if (!import.meta.env.DEV) return "";
  const v = import.meta.env.VITE_CHUYAJI_DEV_TOKEN;
  return String(v || "").trim();
}
