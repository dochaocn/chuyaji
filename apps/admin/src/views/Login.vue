<template>
  <div class="login-container">
    <div class="login-bg"></div>
    <el-card class="login-card" shadow="always">
      <div class="login-logo">
        <div class="logo-icon">芽</div>
        <h2>初芽记</h2>
        <p>管理后台</p>
      </div>
      <el-form @submit.prevent="handleLogin">
        <el-form-item>
          <el-input
            v-model="password"
            type="password"
            placeholder="请输入管理密码"
            show-password
            size="large"
            prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            style="width: 100%"
            :loading="loading"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../store/auth'
import { adminLogin } from '../api/admin'

const router = useRouter()
const auth = useAuthStore()
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  if (!password.value) {
    ElMessage.warning('请输入密码')
    return
  }
  loading.value = true
  try {
    const res = await adminLogin(password.value)
    auth.setToken(res.data.token)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5ece4;
  position: relative;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(ellipse 140% 80% at 100% -8%, rgba(232, 184, 150, 0.28) 0%, transparent 50%),
    radial-gradient(ellipse 90% 60% at -10% 100%, rgba(125, 171, 152, 0.16) 0%, transparent 45%),
    radial-gradient(ellipse 50% 35% at 50% 50%, rgba(255, 253, 249, 0.45) 0%, transparent 100%);
  pointer-events: none;
}

.login-card {
  width: 380px;
  max-width: calc(100vw - 32px);
  border-radius: 16px;
  border: 1px solid #f0e4d8;
  background: #fffdf9;
  box-shadow: 0 2px 8px rgba(42, 36, 32, 0.03), 0 12px 40px rgba(42, 36, 32, 0.06);
  position: relative;
  z-index: 1;
}

.login-logo {
  text-align: center;
  margin-bottom: 28px;
}

.login-logo .logo-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  background: linear-gradient(160deg, #da8a7e 0%, #a84f42 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 700;
  margin: 0 auto 12px;
  box-shadow: 0 4px 12px rgba(201, 107, 92, 0.2);
}

.login-logo h2 {
  font-size: 22px;
  font-weight: 700;
  color: #3a322d;
  margin: 0;
}

.login-logo p {
  color: #9a9088;
  font-size: 13px;
  margin-top: 4px;
}
</style>
