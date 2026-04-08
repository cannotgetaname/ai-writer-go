<template>
  <div class="system-view">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 模型配置 -->
      <el-tab-pane label="模型配置" name="config">
        <el-card>
          <el-form :model="config" label-width="140px">
            <el-divider content-position="left">基本设置</el-divider>
            <el-form-item label="提供商">
              <el-select v-model="config.provider" style="width: 200px;">
                <el-option label="DeepSeek" value="deepseek" />
                <el-option label="OpenAI" value="openai" />
                <el-option label="Ollama (本地)" value="ollama" />
              </el-select>
            </el-form-item>
            <el-form-item label="API Key">
              <el-input v-model="config.api_key" type="password" show-password placeholder="输入新的 API Key" style="width: 400px;">
                <template #prepend>
                  <el-tag v-if="config.api_key_set" type="success">已设置 ({{ config.api_key_display }})</el-tag>
                  <el-tag v-else type="danger">未设置</el-tag>
                </template>
              </el-input>
              <div class="el-form-item__tip" style="color: #909399; font-size: 12px;">
                留空表示不修改现有 API Key
              </div>
            </el-form-item>
            <el-form-item label="API Base URL">
              <el-input v-model="config.base_url" placeholder="可选，自定义 API 地址" style="width: 400px;" />
            </el-form-item>
            <el-form-item label="请求超时">
              <el-input-number v-model="config.timeout" :min="30" :max="600" /> 秒
            </el-form-item>
            <el-form-item label="最大重试次数">
              <el-input-number v-model="config.max_retries" :min="0" :max="10" />
            </el-form-item>

            <el-divider content-position="left">模型映射</el-divider>
            <el-table :data="modelList" border style="margin-bottom: 20px;">
              <el-table-column prop="name" label="任务类型" width="120">
                <template #default="{ row }">
                  {{ taskNames[row.key] || row.key }}
                </template>
              </el-table-column>
              <el-table-column prop="model" label="模型">
                <template #default="{ row }">
                  <el-input v-model="config.models[row.key]" size="small" style="width: 180px;" />
                </template>
              </el-table-column>
              <el-table-column prop="temperature" label="温度">
                <template #default="{ row }">
                  <el-input-number v-model="config.temperatures[row.key]" :min="0" :max="2" :step="0.1" size="small" style="width: 120px;" />
                </template>
              </el-table-column>
            </el-table>

            <el-divider content-position="left">向量生成 (Embedding)</el-divider>
            <el-form-item label="提供商">
              <el-select v-model="config.embedding.provider" style="width: 200px;">
                <el-option label="TEI (内置)" value="tei" />
                <el-option label="Ollama (本地)" value="ollama" />
                <el-option label="OpenAI" value="openai" />
                <el-option label="DeepSeek" value="deepseek" />
                <el-option label="自定义 API" value="custom" />
              </el-select>
            </el-form-item>
            <el-form-item label="模型">
              <el-input
                v-model="config.embedding.model"
                :disabled="config.embedding.provider === 'tei'"
                placeholder="bge-base-zh-v1.5"
                style="width: 300px;"
              />
            </el-form-item>
            <el-form-item v-if="config.embedding.provider !== 'tei'" label="API 地址">
              <el-input v-model="config.embedding.base_url" style="width: 400px;" />
            </el-form-item>
            <el-form-item v-if="['openai', 'deepseek', 'custom'].includes(config.embedding.provider)" label="API Key">
              <el-input v-model="config.embedding.api_key" type="password" show-password style="width: 400px;" />
            </el-form-item>

            <el-divider content-position="left">向量存储</el-divider>
            <el-form-item label="分块大小">
              <el-input-number v-model="config.vector_store.chunk_size" :min="100" :max="2000" :step="100" />
            </el-form-item>
            <el-form-item label="分块重叠">
              <el-input-number v-model="config.vector_store.overlap" :min="0" :max="500" :step="50" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="saveConfig">保存配置</el-button>
        </el-card>
      </el-tab-pane>

      <!-- 提示词配置 -->
      <el-tab-pane label="提示词配置" name="prompts">
        <el-card>
          <el-collapse v-model="activePrompt">
            <el-collapse-item title="写作提示词 (writer)" name="writer">
              <el-input v-model="prompts.writer_system" type="textarea" :rows="8" />
            </el-collapse-item>
            <el-collapse-item title="架构师提示词 (architect)" name="architect">
              <el-input v-model="prompts.architect_system" type="textarea" :rows="8" />
            </el-collapse-item>
            <el-collapse-item title="审稿提示词 (reviewer)" name="reviewer">
              <el-input v-model="prompts.reviewer_system" type="textarea" :rows="8" />
            </el-collapse-item>
            <el-collapse-item title="审计提示词 (auditor)" name="auditor">
              <el-input v-model="prompts.auditor_system" type="textarea" :rows="8" />
            </el-collapse-item>
            <el-collapse-item title="时间记录提示词 (timekeeper)" name="timekeeper">
              <el-input v-model="prompts.timekeeper_system" type="textarea" :rows="6" />
            </el-collapse-item>
            <el-collapse-item title="章节摘要提示词" name="summary_chapter">
              <el-input v-model="prompts.summary_chapter_system" type="textarea" :rows="6" />
            </el-collapse-item>
            <el-collapse-item title="全书摘要提示词" name="summary_book">
              <el-input v-model="prompts.summary_book_system" type="textarea" :rows="6" />
            </el-collapse-item>
            <el-collapse-item title="知识过滤提示词" name="knowledge_filter">
              <el-input v-model="prompts.knowledge_filter_system" type="textarea" :rows="6" />
            </el-collapse-item>
            <el-collapse-item title="灵感助手提示词" name="inspiration">
              <el-input v-model="prompts.inspiration_assistant_system" type="textarea" :rows="4" />
            </el-collapse-item>
          </el-collapse>
          <el-button type="primary" @click="savePrompts" style="margin-top: 20px;">保存提示词</el-button>
        </el-card>
      </el-tab-pane>

      <!-- 费用统计 -->
      <el-tab-pane label="费用统计" name="billing">
        <el-card>
          <el-descriptions title="费用概览" :column="3" border>
            <el-descriptions-item label="总Token数">{{ billing.total_tokens?.toLocaleString() || 0 }}</el-descriptions-item>
            <el-descriptions-item label="预估费用">${{ billing.total_cost?.toFixed(4) || '0.0000' }}</el-descriptions-item>
            <el-descriptions-item label="本月Token">{{ billing.monthly_tokens?.toLocaleString() || 0 }}</el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">模型定价</el-divider>
          <el-table :data="pricingList" border style="margin-top: 20px;">
            <el-table-column prop="model" label="模型" width="180" />
            <el-table-column prop="input" label="输入价格 ($/1K tokens)">
              <template #default="{ row }">
                ${{ row.input.toFixed(4) }}
              </template>
            </el-table-column>
            <el-table-column prop="output" label="输出价格 ($/1K tokens)">
              <template #default="{ row }">
                ${{ row.output.toFixed(4) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 写作目标 -->
      <el-tab-pane label="写作目标" name="goals">
        <el-card>
          <el-form :model="goals" label-width="140px">
            <el-form-item label="每日目标字数">
              <el-input-number v-model="goals.daily_words" :min="0" :max="50000" :step="500" />
            </el-form-item>
            <el-form-item label="每周目标章节">
              <el-input-number v-model="goals.weekly_chapters" :min="0" :max="50" />
            </el-form-item>
            <el-form-item label="目标完成日期">
              <el-date-picker v-model="goals.target_date" type="date" placeholder="选择日期" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="saveGoals">保存目标</el-button>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { systemApi } from '@/api'

const activeTab = ref('config')
const activePrompt = ref(['writer'])

const taskNames = {
  writer: '写作',
  architect: '架构师',
  editor: '编辑',
  reviewer: '审稿',
  auditor: '审计',
  timekeeper: '时间记录',
  summary: '摘要'
}

const taskKeys = ['writer', 'architect', 'editor', 'reviewer', 'auditor', 'timekeeper', 'summary']

const config = ref({
  provider: 'deepseek',
  api_key: '',
  base_url: 'https://api.deepseek.com',
  timeout: 120,
  max_retries: 3,
  models: {
    writer: 'deepseek-chat',
    architect: 'deepseek-reasoner',
    editor: 'deepseek-chat',
    reviewer: 'deepseek-chat',
    auditor: 'deepseek-reasoner',
    timekeeper: 'deepseek-chat',
    summary: 'deepseek-chat'
  },
  temperatures: {
    writer: 1.5,
    architect: 1.0,
    editor: 0.7,
    reviewer: 0.5,
    auditor: 0.6,
    timekeeper: 0.1,
    summary: 0.5
  },
  vector_store: {
    chunk_size: 500,
    overlap: 100
  },
  embedding: {
    provider: 'tei',
    model: 'bge-base-zh-v1.5',
    base_url: 'http://127.0.0.1:8081',
    api_key: ''
  }
})

const modelList = computed(() => {
  return taskKeys.map(key => ({
    key,
    name: taskNames[key] || key,
    model: config.value.models[key] || '',
    temperature: config.value.temperatures[key] || 1.0
  }))
})

const billing = ref({
  total_tokens: 0,
  total_cost: 0,
  monthly_tokens: 0,
  monthly_cost: 0,
  pricing: {}
})

const pricingList = computed(() => {
  const pricing = billing.value.pricing || {}
  return Object.entries(pricing).map(([model, p]) => ({
    model,
    input: p.input || 0,
    output: p.output || 0
  }))
})

const goals = ref({
  daily_words: 2000,
  weekly_chapters: 2,
  target_date: null
})

const prompts = ref({
  writer_system: '',
  architect_system: '',
  reviewer_system: '',
  auditor_system: '',
  timekeeper_system: '',
  summary_system: '',
  summary_chapter_system: '',
  summary_book_system: '',
  knowledge_filter_system: '',
  json_only_architect_system: '',
  inspiration_assistant_system: ''
})

const loadConfig = async () => {
  try {
    const res = await systemApi.getConfig()
    const data = res.data || {}
    config.value = {
      provider: data.provider || 'deepseek',
      api_key: '', // 不从服务器获取实际 key
      api_key_set: data.api_key_set || false,
      api_key_display: data.api_key_display || '',
      base_url: data.base_url || 'https://api.deepseek.com',
      timeout: data.timeout || 120,
      max_retries: data.max_retries || 3,
      models: data.models || config.value.models,
      temperatures: data.temperatures || config.value.temperatures,
      vector_store: data.vector_store || config.value.vector_store,
      embedding: data.embedding || config.value.embedding
    }
  } catch (error) {
    console.error('加载配置失败:', error)
  }
}

const saveConfig = async () => {
  try {
    await systemApi.updateConfig(config.value)
    ElMessage.success('配置保存成功')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

const loadBilling = async () => {
  try {
    const res = await systemApi.getBilling()
    billing.value = res.data || billing.value
  } catch (error) {
    console.error('加载费用统计失败:', error)
  }
}

const loadGoals = async () => {
  try {
    const res = await systemApi.getGoals()
    goals.value = res.data || goals.value
  } catch (error) {
    console.error('加载写作目标失败:', error)
  }
}

const saveGoals = async () => {
  try {
    await systemApi.updateGoals(goals.value)
    ElMessage.success('目标保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const loadPrompts = async () => {
  try {
    const res = await systemApi.getPrompts()
    prompts.value = res.data || prompts.value
  } catch (error) {
    console.error('加载提示词失败:', error)
  }
}

const savePrompts = async () => {
  try {
    await systemApi.updatePrompts(prompts.value)
    ElMessage.success('提示词保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

onMounted(() => {
  loadConfig()
  loadBilling()
  loadGoals()
  loadPrompts()
})
</script>

<style scoped>
.system-view {
  max-width: 1000px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

:deep(.el-collapse-item__header) {
  font-weight: 500;
}

:deep(.el-textarea__inner) {
  font-family: monospace;
  font-size: 13px;
}
</style>