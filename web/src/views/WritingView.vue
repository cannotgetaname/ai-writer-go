<template>
  <div class="writing-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - AI 写作</h2>
      <div class="header-actions">
        <el-button @click="showParagraphDialog">
          <el-icon><List /></el-icon>
          段落管理
        </el-button>
        <el-button type="primary" @click="showGenerateDialog">
          <el-icon><MagicStick /></el-icon>
          AI 生成
        </el-button>
        <el-button type="success" @click="showContinueDialog">
          <el-icon><Plus /></el-icon>
          续写
        </el-button>
        <el-button type="warning" @click="review" :loading="reviewing">
          <el-icon><DocumentChecked /></el-icon>
          审稿
        </el-button>
        <el-button type="info" @click="showWorldAuditDialog" :disabled="!currentChapter">
          <el-icon><DataAnalysis /></el-icon>
          审计世界状态
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
            <div class="card-header">
              <span>审稿结果</span>
              <el-button size="small" @click="review" :loading="reviewing">重新审稿</el-button>
            </div>
          </template>
          <div class="review-score">
            综合评分:
            <el-progress :percentage="reviewResult.overall_score" :color="getScoreColor(reviewResult.overall_score)" />
          </div>

          <div v-if="reviewResult.paragraph_issues && reviewResult.paragraph_issues.length > 0" class="review-issues">
            <el-divider content-position="left">段落问题</el-divider>
            <el-collapse>
              <el-collapse-item
                v-for="(issue, idx) in reviewResult.paragraph_issues"
                :key="idx"
                :name="idx"
              >
                <template #title>
                  <div class="issue-title">
                    <el-tag size="small" color="#e6f7ff" border-color="#91d5ff">段落{{ issue.paragraph_index }}</el-tag>
                    <el-tag v-if="issue.related_paragraphs && issue.related_paragraphs.length > 0" size="small" color="#f9f0ff" border-color="#d3adf7">
                      相关: {{ issue.related_paragraphs.join(', ') }}
                    </el-tag>
                    <el-tag :type="getSeverityType(issue.severity)" size="small">{{ issue.severity }}</el-tag>
                    <el-tag size="small" color="#fff7e6" border-color="#ffd591">{{ issue.type }}</el-tag>
                  </div>
                </template>
                <div class="issue-detail">
                  <p class="issue-desc">{{ issue.description }}</p>
                  <p class="issue-suggestion"><strong>建议:</strong> {{ issue.suggestion }}</p>
                  <div class="issue-original" v-if="issue.original_text">
                    <el-text type="info" size="small">原文预览: {{ issue.original_text }}</el-text>
                  </div>
                  <el-button type="primary" size="small" @click="showRewriteDialog(issue)">
                    重绘此段落
                  </el-button>
                </div>
              </el-collapse-item>
            </el-collapse>
          </div>

          <div v-if="reviewResult.suggestions && reviewResult.suggestions.length > 0" class="review-suggestions">
            <el-divider content-position="left">整体建议</el-divider>
            <ul>
              <li v-for="(s, idx) in reviewResult.suggestions" :key="idx">{{ s }}</li>
            </ul>
          </div>
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

    <!-- 段落重绘对话框 -->
    <el-dialog v-model="rewriteDialogVisible" title="段落重绘" width="600px">
      <el-form label-width="80px">
        <el-form-item label="段落序号">
          <el-tag>段落 {{ rewriteParams.paragraph_index }}</el-tag>
        </el-form-item>
        <el-form-item label="原文预览">
          <el-input :model-value="rewriteParams.original_text" type="textarea" :rows="3" readonly />
        </el-form-item>
        <el-form-item label="问题类型">
          <el-tag>{{ rewriteParams.issue_type }}</el-tag>
        </el-form-item>
        <el-form-item label="问题描述">
          <el-text>{{ rewriteParams.issue_description }}</el-text>
        </el-form-item>
        <el-form-item label="重绘指令">
          <el-input v-model="rewriteParams.instruction" type="textarea" :rows="3" placeholder="可编辑修改建议，或输入自定义重绘指令" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rewriteDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="rewriteParagraph" :loading="rewriting">
          开始重绘
        </el-button>
      </template>
    </el-dialog>

    <!-- 重绘结果对话框 -->
    <el-dialog v-model="rewriteResultVisible" title="重绘结果" width="700px">
      <el-row :gutter="20">
        <el-col :span="12">
          <h4>原文</h4>
          <div class="rewrite-content">{{ rewriteResult.original }}</div>
        </el-col>
        <el-col :span="12">
          <h4>新文</h4>
          <div class="rewrite-content">{{ rewriteResult.new }}</div>
        </el-col>
      </el-row>
      <template #footer>
        <el-button @click="rewriteResultVisible = false">放弃</el-button>
        <el-button type="primary" @click="applyRewrite">应用修改</el-button>
      </template>
    </el-dialog>

    <!-- 段落管理对话框 -->
    <el-dialog v-model="paragraphDialogVisible" title="段落管理" width="800px">
      <el-alert type="info" :closable="false" style="margin-bottom: 15px;">
        拖拽段落可调整顺序，点击操作按钮可删除或编辑段落
      </el-alert>
      <el-table :data="paragraphs" stripe>
        <el-table-column type="index" label="序号" width="60" />
        <el-table-column label="内容预览">
          <template #default="{ row }">
            <el-text line-clamp="2">{{ row.text }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="word_count" label="字数" width="80" />
        <el-table-column label="操作" width="180">
          <template #default="{ row, $index }">
            <el-button-group>
              <el-button size="small" :disabled="$index === 0" @click="moveParagraph($index, -1)">
                <el-icon><ArrowUp /></el-icon>
              </el-button>
              <el-button size="small" :disabled="$index === paragraphs.length - 1" @click="moveParagraph($index, 1)">
                <el-icon><ArrowDown /></el-icon>
              </el-button>
              <el-button size="small" type="primary" @click="editParagraph(row, $index)">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button size="small" type="danger" @click="deleteParagraph($index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="paragraphDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="saveParagraphs">保存修改</el-button>
      </template>
    </el-dialog>

    <!-- 编辑段落对话框 -->
    <el-dialog v-model="editParagraphVisible" title="编辑段落" width="600px">
      <el-input v-model="editParagraphContent" type="textarea" :rows="8" />
      <template #footer>
        <el-button @click="editParagraphVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEditParagraph">保存</el-button>
      </template>
    </el-dialog>

    <!-- 世界状态审计对话框 -->
    <WorldAuditDialog
      v-model="worldAuditDialogVisible"
      :book-id="bookId"
      :chapter-id="currentChapter?.id || 0"
      @applied="loadForeshadows"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { chapterApi, aiApi, foreshadowApi } from '@/api'
import WorldAuditDialog from '@/components/WorldAuditDialog.vue'

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

// 段落重绘相关
const rewriteDialogVisible = ref(false)
const rewriting = ref(false)
const rewriteParams = ref({
  paragraph_id: '',
  paragraph_index: 0,
  original_text: '',
  issue_type: '',
  issue_description: '',
  instruction: ''
})
const rewriteResultVisible = ref(false)
const rewriteResult = ref({
  original: '',
  new: '',
  paragraph_id: ''
})

// 段落管理
const paragraphDialogVisible = ref(false)
const paragraphs = ref([])
const editParagraphVisible = ref(false)
const editParagraphContent = ref('')
const editParagraphIndex = ref(-1)

// 世界状态审计对话框
const worldAuditDialogVisible = ref(false)

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
    // 加载已保存的审稿结果
    try {
      const reviewRes = await aiApi.getReview(bookId.value, ch.id)
      if (reviewRes.data?.paragraph_issues) {
        reviewResult.value = reviewRes.data
      } else {
        reviewResult.value = null
      }
    } catch (error) {
      reviewResult.value = null
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
  if (!currentChapter.value) {
    ElMessage.warning('请先选择章节')
    return
  }
  reviewing.value = true
  try {
    const res = await aiApi.reviewByParagraph({
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

const getScoreColor = (score) => {
  if (score >= 80) return '#67c23a'
  if (score >= 60) return '#e6a23c'
  return '#f56c6c'
}

const getSeverityType = (severity) => {
  if (severity === '严重' || severity === '高') return 'danger'
  if (severity === '中等' || severity === '中') return 'warning'
  if (severity === '轻微') return 'success'
  return 'info'
}

const showRewriteDialog = (issue) => {
  rewriteParams.value = {
    paragraph_id: issue.paragraph_id,
    paragraph_index: issue.paragraph_index,
    original_text: issue.original_text || '',
    issue_type: issue.type,
    issue_description: issue.description,
    instruction: issue.suggestion || ''
  }
  rewriteDialogVisible.value = true
}

const rewriteParagraph = async () => {
  if (!rewriteParams.value.instruction) {
    ElMessage.warning('请输入重绘指令')
    return
  }
  rewriting.value = true
  try {
    const res = await aiApi.rewriteParagraph({
      book_name: bookId.value,
      chapter_id: currentChapter.value.id,
      paragraph_id: rewriteParams.value.paragraph_id,
      instruction: rewriteParams.value.instruction
    })
    if (res.data?.content) {
      rewriteResult.value = {
        original: rewriteParams.value.original_text,
        new: res.data.content,
        paragraph_id: rewriteParams.value.paragraph_id
      }
      rewriteDialogVisible.value = false
      rewriteResultVisible.value = true
    } else {
      ElMessage.info(res.data?.message || '重绘完成')
    }
  } catch (error) {
    ElMessage.error('重绘失败: ' + (error.response?.data?.error || error.message))
  }
  rewriting.value = false
}

const applyRewrite = async () => {
  try {
    // 更新单个段落
    await chapterApi.updateParagraph(
      bookId.value,
      currentChapter.value.id,
      rewriteResult.value.paragraph_id,
      rewriteResult.value.new
    )
    ElMessage.success('已应用修改')
    rewriteResultVisible.value = false
    // 重新加载内容
    selectChapter(currentChapter.value.id.toString())
  } catch (error) {
    ElMessage.error('应用失败: ' + (error.response?.data?.error || error.message))
  }
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

// 段落管理功能
const showParagraphDialog = () => {
  parseParagraphs()
  paragraphDialogVisible.value = true
}

const parseParagraphs = () => {
  // 将内容按空行分割成段落
  const parts = content.value.split(/\n\n+/).filter(p => p.trim())
  paragraphs.value = parts.map((text, index) => ({
    id: `para_${index}`,
    text: text.trim(),
    word_count: (text.match(/[\u4e00-\u9fa5]/g) || []).length
  }))
}

const moveParagraph = (index, direction) => {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= paragraphs.value.length) return
  const temp = paragraphs.value[index]
  paragraphs.value[index] = paragraphs.value[newIndex]
  paragraphs.value[newIndex] = temp
}

const editParagraph = (row, index) => {
  editParagraphContent.value = row.text
  editParagraphIndex.value = index
  editParagraphVisible.value = true
}

const saveEditParagraph = () => {
  if (editParagraphIndex.value >= 0) {
    paragraphs.value[editParagraphIndex.value].text = editParagraphContent.value
    paragraphs.value[editParagraphIndex.value].word_count = (editParagraphContent.value.match(/[\u4e00-\u9fa5]/g) || []).length
  }
  editParagraphVisible.value = false
}

const deleteParagraph = (index) => {
  paragraphs.value.splice(index, 1)
}

const saveParagraphs = async () => {
  // 将段落合并回内容
  content.value = paragraphs.value.map(p => p.text).join('\n\n')
  updateWordCount()
  await saveContent()
  paragraphDialogVisible.value = false
  ElMessage.success('段落已更新')
}

// 世界状态审计
const showWorldAuditDialog = () => {
  if (!currentChapter.value) {
    ElMessage.warning('请先选择章节')
    return
  }
  worldAuditDialogVisible.value = true
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
  margin-bottom: 15px;
}

.review-issues {
  margin-top: 10px;
}

.issue-title {
  display: flex;
  gap: 8px;
  align-items: center;
}

.issue-detail {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.issue-desc {
  margin: 0 0 10px 0;
  color: #303133;
}

.issue-suggestion {
  margin: 0 0 10px 0;
  color: #606266;
}

.issue-original {
  margin-bottom: 10px;
}

.review-suggestions {
  margin-top: 10px;
}

.review-suggestions ul {
  margin: 0;
  padding-left: 20px;
}

.review-suggestions li {
  margin: 5px 0;
  color: #606266;
}

.rewrite-content {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  max-height: 300px;
  overflow-y: auto;
  white-space: pre-wrap;
  font-size: 14px;
  line-height: 1.6;
}
</style>