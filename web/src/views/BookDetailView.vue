<template>
  <div class="book-detail-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ book.name || bookId }}
        <el-button size="small" @click="showEditDialog">编辑</el-button>
      </h2>
      <div class="header-actions">
        <el-button @click="goToWrite">
          <el-icon><Edit /></el-icon>
          写作
        </el-button>
        <el-button @click="goToBatch" type="success">
          <el-icon><VideoPlay /></el-icon>
          批量生成
        </el-button>
        <el-dropdown trigger="click">
          <el-button>
            <el-icon><Download /></el-icon>
            导出
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="exportBook('txt')">TXT 文本</el-dropdown-item>
              <el-dropdown-item @click="exportBook('markdown')">Markdown</el-dropdown-item>
              <el-dropdown-item @click="exportBook('json')">JSON</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button @click="goToSettings">
          <el-icon><Tools /></el-icon>
          设定
        </el-button>
        <el-button @click="goToTimeline">
          <el-icon><Clock /></el-icon>
          时间线
        </el-button>
        <el-button @click="goToGraph">
          <el-icon><Share /></el-icon>
          图谱
        </el-button>
        <el-button @click="goToSync">
          <el-icon><Refresh /></el-icon>
          状态同步
        </el-button>
        <el-button @click="indexBook" :loading="indexing">
          <el-icon><Collection /></el-icon>
          索引全书
        </el-button>
        <el-button @click="goToAnalysis">
          <el-icon><DataAnalysis /></el-icon>
          分析
        </el-button>
        <el-button type="primary" @click="goToArchitect">
          <el-icon><Grid /></el-icon>
          架构师
        </el-button>
      </div>
    </div>

    <el-row :gutter="20">
      <!-- 基本信息 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>书籍概览</span>
          </template>
          <el-descriptions :column="3" border>
            <el-descriptions-item label="章节数">{{ book.chapters?.length || 0 }}</el-descriptions-item>
            <el-descriptions-item label="人物数">{{ book.characters?.length || 0 }}</el-descriptions-item>
            <el-descriptions-item label="物品数">{{ book.items?.length || 0 }}</el-descriptions-item>
            <el-descriptions-item label="地点数">{{ book.locations?.length || 0 }}</el-descriptions-item>
            <el-descriptions-item label="分卷数">{{ book.volumes?.length || 0 }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 章节列表 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>章节列表</span>
              <el-button size="small" type="primary" @click="goToWrite">
                <el-icon><Plus /></el-icon>
                新增章节
              </el-button>
            </div>
          </template>
          <el-table :data="book.chapters" stripe max-height="400">
            <el-table-column prop="id" label="章节" width="80" />
            <el-table-column prop="title" label="标题" />
            <el-table-column prop="outline" label="大纲" show-overflow-tooltip />
            <el-table-column prop="word_count" label="字数" width="100" />
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="openChapter(row.id)">查看</el-button>
                <el-button size="small" type="danger" @click="deleteChapter(row.id)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="8">
        <!-- 人物卡片 -->
        <el-card>
          <template #header>
            <div class="card-header">
              <span>主要人物</span>
              <el-button size="small" @click="goToSettings">管理</el-button>
            </div>
          </template>
          <div v-for="char in topCharacters" :key="char.id" class="character-item">
            <el-avatar :size="40">{{ char.name?.charAt(0) }}</el-avatar>
            <div class="char-info">
              <strong>{{ char.name }}</strong>
              <span class="char-role">{{ char.role }}</span>
            </div>
          </div>
        </el-card>

        <!-- 世界观 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>世界观</span>
              <el-button size="small" @click="goToSettings">编辑</el-button>
            </div>
          </template>
          <div v-if="book.worldview" class="worldview-summary">
            <p><strong>题材:</strong> {{ book.worldview.basic_info?.genre }}</p>
            <p><strong>力量体系:</strong> {{ book.worldview.core_settings?.power_system }}</p>
          </div>
          <el-empty v-else description="暂无世界观设定" :image-size="60" />
        </el-card>

        <!-- 向量搜索 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <span>相似内容搜索</span>
          </template>
          <el-input
            v-model="searchQuery"
            placeholder="输入关键词搜索相似内容..."
            @keyup.enter="searchSimilar"
            @input="onSearchInput"
            style="margin-bottom: 10px;"
          >
            <template #append>
              <el-button @click="searchSimilar" :loading="searching">
                <el-icon><Search /></el-icon>
              </el-button>
            </template>
          </el-input>
          <div v-if="searchResults.length > 0" class="search-results">
            <div v-for="(result, idx) in searchResults" :key="idx" class="search-result-item">
              <div class="result-header">
                <el-tag size="small">第{{ result.chapter_id }}章</el-tag>
                <span class="result-score">相似度: {{ ((result.score || 0) * 100).toFixed(1) }}%</span>
              </div>
              <div class="result-content">{{ result.content }}</div>
            </div>
          </div>
          <el-empty v-else-if="searchQuery && !searching && hasSearched" description="无搜索结果" :image-size="40" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑书籍" width="400px">
      <el-form :model="editName" label-width="80px">
        <el-form-item label="书籍名称">
          <el-input v-model="editName" placeholder="请输入新书名" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="renameBook">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { bookApi, chapterApi, exportApi, vectorApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const book = ref({})
const editDialogVisible = ref(false)
const editName = ref('')
const indexing = ref(false)
const searchQuery = ref('')
const searchResults = ref([])
const searching = ref(false)
const hasSearched = ref(false)

const topCharacters = computed(() => {
  const chars = book.value.characters || []
  return chars.filter(c => c.role === '主角' || c.role === '配角').slice(0, 5)
})

const goBack = () => {
  router.push('/books')
}

const showEditDialog = () => {
  editName.value = book.value.name || bookId.value
  editDialogVisible.value = true
}

const renameBook = async () => {
  if (!editName.value) {
    ElMessage.warning('请输入书籍名称')
    return
  }
  try {
    await bookApi.update(bookId.value, { name: editName.value })
    ElMessage.success('书名修改成功')
    editDialogVisible.value = false
    router.push(`/books/${editName.value}`)
  } catch (error) {
    ElMessage.error('修改失败: ' + (error.response?.data?.error || error.message))
  }
}

const goToWrite = () => {
  router.push(`/books/${bookId.value}/write`)
}

const goToBatch = () => {
  router.push(`/books/${bookId.value}/batch`)
}

const goToSettings = () => {
  router.push(`/books/${bookId.value}/settings`)
}

const goToTimeline = () => {
  router.push(`/books/${bookId.value}/timeline`)
}

const goToGraph = () => {
  router.push(`/books/${bookId.value}/graph`)
}

const goToArchitect = () => {
  router.push(`/books/${bookId.value}/architect`)
}

const goToSync = () => {
  router.push(`/books/${bookId.value}/sync`)
}

const goToAnalysis = () => {
  router.push(`/books/${bookId.value}/analysis`)
}

const exportBook = (format) => {
  const url = exportApi[format](bookId.value)
  window.open(url, '_blank')
}

const openChapter = (chapterId) => {
  router.push(`/books/${bookId.value}/write?chapter=${chapterId}`)
}

const deleteChapter = async (chapterId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个章节吗？删除后无法恢复。', '删除确认', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await chapterApi.delete(bookId.value, chapterId)
    ElMessage.success('章节删除成功')
    loadBook() // 刷新列表
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.error || error.message))
    }
  }
}

const indexBook = async () => {
  indexing.value = true
  try {
    const res = await vectorApi.indexBook(bookId.value)
    if (res.data?.error) {
      ElMessage.error(res.data.error)
    } else {
      ElMessage.success(`索引完成，共 ${res.data?.chunks || 0} 个分块`)
    }
  } catch (error) {
    ElMessage.error('索引失败: ' + (error.response?.data?.error || error.message))
  }
  indexing.value = false
}

const searchSimilar = async () => {
  if (!searchQuery.value.trim()) return
  searching.value = true
  hasSearched.value = false
  try {
    const res = await vectorApi.search(bookId.value, searchQuery.value, 5)
    if (res.data?.results) {
      searchResults.value = res.data.results
    } else if (res.data?.error) {
      ElMessage.error(res.data.error)
    }
  } catch (error) {
    ElMessage.error('搜索失败: ' + (error.response?.data?.error || error.message))
  }
  searching.value = false
  hasSearched.value = true
}

const onSearchInput = () => {
  // 输入时清除之前的搜索结果
  searchResults.value = []
  hasSearched.value = false
}

const loadBook = async () => {
  try {
    const res = await bookApi.get(bookId.value)
    book.value = res.data || {}
  } catch (error) {
    console.error('加载书籍失败:', error)
  }
}

onMounted(() => {
  loadBook()
})
</script>

<style scoped>
.book-detail-view {
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
  flex: 1;
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

.character-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
}

.char-info {
  flex: 1;
}

.char-info strong {
  margin-right: 10px;
}

.char-role {
  color: #666;
  font-size: 12px;
}

.worldview-summary p {
  margin: 5px 0;
  color: #666;
}

.search-results {
  max-height: 400px;
  overflow-y: auto;
}

.search-result-item {
  padding: 10px;
  border-bottom: 1px solid #ebeef5;
}

.search-result-item:last-child {
  border-bottom: none;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}

.result-score {
  font-size: 12px;
  color: #909399;
}

.result-content {
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>