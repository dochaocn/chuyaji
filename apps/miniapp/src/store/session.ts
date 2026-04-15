import { defineStore } from "pinia";

export const useSessionStore = defineStore("session", {
  state: () => ({
    babyId: 0 as number,
    motherId: 0 as number,
    privacyOk: false as boolean,
  }),
  actions: {
    load() {
      try {
        this.babyId = Number(uni.getStorageSync("chuyaji_baby_id") || 0);
        this.motherId = Number(uni.getStorageSync("chuyaji_mother_id") || 0);
        this.privacyOk = !!uni.getStorageSync("chuyaji_privacy_ok");
      } catch {
        /* ignore */
      }
    },
    setBaby(id: number) {
      this.babyId = id;
      uni.setStorageSync("chuyaji_baby_id", id);
    },
    setMother(id: number) {
      this.motherId = id;
      uni.setStorageSync("chuyaji_mother_id", id);
    },
    setPrivacyOk() {
      this.privacyOk = true;
      uni.setStorageSync("chuyaji_privacy_ok", 1);
    },
    /** 新登录成功时调用：避免沿用上一账号在本机缓存的 baby_id / mother_id */
    clearFamilyCache() {
      this.babyId = 0;
      this.motherId = 0;
      try {
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
      } catch {
        /* ignore */
      }
    },
  },
});
