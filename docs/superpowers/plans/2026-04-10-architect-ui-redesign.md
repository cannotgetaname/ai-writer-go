# Architect UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor ArchitectView with left sidebar navigation cards and right workspace for better UX.

**Architecture:** Split into 3 components: main view (state management), StepCard (navigation), and Workspace (content). Use Vue 3 Composition API with reactive state. Maintain backward compatibility with existing data format.

**Tech Stack:** Vue 3, Element Plus, Pinia (implicit), existing API layer

---

## File Structure

```
web/src/
├── views/
│   └── ArchitectView.vue          # Main container, state management
├── components/
│   ├── ArchitectStepCard.vue      # Step navigation card (new)
│   └── ArchitectWorkspace.vue     # Workspace container (new)
└── api/
    └── index.js                   # API calls (existing, minor updates)
```

---

## Task 1: Create ArchitectStepCard Component

**Files:**
- Create: `web/src/components/ArchitectStepCard.vue`

- [ ] **Step 1: Create the step card component file**

```vue
<template>
  <div
    class="step-card"
    :class="{
      'step-card--active': isActive,
      'step-card--completed': status === 'completed',
      'step-card--outdated': status === 'outdated',
      'step-card--generating': status === 'generating'
    }"
    @click="$emit('click')"
  >
    <div class="step-card__header">
      <span class="step-card__number">{{ index + 1 }}</span>
      <span class="step-card__title">{{ title }}</span>
    </div>
    <div class="step-card__status">
      <el-tag :type="statusTagType" size="small">
        <el-icon v-if="status === 'generating'" class="is-loading"><Loading /></el-icon>
        <span v-else>{{ statusIcon }}</span>
        {{ statusText }}
      </el-tag>
    </div>
    <div v-if="summary1" class="step-card__summary">
      <span class="step-card__summary-line">{{ summary1 }}</span>
      <span v-if="summary2" class="step-card__summary-line step-card__summary-line--secondary">
        {{ summary2 }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Loading } from '@element-plus/icons-vue'

const props = defineProps({
  index: { type: Number, required: true },
  title: { type: String, required: true },
  status: { type: String, default: 'pending' }, // pending, current, completed, outdated, generating
  isActive: { type: Boolean, default: false },
  summary1: { type: String, default: '' },
  summary2: { type: String, default: '' }
})

defineEmits(['click'])

const statusIcon = computed(() => {
  const icons = {
    pending: '○',
    current: '●',
    completed: '✓',
    outdated: '⚠️'
  }
  return icons[props.status] || '○'
})

const statusText = computed(() => {
  const texts = {
    pending: '未开始',
    current: '当前',
    completed: '已完成',
    outdated: '可能过时',
    generating: '生成中'
  }
  return texts[props.status] || '未开始'
})

const statusTagType = computed(() => {
  const types = {
    pending: 'info',
    current: 'primary',
    completed: 'success',
    outdated: 'warning',
    generating: 'primary'
  }
  return types[props.status] || 'info'
})
</script>

<style scoped>
.step-card {
  padding: 16px;
  background: #ffffff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-left: 4px solid transparent;
  margin-bottom: 12px;
}

.step-card:hover {
  background: #f5f7fa;
}

.step-card--active {
  border-left-color: #0071e3;
  background: #f5f7fa;
}

.step-card--completed .step-card__number {
  background: #67c23a;
  color: white;
}

.step-card--outdated .step-card__number {
  background: #e6a23c;
  color: white;
}

.step-card--generating .step-card__number {
  background: #0071e3;
  color: white;
}

.step-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.step-card__number {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #909399;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
}

.step-card__title {
  font-size: 14px;
  font-weight: 600;
  color: #1d1d1f;
}

.step-card__status {
  margin-bottom: 8px;
}

.step-card__summary {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.step-card__summary-line {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.8);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.step-card__summary-line--secondary {
  color: rgba(0, 0, 0, 0.48);
}
</style>
```

- [ ] **Step 2: Commit the step card component**

```bash
git add web/src/components/ArchitectStepCard.vue
git commit -m "feat(architect): add step navigation card component

- Display step number, title, status
- Support pending/current/completed/outdated/generating states
- Show summary info for completed steps
- Click to switch step

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 2: Create ArchitectWorkspace Component

**Files:**
- Create: `web/src/components/ArchitectWorkspace.vue`

- [ ] **Step 1: Create the workspace container component**

```vue
<template>
  <div class="workspace">
    <!-- Header -->
    <div class="workspace__header">
      <h3 class="workspace__title">第{{ stepIndex + 1 }}步：{{ title }}</h3>
      <p class="workspace__description">{{ description }}</p>
    </div>

    <!-- Outdated Warning -->
    <el-alert
      v-if="isOutdated"
      type="warning"
      :closable="false"
      show-icon
      class="workspace__warning"
    >
      <template #title>
        {{ outdatedMessage }}
        <el-button type="primary" link size="small" @click="$emit('regenerate')">
          重新生成
        </el-button>
      </template>
    </el-alert>

    <!-- Context Preview (collapsible) -->
    <div v-if="contextPreview" class="workspace__context">
      <div class="workspace__context-header" @click="contextExpanded = !contextExpanded">
        <el-icon><ArrowRight :class="{ 'is-rotate': contextExpanded }" /></el-icon>
        <span>上下文预览</span>
      </div>
      <div v-show="contextExpanded" class="workspace__context-content">
        {{ contextPreview }}
      </div>
    </div>

    <!-- Action Buttons -->
    <div class="workspace__actions">
      <slot name="actions"></slot>
    </div>

    <!-- Content Area -->
    <div class="workspace__content">
      <slot></slot>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="workspace__loading">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <p>{{ loadingText }}</p>
      <el-button @click="$emit('cancel')">取消</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Loading, ArrowRight } from '@element-plus/icons-vue'

const props = defineProps({
  stepIndex: { type: Number, required: true },
  title: { type: String, required: true },
  description: { type: String, default: '' },
  contextPreview: { type: String, default: '' },
  isOutdated: { type: Boolean, default: false },
  outdatedMessage: { type: String, default: '前置数据已更新，当前内容可能过时' },
  loading: { type: Boolean, default: false },
  loadingText: { type: String, default: '正在生成...' }
})

defineEmits(['regenerate', 'cancel'])

const contextExpanded = ref(false)
</script>

<style scoped>
.workspace {
  background: #ffffff;
  border-radius: 8px;
  padding: 24px;
  min-height: calc(100vh - 140px);
  position: relative;
}

.workspace__header {
  margin-bottom: 20px;
}

.workspace__title {
  font-size: 20px;
  font-weight: 600;
  color: #1d1d1f;
  margin: 0 0 8px 0;
}

.workspace__description {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.8);
  margin: 0;
}

.workspace__warning {
  margin-bottom: 16px;
}

.workspace__context {
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}

.workspace__context-header {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.8);
}

.workspace__context-header:hover {
  background: #ebeef5;
}

.workspace__context-header .is-rotate {
  transform: rotate(90deg);
}

.workspace__context-content {
  padding: 0 16px 12px 32px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.6);
  line-height: 1.6;
}

.workspace__actions {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
}

.workspace__content {
  flex: 1;
}

.workspace__loading {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  z-index: 10;
}

.workspace__loading p {
  color: #0071e3;
  font-size: 14px;
}
</style>
```

- [ ] **Step 2: Commit the workspace component**

```bash
git add web/src/components/ArchitectWorkspace.vue
git commit -m "feat(architect): add workspace container component

- Header with title and description
- Collapsible context preview
- Outdated warning banner
- Loading overlay with cancel button
- Slot for actions and content

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 3: Refactor ArchitectView - Layout Structure

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Replace the entire template with new layout**

Replace the `<template>` section with:

```vue
<template>
  <div class="architect-view">
    <!-- Page Header -->
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 架构师</h2>
      <el-tag v-if="hasOutdatedSteps" type="warning">⚠️ 有数据待更新</el-tag>
      <el-tag v-else type="info">分形写作</el-tag>
    </div>

    <div class="architect-body">
      <!-- Left Sidebar: Step Navigation -->
      <div class="step-nav">
        <ArchitectStepCard
          v-for="(step, idx) in steps"
          :key="idx"
          :index="idx"
          :title="step.title"
          :status="getStepStatus(idx)"
          :is-active="currentStep === idx"
          :summary1="getStepSummary1(idx)"
          :summary2="getStepSummary2(idx)"
          @click="switchStep(idx)"
        />
      </div>

      <!-- Right Workspace -->
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
        @cancel="handleCancel"
      >
        <!-- Step-specific content -->
        <template #actions>
          <slot name="actions"></slot>
        </template>

        <!-- Step Content -->
        <component :is="currentStepComponent" />
      </ArchitectWorkspace>
    </div>
  </div>
</template>
```

- [ ] **Step 2: Update script setup with new state management**

Replace the `<script setup>` section:

```vue
<script setup>
import { ref, computed, onMounted, watch, markRaw } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { architectApi } from '@/api'
import ArchitectStepCard from '@/components/ArchitectStepCard.vue'
import ArchitectWorkspace from '@/components/ArchitectWorkspace.vue'

// Step definitions
const steps = [
  { title: '全书总纲', description: '生成故事梗概、题材、主题等基础设定' },
  { title: '世界观', description: '基于总纲生成世界设定、力量体系、社会结构' },
  { title: '分卷大纲', description: '设计各卷主要内容、情感弧线' },
  { title: '章节大纲', description: '展开各分卷的章节内容' },
  { title: '章节细纲', description: '生成场景、对话、伏笔等详细设计' }
]

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

// Core state
const currentStep = ref(0)
const loading = ref(false)
const loadingText = ref('')

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
const chapterDetails = ref({})

// Outdated tracking
const outdatedSteps = ref(new Set())

// Edit modes
const editMode = ref({
  synopsis: false,
  worldview: false,
  volumes: {}
})

// Step status computation
const hasOutdatedSteps = computed(() => outdatedSteps.value.size > 0)

function getStepStatus(stepIndex) {
  if (loading.value && currentStep.value === stepIndex) return 'generating'
  if (outdatedSteps.value.has(stepIndex)) return 'outdated'
  if (currentStep.value === stepIndex) return 'current'
  if (hasStepData(stepIndex)) return 'completed'
  return 'pending'
}

function hasStepData(stepIndex) {
  switch (stepIndex) {
    case 0: return synopsisResult.value !== null
    case 1: return worldViewResult.value !== null
    case 2: return volumeResults.value.length > 0
    case 3: return volumeResults.value.some(v => v.chapters?.length > 0)
    case 4: return Object.keys(chapterDetails.value).length > 0
    default: return false
  }
}

function getStepSummary1(stepIndex) {
  if (!hasStepData(stepIndex)) return ''
  switch (stepIndex) {
    case 0: return synopsisResult.value?.title || ''
    case 1: return worldViewResult.value?.era || ''
    case 2: return `共 ${volumeResults.value.length} 卷`
    case 3:
      const expandedCount = volumeResults.value.filter(v => v.chapters?.length > 0).length
      return `已展开 ${expandedCount}/${volumeResults.value.length} 卷`
    case 4: return `已生成 ${Object.keys(chapterDetails.value).length} 章`
    default: return ''
  }
}

function getStepSummary2(stepIndex) {
  if (!hasStepData(stepIndex)) return ''
  switch (stepIndex) {
    case 0:
      const wordCount = synopsisResult.value?.word_count
      return `${synopsisResult.value?.genre || ''} · ${wordCount ? (wordCount / 10000).toFixed(0) + '万字' : ''}`
    case 1:
      const power = worldViewResult.value?.power_system || ''
      return power.length > 20 ? power.substring(0, 20) + '...' : power
    case 2:
      const expandedCount = volumeResults.value.filter(v => v.chapters?.length > 0).length
      return expandedCount > 0 ? `已展开 ${expandedCount} 卷` : ''
    case 3:
      const totalChapters = volumeResults.value.reduce((sum, v) => sum + (v.chapters?.length || 0), 0)
      return `共 ${totalChapters} 章`
    default: return ''
  }
}

function getContextPreview() {
  if (currentStep.value === 0) return ''
  if (currentStep.value === 1 && synopsisResult.value) {
    return `总纲: ${synopsisResult.value.synopsis?.substring(0, 100)}...`
  }
  if (currentStep.value === 2 && worldViewResult.value) {
    return `世界观: ${worldViewResult.value.power_system?.substring(0, 100)}...`
  }
  if (currentStep.value === 3 && synopsisResult.value) {
    return `总纲: ${synopsisResult.value.synopsis?.substring(0, 50)}... | 共 ${volumeResults.value.length} 卷`
  }
  if (currentStep.value === 4 && volumeResults.value.length > 0) {
    const expanded = volumeResults.value.filter(v => v.chapters?.length > 0).length
    return `已展开 ${expanded}/${volumeResults.value.length} 卷`
  }
  return ''
}

function getOutdatedMessage() {
  if (!outdatedSteps.value.has(currentStep.value)) return ''
  const stepNames = ['总纲', '世界观', '分卷大纲', '章节大纲', '章节细纲']
  // Find which earlier step was modified
  for (let i = 0; i < currentStep.value; i++) {
    // The message should indicate what changed
  }
  return `${steps[currentStep.value].title}可能过时，建议重新生成`
}

// Step switching
function switchStep(stepIndex) {
  currentStep.value = stepIndex
}

function handleRegenerate() {
  // Trigger regenerate for current step
  executeCurrentStep()
}

function handleCancel() {
  loading.value = false
}

function goBack() {
  router.push(`/books/${bookId.value}`)
}

// Mark subsequent steps as outdated
function markSubsequentOutdated(stepIndex) {
  for (let i = stepIndex + 1; i < 5; i++) {
    outdatedSteps.value.add(i)
  }
}

// Dynamic component for current step
const currentStepComponent = computed(() => {
  // Will be implemented in next tasks
  return null
})

// Data persistence
async function saveArchitectData() {
  try {
    await architectApi.saveData({
      book_name: bookId.value,
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

async function loadArchitectData() {
  try {
    const res = await architectApi.loadData(bookId.value)
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
      if (res.data.volumes?.length > 0) {
        volumeResults.value = res.data.volumes
      }
      if (res.data.chapter_details) {
        chapterDetails.value = res.data.chapter_details
      }
      if (res.data.outdated_steps) {
        outdatedSteps.value = new Set(res.data.outdated_steps)
      }
      if (res.data.current_step !== undefined) {
        currentStep.value = res.data.current_step
      }
    }
  } catch (error) {
    console.error('加载架构师数据失败:', error)
  }
}

// Watch for data changes
watch([synopsisResult, worldViewResult, volumeResults, chapterDetails], () => {
  if (synopsisResult.value || worldViewResult.value || volumeResults.value.length > 0) {
    saveArchitectData()
  }
}, { deep: true })

onMounted(() => {
  loadArchitectData()
})

// Placeholder for step execution - will be implemented in next tasks
function executeCurrentStep() {
  // To be implemented
}
</script>
```

- [ ] **Step 3: Update styles for new layout**

Replace the `<style scoped>` section:

```vue
<style scoped>
.architect-view {
  max-width: 1600px;
  margin: 0 auto;
  padding: 24px;
  background: #f5f5f7;
  min-height: 100vh;
  box-sizing: border-box;
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

.page-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1d1d1f;
}

.architect-body {
  display: flex;
  gap: 24px;
}

.step-nav {
  width: 320px;
  flex-shrink: 0;
}

/* Responsive */
@media (max-width: 1200px) {
  .step-nav {
    width: 280px;
  }
}

@media (max-width: 900px) {
  .architect-body {
    flex-direction: column;
  }
  .step-nav {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
  .step-nav > * {
    flex: 1;
    min-width: 200px;
  }
}
</style>
```

- [ ] **Step 4: Commit the layout refactor**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "refactor(architect): implement new layout structure

- Left sidebar with step navigation cards
- Right workspace with header and context preview
- Outdated steps tracking
- Responsive layout for smaller screens

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 4: Implement Step 1 (Synopsis) Content

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Add synopsis form and result components in template**

Add inside `<ArchitectWorkspace>` after the actions slot, as a conditional section:

```vue
        <!-- Step 1: Synopsis -->
        <div v-if="currentStep === 0" class="step-content">
          <el-form v-if="!synopsisResult" :model="synopsisParams" label-width="100px" class="step-form">
            <el-form-item label="题材类型">
              <el-select v-model="synopsisParams.genre" style="width: 100%;">
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
              <el-input-number v-model="synopsisParams.target_words" :min="100000" :max="5000000" :step="100000" style="width: 100%;" />
            </el-form-item>
          </el-form>

          <div class="step-actions">
            <el-button v-if="!synopsisResult" type="primary" @click="generateSynopsis" :loading="loading">
              生成全书总纲
            </el-button>
            <el-button v-if="synopsisResult" type="primary" @click="regenerateSynopsis" :loading="loading">
              重新生成
            </el-button>
          </div>

          <!-- Synopsis Result -->
          <div v-if="synopsisResult" class="result-section">
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
              <el-form-item label="支线剧情">
                <div v-for="(plot, idx) in synopsisResult.sub_plots" :key="idx" class="list-item">
                  <el-input v-model="synopsisResult.sub_plots[idx]" style="flex: 1;" />
                  <el-button type="danger" size="small" @click="synopsisResult.sub_plots.splice(idx, 1)">删除</el-button>
                </div>
                <el-button size="small" @click="synopsisResult.sub_plots.push('')">添加支线</el-button>
              </el-form-item>
            </el-form>
            <div class="result-actions">
              <el-button type="success" @click="saveArchitectData">保存</el-button>
            </div>
          </div>
        </div>
```

- [ ] **Step 2: Add synopsis generation function**

Add to script setup:

```javascript
// Step 1: Generate Synopsis
async function generateSynopsis() {
  loading.value = true
  loadingText.value = '正在生成全书总纲...'
  try {
    const res = await architectApi.generateSynopsis(synopsisParams.value)
    if (res.data?.synopsis) {
      synopsisResult.value = res.data
      currentStep.value = 1
      outdatedSteps.value.clear() // New project, clear outdated
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

async function regenerateSynopsis() {
  await generateSynopsis()
  markSubsequentOutdated(0)
}
```

- [ ] **Step 3: Add styles for step content**

Add to style section:

```css
.step-content {
  min-height: 400px;
}

.step-form {
  max-width: 600px;
  margin-bottom: 20px;
}

.step-actions {
  margin-bottom: 24px;
}

.result-section {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 20px;
}

.result-actions {
  margin-top: 20px;
  display: flex;
  gap: 12px;
}

.list-item {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}
```

- [ ] **Step 4: Commit step 1 implementation**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "feat(architect): implement step 1 synopsis form and result

- Input form for genre, theme, main_char, target_words
- Editable result display with all synopsis fields
- Support adding/removing sub-plots
- Regenerate with outdated marking

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 5: Implement Step 2 (WorldView) Content

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Add worldview content section**

Add after the synopsis section:

```vue
        <!-- Step 2: WorldView -->
        <div v-if="currentStep === 1" class="step-content">
          <div v-if="!worldViewResult" class="step-actions">
            <el-alert type="info" :closable="false" style="margin-bottom: 16px;">
              将基于总纲《{{ synopsisResult?.title }}》生成世界观设定
            </el-alert>
            <el-button type="primary" @click="generateWorldView" :loading="loading">
              生成世界观设定
            </el-button>
          </div>

          <div v-else class="result-section">
            <el-form :model="worldViewResult" label-width="100px">
              <el-form-item label="时代背景">
                <el-input v-model="worldViewResult.era" />
              </el-form-item>
              <el-form-item label="科技水平">
                <el-input v-model="worldViewResult.tech_level" />
              </el-form-item>
              <el-form-item label="力量体系">
                <el-input v-model="worldViewResult.power_system" type="textarea" :rows="4" />
              </el-form-item>
              <el-form-item label="社会结构">
                <el-input v-model="worldViewResult.social_structure" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item label="特殊规则">
                <el-input v-model="worldViewResult.special_rules" type="textarea" :rows="3" />
              </el-form-item>
              <el-form-item label="重要物品">
                <el-input v-model="worldViewResult.important_items" type="textarea" :rows="2" />
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
              <el-form-item label="发展趋势">
                <el-input v-model="worldViewResult.development" type="textarea" :rows="2" />
              </el-form-item>
            </el-form>
            <div class="result-actions">
              <el-button type="success" @click="saveWorldViewToBook">保存世界观</el-button>
              <el-button @click="regenerateWorldView" :loading="loading">重新生成</el-button>
            </div>
          </div>
        </div>
```

- [ ] **Step 2: Add worldview generation functions**

Add to script setup:

```javascript
// Step 2: Generate WorldView
async function generateWorldView() {
  if (!synopsisResult.value) {
    ElMessage.error('请先生成全书总纲')
    return
  }
  loading.value = true
  loadingText.value = '正在生成世界观设定...'
  try {
    const res = await architectApi.generateWorldView({
      book_name: bookId.value,
      genre: synopsisResult.value.genre,
      theme: synopsisResult.value.theme,
      synopsis: synopsisResult.value.synopsis
    })
    if (res.data?.power_system) {
      worldViewResult.value = res.data
      currentStep.value = 2
      await saveArchitectData()
      ElMessage.success('世界观生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

async function regenerateWorldView() {
  await generateWorldView()
  markSubsequentOutdated(1)
}

async function saveWorldViewToBook() {
  try {
    await architectApi.saveWorldView({
      book_name: bookId.value,
      world_view: worldViewResult.value
    })
    await saveArchitectData()
    ElMessage.success('世界观已保存')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}
```

- [ ] **Step 3: Commit step 2 implementation**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "feat(architect): implement step 2 worldview form and result

- All worldview fields editable as form inputs
- Save to book worldview setting
- Regenerate with outdated marking for subsequent steps

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 6: Implement Step 3 (Volumes) Content

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Add volumes content section**

Add after the worldview section:

```vue
        <!-- Step 3: Volumes -->
        <div v-if="currentStep === 2" class="step-content">
          <div v-if="volumeResults.length === 0" class="step-actions">
            <el-form label-width="100px">
              <el-form-item label="分卷数量">
                <el-input-number v-model="synopsisResult.volume_count" :min="1" :max="20" />
              </el-form-item>
            </el-form>
            <el-button type="primary" @click="generateVolumes" :loading="loading">
              生成分卷大纲
            </el-button>
          </div>

          <div v-else class="result-section">
            <el-collapse v-model="expandedVolumes">
              <el-collapse-item
                v-for="(vol, idx) in volumeResults"
                :key="idx"
                :name="idx"
              >
                <template #title>
                  <div class="volume-title">
                    <el-tag>{{ vol.title }}</el-tag>
                    <span class="volume-chapters">{{ vol.chapter_count || 0 }}章</span>
                  </div>
                </template>
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
              </el-collapse-item>
            </el-collapse>
            <div class="result-actions">
              <el-button type="success" @click="saveArchitectData">保存大纲</el-button>
              <el-button @click="regenerateVolumes" :loading="loading">重新生成</el-button>
            </div>
          </div>
        </div>
```

- [ ] **Step 2: Add volume-related state and functions**

Add to script setup:

```javascript
// Step 3: Volume management
const expandedVolumes = ref([])

async function generateVolumes() {
  if (!synopsisResult.value || !worldViewResult.value) {
    ElMessage.error('请先完成总纲和世界观')
    return
  }
  loading.value = true
  loadingText.value = '正在生成分卷大纲...'
  try {
    const res = await architectApi.generateVolumes({
      book_name: bookId.value,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value
    })
    if (res.data?.volumes) {
      volumeResults.value = res.data.volumes
      currentStep.value = 3
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

async function regenerateVolumes() {
  volumeResults.value = []
  await generateVolumes()
  markSubsequentOutdated(2)
}
```

- [ ] **Step 3: Add volume styles**

Add to style section:

```css
.volume-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.volume-chapters {
  color: rgba(0, 0, 0, 0.48);
  font-size: 12px;
}
```

- [ ] **Step 4: Commit step 3 implementation**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "feat(architect): implement step 3 volumes list

- Collapsible volume cards
- Editable volume details
- Volume count configuration
- Regenerate with outdated marking

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 7: Implement Step 4 (Chapters) Content

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Add chapters content section**

Add after the volumes section:

```vue
        <!-- Step 4: Chapters -->
        <div v-if="currentStep === 3" class="step-content">
          <div class="chapters-header">
            <el-text>选择要展开的分卷：</el-text>
            <el-select v-model="selectedVolumeIndex" placeholder="选择分卷" style="width: 200px;">
              <el-option
                v-for="(vol, idx) in volumeResults"
                :key="idx"
                :label="vol.title"
                :value="idx"
              />
            </el-select>
          </div>
          <div class="step-actions">
            <el-button
              type="primary"
              @click="expandSelectedVolume"
              :loading="loading"
              :disabled="selectedVolumeIndex < 0"
            >
              展开选中分卷
            </el-button>
            <el-button type="success" @click="expandAllVolumes" :loading="loading">
              展开全部
            </el-button>
          </div>

          <!-- Expanded Volumes List -->
          <div class="volumes-list">
            <div
              v-for="(vol, vIdx) in volumeResults"
              :key="vIdx"
              class="volume-card"
              :class="{ 'volume-card--expanded': vol.chapters?.length > 0 }"
            >
              <div class="volume-card__header" @click="toggleVolumeExpand(vIdx)">
                <el-icon><ArrowRight :class="{ 'is-rotate': expandedVolumeCards[vIdx] }" /></el-icon>
                <span class="volume-card__title">{{ vol.title }}</span>
                <el-tag size="small" :type="vol.chapters?.length > 0 ? 'success' : 'info'">
                  {{ vol.chapters?.length || 0 }} 章
                </el-tag>
                <el-button
                  v-if="!vol.chapters || vol.chapters.length === 0"
                  size="small"
                  type="primary"
                  link
                  @click.stop="expandVolume(vIdx)"
                  :loading="expandingVolume === vIdx"
                >
                  展开
                </el-button>
              </div>
              <div v-if="expandedVolumeCards[vIdx] && vol.chapters?.length > 0" class="volume-card__content">
                <el-table :data="vol.chapters" size="small" max-height="400">
                  <el-table-column prop="index" label="#" width="50" />
                  <el-table-column prop="title" label="章节名">
                    <template #default="{ row }">
                      <el-input v-model="row.title" size="small" />
                    </template>
                  </el-table-column>
                  <el-table-column prop="synopsis" label="梗概">
                    <template #default="{ row }">
                      <el-input v-model="row.synopsis" size="small" />
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
          </div>

          <div class="result-actions">
            <el-button type="success" @click="saveOutline">保存为章节</el-button>
          </div>
        </div>
```

- [ ] **Step 2: Add chapter-related state and functions**

Add to script setup:

```javascript
// Step 4: Chapter expansion
const selectedVolumeIndex = ref(-1)
const expandedVolumeCards = ref({})
const expandingVolume = ref(-1)

function toggleVolumeExpand(idx) {
  expandedVolumeCards.value[idx] = !expandedVolumeCards.value[idx]
}

async function expandSelectedVolume() {
  if (selectedVolumeIndex.value < 0) {
    ElMessage.warning('请选择要展开的分卷')
    return
  }
  await expandVolume(selectedVolumeIndex.value)
}

async function expandVolume(vIdx) {
  if (!synopsisResult.value || !worldViewResult.value) {
    ElMessage.error('缺少前置数据')
    return
  }
  expandingVolume.value = vIdx
  loading.value = true
  loadingText.value = `正在展开 ${volumeResults.value[vIdx].title}...`
  try {
    const vol = volumeResults.value[vIdx]
    const res = await architectApi.expandVolume({
      book_name: bookId.value,
      volume: vol,
      synopsis: synopsisResult.value,
      world_view: worldViewResult.value
    })
    if (res.data?.chapters) {
      volumeResults.value[vIdx].chapters = res.data.chapters
      expandedVolumeCards.value[vIdx] = true
      await saveArchitectData()
      // Check if all volumes expanded
      const allExpanded = volumeResults.value.every(v => v.chapters?.length > 0)
      if (allExpanded) {
        currentStep.value = 4
        ElMessage.success('所有分卷已展开，进入章节细纲阶段')
      } else {
        ElMessage.success('章节展开成功')
      }
    }
  } catch (error) {
    ElMessage.error('展开失败: ' + (error.response?.data?.error || error.message))
  }
  expandingVolume.value = -1
  loading.value = false
}

async function expandAllVolumes() {
  for (let i = 0; i < volumeResults.value.length; i++) {
    if (!volumeResults.value[i].chapters || volumeResults.value[i].chapters.length === 0) {
      await expandVolume(i)
    }
  }
}

async function saveOutline() {
  try {
    await architectApi.saveOutline({
      book_name: bookId.value,
      volumes: volumeResults.value
    })
    ElMessage.success('大纲已保存为章节')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}
```

- [ ] **Step 3: Add chapter styles**

Add to style section:

```css
.chapters-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.volumes-list {
  margin-top: 20px;
}

.volume-card {
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}

.volume-card--expanded {
  background: #ffffff;
  border: 1px solid #e4e7ed;
}

.volume-card__header {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.volume-card__header:hover {
  background: #ebeef5;
}

.volume-card__title {
  flex: 1;
  font-weight: 500;
}

.volume-card__content {
  padding: 16px;
  border-top: 1px solid #e4e7ed;
}
```

- [ ] **Step 4: Commit step 4 implementation**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "feat(architect): implement step 4 chapter expansion

- Volume selector for individual expansion
- Expand all volumes at once
- Collapsible volume cards with chapter tables
- Inline editing for chapter title and synopsis
- Save outline to book chapters

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 8: Implement Step 5 (Chapter Detail) Content

**Files:**
- Modify: `web/src/views/ArchitectView.vue`

- [ ] **Step 1: Add chapter detail content section**

Add after the chapters section:

```vue
        <!-- Step 5: Chapter Detail -->
        <div v-if="currentStep === 4" class="step-content">
          <div class="chapters-header">
            <el-text>选择要细化的章节：</el-text>
            <el-cascader
              v-model="selectedChapterPath"
              :options="chapterOptions"
              :props="{ value: 'id', label: 'title', children: 'chapters' }"
              placeholder="选择章节"
              clearable
              style="width: 300px;"
            />
          </div>
          <div class="step-actions">
            <el-button
              type="primary"
              @click="expandChapterDetail"
              :loading="loading"
              :disabled="!selectedChapterPath"
            >
              生成章节细纲
            </el-button>
          </div>

          <!-- Chapter Detail Result -->
          <div v-if="chapterDetail" class="result-section">
            <el-form label-width="100px">
              <el-form-item label="目标字数">
                <el-input-number v-model="chapterDetail.word_target" :min="500" :max="10000" :step="500" />
              </el-form-item>
            </el-form>

            <!-- Scenes -->
            <div v-if="chapterDetail.scenes?.length > 0" class="detail-section">
              <h4>场景设计</h4>
              <el-timeline>
                <el-timeline-item
                  v-for="(scene, idx) in chapterDetail.scenes"
                  :key="idx"
                  :timestamp="'场景' + (idx + 1)"
                >
                  <div class="scene-card">
                    <el-form label-width="80px" size="small">
                      <el-form-item label="地点">
                        <el-input v-model="scene.location" />
                      </el-form-item>
                      <el-form-item label="人物">
                        <el-input v-model="scene.characters" />
                      </el-form-item>
                      <el-form-item label="事件">
                        <el-input v-model="scene.event" />
                      </el-form-item>
                      <el-form-item label="氛围">
                        <el-input v-model="scene.mood" />
                      </el-form-item>
                    </el-form>
                  </div>
                </el-timeline-item>
              </el-timeline>
            </div>

            <!-- Dialogues -->
            <div v-if="chapterDetail.dialogues?.length > 0" class="detail-section">
              <h4>关键对话</h4>
              <div v-for="(d, idx) in chapterDetail.dialogues" :key="idx" class="list-item">
                <el-input v-model="chapterDetail.dialogues[idx]" style="flex: 1;" />
                <el-button type="danger" size="small" @click="chapterDetail.dialogues.splice(idx, 1)">删除</el-button>
              </div>
              <el-button size="small" @click="chapterDetail.dialogues.push('')">添加对话</el-button>
            </div>

            <!-- Foreshadows -->
            <div v-if="chapterDetail.foreshadows?.length > 0" class="detail-section">
              <h4>伏笔设置</h4>
              <div v-for="(f, idx) in chapterDetail.foreshadows" :key="idx" class="list-item">
                <el-input v-model="chapterDetail.foreshadows[idx]" style="flex: 1;" />
                <el-button type="danger" size="small" @click="chapterDetail.foreshadows.splice(idx, 1)">删除</el-button>
              </div>
              <el-button size="small" @click="chapterDetail.foreshadows.push('')">添加伏笔</el-button>
            </div>

            <div class="result-actions">
              <el-button type="success" @click="saveChapterDetail">保存细纲</el-button>
            </div>
          </div>

          <!-- Saved Chapter Details -->
          <div v-if="Object.keys(chapterDetails).length > 0" class="saved-details">
            <h4>已生成的章节细纲</h4>
            <el-tag
              v-for="(detail, key) in chapterDetails"
              :key="key"
              class="detail-tag"
              @click="loadChapterDetail(key)"
            >
              {{ detail.chapterTitle || key }}
            </el-tag>
          </div>
        </div>
```

- [ ] **Step 2: Add chapter detail state and functions**

Add to script setup:

```javascript
// Step 5: Chapter detail
const selectedChapterPath = ref(null)
const chapterDetail = ref(null)

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

async function expandChapterDetail() {
  if (!selectedChapterPath.value || selectedChapterPath.value.length < 2) {
    ElMessage.warning('请选择章节')
    return
  }
  const [volId, chapId] = selectedChapterPath.value
  const vIdx = parseInt(volId.split('_')[1])
  const cIdx = parseInt(chapId.split('_')[1])
  const chapter = volumeResults.value[vIdx]?.chapters?.[cIdx]

  if (!chapter) {
    ElMessage.error('找不到章节')
    return
  }

  loading.value = true
  loadingText.value = `正在生成 ${chapter.title} 的细纲...`
  try {
    const res = await architectApi.expandChapter({
      book_name: bookId.value,
      chapter: chapter,
      world_view: worldViewResult.value
    })
    if (res.data?.scenes) {
      chapterDetail.value = {
        ...res.data,
        chapterTitle: chapter.title,
        chapterKey: `${vIdx}_${cIdx}`
      }
      chapterDetails.value[`${vIdx}_${cIdx}`] = chapterDetail.value
      await saveArchitectData()
      ElMessage.success('细纲生成成功')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

function loadChapterDetail(key) {
  chapterDetail.value = chapterDetails.value[key]
}

async function saveChapterDetail() {
  if (!chapterDetail.value) return
  const key = chapterDetail.value.chapterKey
  if (key) {
    chapterDetails.value[key] = chapterDetail.value
  }
  await saveArchitectData()
  ElMessage.success('细纲已保存')
}
```

- [ ] **Step 3: Add chapter detail styles**

Add to style section:

```css
.detail-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  color: #1d1d1f;
}

.scene-card {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 8px;
}

.saved-details {
  margin-top: 24px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.saved-details h4 {
  margin: 0 0 12px 0;
}

.detail-tag {
  margin-right: 8px;
  margin-bottom: 8px;
  cursor: pointer;
}

.detail-tag:hover {
  opacity: 0.8;
}
```

- [ ] **Step 4: Commit step 5 implementation**

```bash
git add web/src/views/ArchitectView.vue
git commit -m "feat(architect): implement step 5 chapter detail

- Cascader for chapter selection
- Scene timeline with editable fields
- Dialogues and foreshadows lists
- Save and load chapter details
- Display saved details as tags

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Task 9: Build and Test

**Files:**
- None (build verification)

- [ ] **Step 1: Build the frontend**

```bash
cd /home/zcz/program/ai-writer-go/web && npm run build
```

Expected: Build succeeds without errors.

- [ ] **Step 2: Verify the build output**

```bash
ls -la /home/zcz/program/ai-writer-go/web/dist/
```

Expected: `index.html` and `assets/` directory exist.

- [ ] **Step 3: Final commit**

```bash
cd /home/zcz/program/ai-writer-go && git add -A && git commit -m "chore: build frontend after architect UI refactor

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## Summary

| Task | Description | Key Files |
|------|-------------|-----------|
| 1 | Create StepCard component | `ArchitectStepCard.vue` |
| 2 | Create Workspace component | `ArchitectWorkspace.vue` |
| 3 | Refactor main layout | `ArchitectView.vue` |
| 4 | Step 1: Synopsis | `ArchitectView.vue` |
| 5 | Step 2: WorldView | `ArchitectView.vue` |
| 6 | Step 3: Volumes | `ArchitectView.vue` |
| 7 | Step 4: Chapters | `ArchitectView.vue` |
| 8 | Step 5: Chapter Detail | `ArchitectView.vue` |
| 9 | Build and verify | - |