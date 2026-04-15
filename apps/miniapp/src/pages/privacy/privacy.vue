<template>
  <view class="page">
    <view class="frame">
      <view class="corner corner--tl" aria-hidden="true" />
      <view class="corner corner--br" aria-hidden="true" />
      <view class="hero">
        <text class="mark">家庭自用</text>
        <text class="h">隐私说明</text>
        <text class="sub">初芽记只为你家服务，读完再开始使用。</text>
      </view>
    </view>

    <view class="card">
      <text class="p">
        初芽记用于家庭自行记录孕期检查与婴幼儿成长信息，不构成医疗诊断或治疗建议。请妥善保管账号与数据，重要健康问题请遵医嘱。
      </text>
    </view>

    <button type="primary" class="btn" @click="agree">我已阅读并同意</button>
  </view>
</template>

<script setup lang="ts">
import { useSessionStore } from "@/store/session";

const session = useSessionStore();

function agree() {
  session.setPrivacyOk();
  uni.navigateBack();
}
</script>

<style lang="scss" scoped>
@keyframes cj-fade-up {
  from {
    opacity: 0;
    transform: translateY(14rpx);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page {
  padding: $cj-page-pad-y $cj-page-pad-x 56rpx;
}

.frame {
  position: relative;
  padding: $cj-gap-lg $cj-gap-md;
  margin-bottom: $cj-gap-lg;
  animation: cj-fade-up 0.55s cubic-bezier(0.22, 1, 0.36, 1) backwards;
}

.corner {
  position: absolute;
  width: 48rpx;
  height: 48rpx;
  border-color: $cj-accent;
  opacity: 0.55;
  pointer-events: none;
}

.corner--tl {
  top: 0;
  left: 0;
  border-top: 3rpx solid;
  border-left: 3rpx solid;
  border-radius: 6rpx 0 0 0;
}

.corner--br {
  bottom: 0;
  right: 0;
  border-bottom: 3rpx solid;
  border-right: 3rpx solid;
  border-radius: 0 0 6rpx 0;
}

.hero {
  text-align: center;
}

.mark {
  display: inline-block;
  font-size: 20rpx;
  font-weight: $cj-fw-title;
  letter-spacing: 3rpx;
  color: $cj-primary-dark;
  background: $cj-primary-soft;
  padding: 6rpx 20rpx;
  border-radius: $cj-radius-pill;
  margin-bottom: $cj-gap-md;
}

.h {
  display: block;
  font-size: 44rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
  letter-spacing: 2rpx;
}

.sub {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.5;
}

.card {
  background: $cj-surface;
  padding: $cj-gap-lg;
  border-radius: $cj-radius-lg;
  margin-bottom: $cj-gap-lg;
  box-shadow: $cj-shadow-card;
  border: 1rpx solid $cj-border-light;
  animation: cj-fade-up 0.5s cubic-bezier(0.22, 1, 0.36, 1) 0.08s backwards;
}

.p {
  font-size: 28rpx;
  color: $cj-text-secondary;
  line-height: 1.85;
  text-align: justify;
}

.btn {
  border-radius: $cj-radius-pill !important;
  animation: cj-fade-up 0.45s cubic-bezier(0.22, 1, 0.36, 1) 0.14s backwards;
}
</style>
