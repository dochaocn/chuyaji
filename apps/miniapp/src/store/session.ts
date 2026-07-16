import { defineStore } from "pinia";
import type { FamilyRole } from "@/api/chuyaji";
import { apiGetCurrentFamily, apiListBabies, apiListMothers } from "@/api/chuyaji";

export const useSessionStore = defineStore("session", {
  state: () => ({
    babyId: 0 as number,
    motherId: 0 as number,
    privacyOk: false as boolean,
    familyId: 0 as number,
    familyRole: "" as FamilyRole | "",
    pendingInviteToken: "" as string,
  }),
  getters: {
    canWrite(state): boolean {
      return state.familyRole === "owner" || state.familyRole === "write" || state.familyRole === "";
    },
    isOwner(state): boolean {
      return state.familyRole === "owner";
    },
  },
  actions: {
    load() {
      try {
        this.babyId = Number(uni.getStorageSync("chuyaji_baby_id") || 0);
        this.motherId = Number(uni.getStorageSync("chuyaji_mother_id") || 0);
        this.privacyOk = !!uni.getStorageSync("chuyaji_privacy_ok");
        this.familyId = Number(uni.getStorageSync("chuyaji_family_id") || 0);
        this.familyRole = (uni.getStorageSync("chuyaji_family_role") || "") as FamilyRole | "";
        this.pendingInviteToken = String(uni.getStorageSync("chuyaji_pending_invite") || "");
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
    setFamilyMeta(id: number, role: FamilyRole) {
      this.familyId = id;
      this.familyRole = role;
      uni.setStorageSync("chuyaji_family_id", id);
      uni.setStorageSync("chuyaji_family_role", role);
    },
    setPendingInvite(token: string) {
      this.pendingInviteToken = token;
      if (token) uni.setStorageSync("chuyaji_pending_invite", token);
      else {
        try {
          uni.removeStorageSync("chuyaji_pending_invite");
        } catch {}
      }
    },
    clearFamilyCache() {
      this.babyId = 0;
      this.motherId = 0;
      this.familyId = 0;
      this.familyRole = "";
      try {
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
        uni.removeStorageSync("chuyaji_family_id");
        uni.removeStorageSync("chuyaji_family_role");
      } catch {}
    },
    async refreshFamilyContext() {
      try {
        const fam = await apiGetCurrentFamily();
        this.setFamilyMeta(fam.id, fam.my_role);
        const [babies, mothers] = await Promise.all([apiListBabies(), apiListMothers()]);
        if (babies.items?.length) {
          const preferred = babies.items.find((b) => b.id === this.babyId) || babies.items[0];
          this.setBaby(preferred.id);
        } else {
          this.setBaby(0);
        }
        if (mothers.items?.length) {
          const preferred = mothers.items.find((m) => m.id === this.motherId) || mothers.items[0];
          this.setMother(preferred.id);
        } else {
          this.setMother(0);
        }
      } catch (e: any) {
        const msg = String(e?.message || "");
        if (msg.includes("404") && msg.includes("no family")) {
          this.clearFamilyCache();
        }
      }
    },
  },
});
