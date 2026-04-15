<template>
  <view class="shell">
    <view class="blob blob-a" />
    <view class="blob blob-b" />

    <view class="hero">
      <view class="hero-inner">
        <view class="mark-row">
          <text class="mark">手记</text>
          <view class="mark-dot" />
        </view>
        <view class="title-row">
          <text class="title-a">初芽</text>
          <text class="title-b">记</text>
        </view>
        <text class="tagline">把孕期与成长的点滴，温柔收进家庭里。</text>
        <view class="title-underline" />
      </view>
    </view>

    <view v-if="!session.privacyOk" class="notice">
      <view class="notice-icon">📋</view>
      <view class="notice-body">
        <text class="notice-title">先读隐私说明</text>
        <text class="notice-desc">数据只属于你的家庭，我们仅用于提供记录服务。</text>
        <button class="notice-btn" size="mini" type="primary" @click="goPrivacy">去阅读</button>
      </view>
    </view>

    <view class="panel">
      <view class="panel-head">
        <view class="panel-bar" />
        <text class="panel-title">当前状态</text>
      </view>
      <view class="status-row">
        <text class="status-label">登录</text>
        <text class="status-val">{{ tokenHint }}</text>
      </view>
    </view>

    <button class="btn-main" type="primary" @click="onLogin">微信登录</button>
    <button class="btn-sub" @click="onLogout">退出账号</button>

    <view class="panel actions">
      <view class="panel-head">
        <view class="panel-bar mint" />
        <text class="panel-title">从这里开始</text>
      </view>
      <view class="grid2">
        <view class="tile" @click="goFamilies">
          <text class="tile-ico">🏠</text>
          <text class="tile-t">家庭</text>
          <text class="tile-d">创建或管理</text>
        </view>
        <view v-if="session.babyId" class="tile tile-accent" @click="goTimeline">
          <text class="tile-ico">🌿</text>
          <text class="tile-t">时间线</text>
          <text class="tile-d">当前宝宝</text>
        </view>
        <view v-else class="tile tile-muted">
          <text class="tile-ico">🌿</text>
          <text class="tile-t">时间线</text>
          <text class="tile-d">登录后可选宝宝</text>
        </view>
      </view>
    </view>

    <view class="foot-note">
      <text>开发联调：本地 Go API + 开发者工具关闭域名校验；附件需 CHUYAJI_UPLOAD_DIR。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onShow } from "@dcloudio/uni-app";
import { computed } from "vue";
import { isDevWechatLogin } from "@/api/config";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";

const auth = useAuthStore();
const session = useSessionStore();

const tokenHint = computed(() => (auth.token ? "已登录" : "未登录"));

onShow(() => {
  auth.loadToken();
  session.load();
});

function goPrivacy() {
  uni.navigateTo({ url: "/pages/privacy/privacy" });
}

async function onLogin() {
  if (isDevWechatLogin()) {
    try {
      await auth.loginWithWeChatCode("dev");
      uni.showToast({ title: "登录成功(dev)", icon: "success" });
    } catch (e) {
      uni.showToast({ title: "登录失败", icon: "none" });
      console.error(e);
    }
    return;
  }
  uni.login({
    provider: "weixin",
    success: async (res) => {
      if (!res.code) {
        uni.showToast({ title: "无 code", icon: "none" });
        return;
      }
      try {
        await auth.loginWithWeChatCode(res.code);
        uni.showToast({ title: "登录成功", icon: "success" });
      } catch (e) {
        uni.showToast({ title: "登录失败", icon: "none" });
        console.error(e);
      }
    },
    fail: (e) => {
      console.error(e);
      uni.showToast({ title: "uni.login 失败", icon: "none" });
    },
  });
}

function onLogout() {
  auth.logout();
  uni.showToast({ title: "已退出", icon: "none" });
}

function goFamilies() {
  uni.navigateTo({ url: "/pages/families/index" });
}

function goTimeline() {
  uni.navigateTo({ url: `/pages/timeline/index?baby_id=${session.babyId}` });
}
</script>

<style lang="scss" scoped>
.shell {
  position: relative;
  min-height: 100vh;
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
  box-sizing: border-box;
  overflow: hidden;
}

.blob {
  position: absolute;
  border-radius: 50%;
  pointer-events: none;
  z-index: 0;
}
.blob-a {
  width: 320rpx;
  height: 320rpx;
  right: -80rpx;
  top: -40rpx;
  background: radial-gradient(circle, rgba(232, 184, 150, 0.45) 0%, transparent 70%);
}
.blob-b {
  width: 260rpx;
  height: 260rpx;
  left: -100rpx;
  bottom: 180rpx;
  background: radial-gradient(circle, rgba(143, 184, 168, 0.28) 0%, transparent 70%);
}

.hero {
  position: relative;
  z-index: 1;
  margin-bottom: $cj-gap-lg;
}

.hero-inner {
  padding: 8rpx 0 12rpx;
}

.mark-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: 12rpx;
}
.mark {
  font-size: 22rpx;
  letter-spacing: 6rpx;
  text-transform: uppercase;
  color: $cj-text-muted;
  font-weight: 500;
}
.mark-dot {
  width: 8rpx;
  height: 8rpx;
  border-radius: 50%;
  background: $cj-accent;
}

.title-row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 8rpx 16rpx;
}
.title-a {
  font-size: 68rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 8rpx;
  line-height: 1.1;
}
.title-b {
  font-size: 68rpx;
  font-weight: $cj-fw-display;
  color: $cj-primary;
  line-height: 1.1;
}

.tagline {
  display: block;
  margin-top: $cj-gap-md;
  font-size: 28rpx;
  line-height: 1.65;
  color: $cj-text-secondary;
  max-width: 92%;
}

.title-underline {
  margin-top: $cj-gap-md;
  width: 120rpx;
  height: 8rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(90deg, $cj-accent, $cj-primary 55%, transparent);
  opacity: 0.9;
}

.notice {
  position: relative;
  z-index: 1;
  display: flex;
  gap: $cj-gap-md;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-lg;
  background: $cj-warn-bg;
  border-radius: $cj-radius-md;
  border: 1rpx solid $cj-warn-border;
  box-shadow: $cj-shadow-card;
}
.notice-icon {
  font-size: 44rpx;
  line-height: 1;
}
.notice-body {
  flex: 1;
  min-width: 0;
}
.notice-title {
  display: block;
  font-size: 28rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
  margin-bottom: 8rpx;
}
.notice-desc {
  display: block;
  font-size: 24rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
  margin-bottom: $cj-gap-sm;
}
.notice-btn {
  border-radius: $cj-radius-pill !important;
}

.panel {
  position: relative;
  z-index: 1;
  background: $cj-surface;
  border-radius: $cj-radius-lg;
  padding: $cj-gap-md $cj-gap-lg;
  margin-bottom: $cj-gap-md;
  border: 1rpx solid $cj-border-light;
  box-shadow: $cj-shadow-card;
}
.panel-head {
  display: flex;
  align-items: center;
  gap: 12rpx;
  margin-bottom: $cj-gap-md;
}
.panel-bar {
  width: 6rpx;
  height: 28rpx;
  border-radius: $cj-radius-pill;
  background: linear-gradient(180deg, $cj-primary, $cj-accent);
}
.panel-bar.mint {
  background: linear-gradient(180deg, $cj-mint, $cj-accent);
}
.panel-title {
  font-size: 26rpx;
  font-weight: $cj-fw-title;
  color: $cj-text-muted;
  letter-spacing: 2rpx;
}

.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status-label {
  font-size: 28rpx;
  color: $cj-text-secondary;
}
.status-val {
  font-size: 32rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}

.btn-main {
  position: relative;
  z-index: 1;
  margin-bottom: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
}
.btn-sub {
  position: relative;
  z-index: 1;
  margin-bottom: $cj-gap-lg;
  background: $cj-surface !important;
  color: $cj-text !important;
  border: 1rpx solid $cj-border !important;
  border-radius: $cj-radius-pill !important;
  font-weight: 500;
}

.actions {
  padding-bottom: $cj-gap-lg;
}

.grid2 {
  display: flex;
  flex-wrap: wrap;
  gap: $cj-gap-md;
}
.tile {
  flex: 1;
  min-width: 280rpx;
  padding: $cj-gap-md;
  border-radius: $cj-radius-md;
  background: $cj-surface-2;
  border: 1rpx solid $cj-border-light;
  box-sizing: border-box;
}
.tile-accent {
  background: linear-gradient(145deg, $cj-accent-soft 0%, $cj-surface 100%);
  border-color: rgba(212, 165, 116, 0.35);
  box-shadow: $cj-shadow-lift;
}
.tile-muted {
  opacity: 0.75;
}
.tile-ico {
  font-size: 40rpx;
  display: block;
  margin-bottom: 8rpx;
}
.tile-t {
  display: block;
  font-size: 30rpx;
  font-weight: $cj-fw-title;
  color: $cj-ink;
}
.tile-d {
  display: block;
  margin-top: 6rpx;
  font-size: 22rpx;
  color: $cj-text-muted;
}

.foot-note {
  position: relative;
  z-index: 1;
  padding: $cj-gap-sm 8rpx;
  font-size: 20rpx;
  color: $cj-text-muted;
  line-height: 1.6;
  opacity: 0.88;
}
</style>
