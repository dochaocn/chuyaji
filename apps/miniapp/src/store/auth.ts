import { defineStore } from "pinia";
import { request } from "@/api/http";
import { isDevWechatLogin } from "@/api/config";
import { useSessionStore } from "@/store/session";
import { clearWeChatLoginCodeCache, getWeChatLoginCode } from "@/utils/wechatLoginCode";

interface LoginResp {
  token: string;
  expires_in: number;
  user: { id: number; nickname: string; avatar_url: string };
}

let ensureWeChatSessionPromise: Promise<void> | null = null;

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: "" as string,
    userId: 0 as number,
    /** 无本地 token 时先为 true，loadToken 读到 token 或静默换票结束后会置 false */
    sessionBooting: true as boolean,
    sessionBootError: "" as string,
  }),
  actions: {
    setToken(t: string) {
      this.token = t;
      try {
        uni.setStorageSync("chuyaji_token", t);
      } catch {}
    },
    loadToken() {
      try {
        const t = uni.getStorageSync("chuyaji_token") as string;
        if (t) {
          this.token = t;
          this.sessionBooting = false;
          this.sessionBootError = "";
        }
      } catch {}
    },
    async loginWithWeChatCode(code: string) {
      const data = await request<LoginResp>({
        path: "/api/v1/auth/wechat",
        method: "POST",
        data: { code },
        token: "",
      });
      this.setToken(data.token);
      this.userId = data.user.id;
      useSessionStore().clearFamilyCache();
      return data;
    },
    /**
     * 启动时完成小程序与后端的就绪流程（uni.login + 后端接口）。多入口并发共用同一 Promise。
     */
    async ensureWeChatSession() {
      this.loadToken();
      if (this.token) {
        this.sessionBootError = "";
        return;
      }
      if (ensureWeChatSessionPromise) {
        await ensureWeChatSessionPromise;
        return;
      }
      this.sessionBooting = true;
      this.sessionBootError = "";
      ensureWeChatSessionPromise = (async () => {
        try {
          if (isDevWechatLogin()) {
            await this.loginWithWeChatCode("dev");
          } else {
            const code = await getWeChatLoginCode();
            await this.loginWithWeChatCode(code);
          }
          this.sessionBootError = "";
        } catch (e) {
          console.error(e);
          this.sessionBootError = "运行环境准备失败，请检查网络后重试";
          this.token = "";
          this.userId = 0;
          try {
            uni.removeStorageSync("chuyaji_token");
          } catch {}
          useSessionStore().clearFamilyCache();
        } finally {
          this.sessionBooting = false;
          ensureWeChatSessionPromise = null;
        }
      })();
      await ensureWeChatSessionPromise;
    },
    logout() {
      this.token = "";
      this.userId = 0;
      this.sessionBootError = "";
      this.sessionBooting = true;
      clearWeChatLoginCodeCache();
      try {
        uni.removeStorageSync("chuyaji_token");
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
      } catch {}
    },
  },
});
