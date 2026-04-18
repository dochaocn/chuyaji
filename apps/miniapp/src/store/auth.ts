import { defineStore } from "pinia";
import { request } from "@/api/http";
import { useSessionStore } from "@/store/session";

interface LoginResp {
  token: string;
  expires_in: number;
  user: { id: number; nickname: string; avatar_url: string };
}

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: "" as string,
    userId: 0 as number,
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
        if (t) this.token = t;
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
    logout() {
      this.token = "";
      this.userId = 0;
      try {
        uni.removeStorageSync("chuyaji_token");
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
      } catch {}
    },
  },
});
