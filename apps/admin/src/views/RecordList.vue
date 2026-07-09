<template>
  <div>
    <div class="page-header">
      <h2>宝宝记录</h2>
    </div>
    <div class="filter-bar">
      <el-select
        v-model="selectedBabyId"
        placeholder="选择宝宝"
        filterable
        clearable
        style="width: 240px"
        @change="onBabyChange"
      >
        <el-option label="全部宝宝" :value="0" />
        <el-option
          v-for="b in allBabies"
          :key="b.id"
          :label="`${b.nickname} (ID:${b.id})`"
          :value="b.id"
        />
      </el-select>
      <el-select v-model="filters.phase" placeholder="阶段" clearable style="width: 120px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option label="孕中" value="prenatal" />
        <el-option label="出生后" value="postnatal" />
      </el-select>
      <el-select v-model="filters.record_type" placeholder="记录类型" clearable style="width: 160px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option v-for="t in recordTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
      </el-select>
      <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" @change="loadData" />
      <el-button type="primary" @click="loadData">查询</el-button>
    </div>

    <el-alert v-if="selectedBabyId && selectedBaby" :title="`当前查看：${selectedBaby.nickname}`" type="info" show-icon :closable="false" style="margin-bottom: 12px">
      <template #default>
        <span v-if="selectedBaby.gender">{{ selectedBaby.gender === 'male' ? '男' : '女' }} · </span>
        <span v-if="selectedBaby.birth_date">出生 {{ selectedBaby.birth_date.slice(0, 10) }}</span>
      </template>
    </el-alert>

    <el-card shadow="hover">
      <el-table :data="items" stripe v-loading="loading" @row-click="(row: any) => showDetail(row)">
        <el-table-column label="宝宝名" width="100">
          <template #default="{ row }">{{ babyNameMap[row.baby_id] || row.baby_id }}</template>
        </el-table-column>
        <el-table-column prop="phase" label="阶段" width="80">
          <template #default="{ row }">{{ phaseLabels[row.phase] || row.phase }}</template>
        </el-table-column>
        <el-table-column prop="record_type" label="类型" width="140">
          <template #default="{ row }">{{ babyTypeLabel(row.record_type) }}</template>
        </el-table-column>
        <el-table-column prop="summary" label="摘要" show-overflow-tooltip />
        <el-table-column prop="occurred_at" label="时间" width="170">
          <template #default="{ row }">{{ formatTime(row.occurred_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-popconfirm title="确定删除?" @confirm.stop="handleDelete(row.id)">
              <template #reference><el-button type="danger" text size="small" @click.stop>删除</el-button></template>
            </el-popconfirm>
          </template>
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

    <el-drawer v-model="drawerVisible" :title="`记录详情 #${detailRecord?.id}`" :size="isMobile ? '100%' : '520px'">
      <template v-if="detailRecord">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="宝宝名">{{ babyNameMap[detailRecord.baby_id] || detailRecord.baby_id }}</el-descriptions-item>
          <el-descriptions-item label="阶段">
            <el-tag :type="detailRecord.phase === 'prenatal' ? 'warning' : 'success'" size="small">{{ phaseLabels[detailRecord.phase] || detailRecord.phase }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="类型">{{ babyTypeLabel(detailRecord.record_type) }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ formatTime(detailRecord.occurred_at) }}</el-descriptions-item>
          <el-descriptions-item label="摘要">{{ detailRecord.summary || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatTime(detailRecord.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(detailRecord.updated_at) }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="payloadEntries.length" style="margin-top: 16px">
          <h4 style="margin-bottom: 8px; font-size: 14px; color: #3a322d">详细数据</h4>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item v-for="[key, val] in payloadEntries" :key="key" :label="key">{{ val }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <div v-if="detailAttachments.length" style="margin-top: 16px">
          <h4 style="margin-bottom: 8px; font-size: 14px; color: #3a322d">图片 ({{ detailAttachments.length }})</h4>
          <div class="attachment-grid">
            <el-image
              v-for="att in detailAttachments"
              :key="att.id"
              :src="att.thumb_url || att.url"
              :preview-src-list="[att.url]"
              fit="cover"
              class="attachment-img"
            />
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { adminListRecords, adminDeleteRecord, adminListBabies, adminListRecordAttachments } from '../api/admin'
import { babyTypeLabel, babyRecordTypeLabels, phaseLabels, babyRecordTypePhase } from '../utils/labels'

const allRecordTypeOptions = Object.entries(babyRecordTypeLabels).map(([value, label]) => ({ value, label }))
const recordTypeOptions = ref(allRecordTypeOptions)

const allBabies = ref<any[]>([])
const selectedBabyId = ref<number>(0)
const selectedBaby = computed(() => allBabies.value.find(b => b.id === selectedBabyId.value) || null)
const babyNameMap = computed(() => {
  const map: Record<number, string> = {}
  for (const b of allBabies.value) {
    map[b.id] = b.nickname || '未命名'
  }
  return map
})

const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const dateRange = ref<string[]>([])
const filters = reactive({ phase: '', record_type: '' })

watch(() => filters.phase, (newPhase) => {
  if (newPhase) {
    recordTypeOptions.value = allRecordTypeOptions.filter(t => babyRecordTypePhase[t.value] === newPhase)
  } else {
    recordTypeOptions.value = allRecordTypeOptions
  }
  if (filters.record_type && !recordTypeOptions.value.some(t => t.value === filters.record_type)) {
    filters.record_type = ''
  }
})

const isMobile = ref(window.innerWidth <= 768)
function onResize() { isMobile.value = window.innerWidth <= 768 }
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))

const drawerVisible = ref(false)
const detailRecord = ref<any>(null)
const detailAttachments = ref<any[]>([])

const payloadEntries = computed(() => {
  if (!detailRecord.value?.payload) return []
  try {
    const obj = typeof detailRecord.value.payload === 'string' ? JSON.parse(detailRecord.value.payload) : detailRecord.value.payload
    if (!obj || typeof obj !== 'object') return []
    return Object.entries(obj).filter(([, v]) => v != null && v !== '')
  } catch {
    return []
  }
})

function formatTime(t: string) {
  if (!t) return '-'
  return t.slice(0, 10)
}

function showDetail(row: any) {
  detailRecord.value = row
  detailAttachments.value = []
  drawerVisible.value = true
  adminListRecordAttachments(row.id).then(res => {
    detailAttachments.value = res.data.items || []
  }).catch(() => {})
}

function onBabyChange() {
  page.value = 1
  loadData()
}

async function loadBabies() {
  try {
    const res = await adminListBabies({ page: 1, page_size: 999 })
    allBabies.value = res.data.items || []
  } catch {}
}

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize, ...filters }
    if (selectedBabyId.value) {
      params.baby_id = selectedBabyId.value
    }
    if (dateRange.value?.length === 2) {
      params.from = dateRange.value[0]
      params.to = dateRange.value[1]
    }
    const res = await adminListRecords(params)
    items.value = res.data.items || []
    total.value = res.data.total || 0
  } catch {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await adminDeleteRecord(id)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadBabies()
  loadData()
})
</script>

<style scoped>
.attachment-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.attachment-img {
  width: 100%;
  height: 120px;
  border-radius: 6px;
  cursor: pointer;
}
@media (max-width: 768px) {
  .attachment-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
