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
      } catch {}
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
    clearFamilyCache() {
      this.babyId = 0;
      this.motherId = 0;
      try {
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
      } catch {}
    },
  },
});
