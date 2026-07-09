<template>
  <div>
    <div class="page-header">
      <h2>用户管理</h2>
    </div>
    <div class="filter-bar">
      <el-input v-model="query" placeholder="搜索用户" clearable style="width: 240px" @clear="loadData" @keyup.enter="loadData" />
      <el-button type="primary" @click="loadData">搜索</el-button>
    </div>
    <el-card shadow="hover">
      <el-table :data="items" stripe v-loading="loading" @row-click="(row: any) => router.push(`/users/${row.id}`)">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="open_id" label="微信OpenID" show-overflow-tooltip />
        <el-table-column prop="baby_count" label="宝宝数" width="80" />
        <el-table-column prop="mother_count" label="宝妈数" width="80" />
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="loadData"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminListUsers } from '../api/admin'

const router = useRouter()
const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const query = ref('')
const loading = ref(false)

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

async function loadData() {
  loading.value = true
  try {
    const res = await adminListUsers({ page: page.value, page_size: pageSize, q: query.value })
    items.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>
