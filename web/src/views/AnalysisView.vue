<template>
  <div class="analysis-view">
    <div class="page-header">
      <h2>拆书分析</h2>
    </div>

    <el-row :gutter="20">
      <!-- 左侧：上传和解析 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>上传作品</span>
          </template>
          <el-upload
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            accept=".txt"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
              拖拽 TXT 文件到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                支持 UTF-8 编码的 TXT 文件，系统会自动识别章节结构
              </div>
            </template>
          </el-upload>
        </el-card>

        <!-- 解析结果 -->
        <el-card v-if="parseResult" style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>解析结果</span>
              <el-tag>{{ parseResult.chapter_count }} 章</el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="章节总数">{{ parseResult.chapter_count }}</el-descriptions-item>
            <el-descriptions-item label="总字数">{{ parseResult.total_words?.toLocaleString() }}</el-descriptions-item>
          </el-descriptions>
          <div class="chapter-list">
            <el-scrollbar height="200px">
              <div v-for="ch in parseResult.chapters?.slice(0, 20)" :key="ch.num" class="chapter-item">
                <span>第{{ ch.num }}章: {{ ch.title }}</span>
                <span class="word-count">{{ ch.word_count }}字</span>
              </div>
              <div v-if="parseResult.chapter_count > 20" class="more-info">
                ... 还有 {{ parseResult.chapter_count - 20 }} 章
              </div>
            </el-scrollbar>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：分析工具 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>分析工具</span>
          </template>
          <div class="analysis-tools">
            <el-button type="primary" @click="analyzeContent" :loading="analyzing" :disabled="!parseResult">
              <el-icon><Search /></el-icon>
              AI 内容分析
            </el-button>
            <el-button type="success" @click="extractOutline" :loading="extracting" :disabled="!parseResult">
              <el-icon><List /></el-icon>
              提取大纲
            </el-button>
            <el-button type="warning" @click="compareWithMine" :loading="comparing" :disabled="!parseResult || !selectedBook">
              <el-icon><Sort /></el-icon>
              与我的作品对比
            </el-button>
          </div>
          <el-select v-model="selectedBook" placeholder="选择对比目标书籍" style="margin-top: 15px; width: 100%;" clearable>
            <el-option v-for="book in myBooks" :key="book.id" :label="book.name" :value="book.id" />
          </el-select>
        </el-card>

        <!-- 分析结果 -->
        <el-card v-if="analysisResult" style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>分析结果</span>
              <el-button size="small" @click="importAnalysis">导入到新书</el-button>
            </div>
          </template>
          <el-tabs>
            <el-tab-pane label="内容摘要">
              <div class="result-section">
                {{ analysisResult.summary }}
              </div>
            </el-tab-pane>
            <el-tab-pane label="人物">
              <div v-for="char in analysisResult.characters" :key="char.name" class="char-item">
                <el-tag>{{ char.role }}</el-tag>
                <strong>{{ char.name }}</strong>
                <p>{{ char.description }}</p>
              </div>
              <el-empty v-if="!analysisResult.characters?.length" description="暂无人物信息" />
            </el-tab-pane>
            <el-tab-pane label="世界观">
              <div class="result-section">
                {{ analysisResult.world_setting }}
              </div>
            </el-tab-pane>
            <el-tab-pane label="写作风格">
              <div class="result-section">
                {{ analysisResult.writing_style }}
              </div>
            </el-tab-pane>
            <el-tab-pane label="大纲结构" v-if="analysisResult.outline">
              <el-tree :data="analysisResult.outline" default-expand-all />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { analysisApi, bookApi } from '@/api'

const parseResult = ref(null)
const analysisResult = ref(null)
const analyzing = ref(false)
const extracting = ref(false)
const comparing = ref(false)
const fileContent = ref('')
const fileName = ref('')
const myBooks = ref([])
const selectedBook = ref('')

const loadMyBooks = async () => {
  try {
    const res = await bookApi.list()
    myBooks.value = res.data || []
  } catch (error) {
    myBooks.value = []
  }
}

const handleFileChange = (file) => {
  fileName.value = file.name
  const reader = new FileReader()
  reader.onload = (e) => {
    fileContent.value = e.target.result
    // 解析章节结构
    const lines = e.target.result.split('\n')
    let chapterCount = 0
    const chapters = []

    lines.forEach(line => {
      if (line.match(/^第[一二三四五六七八九十百千万零0-9]+[章节回卷部篇]/)) {
        chapterCount++
        chapters.push({
          num: chapterCount,
          title: line.trim().substring(0, 30),
          word_count: 0
        })
      }
    })

    // 统计字数
    const chineseChars = (e.target.result.match(/[\u4e00-\u9fa5]/g) || []).length

    parseResult.value = {
      chapter_count: chapterCount || 1,
      total_words: chineseChars,
      chapters: chapters.length > 0 ? chapters : [{ num: 1, title: '全文', word_count: chineseChars }]
    }

    ElMessage.success(`解析完成：${chapterCount || 1} 章，${chineseChars.toLocaleString()} 字`)
  }
  reader.readAsText(file.raw)
}

const analyzeContent = async () => {
  if (!fileContent.value) return
  analyzing.value = true
  try {
    const res = await analysisApi.analyze({
      content: fileContent.value.substring(0, 50000), // 限制长度
      filename: fileName.value
    })
    analysisResult.value = res.data || {}
    ElMessage.success('内容分析完成')
  } catch (error) {
    ElMessage.error('分析失败: ' + (error.response?.data?.error || error.message))
  }
  analyzing.value = false
}

const extractOutline = async () => {
  if (!fileContent.value) return
  extracting.value = true
  try {
    const res = await analysisApi.parse({
      content: fileContent.value.substring(0, 50000),
      filename: fileName.value
    })
    if (analysisResult.value) {
      analysisResult.value.outline = res.data?.outline || []
    } else {
      analysisResult.value = { outline: res.data?.outline || [] }
    }
    ElMessage.success('大纲提取完成')
  } catch (error) {
    ElMessage.error('提取失败: ' + (error.response?.data?.error || error.message))
  }
  extracting.value = false
}

const compareWithMine = async () => {
  if (!fileContent.value || !selectedBook.value) return
  comparing.value = true
  try {
    const res = await analysisApi.analyze({
      content: fileContent.value.substring(0, 30000),
      filename: fileName.value,
      compare_with: selectedBook.value
    })
    analysisResult.value = res.data || {}
    ElMessage.success('对比分析完成')
  } catch (error) {
    ElMessage.error('对比失败: ' + (error.response?.data?.error || error.message))
  }
  comparing.value = false
}

const importAnalysis = async () => {
  if (!analysisResult.value) return
  try {
    // 创建新书并导入分析结果
    const bookName = fileName.value.replace('.txt', '').substring(0, 20)
    const res = await bookApi.create({ name: bookName })
    const bookId = res.data?.id

    if (bookId && analysisResult.value.characters) {
      // 导入人物
      for (const char of analysisResult.value.characters) {
        await bookApi.update(bookId, {
          type: 'character',
          data: {
            name: char.name,
            role: char.role || '配角',
            bio: char.description
          }
        })
      }
    }

    ElMessage.success(`已创建书籍 "${bookName}" 并导入分析结果`)
  } catch (error) {
    ElMessage.error('导入失败')
  }
}

onMounted(() => {
  loadMyBooks()
})
</script>

<style scoped>
.analysis-view {
  max-width: 1400px;
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

.chapter-list {
  margin-top: 15px;
}

.chapter-item {
  display: flex;
  justify-content: space-between;
  padding: 5px 10px;
  border-bottom: 1px solid #eee;
}

.chapter-item:last-child {
  border-bottom: none;
}

.word-count {
  color: #999;
  font-size: 12px;
}

.more-info {
  text-align: center;
  color: #999;
  padding: 10px;
}

.analysis-tools {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.result-section {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  line-height: 1.8;
}

.char-item {
  padding: 10px;
  border-bottom: 1px solid #eee;
}

.char-item:last-child {
  border-bottom: none;
}

.char-item strong {
  margin: 0 10px;
}

.char-item p {
  margin: 5px 0 0 0;
  color: #666;
}

code {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: monospace;
}
</style>