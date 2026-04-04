<template>
  <div class="architect-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 架构师</h2>
    </div>

    <el-row :gutter="20">
      <!-- 左侧：大纲生成 -->
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>大纲生成</span>
            </div>
          </template>
          <el-form :model="outlineParams" label-width="100px">
            <el-form-item label="题材类型">
              <el-select v-model="outlineParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
                <el-option label="科幻" value="科幻" />
                <el-option label="武侠" value="武侠" />
                <el-option label="历史" value="历史" />
              </el-select>
            </el-form-item>
            <el-form-item label="主角设定">
              <el-input v-model="outlineParams.main_char" placeholder="如：少年天才、穿越者" />
            </el-form-item>
            <el-form-item label="故事主题">
              <el-input v-model="outlineParams.theme" placeholder="如：逆袭、复仇、寻宝" />
            </el-form-item>
            <el-form-item label="目标字数">
              <el-input-number v-model="outlineParams.target_words" :min="100000" :max="5000000" :step="100000" />
            </el-form-item>
            <el-form-item label="分卷数量">
              <el-input-number v-model="outlineParams.volumes" :min="1" :max="20" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateOutline" :loading="generating">
            生成大纲
          </el-button>
        </el-card>

        <!-- 分形裂变 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>分形裂变</span>
            </div>
          </template>
          <el-form :model="fissionParams" label-width="100px">
            <el-form-item label="裂变策略">
              <el-select v-model="fissionParams.strategy">
                <el-option label="展开 - 详细展开" value="expand" />
                <el-option label="优化 - 逻辑优化" value="refine" />
                <el-option label="分支 - 剧情分支" value="branch" />
              </el-select>
            </el-form-item>
            <el-form-item label="当前大纲">
              <el-input v-model="fissionParams.outline" type="textarea" :rows="4" placeholder="输入要裂变的大纲内容" />
            </el-form-item>
            <el-form-item label="生成数量">
              <el-input-number v-model="fissionParams.count" :min="1" :max="10" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="runFission" :loading="fissioning">
            执行裂变
          </el-button>
        </el-card>
      </el-col>

      <!-- 右侧：结果展示 -->
      <el-col :span="12">
        <el-card class="result-card">
          <template #header>
            <div class="card-header">
              <span>生成结果</span>
              <el-button size="small" @click="copyResult">复制</el-button>
            </div>
          </template>
          <div class="result-content">
            <pre>{{ resultContent }}</pre>
          </div>
        </el-card>

        <!-- 裂变策略说明 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <span>裂变策略说明</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="展开策略">
              将简单的大纲展开为更详细的章节内容，适合初期规划
            </el-descriptions-item>
            <el-descriptions-item label="优化策略">
              优化现有大纲的逻辑结构，修复剧情漏洞
            </el-descriptions-item>
            <el-descriptions-item label="分支策略">
              生成多条可能的剧情发展方向，探索不同可能性
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const generating = ref(false)
const fissioning = ref(false)
const resultContent = ref('生成的结果将显示在这里...')

const outlineParams = ref({
  genre: '玄幻',
  main_char: '',
  theme: '',
  target_words: 1000000,
  volumes: 3
})

const fissionParams = ref({
  strategy: 'expand',
  outline: '',
  count: 5
})

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const generateOutline = async () => {
  generating.value = true
  try {
    const res = await axios.post('/api/architect/generate', outlineParams.value)
    resultContent.value = res.data.message + '\n\n命令: ' + res.data.command
    ElMessage.info('请使用 CLI 命令生成大纲')
  } catch (error) {
    ElMessage.error('请求失败')
  }
  generating.value = false
}

const runFission = async () => {
  if (!fissionParams.value.outline) {
    ElMessage.warning('请输入要裂变的大纲内容')
    return
  }
  fissioning.value = true
  try {
    const res = await axios.post('/api/architect/fission', fissionParams.value)
    resultContent.value = res.data.message + '\n\n命令: ' + res.data.command
    ElMessage.info('请使用 CLI 命令执行裂变')
  } catch (error) {
    ElMessage.error('请求失败')
  }
  fissioning.value = false
}

const copyResult = () => {
  navigator.clipboard.writeText(resultContent.value)
  ElMessage.success('已复制到剪贴板')
}
</script>

<style scoped>
.architect-view {
  max-width: 1400px;
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

.result-card {
  min-height: 400px;
}

.result-content {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  min-height: 300px;
}

.result-content pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
}
</style>