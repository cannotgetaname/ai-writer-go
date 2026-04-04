<template>
  <div class="batch-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ form.book_name }} - 批量生成</h2>
    </div>

    <el-card>
      <el-form :model="form" label-width="120px">
        <el-form-item label="书籍名称">
          <el-input v-model="form.book_name" disabled style="width: 300px;" />
        </el-form-item>
        <el-form-item label="章节范围">
          <el-col :span="6">
            <el-input-number v-model="form.from" :min="1" :max="1000" placeholder="起始章节" />
          </el-col>
          <el-col :span="2" style="text-align: center;">到</el-col>
          <el-col :span="6">
            <el-input-number v-model="form.to" :min="form.from" :max="1000" placeholder="结束章节" />
          </el-col>
        </el-form-item>
        <el-form-item label="流式输出">
          <el-switch v-model="form.stream" />
        </el-form-item>
        <el-form-item label="重试次数">
          <el-input-number v-model="form.retry" :min="0" :max="5" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="startBatch" :loading="generating" :disabled="!form.book_name">
            <el-icon><VideoPlay /></el-icon>
            开始生成
          </el-button>
          <el-button @click="checkStatus" :disabled="!form.book_name">
            查看进度
          </el-button>
          <el-button type="danger" @click="resetProgress" :disabled="!form.book_name">
            重置进度
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 进度显示 -->
    <el-card v-if="progress" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>生成进度</span>
          <el-tag :type="progress.status === 'running' ? 'warning' : 'success'">
            {{ progress.status === 'running' ? '进行中' : '已完成' }}
          </el-tag>
        </div>
      </template>

      <el-progress
        :percentage="progress.percent"
        :status="progress.status === 'running' ? '' : 'success'"
        :stroke-width="20"
        style="margin-bottom: 20px;"
      />

      <el-descriptions :column="4" border>
        <el-descriptions-item label="当前章节">第 {{ progress.current }} 章</el-descriptions-item>
        <el-descriptions-item label="已完成">{{ progress.completed }} / {{ progress.to - progress.from + 1 }} 章</el-descriptions-item>
        <el-descriptions-item label="失败">{{ progress.failed }} 章</el-descriptions-item>
        <el-descriptions-item label="进度">{{ progress.percent.toFixed(1) }}%</el-descriptions-item>
      </el-descriptions>

      <div v-if="progress.failed_ids && progress.failed_ids.length > 0" style="margin-top: 10px;">
        <el-alert type="warning" :closable="false">
          失败章节: {{ progress.failed_ids.join(', ') }}
        </el-alert>
      </div>
    </el-card>

    <!-- 实时日志 -->
    <el-card v-if="generating" style="margin-top: 20px;">
      <template #header>
        <span>生成日志</span>
      </template>
      <div class="log-container">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <el-tag :type="log.type" size="small">{{ log.event }}</el-tag>
          <span class="log-time">{{ log.time }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </el-card>

    <!-- 章节内容预览 -->
    <el-card v-if="currentContent" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>第 {{ currentChapterId }}章 内容预览</span>
          <el-tag>{{ currentContent.length }} 字</el-tag>
        </div>
      </template>
      <div class="content-preview">{{ currentContent }}</div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { bookApi, chapterApi, batchApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const books = ref([])
const form = ref({
  book_name: '',
  from: 1,
  to: 10,
  stream: true,
  retry: 2
})

const generating = ref(false)
const progress = ref(null)
const logs = ref([])
const currentContent = ref('')
const currentChapterId = ref(0)

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const loadBooks = async () => {
  try {
    const res = await bookApi.list()
    books.value = res.data || []
    // 使用路由参数中的书籍
    if (bookId.value) {
      form.value.book_name = bookId.value
      loadChapterRange()
    } else if (books.value.length > 0) {
      form.value.book_name = books.value[0].name
      loadChapterRange()
    }
  } catch (error) {
    console.error('加载书籍失败:', error)
  }
}

// 监听 bookId 变化
watch(bookId, (newId) => {
  if (newId) {
    form.value.book_name = newId
    loadChapterRange()
  }
})

const loadChapterRange = async () => {
  if (!form.value.book_name) return
  try {
    const res = await chapterApi.list(form.value.book_name)
    const chapters = res.data || []
    if (chapters.length > 0) {
      form.value.from = chapters[0].id
      form.value.to = chapters[chapters.length - 1].id
    }
  } catch (error) {
    console.error('加载章节失败:', error)
  }
}

const checkStatus = async () => {
  try {
    const res = await batchApi.status(form.value.book_name)
    if (res.data.progress) {
      progress.value = res.data
    } else {
      ElMessage.info('没有进行中的批量任务')
      progress.value = null
    }
  } catch (error) {
    ElMessage.error('获取进度失败: ' + error.message)
  }
}

const resetProgress = async () => {
  try {
    await ElMessageBox.confirm('确定要重置进度吗？此操作不可恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await batchApi.reset(form.value.book_name)
    ElMessage.success('进度已重置')
    progress.value = null
    logs.value = []
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重置失败: ' + error.message)
    }
  }
}

const addLog = (event, message, type = 'info') => {
  logs.value.push({
    event,
    message,
    type,
    time: new Date().toLocaleTimeString()
  })
}

const startBatch = async () => {
  generating.value = true
  logs.value = []
  currentContent.value = ''

  addLog('start', `开始批量生成: 第${form.value.from}章 到 第${form.value.to}章`, 'info')

  try {
    const response = await batchApi.generate({
      book_name: form.value.book_name,
      from: form.value.from,
      to: form.value.to,
      stream: form.value.stream,
      retry: form.value.retry
    })

    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      const text = decoder.decode(value)
      const lines = text.split('\n')

      for (const line of lines) {
        if (line.startsWith('data:')) {
          try {
            const data = JSON.parse(line.substring(5).trim())
            handleEvent(data)
          } catch (e) {
            // 忽略解析错误
          }
        }
      }
    }
  } catch (error) {
    addLog('error', error.message, 'danger')
    ElMessage.error('批量生成失败: ' + error.message)
  } finally {
    generating.value = false
    checkStatus()
  }
}

const handleEvent = (data) => {
  switch (data.event || Object.keys(data)[0]) {
    case 'start':
      addLog('start', `书籍: ${data.book_name}, 范围: ${data.from}-${data.to}`, 'info')
      break
    case 'chapter_start':
      currentChapterId.value = data.chapter_id
      currentContent.value = ''
      addLog('chapter', `开始生成第 ${data.chapter_id} 章`, 'info')
      break
    case 'content':
      currentContent.value += data.content
      break
    case 'chapter_done':
      addLog('success', `第 ${data.chapter_id} 章完成, ${data.word_count} 字, 进度: ${data.progress}/${data.total}`, 'success')
      break
    case 'retry':
      addLog('retry', `第 ${data.chapter_id} 章重试 ${data.attempt}`, 'warning')
      break
    case 'chapter_error':
      addLog('error', `第 ${data.chapter_id} 章失败: ${data.error}`, 'danger')
      break
    case 'done':
      addLog('complete', `批量生成完成! 成功: ${data.completed}, 失败: ${data.failed}`, 'success')
      break
  }
}

onMounted(() => {
  loadBooks()
})
</script>

<style scoped>
.batch-view {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
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

.log-container {
  max-height: 400px;
  overflow-y: auto;
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
}

.log-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 13px;
}

.log-time {
  color: #909399;
  font-family: monospace;
}

.log-message {
  color: #303133;
}

.content-preview {
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 14px;
  line-height: 1.6;
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
}
</style>