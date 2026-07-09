<template>
  <div>
    <div class="page-header">
      <h2>提醒管理</h2>
    </div>

    <el-row :gutter="16" class="stats-row">
      <el-col :xs="12" :sm="12" :md="6"><StatsCard label="总提醒数" :value="stats.total" icon="Bell" color="#409eff" /></el-col>
      <el-col :xs="12" :sm="12" :md="6"><StatsCard label="待处理" :value="stats.by_status?.pending || 0" icon="Clock" color="#e6a23c" /></el-col>
      <el-col :xs="12" :sm="12" :md="6"><StatsCard label="已完成" :value="stats.by_status?.done || 0" icon="CircleCheck" color="#67c23a" /></el-col>
      <el-col :xs="12" :sm="12" :md="6"><StatsCard label="已过期" :value="stats.overdue_count" icon="Warning" color="#f56c6c" /></el-col>
    </el-row>

    <div class="filter-bar">
      <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option label="待处理" value="pending" />
        <el-option label="已完成" value="done" />
        <el-option label="已忽略" value="ignored" />
      </el-select>
      <el-select v-model="filters.category" placeholder="分类" clearable style="width: 120px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option label="疫苗" value="vaccine" />
        <el-option label="体检" value="checkup" />
        <el-option label="复查" value="followup" />
        <el-option label="其他" value="other" />
      </el-select>
      <el-select v-model="filters.owner_type" placeholder="归属" clearable style="width: 120px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option label="宝宝" value="baby" />
        <el-option label="宝妈" value="mother" />
      </el-select>
    </div>

    <el-card shadow="hover">
      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="title" label="标题" show-overflow-tooltip />
        <el-table-column label="归属对象" width="140">
          <template #default="{ row }">
            <span v-if="row.owner_name">{{ row.owner_name }}</span>
            <span v-else>{{ row.owner_type === 'baby' ? '宝宝' : row.owner_type === 'mother' ? '宝妈' : '-' }} #{{ row.owner_id }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="source_record_type" label="记录类型" width="120">
          <template #default="{ row }">{{ row.source_record_type || '-' }}</template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="90">
          <template #default="{ row }">{{ categoryLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column prop="owner_type" label="归属" width="70">
          <template #default="{ row }">{{ row.owner_type === 'baby' ? '宝宝' : row.owner_type === 'mother' ? '宝妈' : row.owner_type }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="due_at" label="到期时间" width="170">
          <template #default="{ row }">{{ formatTime(row.due_at) }}</template>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import StatsCard from '../components/StatsCard.vue'
import { adminListReminders, adminReminderStats } from '../api/admin'

const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const filters = reactive({ status: '', category: '', owner_type: '' })
const stats = reactive<any>({ total: 0, by_status: {}, by_category: {}, by_owner_type: {}, overdue_count: 0 })

function formatTime(t: string) {
  return t ? new Date(t).toLocaleString('zh-CN') : '-'
}

function statusType(s: string) {
  return s === 'pending' ? 'warning' : s === 'done' ? 'success' : 'info'
}
function statusLabel(s: string) {
  return s === 'pending' ? '待处理' : s === 'done' ? '已完成' : '已忽略'
}

const categoryLabels: Record<string, string> = {
  vaccine: '疫苗',
  checkup: '体检',
  followup: '复查',
  other: '其他',
}
function categoryLabel(c: string) {
  return categoryLabels[c] || c || '-'
}

async function loadData() {
  loading.value = true
  try {
    const res = await adminListReminders({ page: page.value, page_size: pageSize, ...filters })
    items.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    const res = await adminReminderStats()
    Object.assign(stats, res.data)
  } catch {}
}

onMounted(() => {
  loadData()
  loadStats()
})
</script>

<style scoped>
.stats-row {
  margin-bottom: 16px;
}

.stats-row .el-col {
  margin-bottom: 8px;
}
</style>
