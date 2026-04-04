<template>
  <div class="timeline-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 时间线</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 章节时间线 -->
      <el-tab-pane label="章节时间线" name="chapters">
        <el-card>
          <el-table :data="timelineData" stripe>
            <el-table-column prop="chapter_id" label="章节" width="80" />
            <el-table-column prop="title" label="标题" />
            <el-table-column prop="time_info.label" label="时间标记" />
            <el-table-column prop="time_info.duration" label="时长" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 叙事线程 -->
      <el-tab-pane label="叙事线程" name="threads">
        <div class="tab-header">
          <el-button type="primary" @click="showNewThreadDialog">
            <el-icon><Plus /></el-icon>
            新建线程
          </el-button>
        </div>
        <el-row :gutter="20">
          <el-col :span="8" v-for="thread in threads" :key="thread.id">
            <el-card class="thread-card">
              <div class="thread-header">
                <h4>{{ thread.name }}</h4>
                <el-tag :type="thread.status === 'active' ? 'success' : thread.status === 'paused' ? 'warning' : 'info'" size="small">
                  {{ thread.status }}
                </el-tag>
              </div>
              <div class="thread-info">
                <p><strong>类型:</strong> {{ thread.type }}</p>
                <p><strong>目标:</strong> {{ thread.goal }}</p>
                <p><strong>起始章节:</strong> {{ thread.start_chapter }}</p>
                <p><strong>最后活跃:</strong> {{ thread.last_active_chapter }}</p>
              </div>
              <div class="thread-chapters">
                <span v-for="ch in thread.chapters" :key="ch" class="chapter-dot">
                  {{ ch }}
                </span>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

      <!-- 因果链 -->
      <el-tab-pane label="因果链" name="causal">
        <el-card>
          <div v-for="event in causalEvents" :key="event.id" class="causal-node">
            <div class="cause">
              <el-tag type="info" size="small">因</el-tag>
              <span>{{ event.cause }}</span>
            </div>
            <div class="arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
            <div class="event">
              <el-tag type="primary" size="small">事</el-tag>
              <span>{{ event.event }}</span>
            </div>
            <div class="arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
            <div class="effect">
              <el-tag type="success" size="small">果</el-tag>
              <span>{{ event.effect }}</span>
            </div>
            <div class="arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
            <div class="decision">
              <el-tag type="warning" size="small">决</el-tag>
              <span>{{ event.decision }}</span>
            </div>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 新建线程对话框 -->
    <el-dialog v-model="newThreadDialogVisible" title="新建叙事线程" width="400px">
      <el-form :model="newThread" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="newThread.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="newThread.type">
            <el-option label="主线" value="main" />
            <el-option label="支线" value="sub" />
            <el-option label="并行线" value="parallel" />
            <el-option label="闪回线" value="flashback" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标">
          <el-input v-model="newThread.goal" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newThreadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createThread">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { timelineApi, causalApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const activeTab = ref('chapters')
const timelineData = ref([])
const threads = ref([])
const causalEvents = ref([])

const newThreadDialogVisible = ref(false)
const newThread = ref({ name: '', type: 'sub', goal: '' })

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const loadTimeline = async () => {
  try {
    const res = await timelineApi.get(bookId.value)
    timelineData.value = res.data || []
  } catch (error) {
    timelineData.value = []
  }
}

const loadThreads = async () => {
  try {
    const res = await timelineApi.getThreads(bookId.value)
    threads.value = res.data || []
  } catch (error) {
    threads.value = []
  }
}

const loadCausalEvents = async () => {
  try {
    const res = await causalApi.get(bookId.value)
    causalEvents.value = res.data || []
  } catch (error) {
    causalEvents.value = []
  }
}

const showNewThreadDialog = () => {
  newThread.value = { name: '', type: 'sub', goal: '' }
  newThreadDialogVisible.value = true
}

const createThread = async () => {
  try {
    await timelineApi.createThread(bookId.value, newThread.value)
    ElMessage.success('线程创建成功')
    newThreadDialogVisible.value = false
    loadThreads()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

onMounted(() => {
  loadTimeline()
  loadThreads()
  loadCausalEvents()
})
</script>

<style scoped>
.timeline-view {
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

.tab-header {
  margin-bottom: 15px;
}

.thread-card {
  height: 100%;
}

.thread-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.thread-header h4 {
  margin: 0;
}

.thread-info p {
  margin: 5px 0;
  color: #666;
}

.thread-chapters {
  margin-top: 10px;
  display: flex;
  gap: 5px;
}

.chapter-dot {
  display: inline-block;
  width: 24px;
  height: 24px;
  background: #409eff;
  color: white;
  border-radius: 50%;
  text-align: center;
  line-height: 24px;
  font-size: 12px;
}

.causal-node {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px;
  margin-bottom: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.cause, .event, .effect, .decision {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
}

.arrow {
  color: #409eff;
}
</style>