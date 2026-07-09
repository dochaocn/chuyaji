<template>
  <el-card shadow="hover" class="stats-card">
    <div class="stats-content">
      <div class="stats-info">
        <div class="stats-label">{{ label }}</div>
        <div class="stats-value">{{ value }}</div>
      </div>
      <div class="stats-icon" :style="{ background: bg }">
        <el-icon :size="24"><component :is="icon" /></el-icon>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  value: number | string
  icon: string
  color?: string
}>()

const colorMap: Record<string, { bg: string; text: string }> = {
  '#409eff': { bg: 'linear-gradient(160deg, #da8a7e 0%, #a84f42 100%)', text: '#c96b5c' },
  '#67c23a': { bg: 'linear-gradient(160deg, #a0c4b5 0%, #5e8f7a 100%)', text: '#7dab98' },
  '#e6a23c': { bg: 'linear-gradient(160deg, #e2be98 0%, #b58a58 100%)', text: '#d4a574' },
  '#f56c6c': { bg: 'linear-gradient(160deg, #da8a7e 0%, #a84f42 100%)', text: '#c96b5c' },
  '#909399': { bg: 'linear-gradient(160deg, #d5cec7 0%, #8a8279 100%)', text: '#6b6058' },
}

const bg = computed(() => colorMap[props.color || '']?.bg || 'linear-gradient(160deg, #da8a7e 0%, #a84f42 100%)')
</script>

<style scoped>
.stats-card {
  cursor: default;
}

.stats-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stats-label {
  font-size: 13px;
  color: #9a9088;
  margin-bottom: 6px;
}

.stats-value {
  font-size: 28px;
  font-weight: 700;
  color: #3a322d;
}

.stats-icon {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 0 4px 12px rgba(201, 107, 92, 0.15);
}
</style>
