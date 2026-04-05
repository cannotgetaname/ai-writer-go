<template>
  <div class="analysis-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 智能分析</h2>
    </div>

    <el-tabs v-model="activeTab" class="analysis-tabs">
      <!-- 一致性检查 -->
      <el-tab-pane label="一致性检查" name="consistency">
        <el-card>
          <template #header>
            <span>章节一致性检查</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="起始章节">
              <el-input-number v-model="consistencyFrom" :min="1" />
            </el-form-item>
            <el-form-item label="结束章节">
              <el-input-number v-model="consistencyTo" :min="1" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="runConsistencyCheck" :loading="consistencyLoading">
                开始检查
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="consistencyReport" class="report-section">
            <el-alert :title="consistencyReport.summary" type="info" show-icon :closable="false" />
            <el-table :data="consistencyReport.issues" stripe style="margin-top: 15px;">
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="getIssueType(row.type)">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="entity" label="实体" width="120" />
              <el-table-column prop="severity" label="严重程度" width="100">
                <template #default="{ row }">
                  <el-tag :type="getSeverityType(row.severity)">{{ row.severity }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="问题描述" show-overflow-tooltip />
              <el-table-column prop="suggestion" label="修改建议" show-overflow-tooltip />
            </el-table>
            <el-empty v-if="!consistencyReport.issues?.length" description="未发现一致性问题" />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 情感弧线 -->
      <el-tab-pane label="情感追踪" name="emotion">
        <el-card>
          <template #header>
            <span>角色情感弧线追踪</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="章节">
              <el-input-number v-model="emotionChapter" :min="1" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="trackEmotion" :loading="emotionLoading">
                追踪情感
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="emotionData" class="emotion-section">
            <h4>第 {{ emotionChapter }} 章角色情感</h4>
            <el-row :gutter="20">
              <el-col :span="8" v-for="(point, name) in emotionData" :key="name">
                <el-card shadow="hover" class="emotion-card">
                  <div class="emotion-header">
                    <el-avatar>{{ name?.charAt(0) }}</el-avatar>
                    <span class="char-name">{{ name }}</span>
                  </div>
                  <el-descriptions :column="1" border size="small">
                    <el-descriptions-item label="情绪">{{ point.emotion }}</el-descriptions-item>
                    <el-descriptions-item label="强度">
                      <el-progress :percentage="point.intensity * 10" :stroke-width="8" />
                    </el-descriptions-item>
                    <el-descriptions-item label="触发">{{ point.trigger }}</el-descriptions-item>
                  </el-descriptions>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-card>

        <el-card style="margin-top: 20px;">
          <template #header>
            <span>角色情感弧线查询</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="角色名">
              <el-select v-model="selectedCharacter" placeholder="选择角色">
                <el-option v-for="char in characters" :key="char.name" :label="char.name" :value="char.name" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="getEmotionArc" :loading="arcLoading">
                获取弧线
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="arcData" class="arc-section">
            <el-alert :title="arcData.summary" type="info" show-icon :closable="false" />
            <div v-if="arcData.arc?.length" style="margin-top: 15px;">
              <el-table :data="arcData.arc" stripe>
                <el-table-column prop="chapter_id" label="章节" width="80" />
                <el-table-column prop="emotion" label="情绪" />
                <el-table-column prop="intensity" label="强度" width="80" />
                <el-table-column prop="trigger" label="触发事件" show-overflow-tooltip />
              </el-table>
            </div>
            <el-empty v-else description="暂无情感数据" />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- 信息边界 -->
      <el-tab-pane label="信息边界" name="infoBoundary">
        <el-card>
          <template #header>
            <span>信息越界检测</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="章节">
              <el-input-number v-model="infoChapter" :min="1" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="checkInfoLeak" :loading="infoLoading">
                检测越界
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="leakData" class="leak-section">
            <el-alert
              :title="`发现 ${leakData.count} 个信息越界问题`"
              :type="leakData.count > 0 ? 'warning' : 'success'"
              show-icon
              :closable="false"
            />
            <div v-if="leakData.leaks?.length" style="margin-top: 15px;">
              <el-card v-for="(leak, index) in leakData.leaks" :key="index" shadow="hover" style="margin-bottom: 10px;">
                <p>{{ leak }}</p>
              </el-card>
            </div>
          </div>
        </el-card>

        <el-card style="margin-top: 20px;">
          <template #header>
            <span>角色信息提取</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="章节">
              <el-input-number v-model="extractChapter" :min="1" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="extractInfo" :loading="extractLoading">
                提取信息
              </el-button>
            </el-form-item>
          </el-form>

          <div v-if="extractData" class="extract-section">
            <el-alert
              :title="`从第 ${extractChapter} 章提取角色信息`"
              type="info"
              show-icon
              :closable="false"
            />
            <div v-if="Object.keys(extractData.info_map || {}).length" style="margin-top: 15px;">
              <el-collapse>
                <el-collapse-item v-for="(infos, charName) in extractData.info_map" :key="charName" :title="charName">
                  <el-table :data="infos" stripe size="small">
                    <el-table-column prop="info_key" label="信息标识" width="120" />
                    <el-table-column prop="content" label="内容" />
                    <el-table-column prop="source" label="来源" width="100" />
                  </el-table>
                </el-collapse-item>
              </el-collapse>
            </div>
            <el-empty v-else description="未提取到新信息" />
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { consistencyApi, emotionApi, infoBoundaryApi, settingsApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const activeTab = ref('consistency')

// 一致性检查
const consistencyFrom = ref(1)
const consistencyTo = ref(1)
const consistencyLoading = ref(false)
const consistencyReport = ref(null)

// 情感追踪
const emotionChapter = ref(1)
const emotionLoading = ref(false)
const emotionData = ref(null)
const selectedCharacter = ref('')
const arcLoading = ref(false)
const arcData = ref(null)
const characters = ref([])

// 信息边界
const infoChapter = ref(1)
const infoLoading = ref(false)
const leakData = ref(null)
const extractChapter = ref(1)
const extractLoading = ref(false)
const extractData = ref(null)

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

// 一致性检查
const runConsistencyCheck = async () => {
  consistencyLoading.value = true
  try {
    const res = await consistencyApi.check(bookId.value, consistencyFrom.value, consistencyTo.value)
    consistencyReport.value = res.data
    ElMessage.success('一致性检查完成')
  } catch (error) {
    ElMessage.error('检查失败: ' + error.message)
  }
  consistencyLoading.value = false
}

const getIssueType = (type) => {
  const types = { character: 'primary', item: 'success', location: 'warning', plot: 'danger', time: 'info' }
  return types[type] || ''
}

const getSeverityType = (severity) => {
  const types = { high: 'danger', medium: 'warning', low: 'info' }
  return types[severity] || ''
}

// 情感追踪
const trackEmotion = async () => {
  emotionLoading.value = true
  try {
    const res = await emotionApi.track(bookId.value, emotionChapter.value)
    emotionData.value = res.data
    ElMessage.success('情感追踪完成')
  } catch (error) {
    ElMessage.error('追踪失败: ' + error.message)
  }
  emotionLoading.value = false
}

const getEmotionArc = async () => {
  if (!selectedCharacter.value) {
    ElMessage.warning('请选择角色')
    return
  }
  arcLoading.value = true
  try {
    const res = await emotionApi.getArc(bookId.value, selectedCharacter.value)
    arcData.value = res.data
    ElMessage.success('获取成功')
  } catch (error) {
    ElMessage.error('获取失败: ' + error.message)
  }
  arcLoading.value = false
}

// 信息边界
const checkInfoLeak = async () => {
  infoLoading.value = true
  try {
    const res = await infoBoundaryApi.checkLeak(bookId.value, infoChapter.value)
    leakData.value = res.data
    ElMessage.success('检测完成')
  } catch (error) {
    ElMessage.error('检测失败: ' + error.message)
  }
  infoLoading.value = false
}

const extractInfo = async () => {
  extractLoading.value = true
  try {
    const res = await infoBoundaryApi.extractInfo(bookId.value, extractChapter.value)
    extractData.value = res.data
    ElMessage.success('提取完成')
  } catch (error) {
    ElMessage.error('提取失败: ' + error.message)
  }
  extractLoading.value = false
}

// 加载角色列表
const loadCharacters = async () => {
  try {
    const res = await settingsApi.getCharacters(bookId.value)
    characters.value = res.data || []
  } catch (error) {
    console.error('加载角色失败:', error)
  }
}

onMounted(() => {
  loadCharacters()
})
</script>

<style scoped>
.analysis-view {
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

.analysis-tabs {
  margin-top: 20px;
}

.report-section,
.emotion-section,
.arc-section,
.leak-section,
.extract-section {
  margin-top: 20px;
}

.emotion-card {
  margin-bottom: 15px;
}

.emotion-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.char-name {
  font-weight: bold;
}
</style>