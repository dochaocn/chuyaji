<template>
  <view class="page">
    <view class="card">
      <text class="title">{{ title }}</text>
      <text class="desc">{{ desc }}</text>

      <view v-if="preview && phase === 'preview'" class="preview-box">
        <view class="preview-row">
          <text class="preview-label">邀请人</text>
          <text class="preview-value">{{ inviterLabel }}</text>
        </view>
        <view class="preview-row">
          <text class="preview-label">家庭名称</text>
          <text class="preview-value">{{ preview.family_name }}</text>
        </view>
        <view class="preview-row">
          <text class="preview-label">你的权限</text>
          <text class="preview-value">{{ roleLabel }}</text>
        </view>
        <view class="preview-row">
          <text class="preview-label">家庭成员</text>
          <text class="preview-value">{{ preview.member_count }} 位</text>
        </view>
      </view>

      <view v-if="conflictHint" class="warn-box">
        <text class="warn-text">{{ conflictHint }}</text>
      </view>

      <button
        v-if="phase === 'preview' && !preview?.already_member"
        class="main-btn"
        :loading="loading"
        @click="confirmJoin"
      >
        确认加入
      </button>
      <button v-else-if="phase === 'success' || preview?.already_member" class="main-btn" @click="goHome">
        进入工作台
      </button>
      <button v-else-if="phase === 'error'" class="main-btn" :loading="loading" @click="loadPreview">
        重试
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import {
  apiAcceptFamilyInvite,
  apiPreviewFamilyInvite,
  type FamilyInvitePreview,
} from "@/api/chuyaji";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";
import { parseFamilyError } from "@/utils/familyError";

const auth = useAuthStore();
const session = useSessionStore();
const token = ref("");
const loading = ref(false);
const phase = ref<"loading" | "preview" | "success" | "error">("loading");
const title = ref("正在加载邀请…");
const desc = ref("请稍候");
const preview = ref<FamilyInvitePreview | null>(null);

const inviterLabel = computed(() => {
  const name = String(preview.value?.inviter_nickname || "").trim();
  return name || "家人";
});

const roleLabel = computed(() => {
  if (preview.value?.role === "read") return "只读（可查看）";
  if (preview.value?.role === "write") return "可读写（可查看并添加记录）";
  return preview.value?.role || "";
});

const conflictIsOwner = computed(() => preview.value?.conflict?.my_role === "owner");

const conflictHint = computed(() => {
  const c = preview.value?.conflict;
  if (!c?.has_existing_family) return "";
  if (c.has_other_members) {
    return "你是当前家庭的主人且还有其他成员，请先转移管理权或移除成员后再加入。";
  }
  const parts: string[] = [];
  if (c.baby_count > 0) parts.push(`${c.baby_count} 个宝宝档案`);
  if (c.mother_count > 0) parts.push(`${c.mother_count} 个宝妈档案`);
  if (parts.length) {
    if (conflictIsOwner.value) {
      return `你当前账号里有${parts.join("、")}。确认加入后，这些档案将被删除，改为使用邀请方的共享档案。`;
    }
    return `你当前家庭里有${parts.join("、")}。加入新家庭后你将无法访问这些档案。`;
  }
  if (c.family_name) {
    return `你将离开「${c.family_name}」并加入新家庭。`;
  }
  return "";
});

async function loadPreview() {
  if (!token.value) {
    phase.value = "error";
    title.value = "邀请无效";
    desc.value = "缺少邀请参数，请让家人重新分享。";
    return;
  }
  loading.value = true;
  phase.value = "loading";
  title.value = "正在加载邀请…";
  desc.value = "请稍候";
  try {
    if (!auth.token) {
      await auth.ensureWeChatSession();
    }
    const data = await apiPreviewFamilyInvite(token.value);
    preview.value = data;
    phase.value = "preview";
    if (data.already_member) {
      loading.value = false;
      await enterFamilyHome(data.family_name);
      return;
    }
    title.value = `${inviterLabel.value} 邀请你加入`;
    desc.value = `加入「${data.family_name}」后，可以和家人一起查看与记录。`;
  } catch (e: any) {
    phase.value = "error";
    title.value = "无法加载邀请";
    desc.value = parseFamilyError(e);
  } finally {
    loading.value = false;
  }
}

async function confirmJoin() {
  if (!token.value || !preview.value) return;
  if (preview.value.conflict?.has_other_members) {
    uni.showToast({ title: "请先处理现有家庭", icon: "none" });
    return;
  }

  const c = preview.value.conflict;
  if (c?.has_existing_family && conflictHint.value) {
    const ok = await new Promise<boolean>((resolve) => {
      uni.showModal({
        title: "确认加入新家庭",
        content: conflictHint.value,
        confirmColor: "#c96b5c",
        success: (r) => resolve(!!r.confirm),
      });
    });
    if (!ok) return;
  }

  loading.value = true;
  try {
    const fam = await apiAcceptFamilyInvite(token.value);
    session.clearFamilyCache();
    session.setFamilyMeta(fam.id, fam.my_role);
    session.setPendingInvite("");
    await session.refreshFamilyContext();
    phase.value = "success";
    title.value = "已加入家庭";
    desc.value = `欢迎加入「${fam.name}」，现在可以查看共享档案与记录。`;
    setTimeout(() => goHome(), 800);
  } catch (e: any) {
    phase.value = "error";
    title.value = "无法加入";
    desc.value = parseFamilyError(e);
  } finally {
    loading.value = false;
  }
}

function goHome() {
  session.setPendingInvite("");
  uni.switchTab({ url: "/pages/baby/home" });
}

async function enterFamilyHome(familyName: string) {
  session.setPendingInvite("");
  await session.refreshFamilyContext();
  title.value = "欢迎回来";
  desc.value = `进入「${familyName}」，和家人一起查看记录。`;
  phase.value = "success";
  setTimeout(() => goHome(), 400);
}

onLoad((query) => {
  const fromQuery = String(query?.token || "");
  const fromSession = session.pendingInviteToken || "";
  const t = fromQuery || fromSession;
  token.value = t;
  if (t) session.setPendingInvite(t);
  loadPreview();
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40rpx;
  background: #f4ebe3;
}
.card {
  width: 100%;
  padding: 40rpx 32rpx;
  background: #fffdf9;
  border-radius: 24rpx;
  border: 1rpx solid #eadfd4;
}
.title {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: #5c4a3f;
}
.desc {
  display: block;
  margin: 16rpx 0 24rpx;
  color: #8b8077;
  font-size: 28rpx;
  line-height: 1.5;
}
.preview-box {
  margin-bottom: 24rpx;
  padding: 20rpx;
  background: rgba(255, 255, 255, 0.8);
  border: 1rpx solid #eadfd4;
  border-radius: 16rpx;
}
.preview-row {
  display: flex;
  justify-content: space-between;
  gap: 16rpx;
  padding: 10rpx 0;
}
.preview-label {
  font-size: 26rpx;
  color: #8b8077;
}
.preview-value {
  font-size: 26rpx;
  color: #5c4a3f;
  font-weight: 600;
  text-align: right;
}
.warn-box {
  margin-bottom: 24rpx;
  padding: 16rpx 20rpx;
  background: #fff7ed;
  border: 1rpx solid #f0dcc8;
  border-radius: 12rpx;
}
.warn-text {
  font-size: 24rpx;
  color: #9a6b4a;
  line-height: 1.5;
}
.main-btn {
  background: #c96b5c;
  color: #fff;
}
</style>
