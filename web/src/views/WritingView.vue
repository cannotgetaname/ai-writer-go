<template>
  <div class="writing-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - AI 写作</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showGenerateDialog">
          <el-icon><MagicStick /></el-icon>
          AI 生成
        </el-button>
        <el-button type="success" @click="showContinueDialog">
          <el-icon><Plus /></el-icon>
          续写
        </el-button>
        <el-button @click="saveContent" :loading="saving">
          <el-icon><Save /></el-icon>
          保存
        </el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <!-- 章节列表 -->
      <el-col :span="6">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>章节列表</span>
              <el-button size="small" @click="showNewChapterDialog">
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
          </template>
          <el-menu :default-active="currentChapter?.id?.toString()" @select="selectChapter">
            <el-menu-item v-for="ch in chapters" :key="ch.id" :index="ch.id.toString()">
              <span>第{{ ch.id }}章: {{ ch.title }}</span>
            </el-menu-item>
          </el-menu>
        </el-card>

        <!-- 伏笔面板 -->
        <el-card class="foreshadow-panel" v-if="currentChapter">
          <template #header>
            <span>伏笔追踪</span>
          </template>
          <div v-for="fs in foreshadows" :key="fs.id" class="foreshadow-item">
            <el-tag :type="fs.status === 'resolved' ? 'success' : 'warning'" size="small">
              {{ fs.status }}
            </el-tag>
            <span>{{ fs.content }}</span>
          </div>
        </el-card>
      </el-col>

      <!-- 编辑器 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>第{{ currentChapter?.id }}章: {{ currentChapter?.title }}</span>
              <el-tag v-if="wordCount > 0">{{ wordCount }} 字</el-tag>
            </div>
          </template>
          <el-input
            v-model="content"
            type="textarea"
            :rows="20"
            placeholder="在这里输入章节内容..."
            @input="updateWordCount"
          />
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="6">
        <!-- 大纲 -->
        <el-card v-if="currentChapter">
          <template #header>
            <span>章节大纲</span>
          </template>
          <el-input
            v-model="currentChapter.outline"
            type="textarea"
            :rows="4"
            placeholder="章节大纲..."
          />
        </el-card>

        <!-- 审稿结果 -->
        <el-card v-if="reviewResult" class="review-panel">
          <template #header>
            <span>审稿结果</span>
          </template>
          <div class="review-score">
            综合评分: <el-rate :model-value="reviewResult.overall_score / 20" disabled />
          </div>
          <div v-for="issue in reviewResult.issues" :key="issue.type" class="review-issue">
            <el-tag :type="issue.severity === '严重' ? 'danger' : 'warning'" size="small">
              {{ issue.type }}
            </el-tag>
            <p>{{ issue.description }}</p>
          </div>
          <el-button size="small" @click="review">重新审稿</el-button>
        </el-card>
      </el-col>
    </el-row>

    <!-- AI 生成对话框 -->
    <el-dialog v-model="generateDialogVisible" title="AI 生成章节" width="500px">
      <el-form :model="generateParams" label-width="100px">
        <el-form-item label="章节号">
          <el-input-number v-model="generateParams.chapter_id" :min="1" />
        </el-form-item>
        <el-form-item label="大纲">
          <el-input v-model="generateParams.outline" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="目标字数">
          <el-input-number v-model="generateParams.words" :min="500" :max="10000" :step="500" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="generate" :loading="generating">
          开始生成
        </el-button>
      </template>
    </el-dialog>

    <!-- 续写对话框 -->
    <el-dialog v-model="continueDialogVisible" title="AI 续写章节" width="500px">
      <el-form :model="continueParams" label-width="100px">
        <el-form-item label="续写字数">
          <el-input-number v-model="continueParams.write_words" :min="100" :max="5000" :step="100" />
        </el-form-item>
        <el-form-item label="当前字数">
          <el-tag>{{ wordCount }} 字</el-tag>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="continueDialogVisible = false">取消</el-button>
        <el-button type="success" @click="continueWrite" :loading="continuing">
          开始续写
        </el-button>
      </template>
    </el-dialog>

    <!-- 新建章节对话框 -->
    <el-dialog v-model="newChapterDialogVisible" title="新建章节" width="400px">
      <el-form :model="newChapter" label-width="80px">
        <el-form-item label="章节标题">
          <el-input v-model="newChapter.title" />
        </el-form-item>
        <el-form-item label="大纲">
          <el-input v-model="newChapter.outline" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newChapterDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createChapter">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { chapterApi, aiApi, foreshadowApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const chapters = ref([])
const currentChapter = ref(null)
const content = ref('')
const wordCount = ref(0)
const saving = ref(false)
const foreshadows = ref([])
const reviewResult = ref(null)
const reviewing = ref(false)

const generateDialogVisible = ref(false)
const generating = ref(false)
const generateParams = ref({
  chapter_id: 1,
  outline: '',
  words: 3000
})

const continueDialogVisible = ref(false)
const continuing = ref(false)
const continueParams = ref({
  write_words: 500
})

const newChapterDialogVisible = ref(false)
const newChapter = ref({ title: '', outline: '' })

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const loadChapters = async () => {
  try {
    const res = await chapterApi.list(bookId.value)
    chapters.value = res.data || []
    if (chapters.value.length > 0) {
      selectChapter(chapters.value[0].id.toString())
    }
  } catch (error) {
    console.error('加载章节失败:', error)
  }
}

const selectChapter = async (chapterId) => {
  const ch = chapters.value.find(c => c.id.toString() === chapterId)
  if (ch) {
    currentChapter.value = ch
    try {
      const res = await chapterApi.getContent(bookId.value, ch.id)
      content.value = res.data?.content || ''
      updateWordCount()
    } catch (error) {
      content.value = ''
    }
  }
}

const updateWordCount = () => {
  // 统计中文字符
  const chineseChars = (content.value.match(/[\u4e00-\u9fa5]/g) || []).length
  wordCount.value = chineseChars
}

const saveContent = async () => {
  if (!currentChapter.value) return
  saving.value = true
  try {
    await chapterApi.saveContent(bookId.value, currentChapter.value.id, content.value)
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
  saving.value = false
}

const showGenerateDialog = () => {
  if (currentChapter.value) {
    generateParams.value.chapter_id = currentChapter.value.id
    generateParams.value.outline = currentChapter.value.outline || ''
  }
  generateDialogVisible.value = true
}

const generate = async () => {
  generating.value = true
  try {
    const res = await aiApi.generate({
      book_name: bookId.value,
      chapter_id: generateParams.value.chapter_id,
      outline: generateParams.value.outline
    })
    if (res.data?.content) {
      content.value = res.data.content
      updateWordCount()
      ElMessage.success('内容生成成功')
    } else {
      ElMessage.info(res.data?.message || '生成完成，请查看结果')
    }
    generateDialogVisible.value = false
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  generating.value = false
}

const showContinueDialog = () => {
  if (!currentChapter.value) {
    ElMessage.warning('请先选择章节')
    return
  }
  if (!content.value) {
    ElMessage.warning('当前章节没有内容，请先生成内容')
    return
  }
  continueDialogVisible.value = true
}

const continueWrite = async () => {
  continuing.value = true
  try {
    const res = await aiApi.continue({
      book_name: bookId.value,
      chapter_id: currentChapter.value.id,
      write_words: continueParams.value.write_words
    })
    if (res.data?.content) {
      // 续写内容追加到现有内容
      content.value += '\n\n' + res.data.content
      updateWordCount()
      ElMessage.success('续写成功')
    } else {
      ElMessage.info(res.data?.message || '续写完成')
    }
    continueDialogVisible.value = false
  } catch (error) {
    ElMessage.error('续写失败: ' + (error.response?.data?.error || error.message))
  }
  continuing.value = false
}

const review = async () => {
  if (!currentChapter.value) return
  reviewing.value = true
  try {
    const res = await aiApi.review({
      book_name: bookId.value,
      chapter_id: currentChapter.value.id
    })
    if (res.data?.overall_score !== undefined) {
      reviewResult.value = res.data
      ElMessage.success('审稿完成')
    } else {
      ElMessage.info(res.data?.message || '审稿完成')
    }
  } catch (error) {
    ElMessage.error('审稿失败: ' + (error.response?.data?.error || error.message))
  }
  reviewing.value = false
}

const showNewChapterDialog = () => {
  newChapter.value = { title: '', outline: '' }
  newChapterDialogVisible.value = true
}

const createChapter = async () => {
  try {
    await chapterApi.create(bookId.value, newChapter.value)
    ElMessage.success('章节创建成功')
    newChapterDialogVisible.value = false
    loadChapters()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const loadForeshadows = async () => {
  try {
    const res = await foreshadowApi.list(bookId.value)
    foreshadows.value = res.data || []
  } catch (error) {
    foreshadows.value = []
  }
}

onMounted(() => {
  loadChapters()
  loadForeshadows()
})
</script>

<style scoped>
.writing-view {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.foreshadow-panel {
  margin-top: 20px;
}

.foreshadow-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 5px 0;
}

.review-panel {
  margin-top: 20px;
}

.review-score {
  margin-bottom: 10px;
}

.review-issue {
  margin-bottom: 10px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.review-issue p {
  margin: 5px 0 0 0;
  color: #666;
}
</style>