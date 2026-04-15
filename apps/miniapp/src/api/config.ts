/** 后端 API 根地址，开发可在 .env.development 中配置 VITE_CHUYAJI_API_BASE */
export const API_BASE = import.meta.env.VITE_CHUYAJI_API_BASE || "http://127.0.0.1:8282";

/**
 * 首页是否走「开发登录」（请求 body 为 code=dev，须后端 CHUYAJI_DEV_MODE=true）。
 * - 显式 `VITE_DEV_WECHAT_LOGIN=false`：始终用真实 wx.login code（易遇 40029，需 AppID/Secret 与小程序一致）。
 * - 显式 `true` / 未设置且为 Vite 开发编译：走 dev，避免本地联调被微信换票卡住。
 * - 生产构建（npm run build）：未设置时为 false。
 */
export function isDevWechatLogin(): boolean {
  const v = import.meta.env.VITE_DEV_WECHAT_LOGIN;
  if (v === "false" || v === "0") {
    return false;
  }
  if (v === "true" || v === "1" || v === "yes") {
    return true;
  }
  return import.meta.env.DEV === true;
}
