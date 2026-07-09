<template>
  <div>
    <div class="stats-grid">
      <StatsCard label="用户总数" :value="overview.user_count" icon="User" color="#409eff" />
      <StatsCard label="宝宝档案" :value="overview.baby_count" icon="UserFilled" color="#67c23a" />
      <StatsCard label="宝妈档案" :value="overview.mother_count" icon="Female" color="#e6a23c" />
      <StatsCard label="宝宝记录" :value="overview.record_count" icon="Document" color="#f56c6c" />
      <StatsCard label="宝妈记录" :value="overview.mother_record_count" icon="Notebook" color="#909399" />
      <StatsCard label="待处理提醒" :value="overview.reminder_pending_count" icon="Bell" color="#e6a23c" />
    </div>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :xs="24" :sm="24" :md="12">
        <el-card shadow="hover" style="margin-bottom: 16px">
          <template #header><span style="font-weight: 600">最近用户</span></template>
          <el-table :data="overview.recent_users" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="open_id" label="微信OpenID" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="12">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600">最近记录</span></template>
          <el-table :data="overview.recent_records" stripe size="small">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="record_type" label="类型" width="120">
              <template #default="{ row }">{{ babyTypeLabel(row.record_type) }}</template>
            </el-table-column>
            <el-table-column prop="summary" label="摘要" show-overflow-tooltip />
            <el-table-column prop="occurred_at" label="时间" width="170">
              <template #default="{ row }">{{ formatTime(row.occurred_at) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import StatsCard from '../components/StatsCard.vue'
import { adminOverview } from '../api/admin'
import { babyTypeLabel } from '../utils/labels'

const overview = reactive({
  user_count: 0,
  baby_count: 0,
  mother_count: 0,
  record_count: 0,
  mother_record_count: 0,
  reminder_pending_count: 0,
  recent_users: [] as any[],
  recent_records: [] as any[],
})

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(async () => {
  try {
    const res = await adminOverview()
    Object.assign(overview, res.data)
  } catch {
    ElMessage.error('加载数据失败')
  }
})
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
