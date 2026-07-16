<template>
  <view class="page">
    <view v-if="error && !noFamily" class="notice">
      <text class="notice-title">加载失败</text>
      <text class="notice-desc">{{ error }}</text>
      <button size="mini" class="notice-btn" @click="load">重试</button>
    </view>

    <view v-else-if="noFamily" class="empty-card">
      <text class="empty-title">还没有加入家庭</text>
      <text class="empty-desc">你可以创建自己的家庭，或通过家人分享的邀请链接加入。</text>
      <button class="main-btn" :loading="creating" @click="createFamily">创建我的家庭</button>
    </view>

    <template v-else>
      <view class="hero">
        <text class="hero-kicker">家人共享</text>
        <text class="hero-title">{{ family?.name || "我的家庭" }}</text>
        <view class="hero-meta">
          <text class="hero-count">{{ memberCount }} 位成员</text>
          <view class="hero-role-pill">{{ roleLabel }}</view>
        </view>
        <text class="hero-desc">一起记录宝宝与宝妈的成长点滴</text>
      </view>

      <view class="panel">
        <view class="panel-head">
          <view class="panel-head-text">
            <text class="panel-kicker">家庭成员</text>
            <text class="panel-title">成员列表</text>
          </view>
          <view class="panel-head-rule" />
        </view>

        <view class="member-list">
          <view
            v-for="m in sortedMembers"
            :key="m.user_id"
            :class="['member-card', m.user_id === myUserId && 'member-card--self', canTapMember(m) && 'member-card--tap']"
            @click="onMemberTap(m)"
            @longpress="onMemberLongPress(m)"
          >
            <view v-if="m.avatar_url" class="member-avatar member-avatar--img">
              <image class="member-avatar-img" :src="m.avatar_url" mode="aspectFill" />
            </view>
            <view v-else :class="['member-avatar', avatarTone(m.role)]">
              <text class="member-avatar-text">{{ avatarText(m) }}</text>
            </view>

            <view class="member-body">
              <view class="member-name-row">
                <text :class="['member-name', !m.nickname?.trim() && 'member-name--placeholder']">
                  {{ memberLabel(m) }}
                </text>
                <text v-if="m.user_id === myUserId" class="member-self-tag">我</text>
              </view>
              <text class="member-hint">{{ memberHint(m) }}</text>
            </view>

            <view :class="['member-role-badge', roleBadgeClass(m.role)]">
              {{ roleText(m.role) }}
            </view>

            <text v-if="isOwner && m.user_id !== myUserId && m.role !== 'owner'" class="member-chevron">›</text>
          </view>
        </view>
      </view>

      <view v-if="isOwner" class="panel panel--invite">
        <view class="panel-head">
          <view class="panel-head-text">
            <text class="panel-kicker">邀请家人</text>
            <text class="panel-title">分享加入</text>
          </view>
          <view class="panel-head-rule" />
        </view>

        <text class="invite-hint">选择家人权限后，点击下方按钮发送邀请</text>
        <text v-if="inviteError" class="invite-error">{{ inviteError }}</text>

        <view class="invite-role-grid">
          <view
            :class="['invite-role-card', inviteRole === 'write' && 'invite-role-card--on']"
            @click="inviteRole = 'write'"
          >
            <text class="invite-role-icon">✎</text>
            <view class="invite-role-copy">
              <text class="invite-role-name">可读写</text>
              <text class="invite-role-desc">可查看并添加记录</text>
            </view>
            <view v-if="inviteRole === 'write'" class="invite-role-check">✓</view>
          </view>
          <view
            :class="['invite-role-card', inviteRole === 'read' && 'invite-role-card--on']"
            @click="inviteRole = 'read'"
          >
            <text class="invite-role-icon">👁</text>
            <view class="invite-role-copy">
              <text class="invite-role-name">只读</text>
              <text class="invite-role-desc">仅查看，不能修改</text>
            </view>
            <view v-if="inviteRole === 'read'" class="invite-role-check">✓</view>
          </view>
        </view>

        <button class="share-btn" :loading="inviting" open-type="share">
          <text class="share-btn-icon">↗</text>
          <text>{{ inviting ? "正在生成邀请…" : "分享给家人" }}</text>
        </button>
      </view>

      <view v-if="canLeave" class="panel panel--leave">
        <text class="leave-hint">{{ leaveHint }}</text>
        <button class="leave-btn" :loading="leaving" @click="leaveFamily">退出当前家庭</button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { onLoad, onShareAppMessage } from "@dcloudio/uni-app";
import {
  apiCreateCurrentFamily,
  apiCreateFamilyInvite,
  apiGetCurrentFamily,
  apiLeaveCurrentFamily,
  apiPatchFamilyMember,
  apiPatchFamilyMemberNickname,
  apiPatchMe,
  apiRemoveFamilyMember,
  apiTransferFamilyOwner,
  type FamilyInfo,
  type FamilyMember,
  type FamilyRole,
} from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import { parseFamilyError } from "@/utils/familyError";

const auth = useAuthStore();
const session = useSessionStore();
const family = ref<FamilyInfo | null>(null);
const error = ref("");
const noFamily = ref(false);
const creating = ref(false);
const leaving = ref(false);
const inviting = ref(false);
const inviteRole = ref<"write" | "read">("write");
const sharePath = ref("");
const shareToken = ref("");
const shareTokenRole = ref<"write" | "read" | "">("");
const inviteError = ref("");

const myUserId = computed(() => auth.userId || family.value?.my_user_id || 0);
const isOwner = computed(() => family.value?.my_role === "owner");
const roleLabel = computed(() => roleText(family.value?.my_role || session.familyRole || "owner"));
const memberCount = computed(() => family.value?.members?.length || 0);
const canLeave = computed(() => {
  if (!family.value) return false;
  if (family.value.my_role !== "owner") return true;
  return memberCount.value <= 1;
});
const leaveHint = computed(() => {
  if (family.value?.my_role === "owner" && memberCount.value <= 1) {
    return "退出后将删除你当前家庭的所有档案与记录，此操作不可恢复。";
  }
  return "退出后将无法查看该家庭的共享档案，你可以通过邀请重新加入。";
});

const sortedMembers = computed(() => {
  const list = [...(family.value?.members || [])];
  const rank = (m: FamilyMember) => {
    if (m.role === "owner") return 0;
    if (m.user_id === myUserId.value) return 1;
    return 2;
  };
  return list.sort((a, b) => {
    const diff = rank(a) - rank(b);
    if (diff !== 0) return diff;
    return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
  });
});

const ROLE_MAP: Record<string, { text: string; badge: string; avatar: string }> = {
  owner: { text: "主人", badge: "member-role-badge--owner", avatar: "member-avatar--owner" },
  write: { text: "可读写", badge: "member-role-badge--write", avatar: "member-avatar--write" },
  read: { text: "只读", badge: "member-role-badge--read", avatar: "member-avatar--read" },
};

const DEFAULT_ROLE = { text: "未知", badge: "", avatar: "member-avatar--read" };

function getRoleMeta(role: string) {
  return ROLE_MAP[role] || DEFAULT_ROLE;
}

function roleText(role: string) {
  return getRoleMeta(role).text;
}

function roleBadgeClass(role: string) {
  return getRoleMeta(role).badge;
}

function avatarTone(role: string) {
  return getRoleMeta(role).avatar;
}

function avatarText(m: FamilyMember) {
  const name = String(m.nickname || "").trim();
  if (name) return name.slice(0, 1);
  if (m.role === "owner") return "主";
  return "家";
}

function memberLabel(m: FamilyMember) {
  const name = String(m.nickname || "").trim();
  if (name) return name;
  return "未设置昵称";
}

function memberHint(m: FamilyMember) {
  const hasNick = !!m.nickname?.trim();
  const isSelf = m.user_id === myUserId.value;
  if (isSelf) {
    return hasNick ? "点击修改昵称" : "点击设置昵称";
  }
  if (isOwner.value && m.role !== "owner") {
    return hasNick ? "点击修改昵称，长按管理权限" : "点击设置昵称，长按管理权限";
  }
  return hasNick ? "点击修改昵称" : "点击设置昵称";
}

function canTapMember(_m: FamilyMember) {
  return true;
}

function onMemberTap(m: FamilyMember) {
  editMemberNickname(m);
}

function onMemberLongPress(m: FamilyMember) {
  if (!isOwner.value || m.user_id === myUserId.value || m.role === "owner") return;
  uni.showActionSheet({
    itemList: ["转让主人", "修改权限", "移除成员"],
    success: (res) => {
      if (res.tapIndex === 0) transferOwner(m.user_id, memberLabel(m));
      else if (res.tapIndex === 1) changeRole(m.user_id, m.role);
      else if (res.tapIndex === 2) removeMember(m.user_id, memberLabel(m));
    },
  });
}

async function load() {
  error.value = "";
  noFamily.value = false;
  try {
    await auth.ensureUserId();
    family.value = await apiGetCurrentFamily();
    if (family.value?.my_user_id && !auth.userId) {
      auth.setUserId(family.value.my_user_id);
    }
    session.setFamilyMeta(family.value.id, family.value.my_role);
  } catch (e: any) {
    const msg = String(e?.message || "");
    if (msg.includes("404") && msg.includes("no family")) {
      noFamily.value = true;
      family.value = null;
      error.value = "";
      return;
    }
    error.value = e?.message || "无法加载家庭信息";
  }
}

async function createFamily() {
  creating.value = true;
  try {
    family.value = await apiCreateCurrentFamily();
    noFamily.value = false;
    session.setFamilyMeta(family.value.id, family.value.my_role);
    uni.showToast({ title: "家庭已创建", icon: "success" });
  } catch (e: any) {
    uni.showToast({ title: e?.message || "创建失败", icon: "none" });
  } finally {
    creating.value = false;
  }
}

async function leaveFamily() {
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "退出家庭",
      content: leaveHint.value,
      confirmColor: "#c96b5c",
      success: (r) => resolve(!!r.confirm),
    });
  });
  if (!ok) return;
  leaving.value = true;
  try {
    await apiLeaveCurrentFamily();
    session.clearFamilyCache();
    family.value = null;
    noFamily.value = true;
    uni.showToast({ title: "已退出家庭", icon: "success" });
  } catch (e: any) {
    const msg = String(e?.message || "");
    if (msg.includes("family has other members")) {
      uni.showToast({ title: "请先转移或移除其他成员", icon: "none" });
    } else {
      uni.showToast({ title: "退出失败", icon: "none" });
    }
  } finally {
    leaving.value = false;
  }
}

function pickRoleLabel(role: FamilyRole) {
  return getRoleMeta(role).text;
}

function editMemberNickname(m: FamilyMember) {
  const isSelf = m.user_id === myUserId.value;
  uni.showModal({
    title: isSelf ? "设置我的昵称" : "设置成员昵称",
    editable: true,
    placeholderText: "例如：爸爸、妈妈、爷爷",
    content: String(m.nickname || "").trim(),
    success: async (res) => {
      if (!res.confirm) return;
      const nickname = String(res.content || "").trim();
      if (!nickname) {
        uni.showToast({ title: "昵称不能为空", icon: "none" });
        return;
      }
      try {
        if (isSelf) {
          await apiPatchMe({ nickname });
          await load();
        } else {
          family.value = await apiPatchFamilyMemberNickname(m.user_id, nickname);
        }
        uni.showToast({ title: "昵称已更新", icon: "success" });
      } catch {
        uni.showToast({ title: "保存失败", icon: "none" });
      }
    },
  });
}

async function changeRole(userId: number, current: FamilyRole) {
  if (current === "owner") {
    uni.showToast({ title: "不能修改主人角色", icon: "none" });
    return;
  }
  uni.showActionSheet({
    itemList: ["可读写", "只读"],
    success: async (res) => {
      const next: FamilyRole = res.tapIndex === 0 ? "write" : "read";
      if (next === current) return;
      try {
        family.value = await apiPatchFamilyMember(userId, next);
        uni.showToast({ title: `已设为${pickRoleLabel(next)}`, icon: "success" });
      } catch (e: any) {
        const msg = String(e?.message || "");
        if (msg.includes("cannot change own role")) {
          uni.showToast({ title: "不能修改自己的角色", icon: "none" });
        } else {
          uni.showToast({ title: "修改失败", icon: "none" });
        }
      }
    },
  });
}

function buildSharePath(token: string) {
  return `pages/family/join?token=${encodeURIComponent(token)}`;
}


function clearShareInvite() {
  shareToken.value = "";
  sharePath.value = "";
  shareTokenRole.value = "";
}

async function prepareShareInvite(): Promise<boolean> {
  if (!isOwner.value) return false;
  if (shareToken.value && shareTokenRole.value === inviteRole.value) return true;

  inviting.value = true;
  inviteError.value = "";
  try {
    const inv = await apiCreateFamilyInvite(inviteRole.value);
    shareToken.value = inv.token;
    sharePath.value = buildSharePath(inv.token);
    shareTokenRole.value = inviteRole.value;
    return true;
  } catch (e) {
    clearShareInvite();
    inviteError.value = parseFamilyError(e);
    uni.showToast({ title: inviteError.value, icon: "none", duration: 3000 });
    return false;
  } finally {
    inviting.value = false;
  }
}

async function removeMember(userId: number, name: string) {
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "移除成员",
      content: `确定将「${name}」移出家庭吗？`,
      confirmColor: "#c96b5c",
      success: (r) => resolve(!!r.confirm),
    });
  });
  if (!ok) return;
  try {
    await apiRemoveFamilyMember(userId);
    await load();
    uni.showToast({ title: "已移除", icon: "success" });
  } catch {
    uni.showToast({ title: "移除失败", icon: "none" });
  }
}

async function transferOwner(userId: number, name: string) {
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "转让主人",
      content: `确定将家庭主人转让给「${name}」吗？转让后你将变为可读写成员。`,
      confirmColor: "#c96b5c",
      success: (r) => resolve(!!r.confirm),
    });
  });
  if (!ok) return;
  try {
    family.value = await apiTransferFamilyOwner(userId);
    session.setFamilyMeta(family.value.id, family.value.my_role);
    uni.showToast({ title: "已转让", icon: "success" });
  } catch (e: any) {
    uni.showToast({ title: e?.message || "转让失败", icon: "none" });
  }
}

onLoad(async () => {
  await load();
});

watch(inviteRole, () => {
  clearShareInvite();
  inviteError.value = "";
});

onShareAppMessage(() => {
  const self = family.value?.members?.find((m) => m.user_id === myUserId.value);
  const inviter = self?.nickname?.trim() || "家人";
  const title = `${inviter}邀请你加入初芽记家庭`;
  return {
    title,
    path: "pages/family/index",
    promise: prepareShareInvite().then((ok) => {
      if (!ok || !shareToken.value) {
        throw new Error(inviteError.value || "生成邀请失败");
      }
      return {
        title,
        path: sharePath.value || buildSharePath(shareToken.value),
      };
    }),
  };
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  padding: $cj-page-pad-y $cj-page-pad-x 80rpx;
}

.hero {
  padding: 32rpx 28rpx 28rpx;
  background: linear-gradient(155deg, rgba(255, 253, 249, 0.98) 0%, rgba(255, 247, 240, 0.96) 100%);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-xl;
  box-shadow: $cj-shadow-card;
}

.hero-kicker {
  display: block;
  font-size: 22rpx;
  letter-spacing: 4rpx;
  color: $cj-primary;
  font-weight: 500;
}

.hero-title {
  display: block;
  margin-top: 10rpx;
  font-size: 44rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
  line-height: 1.25;
}

.hero-meta {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-top: 16rpx;
}

.hero-count {
  font-size: 24rpx;
  color: $cj-text-secondary;
}

.hero-role-pill {
  padding: 4rpx 16rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  color: $cj-primary-dark;
  background: $cj-primary-soft;
}

.hero-desc {
  display: block;
  margin-top: 14rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
  line-height: 1.5;
}

.panel {
  margin-top: $cj-gap-md;
  padding: 24rpx 24rpx 20rpx;
  background: $cj-surface;
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-lg;
  box-shadow: $cj-shadow-xs;
}

.panel-head {
  display: flex;
  align-items: flex-end;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.panel-head-text {
  flex-shrink: 0;
}

.panel-kicker {
  display: block;
  font-size: 20rpx;
  letter-spacing: 3rpx;
  color: $cj-text-muted;
}

.panel-title {
  display: block;
  margin-top: 4rpx;
  font-size: 30rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.panel-head-rule {
  flex: 1;
  height: 1rpx;
  margin-bottom: 8rpx;
  background: linear-gradient(90deg, $cj-border-light 0%, transparent 100%);
}

.member-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.member-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 18rpx 16rpx;
  background: rgba(255, 255, 255, 0.72);
  border: 1rpx solid $cj-border-faint;
  border-radius: $cj-radius-md;
  transition: background 0.15s ease;
}

.member-card--self {
  background: linear-gradient(135deg, rgba(245, 221, 216, 0.35) 0%, rgba(255, 253, 249, 0.9) 100%);
  border-color: rgba(201, 107, 92, 0.18);
}

.member-card--tap:active {
  background: rgba(201, 107, 92, 0.06);
}

.member-avatar {
  flex-shrink: 0;
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.member-avatar--img {
  background: $cj-surface-2;
}

.member-avatar-img {
  width: 100%;
  height: 100%;
}

.member-avatar-text {
  font-size: 32rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary-dark;
}

.member-avatar--owner {
  background: linear-gradient(145deg, $cj-primary-soft 0%, #fceee9 100%);
}

.member-avatar--write {
  background: linear-gradient(145deg, $cj-mint-soft 0%, #eef7f2 100%);
}

.member-avatar--write .member-avatar-text {
  color: #3d6b58;
}

.member-avatar--read {
  background: linear-gradient(145deg, $cj-accent-soft 0%, #faf3ea 100%);
}

.member-avatar--read .member-avatar-text {
  color: #8a6848;
}

.member-body {
  flex: 1;
  min-width: 0;
}

.member-name-row {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.member-name {
  font-size: 30rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-name--placeholder {
  color: $cj-text-muted;
  font-weight: 500;
}

.member-self-tag {
  flex-shrink: 0;
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
  font-size: 18rpx;
  color: $cj-primary;
  background: rgba(201, 107, 92, 0.12);
}

.member-hint {
  display: block;
  margin-top: 4rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-role-badge {
  flex-shrink: 0;
  padding: 6rpx 14rpx;
  border-radius: $cj-radius-pill;
  font-size: 20rpx;
  font-weight: 500;
}

.member-role-badge--owner {
  color: $cj-primary-dark;
  background: $cj-primary-soft;
}

.member-role-badge--write {
  color: #3d6b58;
  background: $cj-mint-soft;
}

.member-role-badge--read {
  color: #8a6848;
  background: $cj-accent-soft;
}

.member-chevron {
  flex-shrink: 0;
  font-size: 36rpx;
  line-height: 1;
  color: $cj-text-muted;
  margin-left: -4rpx;
}

.invite-hint {
  display: block;
  margin-bottom: 16rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
  line-height: 1.5;
}

.invite-error {
  display: block;
  margin-bottom: 16rpx;
  font-size: 24rpx;
  color: #c96b5c;
  line-height: 1.5;
}

.invite-role-grid {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
  margin-bottom: 20rpx;
}

.invite-role-card {
  display: flex;
  align-items: center;
  gap: 16rpx;
  padding: 20rpx 18rpx;
  background: rgba(255, 255, 255, 0.72);
  border: 2rpx solid $cj-border-faint;
  border-radius: $cj-radius-md;

  &:active {
    opacity: 0.92;
  }
}

.invite-role-card--on {
  border-color: rgba(201, 107, 92, 0.45);
  background: linear-gradient(135deg, rgba(245, 221, 216, 0.28) 0%, rgba(255, 253, 249, 0.95) 100%);
  box-shadow: $cj-shadow-xs;
}

.invite-role-icon {
  flex-shrink: 0;
  width: 56rpx;
  height: 56rpx;
  line-height: 56rpx;
  text-align: center;
  font-size: 28rpx;
  border-radius: 16rpx;
  background: $cj-surface-2;
}

.invite-role-copy {
  flex: 1;
  min-width: 0;
}

.invite-role-name {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.invite-role-desc {
  display: block;
  margin-top: 4rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.invite-role-check {
  flex-shrink: 0;
  width: 40rpx;
  height: 40rpx;
  line-height: 40rpx;
  text-align: center;
  border-radius: 50%;
  font-size: 22rpx;
  color: #fffefb;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
}

.share-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10rpx;
  width: 100%;
  padding: 0;
  height: 88rpx;
  line-height: 88rpx;
  font-size: 30rpx;
  font-weight: 600;
  color: #fffefb;
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
  border-radius: $cj-radius-pill;
  box-shadow: $cj-shadow-soft;
}

.share-btn[disabled] {
  background: #dcc9bf;
  box-shadow: none;
}

.share-btn-icon {
  font-size: 32rpx;
  line-height: 1;
}

.notice {
  margin-top: 24rpx;
  padding: 24rpx;
  background: $cj-surface-2;
  border-radius: $cj-radius-md;
  border: 1rpx solid $cj-warn-border;
}

.notice-title {
  display: block;
  font-weight: 600;
  color: $cj-ink;
}

.notice-desc {
  display: block;
  margin: 8rpx 0 12rpx;
  color: $cj-text-secondary;
  font-size: 26rpx;
}

.notice-btn {
  background: $cj-primary;
  color: #fff;
}

.empty-card {
  margin-top: 24rpx;
  padding: 40rpx 32rpx;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  border: 1rpx solid $cj-border-faint;
}

.empty-title {
  display: block;
  font-size: 34rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.empty-desc {
  display: block;
  margin: 12rpx 0 28rpx;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
}

.main-btn {
  background: linear-gradient(160deg, $cj-primary-gradient-top 0%, $cj-primary-dark 100%);
  color: #fffefb;
  border-radius: $cj-radius-pill;
}

.panel--leave {
  margin-top: $cj-gap-md;
}

.leave-hint {
  display: block;
  margin-bottom: 16rpx;
  font-size: 24rpx;
  color: $cj-text-muted;
  line-height: 1.5;
}

.leave-btn {
  width: 100%;
  background: transparent;
  color: $cj-primary;
  border: 1rpx solid rgba(201, 107, 92, 0.35);
  border-radius: $cj-radius-pill;
}
</style>
