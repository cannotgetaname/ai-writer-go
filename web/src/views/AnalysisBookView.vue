<template>
  <div class="analysis-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 智能分析</h2>
      <el-button type="primary" @click="goToSettings">
        <el-icon><Tools /></el-icon>
        设定管理
      </el-button>
    </div>

    <el-tabs v-model="activeTab" class="analysis-tabs">
      <!-- 分析报告 -->
      <el-tab-pane label="分析报告" name="reports">
        <el-row :gutter="20">
          <!-- 左侧：运行分析 -->
          <el-col :span="6">
            <el-card>
              <template #header>
                <span>运行分析</span>
              </template>
              <el-form label-width="70px">
                <el-form-item label="章节">
                  <el-input-number v-model="runChapterId" :min="1" />
                </el-form-item>
                <el-form-item label="类型">
                  <el-select v-model="runType" style="width: 100%;">
                    <el-option label="手动分析" value="manual" />
                    <el-option label="审稿分析" value="review" />
                  </el-select>
                </el-form-item>
                <el-form-item>
                  <el-button type="primary" @click="runAnalysis" :loading="runningAnalysis">
                    <el-icon><DataAnalysis /></el-icon>
                    运行分析
                  </el-button>
                </el-form-item>
              </el-form>
            </el-card>

            <!-- 快捷导航 -->
            <el-card style="margin-top: 20px;">
              <template #header>
                <span>快捷导航</span>
              </template>
              <div class="nav-links">
                <el-button size="small" @click="goToSettings">角色设定</el-button>
                <el-button size="small" @click="goToTimeline">时间线</el-button>
                <el-button size="small" @click="goToGraph">知识图谱</el-button>
                <el-button size="small" @click="goToSync">状态同步</el-button>
              </div>
            </el-card>
          </el-col>

          <!-- 右侧：报告列表 -->
          <el-col :span="18">
            <el-card>
              <template #header>
                <div class="card-header">
                  <span>历史分析报告 ({{ reports.length }})</span>
                  <el-button size="small" @click="loadReports" :loading="loadingReports">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-button>
                </div>
              </template>

              <el-empty v-if="!reports.length" description="暂无分析报告，点击运行分析生成报告" />

              <div v-else class="reports-list">
                <el-collapse v-model="expandedReport" accordion>
                  <el-collapse-item v-for="report in reports" :key="report.id" :name="report.id">
                    <template #title>
                      <div class="report-title">
                        <el-tag :type="report.type === 'manual' ? 'primary' : 'success'" size="small">
                          {{ report.type === 'manual' ? '手动' : '审稿' }}
                        </el-tag>
                        <span class="report-chapter">第 {{ report.chapter_id }} 章</span>
                        <span class="report-time">{{ formatTime(report.created_at) }}</span>
                        <div class="report-scores">
                          <el-tag v-if="report.foreshadow_analysis?.score"
                                  :type="getScoreType(report.foreshadow_analysis.score)" size="small">
                            伏笔: {{ report.foreshadow_analysis.score }}
                          </el-tag>
                          <el-tag v-if="report.causal_analysis?.score"
                                  :type="getScoreType(report.causal_analysis.score)" size="small">
                            因果: {{ report.causal_analysis.score }}
                          </el-tag>
                          <el-tag v-if="report.thread_analysis?.score"
                                  :type="getScoreType(report.thread_analysis.score)" size="small">
                            线程: {{ report.thread_analysis.score }}
                          </el-tag>
                          <el-tag v-if="report.emotion_analysis?.score"
                                  :type="getScoreType(report.emotion_analysis.score)" size="small">
                            情感: {{ report.emotion_analysis.score }}
                          </el-tag>
                          <el-tag v-if="report.timeline_analysis?.score"
                                  :type="getScoreType(report.timeline_analysis.score)" size="small">
                            时间: {{ report.timeline_analysis.score }}
                          </el-tag>
                        </div>
                      </div>
                    </template>

                    <!-- 报告详情 -->
                    <div class="report-details">
                      <!-- 伏笔分析 -->
                      <el-card v-if="report.foreshadow_analysis" shadow="never" class="detail-card">
                        <template #header>
                          <div class="detail-header">
                            <span>伏笔分析</span>
                            <el-rate v-model="report.foreshadow_analysis.score" disabled :max="10" />
                          </div>
                        </template>
                        <div v-if="report.foreshadow_analysis.warnings?.length" class="issues-section">
                          <el-alert type="warning" :closable="false" style="margin-bottom: 10px;">
                            <template #title>发现问题 ({{ report.foreshadow_analysis.warnings.length }})</template>
                          </el-alert>
                          <el-table :data="report.foreshadow_analysis.warnings" stripe size="small">
                            <el-table-column prop="foreshadow_id" label="ID" width="100" />
                            <el-table-column prop="issue_type" label="类型" width="100">
                              <template #default="{ row }">
                                <el-tag size="small">{{ row.issue_type }}</el-tag>
                              </template>
                            </el-table-column>
                            <el-table-column prop="description" label="描述" show-overflow-tooltip />
                          </el-table>
                        </div>
                        <div v-if="report.foreshadow_analysis.suggestions?.length">
                          <h5 style="margin-top: 10px;">建议</h5>
                          <ul class="suggestions-list">
                            <li v-for="(sug, i) in report.foreshadow_analysis.suggestions" :key="i">{{ sug }}</li>
                          </ul>
                        </div>
                        <el-empty v-if="!report.foreshadow_analysis.warnings?.length && !report.foreshadow_analysis.suggestions?.length"
                                  description="无伏笔问题" :image-size="40" />
                      </el-card>

                      <!-- 因果链分析 -->
                      <el-card v-if="report.causal_analysis" shadow="never" class="detail-card">
                        <template #header>
                          <div class="detail-header">
                            <span>因果链分析</span>
                            <el-rate v-model="report.causal_analysis.score" disabled :max="10" />
                          </div>
                        </template>
                        <div v-if="report.causal_analysis.broken_chains?.length" class="issues-section">
                          <el-alert type="danger" :closable="false" style="margin-bottom: 10px;">
                            <template #title>断裂链 ({{ report.causal_analysis.broken_chains.length }})</template>
                          </el-alert>
                          <el-table :data="report.causal_analysis.broken_chains" stripe size="small">
                            <el-table-column prop="event_name" label="事件" />
                            <el-table-column prop="issue" label="问题" width="100">
                              <template #default="{ row }">
                                <el-tag type="danger" size="small">{{ row.issue }}</el-tag>
                              </template>
                            </el-table-column>
                          </el-table>
                        </div>
                        <div v-if="report.causal_analysis.orphan_events?.length" class="issues-section">
                          <el-alert type="warning" :closable="false" style="margin-bottom: 10px;">
                            <template #title>孤立事件 ({{ report.causal_analysis.orphan_events.length }})</template>
                          </el-alert>
                          <el-table :data="report.causal_analysis.orphan_events" stripe size="small">
                            <el-table-column prop="event_name" label="事件" />
                          </el-table>
                        </div>
                        <div v-if="report.causal_analysis.circular_deps?.length" class="issues-section">
                          <el-alert type="error" :closable="false" style="margin-bottom: 10px;">
                            <template #title>循环依赖 ({{ report.causal_analysis.circular_deps.length }})</template>
                          </el-alert>
                          <div v-for="(dep, i) in report.causal_analysis.circular_deps" :key="i">
                            <el-tag type="danger" size="small">循环 {{ i + 1 }}</el-tag>
                            <span class="dep-chain">{{ dep.chain?.join(' -> ') }}</span>
                          </div>
                        </div>
                        <el-empty v-if="!report.causal_analysis.broken_chains?.length &&
                                       !report.causal_analysis.orphan_events?.length &&
                                       !report.causal_analysis.circular_deps?.length"
                                  description="因果链完整" :image-size="40" />
                      </el-card>

                      <!-- 叙事线程分析 -->
                      <el-card v-if="report.thread_analysis" shadow="never" class="detail-card">
                        <template #header>
                          <div class="detail-header">
                            <span>叙事线程分析</span>
                            <el-rate v-model="report.thread_analysis.score" disabled :max="10" />
                          </div>
                        </template>
                        <div v-if="report.thread_analysis.forgotten_threads?.length" class="issues-section">
                          <el-alert type="warning" :closable="false" style="margin-bottom: 10px;">
                            <template #title>遗忘线程 ({{ report.thread_analysis.forgotten_threads.length }})</template>
                          </el-alert>
                          <el-table :data="report.thread_analysis.forgotten_threads" stripe size="small">
                            <el-table-column prop="thread_name" label="线程" />
                            <el-table-column prop="last_active" label="最后活跃" width="80" />
                            <el-table-column prop="chapters_skipped" label="跳过章节" width="80" />
                          </el-table>
                        </div>
                        <div v-if="report.thread_analysis.pacing_issues?.length" class="issues-section">
                          <el-alert type="info" :closable="false" style="margin-bottom: 10px;">
                            <template #title>节奏问题 ({{ report.thread_analysis.pacing_issues.length }})</template>
                          </el-alert>
                          <el-table :data="report.thread_analysis.pacing_issues" stripe size="small">
                            <el-table-column prop="thread_name" label="线程" />
                            <el-table-column prop="issue" label="问题" show-overflow-tooltip />
                          </el-table>
                        </div>
                        <div v-if="report.thread_analysis.conflicts?.length" class="issues-section">
                          <el-alert type="error" :closable="false" style="margin-bottom: 10px;">
                            <template #title>线程冲突 ({{ report.thread_analysis.conflicts.length }})</template>
                          </el-alert>
                          <el-table :data="report.thread_analysis.conflicts" stripe size="small">
                            <el-table-column prop="chapter_id" label="章节" width="80" />
                            <el-table-column prop="conflict_type" label="冲突类型" />
                          </el-table>
                        </div>
                        <el-empty v-if="!report.thread_analysis.forgotten_threads?.length &&
                                       !report.thread_analysis.pacing_issues?.length &&
                                       !report.thread_analysis.conflicts?.length"
                                  description="线程运转正常" :image-size="40" />
                      </el-card>

                      <!-- 情感弧线分析 -->
                      <el-card v-if="report.emotion_analysis" shadow="never" class="detail-card">
                        <template #header>
                          <div class="detail-header">
                            <span>情感弧线分析</span>
                            <el-rate v-model="report.emotion_analysis.score" disabled :max="10" />
                          </div>
                        </template>
                        <div v-if="report.emotion_analysis.inconsistencies?.length" class="issues-section">
                          <el-alert type="warning" :closable="false" style="margin-bottom: 10px;">
                            <template #title>情感不一致 ({{ report.emotion_analysis.inconsistencies.length }})</template>
                          </el-alert>
                          <el-table :data="report.emotion_analysis.inconsistencies" stripe size="small">
                            <el-table-column prop="character" label="角色" width="100" />
                            <el-table-column prop="chapter_id" label="章节" width="80" />
                            <el-table-column prop="from_emotion" label="原情绪" width="80" />
                            <el-table-column prop="to_emotion" label="新情绪" width="80" />
                            <el-table-column prop="intensity_jump" label="强度跳跃" width="80" />
                          </el-table>
                        </div>
                        <div v-if="report.emotion_analysis.pacing_issues?.length" class="issues-section">
                          <el-alert type="info" :closable="false" style="margin-bottom: 10px;">
                            <template #title>情感节奏问题 ({{ report.emotion_analysis.pacing_issues.length }})</template>
                          </el-alert>
                          <el-table :data="report.emotion_analysis.pacing_issues" stripe size="small">
                            <el-table-column prop="character" label="角色" />
                            <el-table-column prop="issue" label="问题" show-overflow-tooltip />
                          </el-table>
                        </div>
                        <el-empty v-if="!report.emotion_analysis.inconsistencies?.length &&
                                       !report.emotion_analysis.pacing_issues?.length"
                                  description="情感弧线流畅" :image-size="40" />
                        <div v-if="report.emotion_analysis.weaving_score" style="margin-top: 10px;">
                          <span>编织分数: </span>
                          <el-rate v-model="report.emotion_analysis.weaving_score" disabled :max="10" />
                        </div>
                      </el-card>

                      <!-- 时间线分析 -->
                      <el-card v-if="report.timeline_analysis" shadow="never" class="detail-card">
                        <template #header>
                          <div class="detail-header">
                            <span>时间线分析</span>
                            <el-rate v-model="report.timeline_analysis.score" disabled :max="10" />
                          </div>
                        </template>
                        <div v-if="report.timeline_analysis.time_jumps?.length" class="issues-section">
                          <el-alert type="info" :closable="false" style="margin-bottom: 10px;">
                            <template #title>时间跳跃 ({{ report.timeline_analysis.time_jumps.length }})</template>
                          </el-alert>
                          <el-table :data="report.timeline_analysis.time_jumps" stripe size="small">
                            <el-table-column prop="from_chapter" label="起始章节" width="80" />
                            <el-table-column prop="to_chapter" label="目标章节" width="80" />
                            <el-table-column prop="from_time" label="起始时间" />
                            <el-table-column prop="to_time" label="目标时间" />
                            <el-table-column prop="duration" label="跨度" />
                          </el-table>
                        </div>
                        <div v-if="report.timeline_analysis.overlaps?.length" class="issues-section">
                          <el-alert type="warning" :closable="false" style="margin-bottom: 10px;">
                            <template #title>时间重叠 ({{ report.timeline_analysis.overlaps.length }})</template>
                          </el-alert>
                          <el-table :data="report.timeline_analysis.overlaps" stripe size="small">
                            <el-table-column prop="chapter_id" label="章节" width="80" />
                            <el-table-column prop="time_label" label="时间" />
                            <el-table-column prop="events" label="重叠事件">
                              <template #default="{ row }">
                                <span>{{ row.events?.join(', ') }}</span>
                              </template>
                            </el-table-column>
                          </el-table>
                        </div>
                        <div v-if="report.timeline_analysis.inconsistencies?.length" class="issues-section">
                          <el-alert type="error" :closable="false" style="margin-bottom: 10px;">
                            <template #title>时序矛盾 ({{ report.timeline_analysis.inconsistencies.length }})</template>
                          </el-alert>
                          <el-table :data="report.timeline_analysis.inconsistencies" stripe size="small">
                            <el-table-column prop="chapter_id" label="章节" width="80" />
                            <el-table-column prop="event_order" label="事件顺序" />
                            <el-table-column prop="issue" label="问题" show-overflow-tooltip />
                          </el-table>
                        </div>
                        <el-empty v-if="!report.timeline_analysis.time_jumps?.length &&
                                       !report.timeline_analysis.overlaps?.length &&
                                       !report.timeline_analysis.inconsistencies?.length"
                                  description="时间线连贯" :image-size="40" />
                      </el-card>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </el-tab-pane>

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

      <!-- 情感追踪 -->
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
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { consistencyApi, emotionApi, infoBoundaryApi, settingsApi, analysisApi } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const activeTab = ref('reports')

// 分析报告
const reports = ref([])
const loadingReports = ref(false)
const runningAnalysis = ref(false)
const runChapterId = ref(1)
const runType = ref('manual')
const expandedReport = ref('')

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

const goToSettings = () => {
  router.push(`/books/${bookId.value}/settings`)
}

const goToTimeline = () => {
  router.push(`/books/${bookId.value}/timeline`)
}

const goToGraph = () => {
  router.push(`/books/${bookId.value}/graph`)
}

const goToSync = () => {
  router.push(`/books/${bookId.value}/sync`)
}

// 分析报告
const loadReports = async () => {
  loadingReports.value = true
  try {
    const res = await analysisApi.getReports(bookId.value)
    reports.value = res.data || []
    // 默认展开第一个报告
    if (reports.value.length > 0 && !expandedReport.value) {
      expandedReport.value = reports.value[0].id
    }
  } catch (error) {
    console.error('加载报告失败:', error)
    reports.value = []
  }
  loadingReports.value = false
}

const runAnalysis = async () => {
  runningAnalysis.value = true
  try {
    const res = await analysisApi.run(bookId.value, runChapterId.value, runType.value)
    ElMessage.success('分析运行完成')
    // 刷新报告列表
    await loadReports()
    // 展开新报告
    if (res.data?.id) {
      expandedReport.value = res.data.id
    }
  } catch (error) {
    ElMessage.error('分析失败: ' + (error.response?.data?.error || error.message))
  }
  runningAnalysis.value = false
}

const getScoreType = (score) => {
  if (score >= 8) return 'success'
  if (score >= 5) return 'warning'
  return 'danger'
}

const formatTime = (time) => {
  if (!time) return ''
  const date = new Date(time)
  return date.toLocaleString()
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

watch(bookId, () => {
  loadReports()
  loadCharacters()
})

onMounted(() => {
  loadReports()
  loadCharacters()
})
</script>

<style scoped>
.analysis-view {
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
  flex: 1;
}

.analysis-tabs {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-links {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.reports-list {
  margin-top: 10px;
}

.report-title {
  display: flex;
  align-items: center;
  gap: 15px;
  width: 100%;
}

.report-chapter {
  font-weight: bold;
}

.report-time {
  color: #909399;
  font-size: 13px;
}

.report-scores {
  display: flex;
  gap: 5px;
  margin-left: auto;
}

.report-details {
  padding: 10px 0;
}

.detail-card {
  margin-bottom: 15px;
  border: 1px solid #ebeef5;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.issues-section {
  margin-bottom: 15px;
}

.suggestions-list {
  padding-left: 20px;
  color: #606266;
  line-height: 1.8;
}

.suggestions-list li {
  margin-bottom: 5px;
}

.dep-chain {
  margin-left: 10px;
  color: #f56c6c;
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