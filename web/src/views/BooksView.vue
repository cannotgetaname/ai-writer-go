<template>
  <div class="books-view">
    <div class="page-header">
      <h2>书籍管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        新建书籍
      </el-button>
    </div>

    <el-row :gutter="20">
      <el-col :span="8" v-for="book in books" :key="book.id">
        <el-card class="book-card" @click="goToBook(book.id)">
          <div class="book-info">
            <h3>{{ book.name || book.id }}</h3>
            <p class="book-meta">
              <el-icon><Calendar /></el-icon>
              {{ formatDate(book.created_at) }}
            </p>
            <div class="book-stats">
              <el-tag type="info" size="small">
                {{ book.chapters?.length || 0 }} 章节
              </el-tag>
              <el-tag type="success" size="small">
                {{ book.characters?.length || 0 }} 人物
              </el-tag>
            </div>
          </div>
          <div class="book-actions">
            <el-button-group>
              <el-button size="small" @click.stop="goToWrite(book.id)">
                <el-icon><Edit /></el-icon>
                写作
              </el-button>
              <el-button size="small" @click.stop="goToSettings(book.id)">
                <el-icon><Tools /></el-icon>
                设定
              </el-button>
              <el-button size="small" type="danger" @click.stop="deleteBook(book.id)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-button-group>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="books.length === 0" description="暂无书籍，点击上方按钮创建" />

    <!-- 创建对话框 -->
    <el-dialog v-model="createDialogVisible" title="新建书籍" width="400px">
      <el-form :model="newBook" label-width="80px">
        <el-form-item label="书籍名称">
          <el-input v-model="newBook.name" placeholder="请输入书籍名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createBook">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { bookApi } from '@/api'

const router = useRouter()
const books = ref([])
const createDialogVisible = ref(false)
const newBook = ref({ name: '' })

const formatDate = (date) => {
  if (!date) return '未知'
  return new Date(date).toLocaleDateString('zh-CN')
}

const loadBooks = async () => {
  try {
    const res = await bookApi.list()
    books.value = res.data || []
  } catch (error) {
    console.error('加载书籍列表失败:', error)
    ElMessage.error('加载书籍列表失败')
  }
}

const showCreateDialog = () => {
  newBook.value = { name: '' }
  createDialogVisible.value = true
}

const createBook = async () => {
  if (!newBook.value.name) {
    ElMessage.warning('请输入书籍名称')
    return
  }
  try {
    await bookApi.create({ name: newBook.value.name })
    ElMessage.success('书籍创建成功')
    createDialogVisible.value = false
    loadBooks()
  } catch (error) {
    ElMessage.error('创建失败: ' + error.message)
  }
}

const deleteBook = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此书籍？此操作不可恢复', '删除确认', {
      type: 'warning'
    })
    await bookApi.delete(id)
    ElMessage.success('删除成功')
    loadBooks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const goToBook = (id) => {
  router.push(`/books/${id}`)
}

const goToWrite = (id) => {
  router.push(`/books/${id}/write`)
}

const goToSettings = (id) => {
  router.push(`/books/${id}/settings`)
}

onMounted(() => {
  loadBooks()
})
</script>

<style scoped>
.books-view {
  max-width: 1200px;
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

.book-card {
  cursor: pointer;
  transition: all 0.3s;
}

.book-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.book-info h3 {
  margin: 0 0 10px 0;
  font-size: 18px;
}

.book-meta {
  color: #666;
  font-size: 12px;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.book-stats {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.book-actions {
  text-align: center;
}
</style>