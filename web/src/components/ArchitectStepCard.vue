<template>
  <div
    class="step-card"
    :class="{
      'step-card--active': isActive,
      'step-card--completed': status === 'completed',
      'step-card--outdated': status === 'outdated',
      'step-card--generating': status === 'generating'
    }"
    @click="$emit('click')"
  >
    <div class="step-card__header">
      <span class="step-card__number">{{ index + 1 }}</span>
      <span class="step-card__title">{{ title }}</span>
    </div>
    <div class="step-card__status">
      <el-tag :type="statusTagType" size="small">
        <el-icon v-if="status === 'generating'" class="is-loading"><Loading /></el-icon>
        <span v-else>{{ statusIcon }}</span>
        {{ statusText }}
      </el-tag>
    </div>
    <div v-if="summary1" class="step-card__summary">
      <span class="step-card__summary-line">{{ summary1 }}</span>
      <span v-if="summary2" class="step-card__summary-line step-card__summary-line--secondary">
        {{ summary2 }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Loading } from '@element-plus/icons-vue'

const props = defineProps({
  index: { type: Number, required: true },
  title: { type: String, required: true },
  status: { type: String, default: 'pending' }, // pending, current, completed, outdated, generating
  isActive: { type: Boolean, default: false },
  summary1: { type: String, default: '' },
  summary2: { type: String, default: '' }
})

defineEmits(['click'])

const statusIcon = computed(() => {
  const icons = {
    pending: '○',
    current: '●',
    completed: '✓',
    outdated: '⚠️'
  }
  return icons[props.status] || '○'
})

const statusText = computed(() => {
  const texts = {
    pending: '未开始',
    current: '当前',
    completed: '已完成',
    outdated: '可能过时',
    generating: '生成中'
  }
  return texts[props.status] || '未开始'
})

const statusTagType = computed(() => {
  const types = {
    pending: 'info',
    current: 'primary',
    completed: 'success',
    outdated: 'warning',
    generating: 'primary'
  }
  return types[props.status] || 'info'
})
</script>

<style scoped>
.step-card {
  padding: 16px;
  background: #ffffff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-left: 4px solid transparent;
  margin-bottom: 12px;
}

.step-card:hover {
  background: #f5f7fa;
}

.step-card--active {
  border-left-color: #0071e3;
  background: #f5f7fa;
}

.step-card--completed .step-card__number {
  background: #67c23a;
  color: white;
}

.step-card--outdated .step-card__number {
  background: #e6a23c;
  color: white;
}

.step-card--generating .step-card__number {
  background: #0071e3;
  color: white;
}

.step-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.step-card__number {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #909399;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.step-card__title {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
}

.step-card__status {
  margin-bottom: 8px;
}

.step-card__summary {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.step-card__summary-line {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.8);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.step-card__summary-line--secondary {
  color: rgba(0, 0, 0, 0.48);
}
</style>