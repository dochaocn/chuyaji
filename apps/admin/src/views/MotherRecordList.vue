<template>
  <div>
    <div class="page-header">
      <h2>宝妈记录</h2>
    </div>
    <div class="filter-bar">
      <el-select
        v-model="selectedMotherId"
        placeholder="选择宝妈"
        filterable
        clearable
        style="width: 240px"
        @change="onMotherChange"
      >
        <el-option label="全部宝妈" :value="0" />
        <el-option
          v-for="m in allMothers"
          :key="m.id"
          :label="`${m.name || '未命名'} (ID:${m.id})`"
          :value="m.id"
        />
      </el-select>
      <el-select v-model="filters.record_type" placeholder="记录类型" clearable style="width: 160px" @change="loadData">
        <el-option label="全部" value="" />
        <el-option v-for="t in recordTypeOptions" :key="t.value" :label="t.label" :value="t.value" />
      </el-select>
      <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期" end-placeholder="结束日期" value-format="YYYY-MM-DD" @change="loadData" />
      <el-button type="primary" @click="loadData">查询</el-button>
    </div>

    <el-alert v-if="selectedMotherId && selectedMother" :title="`当前查看：${selectedMother.name || '未命名'}`" type="info" show-icon :closable="false" style="margin-bottom: 12px">
      <template #default>
        <span v-if="selectedMother.status">{{ statusLabel(selectedMother.status) }}</span>
      </template>
    </el-alert>

    <el-card shadow="hover">
      <el-table :data="items" stripe v-loading="loading" @row-click="(row: any) => showDetail(row)">
        <el-table-column label="宝妈名" width="100">
          <template #default="{ row }">{{ motherNameMap[row.mother_id] || row.mother_id }}</template>
        </el-table-column>
        <el-table-column prop="record_type" label="类型" width="140">
          <template #default="{ row }">{{ motherTypeLabel(row.record_type) }}</template>
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

    <el-drawer v-model="drawerVisible" :title="`记录详情 #${detailRecord?.id}`" size="520px">
      <template v-if="detailRecord">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="宝妈名">{{ motherNameMap[detailRecord.mother_id] || detailRecord.mother_id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ motherTypeLabel(detailRecord.record_type) }}</el-descriptions-item>
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
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { adminListMotherRecords, adminDeleteMotherRecord, adminListMothers, adminListMotherRecordAttachments } from '../api/admin'
import { motherTypeLabel, motherRecordTypeLabels } from '../utils/labels'

const recordTypeOptions = Object.entries(motherRecordTypeLabels).map(([value, label]) => ({ value, label }))

const allMothers = ref<any[]>([])
const selectedMotherId = ref<number>(0)
const selectedMother = computed(() => allMothers.value.find(m => m.id === selectedMotherId.value) || null)
const motherNameMap = computed(() => {
  const map: Record<number, string> = {}
  for (const m of allMothers.value) {
    map[m.id] = m.name || '未命名'
  }
  return map
})

const items = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const dateRange = ref<string[]>([])
const filters = reactive({ record_type: '' })

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

function statusLabel(s: string) {
  return s === 'pregnant' ? '怀孕中' : s === 'postpartum' ? '产后恢复' : s === 'parenting' ? '育儿期' : s || '-'
}

function showDetail(row: any) {
  detailRecord.value = row
  detailAttachments.value = []
  drawerVisible.value = true
  adminListMotherRecordAttachments(row.id).then(res => {
    detailAttachments.value = res.data.items || []
  }).catch(() => {})
}

function onMotherChange() {
  page.value = 1
  loadData()
}

async function loadMothers() {
  try {
    const res = await adminListMothers({ page: 1, page_size: 999 })
    allMothers.value = res.data.items || []
  } catch {}
}

async function loadData() {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize, ...filters }
    if (selectedMotherId.value) {
      params.mother_id = selectedMotherId.value
    }
    if (dateRange.value?.length === 2) {
      params.from = dateRange.value[0]
      params.to = dateRange.value[1]
    }
    const res = await adminListMotherRecords(params)
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
    await adminDeleteMotherRecord(id)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadMothers()
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
</style>
