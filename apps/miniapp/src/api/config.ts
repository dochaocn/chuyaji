/** 后端 API 根地址，开发可在 .env.development 中配置 VITE_CHUYAJI_API_BASE */
export const API_BASE = import.meta.env.VITE_CHUYAJI_API_BASE || "http://127.0.0.1:8282";

/**
 * 是否走「开发登录」（请求 body 为 code=dev，须后端 CHUYAJI_DEV_MODE=true）。
 * 默认始终使用 uni.login 返回的真实 code，与生产一致。
 * 仅当显式设置 `VITE_DEV_WECHAT_LOGIN=true`（或 1 / yes）时才发 code=dev，供无微信换票时的本地联调。
 */
export function isDevWechatLogin(): boolean {
  const v = import.meta.env.VITE_DEV_WECHAT_LOGIN;
  return v === "true" || v === "1" || v === "yes";
}
