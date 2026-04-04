<template>
  <div class="toolbox-view">
    <div class="page-header">
      <h2>智能工具箱</h2>
    </div>

    <el-row :gutter="20">
      <!-- 命名工具 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><User /></el-icon>
              命名生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="类型">
              <el-select v-model="namingParams.type">
                <el-option label="人名" value="人名" />
                <el-option label="功法" value="功法" />
                <el-option label="法宝" value="法宝" />
                <el-option label="宗门" value="宗门" />
                <el-option label="地点" value="地点" />
              </el-select>
            </el-form-item>
            <el-form-item label="题材">
              <el-select v-model="namingParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
                <el-option label="科幻" value="科幻" />
              </el-select>
            </el-form-item>
            <el-form-item label="数量">
              <el-input-number v-model="namingParams.count" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="性别" v-if="namingParams.type === '人名'">
              <el-select v-model="namingParams.gender">
                <el-option label="男" value="男" />
                <el-option label="女" value="女" />
              </el-select>
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateNames" :loading="loading">
            生成名称
          </el-button>
          <div v-if="namingResults.length > 0" class="results">
            <div v-for="(item, i) in namingResults" :key="i" class="result-item">
              <strong>{{ item.name }}</strong>
              <span class="meaning">{{ item.meaning }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 角色生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Avatar /></el-icon>
              角色生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="角色类型">
              <el-select v-model="characterParams.type">
                <el-option label="主角" value="主角" />
                <el-option label="配角" value="配角" />
                <el-option label="反派" value="反派" />
              </el-select>
            </el-form-item>
            <el-form-item label="性别">
              <el-select v-model="characterParams.gender">
                <el-option label="男" value="男" />
                <el-option label="女" value="女" />
              </el-select>
            </el-form-item>
            <el-form-item label="题材">
              <el-select v-model="characterParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
              </el-select>
            </el-form-item>
            <el-form-item label="主题特点">
              <el-input v-model="characterParams.theme" placeholder="如: 冷血无情" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateCharacter" :loading="loading">
            生成角色
          </el-button>
          <div v-if="characterResult" class="results">
            <p><strong>姓名:</strong> {{ characterResult.name }}</p>
            <p><strong>性格:</strong> {{ characterResult.personality }}</p>
            <p><strong>目标:</strong> {{ characterResult.goal }}</p>
            <p><strong>背景:</strong> {{ characterResult.background }}</p>
          </div>
        </el-card>
      </el-col>

      <!-- 冲突生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Warning /></el-icon>
              冲突生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="冲突类型">
              <el-select v-model="conflictParams.type">
                <el-option label="人物" value="人物" />
                <el-option label="利益" value="利益" />
                <el-option label="情感" value="情感" />
                <el-option label="理念" value="理念" />
              </el-select>
            </el-form-item>
            <el-form-item label="题材">
              <el-select v-model="conflictParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
              </el-select>
            </el-form-item>
            <el-form-item label="背景上下文">
              <el-input v-model="conflictParams.context" type="textarea" :rows="2" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateConflict" :loading="loading">
            生成冲突
          </el-button>
          <div v-if="conflictResult" class="results">
            <p><strong>标题:</strong> {{ conflictResult.title }}</p>
            <p><strong>描述:</strong> {{ conflictResult.description }}</p>
            <p><strong>利害关系:</strong> {{ conflictResult.stakes }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 场景生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Picture /></el-icon>
              场景生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="场景类型">
              <el-select v-model="sceneParams.type">
                <el-option label="战斗" value="战斗" />
                <el-option label="日常" value="日常" />
                <el-option label="对话" value="对话" />
                <el-option label="冒险" value="冒险" />
              </el-select>
            </el-form-item>
            <el-form-item label="地点">
              <el-input v-model="sceneParams.location" />
            </el-form-item>
            <el-form-item label="氛围">
              <el-input v-model="sceneParams.mood" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateScene" :loading="loading">
            生成场景
          </el-button>
          <div v-if="sceneResult" class="results">
            <p><strong>环境:</strong> {{ sceneResult.setting }}</p>
            <p><strong>氛围:</strong> {{ sceneResult.atmosphere }}</p>
            <p>{{ sceneResult.description }}</p>
          </div>
        </el-card>
      </el-col>

      <!-- 金手指生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Star /></el-icon>
              金手指生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="类型">
              <el-select v-model="goldfingerParams.type">
                <el-option label="系统" value="系统" />
                <el-option label="天赋" value="天赋" />
                <el-option label="宝物" value="宝物" />
                <el-option label="传承" value="传承" />
              </el-select>
            </el-form-item>
            <el-form-item label="题材">
              <el-select v-model="goldfingerParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
              </el-select>
            </el-form-item>
            <el-form-item label="强度">
              <el-select v-model="goldfingerParams.level">
                <el-option label="弱" value="弱" />
                <el-option label="中等" value="中等" />
                <el-option label="强" value="强" />
              </el-select>
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateGoldfinger" :loading="loading">
            生成金手指
          </el-button>
          <div v-if="goldfingerResult" class="results">
            <p><strong>名称:</strong> {{ goldfingerResult.name }}</p>
            <p><strong>描述:</strong> {{ goldfingerResult.description }}</p>
            <p><strong>限制:</strong> {{ goldfingerResult.limitations }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <!-- 书名生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Notebook /></el-icon>
              书名生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="题材">
              <el-select v-model="titleParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
                <el-option label="科幻" value="科幻" />
              </el-select>
            </el-form-item>
            <el-form-item label="主题">
              <el-input v-model="titleParams.theme" />
            </el-form-item>
            <el-form-item label="数量">
              <el-input-number v-model="titleParams.count" :min="1" :max="10" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateTitle" :loading="loading">
            生成书名
          </el-button>
          <div v-if="titleResults.length > 0" class="results">
            <div v-for="(t, i) in titleResults" :key="i" class="result-item">
              <strong>{{ t.title }}</strong>
              <p class="meaning">{{ t.meaning }}</p>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 简介生成 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Document /></el-icon>
              简介生成器
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="题材">
              <el-select v-model="synopsisParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
              </el-select>
            </el-form-item>
            <el-form-item label="主角">
              <el-input v-model="synopsisParams.main_char" />
            </el-form-item>
            <el-form-item label="类型">
              <el-select v-model="synopsisParams.type">
                <el-option label="简短" value="short" />
                <el-option label="详细" value="long" />
              </el-select>
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateSynopsis" :loading="loading">
            生成简介
          </el-button>
          <div v-if="synopsisResult" class="results">
            <p>{{ synopsisResult.synopsis }}</p>
          </div>
        </el-card>
      </el-col>

      <!-- 剧情转折 -->
      <el-col :span="8">
        <el-card class="tool-card">
          <template #header>
            <div class="card-header">
              <el-icon><Switch /></el-icon>
              剧情转折
            </div>
          </template>
          <el-form label-width="80px">
            <el-form-item label="类型">
              <el-select v-model="twistParams.type">
                <el-option label="意外转折" value="unexpected" />
                <el-option label="剧情反转" value="reversal" />
              </el-select>
            </el-form-item>
            <el-form-item label="题材">
              <el-select v-model="twistParams.genre">
                <el-option label="玄幻" value="玄幻" />
                <el-option label="仙侠" value="仙侠" />
                <el-option label="都市" value="都市" />
              </el-select>
            </el-form-item>
            <el-form-item label="上下文">
              <el-input v-model="twistParams.context" type="textarea" :rows="2" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateTwist" :loading="loading">
            生成转折
          </el-button>
          <div v-if="twistResult" class="results">
            <p><strong>{{ twistResult.title }}</strong></p>
            <p>{{ twistResult.description }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { toolboxApi } from '@/api'

const loading = ref(false)

const namingParams = ref({ type: '人名', genre: '玄幻', count: 5, gender: '' })
const namingResults = ref([])

const characterParams = ref({ type: '主角', gender: '男', genre: '玄幻', theme: '' })
const characterResult = ref(null)

const conflictParams = ref({ type: '人物', genre: '玄幻', context: '' })
const conflictResult = ref(null)

const sceneParams = ref({ type: '战斗', location: '', mood: '' })
const sceneResult = ref(null)

const goldfingerParams = ref({ type: '系统', genre: '玄幻', level: '中等', theme: '' })
const goldfingerResult = ref(null)

const titleParams = ref({ genre: '玄幻', theme: '', count: 5, style: '' })
const titleResults = ref([])

const synopsisParams = ref({ genre: '玄幻', main_char: '', world_view: '', type: 'short' })
const synopsisResult = ref(null)

const twistParams = ref({ type: 'unexpected', genre: '玄幻', context: '', characters: '' })
const twistResult = ref(null)

const generateNames = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.naming(namingParams.value)
    namingResults.value = res.data?.names || []
    if (!res.data?.names) {
      ElMessage.info(res.data?.message || '请使用 CLI 命令生成名称')
    }
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateCharacter = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.character(characterParams.value)
    characterResult.value = res.data || null
    if (!res.data?.name) {
      ElMessage.info(res.data?.message || '请使用 CLI 命令生成角色')
    }
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateConflict = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.conflict(conflictParams.value)
    conflictResult.value = res.data || null
    if (!res.data?.title) {
      ElMessage.info(res.data?.message || '请使用 CLI 命令生成冲突')
    }
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateScene = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.scene(sceneParams.value)
    sceneResult.value = res.data || null
    if (!res.data?.description) {
      ElMessage.info(res.data?.message || '请使用 CLI 命令生成场景')
    }
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateGoldfinger = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.goldfinger(goldfingerParams.value)
    goldfingerResult.value = res.data || null
    if (!res.data?.name) {
      ElMessage.info(res.data?.message || '请使用 CLI 命令生成金手指')
    }
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateTitle = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.title(titleParams.value)
    titleResults.value = res.data?.titles || []
    ElMessage.info(res.data?.message || '请使用 CLI 命令生成书名')
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateSynopsis = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.synopsis(synopsisParams.value)
    synopsisResult.value = res.data || null
    ElMessage.info(res.data?.message || '请使用 CLI 命令生成简介')
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}

const generateTwist = async () => {
  loading.value = true
  try {
    const res = await toolboxApi.twist(twistParams.value)
    twistResult.value = res.data || null
    ElMessage.info(res.data?.message || '请使用 CLI 命令生成剧情转折')
  } catch (error) {
    ElMessage.error('生成失败')
  }
  loading.value = false
}
</script>

<style scoped>
.toolbox-view {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
}

.tool-card {
  height: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: bold;
}

.results {
  margin-top: 15px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.result-item {
  margin-bottom: 8px;
}

.meaning {
  color: #666;
  margin-left: 10px;
}

.results p {
  margin: 5px 0;
}
</style>