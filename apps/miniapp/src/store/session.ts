import { defineStore } from "pinia";
import type { FamilyRole } from "@/api/chuyaji";
import { apiGetCurrentFamily, apiListBabies } from "@/api/chuyaji";

export const useSessionStore = defineStore("session", {
  state: () => ({
    babyId: 0 as number,
    motherId: 0 as number,
    privacyOk: false as boolean,
    familyId: 0 as number,
    familyRole: "" as FamilyRole | "",
    /** 当前人在家庭里的身份（即家人页设置的昵称，如「爸爸」「妈妈」） */
    myNickname: "" as string,
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
        this.myNickname = String(uni.getStorageSync("chuyaji_my_nickname") || "");
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
    setMyNickname(name: string) {
      this.myNickname = name;
      try {
        if (name) uni.setStorageSync("chuyaji_my_nickname", name);
        else uni.removeStorageSync("chuyaji_my_nickname");
      } catch {}
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
      this.myNickname = "";
      try {
        uni.removeStorageSync("chuyaji_baby_id");
        uni.removeStorageSync("chuyaji_mother_id");
        uni.removeStorageSync("chuyaji_family_id");
        uni.removeStorageSync("chuyaji_family_role");
        uni.removeStorageSync("chuyaji_my_nickname");
      } catch {}
    },
    async refreshFamilyContext() {
      try {
        const fam = await apiGetCurrentFamily();
        this.setFamilyMeta(fam.id, fam.my_role);
        const self = fam.members?.find((m) => m.user_id === fam.my_user_id);
        this.setMyNickname(String(self?.nickname || "").trim());
        const babies = await apiListBabies();
        if (babies.items?.length) {
          const preferred = babies.items.find((b) => b.id === this.babyId) || babies.items[0];
          this.setBaby(preferred.id);
        } else {
          this.setBaby(0);
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
