<template>
  <div class="sync-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 状态同步</h2>
    </div>

    <el-row :gutter="20">
      <!-- 提取面板 -->
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>提取状态变更</span>
          </template>
          <el-form label-width="80px">
            <el-form-item label="章节号">
              <el-input-number v-model="extractChapterId" :min="1" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="extractChanges" :loading="extracting">
                <el-icon><Search /></el-icon>
                提取变更
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 操作按钮 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <span>批量操作</span>
          </template>
          <div class="action-buttons">
            <el-button type="success" @click="applyAll" :disabled="!pendingChanges?.changes?.length">
              应用全部
            </el-button>
            <el-button type="danger" @click="rejectAll" :disabled="!pendingChanges?.changes?.length">
              丢弃全部
            </el-button>
            <el-button @click="loadPending" :disabled="!bookId">
              刷新列表
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 待审核列表 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>待审核变更 ({{ pendingChanges?.changes?.length || 0 }})</span>
              <el-tag v-if="pendingChanges?.extracted_at" type="info" size="small">
                提取时间: {{ formatTime(pendingChanges.extracted_at) }}
              </el-tag>
            </div>
          </template>

          <el-empty v-if="!pendingChanges?.changes?.length" description="暂无待审核变更" />

          <div v-else class="changes-list">
            <el-card v-for="(change, index) in pendingChanges.changes" :key="change.id" class="change-item" shadow="hover">
              <div class="change-header">
                <el-tag :type="getChangeTypeTag(change.type)">{{ change.type }}</el-tag>
                <span class="entity-name">{{ change.entity }}</span>
                <span class="change-id">ID: {{ change.id?.substring(0, 8) }}</span>
              </div>

              <div class="change-content">
                <div v-if="change.field" class="change-field">
                  <strong>字段:</strong> {{ change.field }}
                </div>
                <div v-if="change.old_value" class="change-old">
                  <strong>旧值:</strong> {{ change.old_value }}
                </div>
                <div class="change-new">
                  <strong>新值:</strong> {{ change.new_value }}
                </div>
                <div v-if="change.reason" class="change-reason">
                  <strong>原因:</strong> {{ change.reason }}
                </div>
              </div>

              <div class="change-actions">
                <el-button type="success" size="small" @click="applyChange(change.id)">
                  应用
                </el-button>
                <el-button type="danger" size="small" @click="rejectChange(change.id)">
                  丢弃
                </el-button>
              </div>
            </el-card>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { syncApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const extractChapterId = ref(1)
const extracting = ref(false)
const pendingChanges = ref(null)

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const extractChanges = async () => {
  extracting.value = true
  try {
    const res = await syncApi.extract({
      book_name: bookId.value,
      chapter_id: extractChapterId.value
    })
    if (res.data.changes?.length > 0) {
      ElMessage.success(`提取完成，发现 ${res.data.changes.length} 条变更`)
      pendingChanges.value = res.data
    } else {
      ElMessage.info('未检测到状态变更')
      pendingChanges.value = null
    }
  } catch (error) {
    ElMessage.error('提取失败: ' + (error.response?.data?.error || error.message))
  }
  extracting.value = false
}

const loadPending = async () => {
  try {
    const res = await syncApi.pending(bookId.value)
    if (res.data.changes?.length > 0) {
      pendingChanges.value = res.data
    } else {
      pendingChanges.value = null
    }
  } catch (error) {
    console.error('加载待审核变更失败:', error)
  }
}

const applyChange = async (changeId) => {
  try {
    const res = await syncApi.apply({
      book_name: bookId.value,
      change_id: changeId
    })
    ElMessage.success(`已应用变更`)
    loadPending()
  } catch (error) {
    ElMessage.error('应用失败: ' + (error.response?.data?.error || error.message))
  }
}

const rejectChange = async (changeId) => {
  try {
    await ElMessageBox.confirm('确定要丢弃此变更吗？', '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await syncApi.reject({
      book_name: bookId.value,
      change_id: changeId
    })
    ElMessage.success('已丢弃变更')
    loadPending()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const applyAll = async () => {
  try {
    await ElMessageBox.confirm('确定要应用所有变更吗？', '确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await syncApi.apply({ book_name: bookId.value })
    ElMessage.success(`已应用 ${res.data.applied} 条变更`)
    loadPending()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('应用失败')
    }
  }
}

const rejectAll = async () => {
  try {
    await ElMessageBox.confirm('确定要丢弃所有变更吗？此操作不可恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await syncApi.reject({ book_name: bookId.value })
    ElMessage.success(`已丢弃 ${res.data.rejected} 条变更`)
    pendingChanges.value = null
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

const getChangeTypeTag = (type) => {
  const typeMap = {
    'character_status': 'warning',
    'item_update': 'info',
    'new_character': 'success',
    'new_item': 'success',
    'new_location': 'success',
    'relation_update': 'primary',
    'time_progression': ''
  }
  return typeMap[type] || ''
}

const formatTime = (time) => {
  if (!time) return ''
  return new Date(time).toLocaleString()
}

onMounted(() => {
  loadPending()
})

watch(bookId, () => {
  loadPending()
})
</script>

<style scoped>
.sync-view {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.changes-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.change-item {
  position: relative;
}

.change-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.entity-name {
  font-weight: bold;
  font-size: 16px;
}

.change-id {
  color: #909399;
  font-size: 12px;
  margin-left: auto;
}

.change-content {
  margin-bottom: 10px;
  font-size: 14px;
  color: #606266;
}

.change-content > div {
  margin-bottom: 5px;
}

.change-field, .change-old, .change-new, .change-reason {
  padding: 5px 0;
}

.change-old {
  color: #f56c6c;
}

.change-new {
  color: #67c23a;
}

.change-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}
</style>