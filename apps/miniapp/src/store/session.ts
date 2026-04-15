import { defineStore } from "pinia";

export const useSessionStore = defineStore("session", {
  state: () => ({
    familyId: 0 as number,
    babyId: 0 as number,
    privacyOk: false as boolean,
  }),
  actions: {
    load() {
      try {
        this.familyId = Number(uni.getStorageSync("chuyaji_family_id") || 0);
        this.babyId = Number(uni.getStorageSync("chuyaji_baby_id") || 0);
        this.privacyOk = !!uni.getStorageSync("chuyaji_privacy_ok");
      } catch {
        /* ignore */
      }
    },
    setFamily(id: number) {
      this.familyId = id;
      uni.setStorageSync("chuyaji_family_id", id);
    },
    setBaby(id: number) {
      this.babyId = id;
      uni.setStorageSync("chuyaji_baby_id", id);
    },
    setPrivacyOk() {
      this.privacyOk = true;
      uni.setStorageSync("chuyaji_privacy_ok", 1);
    },
  },
});
