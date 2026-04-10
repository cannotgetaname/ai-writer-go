<template>
  <div class="workspace">
    <!-- Header -->
    <div class="workspace__header">
      <h3 class="workspace__title">第{{ stepIndex + 1 }}步：{{ title }}</h3>
      <p class="workspace__description">{{ description }}</p>
    </div>

    <!-- Outdated Warning -->
    <el-alert
      v-if="isOutdated"
      type="warning"
      :closable="false"
      show-icon
      class="workspace__warning"
    >
      <template #title>
        {{ outdatedMessage }}
        <el-button type="primary" link size="small" @click="$emit('regenerate')">
          重新生成
        </el-button>
      </template>
    </el-alert>

    <!-- Context Preview (collapsible) -->
    <div v-if="contextPreview" class="workspace__context">
      <div class="workspace__context-header" @click="contextExpanded = !contextExpanded">
        <el-icon><ArrowRight :class="{ 'is-rotate': contextExpanded }" /></el-icon>
        <span>上下文预览</span>
      </div>
      <div v-show="contextExpanded" class="workspace__context-content">
        {{ contextPreview }}
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="workspace__actions">
      <slot name="actions"></slot>
    </div>

    <!-- Content Area -->
    <div class="workspace__content">
      <slot></slot>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="workspace__loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <p>{{ loadingText }}</p>
      <el-button @click="$emit('cancel')">取消</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Loading, ArrowRight } from '@element-plus/icons-vue'

const props = defineProps({
  stepIndex: { type: Number, required: true },
  title: { type: String, required: true },
  description: { type: String, default: '' },
  contextPreview: { type: String, default: '' },
  isOutdated: { type: Boolean, default: false },
  outdatedMessage: { type: String, default: '前置数据已更新，当前内容可能过时' },
  loading: { type: Boolean, default: false },
  loadingText: { type: String, default: '正在生成...' }
})

defineEmits(['regenerate', 'cancel'])

const contextExpanded = ref(false)
</script>

<style scoped>
.workspace {
  background: #ffffff;
  border-radius: 8px;
  padding: 24px;
  min-height: calc(100vh - 140px);
  min-width: 600px;
  position: relative;
}

.workspace__header {
  margin-bottom: 20px;
}

.workspace__title {
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0 0 8px 0;
}

.workspace__description {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.8);
  margin: 0;
}

.workspace__warning {
  margin-bottom: 16px;
}

.workspace__context {
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}

.workspace__context-header {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.8);
}

.workspace__context-header:hover {
  background: #ebeef5;
}

.workspace__context-header .is-rotate {
  transform: rotate(90deg);
}

.workspace__context-content {
  padding: 0 16px 12px 32px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.6);
  line-height: 1.6;
}

.workspace__actions {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
}

.workspace__content {
  flex: 1;
}

.workspace__loading {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  z-index: 10;
}

.workspace__loading p {
  color: #0071e3;
  font-size: 14px;
}
</style>