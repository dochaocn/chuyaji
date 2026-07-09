<template>
  <el-container style="height: 100%">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <div class="logo-icon">芽</div>
        <div class="logo-text">
          <span>初芽记</span>
          <small>管理后台</small>
        </div>
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="transparent"
        text-color="#6b6058"
        active-text-color="#c96b5c"
        router
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>数据总览</span>
        </el-menu-item>
        <el-menu-item index="/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/records">
          <el-icon><Document /></el-icon>
          <span>宝宝记录</span>
        </el-menu-item>
        <el-menu-item index="/mother-records">
          <el-icon><Notebook /></el-icon>
          <span>宝妈记录</span>
        </el-menu-item>
        <el-menu-item index="/reminders">
          <el-icon><Bell /></el-icon>
          <span>提醒管理</span>
        </el-menu-item>
        <el-menu-item index="/analytics">
          <el-icon><TrendCharts /></el-icon>
          <span>数据分析</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="topbar">
        <span class="topbar-title">{{ currentTitle }}</span>
        <el-button text @click="handleLogout" class="logout-btn">退出登录</el-button>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/users')) return '/users'
  if (path.startsWith('/records')) return '/records'
  if (path.startsWith('/mother-records')) return '/mother-records'
  if (path.startsWith('/reminders')) return '/reminders'
  return path
})

const currentTitle = computed(() => (route.meta.title as string) || '数据总览')

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.sidebar {
  background: #fffdf9;
  border-right: 1px solid #f0e4d8;
  box-shadow: 2px 0 12px rgba(42, 36, 32, 0.04);
}

.logo {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid #f0e4d8;
  padding: 0 16px;
}

.logo-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(160deg, #da8a7e 0%, #a84f42 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(201, 107, 92, 0.2);
}

.logo-text {
  display: flex;
  flex-direction: column;
}

.logo-text span {
  font-size: 16px;
  font-weight: 700;
  color: #3a322d;
  line-height: 1.2;
}

.logo-text small {
  font-size: 11px;
  color: #9a9088;
}

:deep(.el-menu-item) {
  border-radius: 8px;
  margin: 2px 8px;
  height: 44px;
  line-height: 44px;
  transition: all 0.2s;
}

:deep(.el-menu-item:hover) {
  background: #faf5f0 !important;
}

:deep(.el-menu-item.is-active) {
  background: #faeeeb !important;
  font-weight: 600;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #f0e4d8;
  background: #fffdf9;
  padding: 0 24px;
  height: 56px;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #3a322d;
}

.logout-btn {
  color: #9a9088;
  &:hover {
    color: #c96b5c;
  }
}

.main-content {
  background: #f5ece4;
  overflow-y: auto;
}
</style>
