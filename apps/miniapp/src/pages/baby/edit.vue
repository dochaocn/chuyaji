<template>
  <view class="page">
    <view class="head">
      <text class="head-title">{{ id ? "编辑宝宝" : "添加宝宝" }}</text>
      <text class="head-desc">填写宝宝信息，日期格式建议 yyyy-mm-dd。</text>
    </view>
    <view class="field">
      <text class="lab">昵称</text>
      <input v-model="nickname" class="input" placeholder="宝宝昵称" />
    </view>
    <view class="field">
      <text class="lab">末次月经（可选）</text>
      <input v-model="lmp" class="input" placeholder="yyyy-mm-dd" />
    </view>
    <view class="field">
      <text class="lab">预产期（可选）</text>
      <input v-model="edd" class="input" placeholder="yyyy-mm-dd" />
    </view>
    <view class="field">
      <text class="lab">出生日期（可选）</text>
      <input v-model="birth" class="input" placeholder="yyyy-mm-dd" />
    </view>
    <button type="primary" class="btn" :loading="loading" @click="save">保存</button>
    <button v-if="id" class="danger" @click="remove">删除</button>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from "@dcloudio/uni-app";
import { ref } from "vue";
import { apiCreateBaby, apiDeleteBaby, apiPatchBaby } from "@/api/chuyaji";

const id = ref(0);
const familyId = ref(0);
const nickname = ref("");
const lmp = ref("");
const edd = ref("");
const birth = ref("");
const loading = ref(false);

onLoad((q: Record<string, string | undefined>) => {
  familyId.value = Number(q.family_id || 0);
  id.value = Number(q.id || 0);
});

function iso(s: string): string | undefined {
  const t = Date.parse(s);
  if (!Number.isFinite(t)) return undefined;
  return new Date(t).toISOString();
}

async function save() {
  if (!nickname.value.trim()) {
    uni.showToast({ title: "请输入昵称", icon: "none" });
    return;
  }
  loading.value = true;
  try {
    const body: Record<string, unknown> = {
      nickname: nickname.value.trim(),
      lmp_date: lmp.value ? iso(lmp.value) : undefined,
      edd_date: edd.value ? iso(edd.value) : undefined,
      birth_date: birth.value ? iso(birth.value) : undefined,
    };
    if (id.value) {
      await apiPatchBaby(id.value, body);
    } else {
      if (!familyId.value) {
        uni.showToast({ title: "缺少 family_id", icon: "none" });
        return;
      }
      await apiCreateBaby({
        family_id: familyId.value,
        nickname: nickname.value.trim(),
        lmp_date: body.lmp_date as string | undefined,
        edd_date: body.edd_date as string | undefined,
        birth_date: body.birth_date as string | undefined,
      });
    }
    uni.showToast({ title: "已保存", icon: "success" });
    setTimeout(() => uni.navigateBack(), 400);
  } catch (e) {
    console.error(e);
    uni.showToast({ title: "失败", icon: "none" });
  } finally {
    loading.value = false;
  }
}

async function remove() {
  if (!id.value) return;
  uni.showModal({
    title: "确认删除？",
    success: async (r) => {
      if (!r.confirm) return;
      try {
        await apiDeleteBaby(id.value);
        uni.navigateBack();
      } catch (e) {
        console.error(e);
      }
    },
  });
}
</script>

<style lang="scss" scoped>
.page {
  padding: $cj-page-pad-y $cj-page-pad-x 48rpx;
}

.head {
  margin-bottom: $cj-gap-lg;
}

.head-title {
  display: block;
  font-size: 40rpx;
  font-weight: $cj-fw-display;
  color: $cj-ink;
}

.head-desc {
  display: block;
  margin-top: $cj-gap-sm;
  font-size: 26rpx;
  color: $cj-text-secondary;
  line-height: 1.55;
}

.field {
  background: $cj-surface;
  border-radius: $cj-radius-md;
  padding: $cj-gap-md;
  margin-bottom: $cj-gap-md;
  border: 1rpx solid $cj-border;
  box-shadow: $cj-shadow-card;
}

.lab {
  display: block;
  font-size: 24rpx;
  color: $cj-text-muted;
  margin-bottom: $cj-gap-sm;
}

.input {
  font-size: 28rpx;
  color: $cj-text;
}

.btn {
  margin-top: $cj-gap-sm;
  border-radius: $cj-radius-pill !important;
}

.danger {
  margin-top: $cj-gap-md;
  background: $cj-danger-bg !important;
  color: $cj-danger-text !important;
  border-radius: $cj-radius-pill !important;
  border: none !important;
}
</style>
