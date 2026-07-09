<template>
  <div v-loading="loading">
    <div class="page-header">
      <h2>用户详情 #{{ userId }}</h2>
      <el-button @click="router.push('/users')">返回列表</el-button>
    </div>

    <el-card shadow="hover" style="margin-bottom: 16px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID">{{ user.id }}</el-descriptions-item>
        <el-descriptions-item label="微信OpenID">{{ user.open_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ formatTime(user.created_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600">宝宝档案 ({{ babies.length }})</span></template>
          <el-table :data="babies" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="nickname" label="昵称" />
            <el-table-column prop="gender" label="性别" width="70">
              <template #default="{ row }">{{ row.gender === 'male' ? '男' : row.gender === 'female' ? '女' : '-' }}</template>
            </el-table-column>
            <el-table-column prop="birth_date" label="出生日期" width="120">
              <template #default="{ row }">{{ row.birth_date ? row.birth_date.slice(0, 10) : '-' }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600">宝妈档案 ({{ mothers.length }})</span></template>
          <el-table :data="mothers" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">{{ statusLabel(row.status) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminGetUser } from '../api/admin'

const route = useRoute()
const router = useRouter()
const userId = Number(route.params.id)
const loading = ref(false)
const user = ref<any>({})
const babies = ref<any[]>([])
const mothers = ref<any[]>([])

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

function statusLabel(s: string) {
  return s === 'pregnant' ? '怀孕中' : s === 'postpartum' ? '产后恢复' : s === 'parenting' ? '育儿期' : s || '-'
}

onMounted(async () => {
  loading.value = true
  try {
    const res = await adminGetUser(userId)
    user.value = res.data.user || {}
    babies.value = res.data.babies || []
    mothers.value = res.data.mothers || []
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
})
</script>
