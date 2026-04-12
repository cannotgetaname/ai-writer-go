<template>
  <div class="architect-view">
    <!-- Page Header -->
    <div class="page-header">
      <el-button class="back-btn" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2 class="page-title">{{ bookName }}</h2>
      <el-tag type="primary" effect="plain">架构师</el-tag>
    </div>

    <!-- Main Layout -->
    <div class="architect-layout">
      <!-- Left Sidebar: Step Navigation -->
      <div class="sidebar">
        <ArchitectStepCard
          v-for="(step, index) in steps"
          :key="index"
          :index="index"
          :title="step.title"
          :status="getStepStatus(index)"
          :is-active="currentStep === index"
          :summary1="getStepSummary1(index)"
          :summary2="getStepSummary2(index)"
          @click="switchStep(index)"
        />
      </div>

      <!-- Right Workspace -->
      <div class="workspace-container">
        <ArchitectWorkspace
          :step-index="currentStep"
          :title="steps[currentStep].title"
          :description="steps[currentStep].description"
          :context-preview="getContextPreview()"
          :is-outdated="outdatedSteps.has(currentStep)"
          :outdated-message="getOutdatedMessage()"
          :loading="loading"
          :loading-text="loadingText"
          @regenerate="handleRegenerate"
          @cancel="cancelGeneration"
        >
        <!-- Action Buttons Slot -->
        <template #actions>
          <slot name="actions">
            <!-- Step 0: Synopsis Actions -->
            <el-button
              v-if="currentStep === 0"
              type="primary"
              :loading="loading"
              @click="generateSynopsis"
            >
              生成全书总纲
            </el-button>

            <!-- Step 1: WorldView Actions -->
            <el-button
              v-if="currentStep === 1"
              type="primary"
              :loading="loading"
              :disabled="!synopsisResult"
              @click="generateWorldView"
            >
              生成世界观设定
            </el-button>

            <!-- Step 2: Volumes Actions -->
            <el-button
              v-if="currentStep === 2"
              type="primary"
              :loading="loading"
              :disabled="!worldViewResult"
              @click="generateVolumes"
            >
              生成分卷大纲
            </el-button>

            <!-- Step 3: Chapters Actions -->
            <template v-if="currentStep === 3">
              <el-select
                v-model="selectedVolumeIndex"
                placeholder="选择分卷"
                style="width: 200px"
                :disabled="volumeResults.length === 0"
              >
                <el-option
                  v-for="(vol, idx) in volumeResults"
                  :key="idx"
                  :label="vol.title"
                  :value="idx"
                />
              </el-select>
              <el-button
                type="primary"
                :loading="loading"
                :disabled="selectedVolumeIndex < 0"
                @click="expandVolume"
              >
                展开选中分卷
              </el-button>
              <el-button
                type="success"
                :loading="loading"
                @click="expandAllVolumes"
              >
                展开全部分卷
              </el-button>
            </template>

            <!-- Step 4: Chapter Detail Actions -->
            <template v-if="currentStep === 4">
              <el-cascader
                v-model="selectedChapterPath"
                :options="chapterOptions"
                :props="{ value: 'id', label: 'title', children: 'chapters' }"
                placeholder="选择章节"
                clearable
                style="width: 300px"
              />
              <el-button
                type="primary"
                :loading="loading"
                :disabled="!selectedChapterPath"
                @click="expandChapter"
              >
                生成章节细纲
              </el-button>
            </template>
          </slot>
        </template>

        <!-- Content Slot -->
        <template #default>
          <!-- Step 0: Synopsis Content -->
          <div v-if="currentStep === 0" class="step-content">
            <!-- Input Form -->
            <div v-if="!synopsisResult" class="input-form">
              <el-form :model="synopsisParams" label-width="100px">
                <el-form-item label="题材类型">
                  <el-select v-model="synopsisParams.genre">
                    <el-option label="玄幻" value="玄幻" />
                    <el-option label="仙侠" value="仙侠" />
                    <el-option label="都市" value="都市" />
                    <el-option label="科幻" value="科幻" />
                    <el-option label="武侠" value="武侠" />
                    <el-option label="历史" value="历史" />
                  </el-select>
                </el-form-item>
                <el-form-item label="主题方向">
                  <el-input v-model="synopsisParams.theme" placeholder="如：逆袭、复仇、寻宝、成长" />
                </el-form-item>
                <el-form-item label="主角设定">
                  <el-input v-model="synopsisParams.main_char" placeholder="如：少年天才、穿越者、废柴逆袭" />
                </el-form-item>
                <el-form-item label="目标字数">
                  <el-input-number v-model="synopsisParams.target_words" :min="100000" :max="5000000" :step="100000" />
                </el-form-item>
              </el-form>
            </div>

            <!-- Result Display -->
            <div v-if="synopsisResult" class="result-section">
              <div class="result-header">
                <el-tag v-if="synopsisResult.title" size="large">{{ synopsisResult.title }}</el-tag>
                <el-button size="small" @click="editMode.synopsis = !editMode.synopsis">
                  {{ editMode.synopsis ? '预览' : '编辑' }}
                </el-button>
              </div>

              <!-- Edit Mode -->
              <div v-if="editMode.synopsis">
                <el-form :model="synopsisResult" label-width="100px">
                  <el-form-item label="书名">
                    <el-input v-model="synopsisResult.title" />
                  </el-form-item>
                  <el-form-item label="题材">
                    <el-input v-model="synopsisResult.genre" />
                  </el-form-item>
                  <el-form-item label="主题">
                    <el-input v-model="synopsisResult.theme" />
                  </el-form-item>
                  <el-form-item label="预估字数">
                    <el-input-number v-model="synopsisResult.word_count" :min="100000" :max="5000000" />
                  </el-form-item>
                  <el-form-item label="结局类型">
                    <el-input v-model="synopsisResult.ending_type" />
                  </el-form-item>
                  <el-form-item label="分卷数量">
                    <el-input-number v-model="synopsisResult.volume_count" :min="1" :max="20" />
                  </el-form-item>
                  <el-form-item label="故事梗概">
                    <el-input v-model="synopsisResult.synopsis" type="textarea" :rows="5" />
                  </el-form-item>
                  <el-form-item label="主线剧情">
                    <el-input v-model="synopsisResult.main_plot" type="textarea" :rows="3" />
                  </el-form-item>
                </el-form>
              </div>

              <!-- Preview Mode -->
              <div v-else class="preview-mode">
                <el-descriptions :column="2" border size="small">
                  <el-descriptions-item label="题材">{{ synopsisResult.genre }}</el-descriptions-item>
                  <el-descriptions-item label="主题">{{ synopsisResult.theme }}</el-descriptions-item>
                  <el-descriptions-item label="预估字数">{{ synopsisResult.word_count }}</el-descriptions-item>
                  <el-descriptions-item label="结局类型">{{ synopsisResult.ending_type }}</el-descriptions-item>
                </el-descriptions>
                <div class="synopsis-text">
                  <strong>故事梗概：</strong>
                  <p>{{ synopsisResult.synopsis }}</p>
                </div>
                <div v-if="synopsisResult.main_plot">
                  <strong>主线剧情：</strong>
                  <p>{{ synopsisResult.main_plot }}</p>
                </div>
              </div>
            </div>
          </div>

          <!-- Step 1: WorldView Content -->
          <div v-if="currentStep === 1" class="step-content">
            <div v-if="worldViewResult" class="result-section">
              <div class="result-header">
                <el-button size="small" @click="editMode.worldview = !editMode.worldview">
                  {{ editMode.worldview ? '预览' : '编辑' }}
                </el-button>
                <el-button size="small" type="success" @click="saveWorldView">保存世界观</el-button>
              </div>

              <!-- Edit Mode -->
              <div v-if="editMode.worldview">
                <el-form :model="worldViewResult" label-width="100px">
                  <el-form-item label="时代背景">
                    <el-input v-model="worldViewResult.era" />
                  </el-form-item>
                  <el-form-item label="科技水平">
                    <el-input v-model="worldViewResult.tech_level" />
                  </el-form-item>
                  <el-form-item label="力量体系">
                    <el-input v-model="worldViewResult.power_system" type="textarea" :rows="3" />
                  </el-form-item>
                  <el-form-item label="社会结构">
                    <el-input v-model="worldViewResult.social_structure" type="textarea" :rows="3" />
                  </el-form-item>
                  <el-form-item label="特殊规则">
                    <el-input v-model="worldViewResult.special_rules" type="textarea" :rows="3" />
                  </el-form-item>
                  <el-form-item label="主要势力">
                    <el-input v-model="worldViewResult.organizations" type="textarea" :rows="2" />
                  </el-form-item>
                  <el-form-item label="主要地点">
                    <el-input v-model="worldViewResult.locations" type="textarea" :rows="2" />
                  </el-form-item>
                  <el-form-item label="历史背景">
                    <el-input v-model="worldViewResult.history" type="textarea" :rows="3" />
                  </el-form-item>
                  <el-form-item label="主要矛盾">
                    <el-input v-model="worldViewResult.main_conflict" type="textarea" :rows="2" />
                  </el-form-item>
                </el-form>
              </div>

              <!-- Preview Mode -->
              <el-descriptions v-else :column="1" border size="small">
                <el-descriptions-item label="时代背景">{{ worldViewResult.era }}</el-descriptions-item>
                <el-descriptions-item label="科技水平">{{ worldViewResult.tech_level }}</el-descriptions-item>
                <el-descriptions-item label="力量体系">
                  <div class="long-text">{{ worldViewResult.power_system }}</div>
                </el-descriptions-item>
                <el-descriptions-item label="社会结构">
                  <div class="long-text">{{ worldViewResult.social_structure }}</div>
                </el-descriptions-item>
                <el-descriptions-item label="特殊规则">
                  <div class="long-text">{{ worldViewResult.special_rules }}</div>
                </el-descriptions-item>
                <el-descriptions-item label="主要势力">{{ worldViewResult.organizations }}</el-descriptions-item>
                <el-descriptions-item label="主要地点">{{ worldViewResult.locations }}</el-descriptions-item>
                <el-descriptions-item label="历史背景">
                  <div class="long-text">{{ worldViewResult.history }}</div>
                </el-descriptions-item>
                <el-descriptions-item label="主要矛盾">{{ worldViewResult.main_conflict }}</el-descriptions-item>
              </el-descriptions>
            </div>

            <!-- Empty State -->
            <div v-if="!worldViewResult && !loading" class="empty-state">
              <p>请先生成全书总纲，然后点击"生成世界观设定"按钮</p>
            </div>
          </div>

          <!-- Step 2: Volumes Content -->
          <div v-if="currentStep === 2" class="step-content">
            <!-- Volume Count Input -->
            <div v-if="!volumeResults.length && synopsisResult" class="input-form">
              <el-form label-width="100px">
                <el-form-item label="分卷数量">
                  <el-input-number v-model="synopsisResult.volume_count" :min="1" :max="20" />
                </el-form-item>
              </el-form>
            </div>

            <!-- Result Display -->
            <div v-if="volumeResults.length > 0" class="result-section">
              <div class="result-header">
                <span>分卷大纲 ({{ volumeResults.length }}卷)</span>
                <el-button size="small" type="success" @click="saveOutline">保存为章节</el-button>
              </div>

              <el-collapse>
                <el-collapse-item
                  v-for="(vol, idx) in volumeResults"
                  :key="idx"
                  :name="idx"
                >
                  <template #title>
                    <div class="volume-title">
                      <el-tag>{{ vol.title }}</el-tag>
                      <el-text type="info" size="small">共{{ vol.chapter_count || vol.chapters?.length || 0 }}章</el-text>
                      <el-button size="small" link @click.stop="toggleVolumeEdit(idx)" style="margin-left: 10px;">
                        {{ editMode.volumes[idx] ? '预览' : '编辑' }}
                      </el-button>
                    </div>
                  </template>

                  <!-- Edit Mode -->
                  <div v-if="editMode.volumes[idx]" class="volume-content">
                    <el-form :model="vol" label-width="80px" size="small">
                      <el-form-item label="卷名">
                        <el-input v-model="vol.title" />
                      </el-form-item>
                      <el-form-item label="梗概">
                        <el-input v-model="vol.synopsis" type="textarea" :rows="2" />
                      </el-form-item>
                      <el-form-item label="核心事件">
                        <el-input v-model="vol.main_event" type="textarea" :rows="2" />
                      </el-form-item>
                      <el-form-item label="情感弧线">
                        <el-input v-model="vol.emotion_arc" />
                      </el-form-item>
                      <el-form-item label="章节数">
                        <el-input-number v-model="vol.chapter_count" :min="1" />
                      </el-form-item>
                    </el-form>

                    <!-- Chapters List Edit -->
                    <div v-if="vol.chapters && vol.chapters.length > 0" class="chapters-list">
                      <el-divider content-position="left">章节列表</el-divider>
                      <el-table :data="vol.chapters" size="small" max-height="300">
                        <el-table-column prop="index" label="#" width="50" />
                        <el-table-column label="章节名">
                          <template #default="{ row }">
                            <el-input v-model="row.title" size="small" />
                          </template>
                        </el-table-column>
                        <el-table-column label="梗概">
                          <template #default="{ row }">
                            <el-input v-model="row.synopsis" size="small" />
                          </template>
                        </el-table-column>
                      </el-table>
                    </div>
                  </div>

                  <!-- Preview Mode -->
                  <div v-else class="volume-content">
                    <p><strong>梗概：</strong>{{ vol.synopsis }}</p>
                    <p><strong>核心事件：</strong>{{ vol.main_event }}</p>
                    <p><strong>情感弧线：</strong>{{ vol.emotion_arc }}</p>

                    <!-- Chapters List -->
                    <div v-if="vol.chapters && vol.chapters.length > 0" class="chapters-list">
                      <el-divider content-position="left">章节列表</el-divider>
                      <el-table :data="vol.chapters" size="small" max-height="300">
                        <el-table-column prop="index" label="#" width="50" />
                        <el-table-column prop="title" label="章节名" />
                        <el-table-column prop="synopsis" label="梗概" show-overflow-tooltip />
                      </el-table>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>
            </div>

            <!-- Empty State -->
            <div v-if="!volumeResults.length && !loading" class="empty-state">
              <p>请先生成世界观，然后点击"生成分卷大纲"按钮</p>
            </div>
          </div>

          <!-- Step 3: Chapters Content -->
          <div v-if="currentStep === 3" class="step-content">
            <!-- Progress Display -->
            <div class="expansion-progress">
              <el-progress
                :percentage="expansionProgress"
                :format="progressFormat"
              />
              <p class="progress-text">
                已展开 {{ expandedVolumeCount }}/{{ volumeResults.length }} 个分卷
              </p>
            </div>

            <!-- Volume Cards with Expandable Chapters -->
            <div class="volume-chapters-list">
              <el-card
                v-for="(vol, idx) in volumeResults"
                :key="idx"
                class="volume-chapter-card"
                :class="{ 'volume-chapter-card--expanded': vol.chapters?.length > 0 }"
              >
                <template #header>
                  <div class="volume-chapter-header">
                    <div class="volume-chapter-title-row">
                      <el-tag :type="vol.chapters?.length > 0 ? 'success' : 'info'" size="small">
                        {{ vol.title }}
                      </el-tag>
                      <span class="volume-chapter-count">
                        {{ vol.chapters?.length || vol.chapter_count || 0 }}章
                      </span>
                    </div>
                    <div class="volume-chapter-actions">
                      <el-button
                        v-if="!vol.chapters || vol.chapters.length === 0"
                        size="small"
                        type="primary"
                        :loading="loading && selectedVolumeIndex === idx"
                        @click="selectedVolumeIndex = idx; expandVolume()"
                      >
                        展开章节
                      </el-button>
                      <el-button
                        v-else
                        size="small"
                        link
                        @click="toggleChapterList(idx)"
                      >
                        {{ expandedChapterVolumes.has(idx) ? '收起' : '展开' }}
                      </el-button>
                    </div>
                  </div>
                </template>

                <!-- Volume Synopsis -->
                <p class="volume-synopsis">{{ vol.synopsis }}</p>

                <!-- Chapters List (when expanded) -->
                <div v-if="vol.chapters && vol.chapters.length > 0 && expandedChapterVolumes.has(idx)" class="chapters-table">
                  <el-table :data="vol.chapters" size="small" max-height="400">
                    <el-table-column prop="index" label="#" width="50" />
                    <el-table-column prop="title" label="章节名" width="150" />
                    <el-table-column prop="synopsis" label="梗概" show-overflow-tooltip />
                    <el-table-column prop="main_event" label="核心事件" width="150" show-overflow-tooltip />
                  </el-table>
                </div>

                <!-- Empty state for unexpanded volume -->
                <div v-if="!vol.chapters || vol.chapters.length === 0" class="volume-empty">
                  <el-text type="info" size="small">点击"展开章节"生成章节列表</el-text>
                </div>
              </el-card>
            </div>
          </div>

          <!-- Step 4: Chapter Detail Content -->
          <div v-if="currentStep === 4" class="step-content">
            <div v-if="chapterDetail" class="result-section">
              <div class="result-header" style="margin-bottom: 12px;">
                <el-button size="small" type="success" @click="importToWriting" :loading="importing">
                  导入到写作
                </el-button>
              </div>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="目标字数">{{ chapterDetail.word_target }}</el-descriptions-item>
              </el-descriptions>

              <!-- Scenes -->
              <div v-if="chapterDetail.scenes && chapterDetail.scenes.length > 0" class="scenes-section">
                <h4>场景设计</h4>
                <el-timeline>
                  <el-timeline-item
                    v-for="(scene, idx) in chapterDetail.scenes"
                    :key="idx"
                    :timestamp="'场景' + (idx + 1)"
                  >
                    <el-card>
                      <p><strong>地点：</strong>{{ scene.location }}</p>
                      <p><strong>人物：</strong>{{ scene.characters }}</p>
                      <p><strong>事件：</strong>{{ scene.event }}</p>
                      <p><strong>氛围：</strong>{{ scene.mood }}</p>
                    </el-card>
                  </el-timeline-item>
                </el-timeline>
              </div>

              <!-- Dialogues -->
              <div v-if="chapterDetail.dialogues && chapterDetail.dialogues.length > 0">
                <h4>关键对话</h4>
                <ul class="detail-list">
                  <li v-for="(d, idx) in chapterDetail.dialogues" :key="idx">{{ d }}</li>
                </ul>
              </div>

              <!-- Foreshadows -->
              <div v-if="chapterDetail.foreshadows && chapterDetail.foreshadows.length > 0">
                <h4>伏笔设置</h4>
                <ul class="detail-list">
                  <li v-for="(f, idx) in chapterDetail.foreshadows" :key="idx">{{ f }}</li>
                </ul>
              </div>
            </div>

            <!-- Empty State -->
            <div v-if="!chapterDetail && !loading" class="empty-state">
              <p>请选择一个章节，然后点击"生成章节细纲"按钮</p>
            </div>

            <!-- All Chapter Details Summary -->
            <div v-if="Object.keys(chapterDetails).length > 0 && !chapterDetail" class="chapter-details-summary">
              <h4>已生成的章节细纲</h4>
              <el-table :data="chapterDetailsList" size="small">
                <el-table-column prop="volumeTitle" label="分卷" width="120" />
                <el-table-column prop="chapterTitle" label="章节" width="150" />
                <el-table-column prop="scenesCount" label="场景数" width="80">
                  <template #default="{ row }">{{ row.scenesCount || 0 }}</template>
                </el-table-column>
                <el-table-column label="操作" width="100">
                  <template #default="{ row }">
                    <el-button size="small" link type="primary" @click="viewChapterDetail(row.key)">
                      查看
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </template>
      </ArchitectWorkspace>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { architectApi, chapterApi } from '@/api'
import ArchitectStepCard from '@/components/ArchitectStepCard.vue'
import ArchitectWorkspace from '@/components/ArchitectWorkspace.vue'

const router = useRouter()
const route = useRoute()
const bookName = computed(() => route.params.id)

// Step definitions
const steps = [
  { title: '全书总纲', description: '故事梗概与核心设定' },
  { title: '世界观', description: '力量体系与社会结构' },
  { title: '分卷大纲', description: '各卷主要内容规划' },
  { title: '章节大纲', description: '逐章展开章节列表' },
  { title: '章节细纲', description: '场景与对话设计' }
]

// State management
const currentStep = ref(0)
const loading = ref(false)
const loadingText = ref('正在生成...')
const outdatedSteps = ref(new Set())
const selectedVolumeIndex = ref(-1)
const selectedChapterPath = ref(null)
const expandedChapterVolumes = ref(new Set()) // Track which volumes have chapters expanded

// Edit mode state
const editMode = ref({
  synopsis: false,
  worldview: false,
  volumes: {}
})

// Data state
const synopsisParams = ref({
  genre: '玄幻',
  theme: '',
  main_char: '',
  target_words: 1000000
})
const synopsisResult = ref(null)
const worldViewResult = ref(null)
const volumeResults = ref([])
const chapterDetail = ref(null)
const chapterDetails = ref({}) // key: "volIdx_chapIdx"
const importing = ref(false)

// Chapter options for cascader
const chapterOptions = computed(() => {
  return volumeResults.value.map((vol, vIdx) => ({
    id: `vol_${vIdx}`,
    title: vol.title,
    chapters: (vol.chapters || []).map((ch, cIdx) => ({
      id: `${vIdx}_${cIdx}`,
      title: `第${ch.index}章: ${ch.title}`,
      volumeIndex: vIdx,
      chapterIndex: cIdx,
      data: ch
    }))
  }))
})

// Expansion progress
const expandedVolumeCount = computed(() => {
  return volumeResults.value.filter(v => v.chapters && v.chapters.length > 0).length
})

const expansionProgress = computed(() => {
  if (volumeResults.value.length === 0) return 0
  return Math.round((expandedVolumeCount.value / volumeResults.value.length) * 100)
})

const progressFormat = (percentage) => {
  return `${expandedVolumeCount.value}/${volumeResults.value.length}`
}

// Chapter details list for table
const chapterDetailsList = computed(() => {
  return Object.entries(chapterDetails.value).map(([key, detail]) => {
    const [vIdx, cIdx] = key.split('_').map(Number)
    const vol = volumeResults.value[vIdx]
    const chapter = vol?.chapters?.[cIdx]
    return {
      key,
      volumeTitle: vol?.title || '',
      chapterTitle: chapter?.title || '',
      scenesCount: detail.scenes?.length || 0
    }
  })
})

// Navigation
const goBack = () => {
  router.push(`/books/${bookName.value}`)
}

// Step status logic
const getStepStatus = (index) => {
  if (loading.value && currentStep.value === index) return 'generating'
  if (currentStep.value === index) return 'current'
  if (outdatedSteps.value.has(index)) return 'outdated'

  // Check completion based on data
  switch (index) {
    case 0:
      return synopsisResult.value ? 'completed' : 'pending'
    case 1:
      return worldViewResult.value ? 'completed' : 'pending'
    case 2:
      return volumeResults.value.length > 0 ? 'completed' : 'pending'
    case 3:
      return expandedVolumeCount.value > 0 ? 'completed' : 'pending'
    case 4:
      return Object.keys(chapterDetails.value).length > 0 ? 'completed' : 'pending'
    default:
      return 'pending'
  }
}

// Step summaries for card display
const getStepSummary1 = (index) => {
  switch (index) {
    case 0:
      return synopsisResult.value?.title || ''
    case 1:
      return worldViewResult.value?.era || ''
    case 2:
      return volumeResults.value.length > 0 ? `${volumeResults.value.length}卷` : ''
    case 3:
      return expandedVolumeCount.value > 0 ? `${expandedVolumeCount.value}卷已展开` : ''
    case 4:
      return Object.keys(chapterDetails.value).length > 0
        ? `${Object.keys(chapterDetails.value).length}个细纲`
        : ''
    default:
      return ''
  }
}

const getStepSummary2 = (index) => {
  switch (index) {
    case 0:
      return synopsisResult.value?.genre || ''
    case 1:
      return worldViewResult.value?.power_system?.substring(0, 20) || ''
    case 2:
      return synopsisResult.value?.volume_count
        ? `每卷约${Math.round(synopsisResult.value.word_count / synopsisResult.value.volume_count)}字`
        : ''
    case 3:
      const totalChapters = volumeResults.value.reduce(
        (sum, v) => sum + (v.chapters?.length || 0), 0
      )
      return totalChapters > 0 ? `共${totalChapters}章` : ''
    case 4:
      return ''
    default:
      return ''
  }
}

// Context preview for each step
const getContextPreview = () => {
  switch (currentStep.value) {
    case 0:
      return ''
    case 1:
      if (!synopsisResult.value) return ''
      return `题材：${synopsisResult.value.genre} | 主题：${synopsisResult.value.theme}\n梗概：${synopsisResult.value.synopsis?.substring(0, 100)}...`
    case 2:
      if (!worldViewResult.value) return ''
      return `时代：${worldViewResult.value.era} | 力量体系：${worldViewResult.value.power_system?.substring(0, 30)}...`
    case 3:
      if (!synopsisResult.value || !worldViewResult.value) return ''
      return `总纲：${synopsisResult.value.title} | ${volumeResults.value.length}卷`
    case 4:
      const totalChapters = volumeResults.value.reduce(
        (sum, v) => sum + (v.chapters?.length || 0), 0
      )
      return `已展开：${expandedVolumeCount.value}/${volumeResults.value.length}卷，共${totalChapters}章`
    default:
      return ''
  }
}

// Outdated message
const getOutdatedMessage = () => {
  const previousSteps = []
  for (let i = 0; i < currentStep.value; i++) {
    if (outdatedSteps.value.has(i)) {
      previousSteps.push(steps[i].title)
    }
  }
  if (previousSteps.length > 0) {
    return `${previousSteps.join('、')}已重新生成，当前内容可能过时`
  }
  return '前置数据已更新，当前内容可能过时'
}

// Switch step
const switchStep = (index) => {
  currentStep.value = index
  // Clear chapter detail when switching away from step 4
  if (index !== 4) {
    chapterDetail.value = null
    selectedChapterPath.value = null
  }
  // Clear volume selection when switching away from step 3
  if (index !== 3) {
    selectedVolumeIndex.value = -1
  }
}

// Mark subsequent steps as outdated
const markOutdated = (fromStep) => {
  for (let i = fromStep + 1; i < 5; i++) {
    outdatedSteps.value.add(i)
  }
  saveArchitectData()
}

// Clear outdated for current step when regenerated
const clearOutdated = (step) => {
  outdatedSteps.value.delete(step)
}

// Handle regenerate button click
const handleRegenerate = () => {
  // Trigger regeneration based on current step
  switch (currentStep.value) {
    case 0:
      generateSynopsis()
      break
    case 1:
      generateWorldView()
      break
    case 2:
      generateVolumes()
      break
    case 3:
      if (selectedVolumeIndex.value >= 0) {
        expandVolume()
      } else {
        expandAllVolumes()
      }
      break
    case 4:
      expandChapter()
      break
  }
}

// Cancel generation
const cancelGeneration = () => {
  // Note: The actual API doesn't support cancellation,
  // this is just a UI state reset
  loading.value = false
  ElMessage.info('已取消（后台生成可能仍在继续）')
}

// Step 1: Generate synopsis
const generateSynopsis = async () => {
  loading.value = true
  loadingText.value = '正在生成全书总纲...'
  try {
    const res = await architectApi.generateSynopsis(synopsisParams.value)
    if (res.data?.synopsis) {
      synopsisResult.value = res.data
      currentStep.value = 1
      clearOutdated(0)
      markOutdated(0)
      await saveArchitectData()
      ElMessage.success('总纲生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 2: Generate worldview
const generateWorldView = async () => {
  if (!synopsisResult.value) {
    ElMessage.error('请先生成全书总纲')
    return
  }
  loading.value = true
  loadingText.value = '正在生成世界观设定...'
  try {
    const res = await architectApi.generateWorldView({
      book_name: bookName.value,
      genre: synopsisResult.value.genre,
      theme: synopsisResult.value.theme,
      synopsis: synopsisResult.value.synopsis
    })
    // 更宽松的检查条件：只要有返回数据就更新
    if (res.data && Object.keys(res.data).length > 0) {
      worldViewResult.value = res.data
      currentStep.value = 2
      clearOutdated(1)
      markOutdated(1)
      await saveArchitectData()
      ElMessage.success('世界观生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败，返回数据为空')
    }
  } catch (error) {
    console.error('generateWorldView error:', error)
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 3: Generate volumes
const generateVolumes = async () => {
  if (!synopsisResult.value || !synopsisResult.value.synopsis) {
    ElMessage.error('请先生成全书总纲')
    return
  }
  if (!worldViewResult.value || !worldViewResult.value.power_system) {
    ElMessage.error('请先生成世界观')
    return
  }

  loading.value = true
  loadingText.value = '正在生成分卷大纲...'
  try {
    const res = await architectApi.generateVolumes({
      book_name: bookName.value,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value
    })
    if (res.data?.volumes) {
      volumeResults.value = res.data.volumes
      currentStep.value = 3
      clearOutdated(2)
      markOutdated(2)
      await saveArchitectData()
      ElMessage.success('分卷大纲生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Expand single volume
const expandVolume = async () => {
  if (selectedVolumeIndex.value < 0) {
    ElMessage.warning('请选择要展开的分卷')
    return
  }
  if (!synopsisResult.value || !worldViewResult.value) {
    ElMessage.error('缺少前置数据')
    return
  }

  loading.value = true
  loadingText.value = `正在展开《${volumeResults.value[selectedVolumeIndex.value].title}》...`
  try {
    const vol = volumeResults.value[selectedVolumeIndex.value]
    console.log('expandVolume request:', {
      book_name: bookName.value,
      volume: vol,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value
    })
    const res = await architectApi.expandVolume({
      book_name: bookName.value,
      volume: vol,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value
    })
    console.log('expandVolume response:', res.data)
    // 检查返回数据，可能是 chapters 数组也可能是 Chapters（大写）
    const chapters = res.data?.chapters || res.data?.Chapters
    if (chapters && chapters.length > 0) {
      // 使用 Vue.set 或 splice 来确保响应式更新
      const newVolume = { ...vol, chapters: chapters }
      volumeResults.value.splice(selectedVolumeIndex.value, 1, newVolume)
      clearOutdated(3)
      await saveArchitectData()

      const allExpanded = volumeResults.value.every(v => v.chapters && v.chapters.length > 0)
      if (allExpanded) {
        currentStep.value = 4
        ElMessage.success('所有分卷已展开，进入章节细纲阶段')
      } else {
        ElMessage.success(`展开成功，共${chapters.length}章`)
      }
    } else {
      console.warn('expandVolume: no chapters returned', res.data)
      ElMessage.warning('展开失败：返回数据中没有章节')
    }
  } catch (error) {
    console.error('expandVolume error:', error)
    ElMessage.error('展开失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Expand all volumes
const expandAllVolumes = async () => {
  if (!synopsisResult.value || !worldViewResult.value) {
    ElMessage.error('缺少前置数据')
    return
  }

  loading.value = true
  loadingText.value = '正在展开全部分卷...'
  try {
    for (let i = 0; i < volumeResults.value.length; i++) {
      const vol = volumeResults.value[i]
      if (!vol.chapters || vol.chapters.length === 0) {
        loadingText.value = `正在展开《${vol.title}》...`
        const res = await architectApi.expandVolume({
          book_name: bookName.value,
          volume: vol,
          synopsis: synopsisResult.value,
          world_view: worldViewResult.value
        })
        const chapters = res.data?.chapters || res.data?.Chapters
        if (chapters && chapters.length > 0) {
          // 使用 splice 确保响应式更新
          const newVolume = { ...vol, chapters: chapters }
          volumeResults.value.splice(i, 1, newVolume)
        }
      }
    }
    currentStep.value = 4
    clearOutdated(3)
    await saveArchitectData()
    ElMessage.success('全部分卷展开完成，进入章节细纲阶段')
  } catch (error) {
    console.error('expandAllVolumes error:', error)
    ElMessage.error('展开失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 5: Expand chapter detail
const expandChapter = async () => {
  if (!selectedChapterPath.value || selectedChapterPath.value.length < 2) {
    ElMessage.warning('请选择章节')
    return
  }
  if (!worldViewResult.value) {
    ElMessage.error('缺少世界观数据')
    return
  }

  loading.value = true
  const [volId, chapId] = selectedChapterPath.value
  const vIdx = parseInt(volId.split('_')[1])
  const cIdx = parseInt(chapId.split('_')[1])
  const chapter = volumeResults.value[vIdx]?.chapters?.[cIdx]

  if (!chapter) {
    ElMessage.error('找不到章节')
    loading.value = false
    return
  }

  loadingText.value = `正在生成《${chapter.title}》细纲...`
  try {
    const res = await architectApi.expandChapter({
      book_name: bookName.value,
      chapter: chapter,
      world_view: worldViewResult.value
    })
    if (res.data?.scenes) {
      chapterDetail.value = res.data
      const key = `${vIdx}_${cIdx}`
      chapterDetails.value[key] = {
        chapter_key: key,
        ...res.data
      }
      clearOutdated(4)
      await saveArchitectData()
      ElMessage.success('细纲生成成功')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// View existing chapter detail
const viewChapterDetail = (key) => {
  chapterDetail.value = chapterDetails.value[key]
}

// Import chapter detail to writing page (as chapter outline)
const importToWriting = async () => {
  if (!chapterDetail.value || !selectedChapterPath.value || selectedChapterPath.value.length < 2) {
    ElMessage.warning('请先选择章节并生成细纲')
    return
  }

  importing.value = true
  try {
    const [volId, chapId] = selectedChapterPath.value
    const vIdx = parseInt(volId.split('_')[1])
    const cIdx = parseInt(chapId.split('_')[1])
    const chapter = volumeResults.value[vIdx]?.chapters?.[cIdx]

    if (!chapter || !chapter.id) {
      ElMessage.error('找不到章节信息')
      importing.value = false
      return
    }

    // 检查 chapter.id 是否是有效的数字 ID（不是字符串如 'chap_0_1'）
    if (typeof chapter.id === 'string' && chapter.id.startsWith('chap_')) {
      ElMessage.warning('请先点击"保存为章节"按钮，将大纲保存后再导入细纲')
      importing.value = false
      return
    }

    // Convert chapter detail to outline text
    const outlineText = generateOutlineFromDetail(chapterDetail.value, chapter)

    // Update chapter outline
    await chapterApi.update(bookName.value, chapter.id, {
      outline: outlineText
    })

    ElMessage.success('已导入到写作页面，可前往写作查看')
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.response?.data?.error || error.message))
  }
  importing.value = false
}

// Generate outline text from chapter detail
const generateOutlineFromDetail = (detail, chapter) => {
  let outline = `【章节：${chapter.title || '未命名'}】\n\n`

  if (detail.scenes && detail.scenes.length > 0) {
    outline += `【场景设计】\n`
    detail.scenes.forEach((scene, idx) => {
      outline += `场景${idx + 1}: ${scene.location}\n`
      outline += `  - 人物: ${scene.characters}\n`
      outline += `  - 事件: ${scene.event}\n`
      outline += `  - 氛围: ${scene.mood}\n\n`
    })
  }

  if (detail.dialogues && detail.dialogues.length > 0) {
    outline += `【关键对话】\n`
    detail.dialogues.forEach(d => {
      outline += `- ${d}\n`
    })
    outline += '\n'
  }

  if (detail.actions && detail.actions.length > 0) {
    outline += `【动作设计】\n`
    detail.actions.forEach(a => {
      outline += `- ${a}\n`
    })
    outline += '\n'
  }

  if (detail.foreshadows && detail.foreshadows.length > 0) {
    outline += `【伏笔设置】\n`
    detail.foreshadows.forEach(f => {
      outline += `- ${f}\n`
    })
    outline += '\n'
  }

  if (detail.emotions && detail.emotions.length > 0) {
    outline += `【情感变化】\n`
    detail.emotions.forEach(e => {
      outline += `- ${e}\n`
    })
    outline += '\n'
  }

  if (detail.word_target) {
    outline += `【目标字数】${detail.word_target}字\n`
  }

  return outline
}

// Save worldview to book settings
const saveWorldView = async () => {
  if (!worldViewResult.value) return
  if (!bookName.value) {
    ElMessage.error('请先选择一本书籍')
    return
  }
  try {
    await architectApi.saveWorldView({
      book_name: bookName.value,
      world_view: worldViewResult.value
    })
    ElMessage.success('世界观已保存到书籍设定')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

// Toggle volume edit mode
const toggleVolumeEdit = (idx) => {
  editMode.value.volumes[idx] = !editMode.value.volumes[idx]
}

// Toggle chapter list visibility for a volume
const toggleChapterList = (idx) => {
  if (expandedChapterVolumes.value.has(idx)) {
    expandedChapterVolumes.value.delete(idx)
  } else {
    expandedChapterVolumes.value.add(idx)
  }
}

// Save outline as chapters
const saveOutline = async () => {
  if (volumeResults.value.length === 0) return
  if (!bookName.value) {
    ElMessage.error('请先选择一本书籍')
    return
  }
  try {
    await architectApi.saveOutline({
      book_name: bookName.value,
      volumes: volumeResults.value
    })
    // 保存成功后重新加载章节数据，更新 volumeResults 中的 chapter.id 为实际 ID
    const chaptersRes = await chapterApi.list(bookName.value)
    const savedChapters = chaptersRes.data || []

    // 更新 volumeResults 中每个章节的 id 为数据库中的实际 ID
    let chapterIdx = 0
    for (let vIdx = 0; vIdx < volumeResults.value.length; vIdx++) {
      const vol = volumeResults.value[vIdx]
      if (vol.chapters && vol.chapters.length > 0) {
        for (let cIdx = 0; cIdx < vol.chapters.length; cIdx++) {
          if (chapterIdx < savedChapters.length) {
            vol.chapters[cIdx].id = savedChapters[chapterIdx].id
            chapterIdx++
          }
        }
      }
    }

    // 保存更新后的数据
    await saveArchitectData()
    ElMessage.success('大纲已保存为章节，现在可以导入细纲到写作页面')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

// Save architect data with persistence
const saveArchitectData = async () => {
  if (!bookName.value) return // 静默跳过，不提示
  try {
    await architectApi.saveData({
      book_name: bookName.value,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value,
      volumes: volumeResults.value,
      chapter_details: chapterDetails.value,
      current_step: currentStep.value,
      outdated_steps: Array.from(outdatedSteps.value)
    })
  } catch (error) {
    console.error('保存架构师数据失败:', error)
  }
}

// Load architect data from persistence
const loadArchitectData = async () => {
  if (!bookName.value) return // 静默跳过
  try {
    const res = await architectApi.loadData(bookName.value)
    if (res.data) {
      if (res.data.synopsis) {
        synopsisResult.value = res.data.synopsis
        synopsisParams.value.genre = res.data.synopsis.genre || '玄幻'
        synopsisParams.value.theme = res.data.synopsis.theme || ''
        synopsisParams.value.main_char = res.data.synopsis.main_chars?.[0] || ''
        synopsisParams.value.target_words = res.data.synopsis.word_count || 1000000
      }
      if (res.data.world_view) {
        worldViewResult.value = res.data.world_view
      }
      if (res.data.volumes && res.data.volumes.length > 0) {
        volumeResults.value = res.data.volumes
        // 同步真实的章节 ID（如果大纲已保存）
        syncChapterIds()
      }
      if (res.data.chapter_details) {
        chapterDetails.value = res.data.chapter_details
      }
      if (res.data.outdated_steps) {
        outdatedSteps.value = new Set(res.data.outdated_steps)
      }
      // Restore current step based on data
      restoreCurrentStep()
    }
  } catch (error) {
    console.error('加载架构师数据失败:', error)
  }
}

// 同步真实的章节 ID（从数据库加载）
const syncChapterIds = async () => {
  if (!bookName.value || volumeResults.value.length === 0) return
  try {
    const chaptersRes = await chapterApi.list(bookName.value)
    const savedChapters = chaptersRes.data || []

    if (savedChapters.length > 0) {
      // 更新 volumeResults 中每个章节的 id 为数据库中的实际 ID
      let chapterIdx = 0
      for (let vIdx = 0; vIdx < volumeResults.value.length; vIdx++) {
        const vol = volumeResults.value[vIdx]
        if (vol.chapters && vol.chapters.length > 0) {
          for (let cIdx = 0; cIdx < vol.chapters.length; cIdx++) {
            if (chapterIdx < savedChapters.length) {
              // 如果当前 id 是字符串（如 chap_0_1），替换为真实 ID
              if (typeof vol.chapters[cIdx].id === 'string' && vol.chapters[cIdx].id.startsWith('chap_')) {
                vol.chapters[cIdx].id = savedChapters[chapterIdx].id
              }
              chapterIdx++
            }
          }
        }
      }
    }
  } catch (error) {
    console.error('同步章节 ID 失败:', error)
  }
}

// Restore current step based on data completeness
const restoreCurrentStep = () => {
  // If there are outdated steps, keep the current step from saved data
  if (outdatedSteps.value.size > 0) {
    // Find the minimum outdated step as a reference
    const minOutdated = Math.min(...Array.from(outdatedSteps.value))
    // Set current step to the step just before the first outdated step
    // This ensures user sees what needs to be regenerated
    currentStep.value = Math.max(0, minOutdated - 1)
    return
  }

  // Otherwise, restore based on completion
  if (volumeResults.value.length > 0) {
    const hasChapters = volumeResults.value.some(v => v.chapters && v.chapters.length > 0)
    const allExpanded = volumeResults.value.every(v => v.chapters && v.chapters.length > 0)

    if (allExpanded) {
      currentStep.value = 4
    } else if (hasChapters) {
      currentStep.value = 3
    } else {
      currentStep.value = 2
    }
  } else if (worldViewResult.value) {
    currentStep.value = 2
  } else if (synopsisResult.value) {
    currentStep.value = 1
  } else {
    currentStep.value = 0
  }
}

// Auto-save on data changes
watch(
  [synopsisResult, worldViewResult, volumeResults, chapterDetails, outdatedSteps],
  () => {
    if (synopsisResult.value || worldViewResult.value || volumeResults.value.length > 0 || Object.keys(chapterDetails.value).length > 0) {
      saveArchitectData()
    }
  },
  { deep: true }
)

// Watch selectedChapterPath to auto-show existing chapter detail
watch(selectedChapterPath, (newPath) => {
  if (currentStep.value === 4 && newPath && newPath.length >= 2) {
    const [volId, chapId] = newPath
    const vIdx = parseInt(volId.split('_')[1])
    const cIdx = parseInt(chapId.split('_')[1])
    const key = `${vIdx}_${cIdx}`

    // Check if chapter detail already exists
    if (chapterDetails.value[key]) {
      chapterDetail.value = chapterDetails.value[key]
    } else {
      // Clear current detail if no existing detail
      chapterDetail.value = null
    }
  } else if (currentStep.value === 4 && !newPath) {
    // Clear detail when selection is cleared
    chapterDetail.value = null
  }
})

// Initialize on mount
onMounted(() => {
  // 检查 bookName 是否有效
  if (!bookName.value) {
    ElMessage.warning('请先从书籍列表选择一本书籍')
    router.push('/books')
    return
  }
  loadArchitectData()
})
</script>

<style scoped>
.architect-view {
  min-height: 100vh;
  background: #f5f5f7;
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 16px 24px;
  background: #ffffff;
  border-radius: 8px;
}

.back-btn {
  font-size: 14px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0;
}

.architect-layout {
  display: flex;
  gap: 20px;
  max-width: 1600px;
  margin: 0 auto;
}

.sidebar {
  width: 280px;
  flex-shrink: 0;
}

.workspace-container {
  flex: 1;
  min-width: 800px;
}

/* Step Content Styles */
.step-content {
  min-height: 400px;
}

.input-form {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.result-section {
  margin-top: 20px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.preview-mode {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
}

.synopsis-text {
  margin-top: 16px;
  padding: 12px;
  background: #ffffff;
  border-radius: 4px;
}

.synopsis-text p {
  margin: 8px 0 0 0;
  line-height: 1.6;
  color: rgba(0, 0, 0, 0.8);
}

.long-text {
  white-space: pre-wrap;
  max-height: 100px;
  overflow-y: auto;
}

.volume-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.volume-content {
  padding: 12px;
}

.chapters-list {
  margin-top: 12px;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: rgba(0, 0, 0, 0.48);
}

/* Expansion Progress */
.expansion-progress {
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.progress-text {
  margin: 8px 0 0 0;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.6);
  text-align: center;
}

/* Volume Cards */
.volume-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
}

.volume-card {
  border-radius: 8px;
}

.volume-card--expanded {
  border-left: 4px solid #67c23a;
}

.volume-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.volume-card-title {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
}

.volume-card-synopsis {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.6);
  margin: 0;
  line-height: 1.5;
}

/* Volume Chapter Cards */
.volume-chapters-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.volume-chapter-card {
  border-radius: 8px;
}

.volume-chapter-card--expanded {
  border-left: 4px solid #67c23a;
}

.volume-chapter-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.volume-chapter-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.volume-chapter-count {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.6);
}

.volume-chapter-actions {
  display: flex;
  gap: 8px;
}

.volume-synopsis {
  margin: 0 0 12px 0;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.8);
  line-height: 1.6;
}

.volume-empty {
  text-align: center;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.chapters-table {
  margin-top: 12px;
}

/* Scenes Section */
.scenes-section {
  margin-top: 20px;
}

.scenes-section h4 {
  margin: 0 0 16px 0;
  color: #1d1d1f;
}

.detail-list {
  padding-left: 20px;
  margin: 10px 0;
}

.detail-list li {
  margin: 6px 0;
  color: rgba(0, 0, 0, 0.8);
  line-height: 1.5;
}

/* Chapter Details Summary */
.chapter-details-summary {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.chapter-details-summary h4 {
  margin: 0 0 12px 0;
  color: #1d1d1f;
}

/* Responsive */
@media (max-width: 1200px) {
  .sidebar {
    width: 240px;
  }

  .workspace-container {
    min-width: 600px;
  }
}

@media (max-width: 1024px) {
  .architect-layout {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
  }

  .workspace-container {
    min-width: 100%;
  }
}
</style>