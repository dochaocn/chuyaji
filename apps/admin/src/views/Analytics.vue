<template>
  <div>
    <div class="page-header">
      <h2>数据分析</h2>
      <el-select v-model="days" style="width: 120px" @change="loadAll">
        <el-option :value="7" label="近 7 天" />
        <el-option :value="30" label="近 30 天" />
        <el-option :value="90" label="近 90 天" />
      </el-select>
    </div>

    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600">记录类型分布</span></template>
          <v-chart :option="pieOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header><span style="font-weight: 600">每日记录数量</span></template>
          <v-chart :option="lineOption" style="height: 350px" autoresize />
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover">
      <template #header><span style="font-weight: 600">每日活跃用户</span></template>
      <v-chart :option="barOption" style="height: 350px" autoresize />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart, LineChart, BarChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent, GridComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { adminAnalyticsRecords, adminAnalyticsActivity } from '../api/admin'
import { babyTypeLabel } from '../utils/labels'

use([PieChart, LineChart, BarChart, TitleComponent, TooltipComponent, LegendComponent, GridComponent, CanvasRenderer])

const days = ref(30)
const recordData = ref<any>({ daily_counts: [], type_distribution: [] })
const activityData = ref<any>({ daily_active_users: [] })

const WARM_COLORS = ['#c96b5c', '#d4a574', '#7dab98', '#da8a7e', '#b58a58', '#5e8f7a', '#e2be98', '#a84f42', '#a0c4b5', '#8a8279', '#e8aaa2', '#d5cec7', '#ecd6df', '#cce5da', '#f5ddd8']

const pieOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
  color: WARM_COLORS,
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    data: (recordData.value.type_distribution || []).map((d: any) => ({ name: babyTypeLabel(d.record_type), value: d.count })),
    emphasis: { itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(42, 36, 32, 0.2)' } },
  }],
}))

const lineOption = computed(() => {
  const rows = recordData.value.daily_counts || []
  return {
    tooltip: { trigger: 'axis', backgroundColor: '#fffdf9', borderColor: '#e8d8cc', textStyle: { color: '#3a322d' } },
    legend: { data: ['宝宝记录', '宝妈记录'] },
    grid: { left: 50, right: 20, bottom: 30, top: 40 },
    xAxis: { type: 'category', data: rows.map((r: any) => r.date), axisLine: { lineStyle: { color: '#e8d8cc' } }, axisLabel: { color: '#9a9088' } },
    yAxis: { type: 'value', minInterval: 1, axisLine: { show: false }, splitLine: { lineStyle: { color: '#f0e4d8' } }, axisLabel: { color: '#9a9088' } },
    color: ['#c96b5c', '#7dab98'],
    series: [
      { name: '宝宝记录', type: 'line', data: rows.map((r: any) => r.baby_records), smooth: true, lineStyle: { width: 2.5 }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(201, 107, 92, 0.15)' }, { offset: 1, color: 'rgba(201, 107, 92, 0)' }] } } },
      { name: '宝妈记录', type: 'line', data: rows.map((r: any) => r.mother_records), smooth: true, lineStyle: { width: 2.5 }, areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(125, 171, 152, 0.15)' }, { offset: 1, color: 'rgba(125, 171, 152, 0)' }] } } },
    ],
  }
})

const barOption = computed(() => {
  const rows = activityData.value.daily_active_users || []
  return {
    tooltip: { trigger: 'axis', backgroundColor: '#fffdf9', borderColor: '#e8d8cc', textStyle: { color: '#3a322d' } },
    grid: { left: 50, right: 20, bottom: 30, top: 20 },
    xAxis: { type: 'category', data: rows.map((r: any) => r.date), axisLine: { lineStyle: { color: '#e8d8cc' } }, axisLabel: { color: '#9a9088' } },
    yAxis: { type: 'value', minInterval: 1, axisLine: { show: false }, splitLine: { lineStyle: { color: '#f0e4d8' } }, axisLabel: { color: '#9a9088' } },
    series: [{ type: 'bar', data: rows.map((r: any) => r.count), itemStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: '#da8a7e' }, { offset: 1, color: '#c96b5c' }] }, borderRadius: [4, 4, 0, 0] } }],
  }
})

async function loadAll() {
  try {
    const [recRes, actRes] = await Promise.all([
      adminAnalyticsRecords(days.value),
      adminAnalyticsActivity(days.value),
    ])
    recordData.value = recRes.data
    activityData.value = actRes.data
  } catch {
    ElMessage.error('加载分析数据失败')
  }
}

onMounted(loadAll)
</script>
