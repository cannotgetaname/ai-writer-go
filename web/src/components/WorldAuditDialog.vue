<template>
  <el-dialog
    v-model="visible"
    title="审计世界状态"
    width="90%"
    top="5vh"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <!-- Loading state -->
    <div v-if="loading" class="audit-loading">
      <el-icon class="is-loading" :size="48"><Loading /></el-icon>
      <p>正在分析章节内容，提取世界状态...</p>
      <el-progress :percentage="loadingProgress" :stroke-width="8" />
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="audit-error">
      <el-icon :size="48"><WarningFilled /></el-icon>
      <p>{{ error }}</p>
      <el-button type="primary" @click="retryExtract">重新提取</el-button>
    </div>

    <!-- Main content -->
    <div v-else-if="hasData" class="audit-content">
      <!-- Tabs -->
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="状态变更" name="stateChanges">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllStateChanges"
                :indeterminate="indeterminateStateChanges"
                @change="toggleAllStateChanges"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ stateChanges.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in stateChanges"
                :key="item.id || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateStateChangesSelection"
                />
                <div class="item-content">
                  <div class="item-header">
                    <el-tag size="small" :type="getStateChangeType(item.type)">
                      {{ getStateChangeTypeLabel(item.type) }}
                    </el-tag>
                    <span class="entity-name">{{ item.entity }}</span>
                  </div>
                  <div class="item-detail">
                    <span class="field-name">{{ item.field }}:</span>
                    <span class="old-value">{{ item.old_value || '无' }}</span>
                    <el-icon><ArrowRight /></el-icon>
                    <span class="new-value">{{ item.new_value }}</span>
                  </div>
                  <div class="item-reason">
                    <el-text size="small" type="info">原因: {{ item.reason }}</el-text>
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('stateChanges', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('stateChanges', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="stateChanges.length === 0" description="暂无状态变更" />
            </el-scrollbar>
          </div>
        </el-tab-pane>

        <el-tab-pane label="因果链" name="causalEvents">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllCausalEvents"
                :indeterminate="indeterminateCausalEvents"
                @change="toggleAllCausalEvents"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ causalEvents.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in causalEvents"
                :key="item.id || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateCausalEventsSelection"
                />
                <div class="item-content causal-item">
                  <div class="causal-flow">
                    <el-tag size="small" type="info">因</el-tag>
                    <span>{{ item.cause }}</span>
                  </div>
                  <el-icon class="flow-arrow"><ArrowDown /></el-icon>
                  <div class="causal-flow">
                    <el-tag size="small" type="warning">事</el-tag>
                    <span>{{ item.event }}</span>
                  </div>
                  <el-icon class="flow-arrow"><ArrowDown /></el-icon>
                  <div class="causal-flow">
                    <el-tag size="small" type="success">果</el-tag>
                    <span>{{ item.effect }}</span>
                  </div>
                  <div v-if="item.decision" class="causal-decision">
                    <el-tag size="small" type="primary">决</el-tag>
                    <span>{{ item.decision }}</span>
                  </div>
                  <div v-if="item.characters && item.characters.length" class="causal-characters">
                    涉及角色: {{ item.characters.join(', ') }}
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('causalEvents', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('causalEvents', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="causalEvents.length === 0" description="暂无因果链" />
            </el-scrollbar>
          </div>
        </el-tab-pane>

        <el-tab-pane label="伏笔" name="foreshadows">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllForeshadows"
                :indeterminate="indeterminateForeshadows"
                @change="toggleAllForeshadows"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ foreshadows.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in foreshadows"
                :key="item.id || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateForeshadowsSelection"
                />
                <div class="item-content">
                  <div class="item-header">
                    <el-tag size="small" :type="getForeshadowTypeTag(item.type)">
                      {{ item.type }}
                    </el-tag>
                    <el-tag size="small" :type="getImportanceTypeTag(item.importance)">
                      {{ item.importance }}
                    </el-tag>
                    <span class="foreshadow-content">{{ item.content }}</span>
                  </div>
                  <div class="foreshadow-meta">
                    <el-text size="small">
                      埋设章节: {{ item.source_chapter }}
                      <span v-if="item.target_chapter"> | 预期回收: 第{{ item.target_chapter }}章</span>
                    </el-text>
                  </div>
                  <div v-if="item.source_context" class="item-context">
                    <el-text size="small" type="info">原文: {{ item.source_context }}</el-text>
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('foreshadows', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('foreshadows', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="foreshadows.length === 0" description="暂无伏笔" />
            </el-scrollbar>
          </div>
        </el-tab-pane>

        <el-tab-pane label="叙事线程" name="threadUpdates">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllThreadUpdates"
                :indeterminate="indeterminateThreadUpdates"
                @change="toggleAllThreadUpdates"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ threadUpdates.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in threadUpdates"
                :key="item.thread_name || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateThreadUpdatesSelection"
                />
                <div class="item-content">
                  <div class="item-header">
                    <el-tag size="small" :type="getThreadUpdateType(item.update_type)">
                      {{ getThreadUpdateTypeLabel(item.update_type) }}
                    </el-tag>
                    <span class="thread-name">{{ item.thread_name }}</span>
                  </div>
                  <div class="thread-meta">
                    <el-text size="small">
                      涉及章节: {{ item.chapters?.join(', ') || '无' }}
                    </el-text>
                  </div>
                  <div v-if="item.pov_characters && item.pov_characters.length" class="thread-pov">
                    POV角色: {{ item.pov_characters.join(', ') }}
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('threadUpdates', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('threadUpdates', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="threadUpdates.length === 0" description="暂无叙事线程" />
            </el-scrollbar>
          </div>
        </el-tab-pane>

        <el-tab-pane label="情感弧线" name="emotionPoints">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllEmotionPoints"
                :indeterminate="indeterminateEmotionPoints"
                @change="toggleAllEmotionPoints"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ emotionPoints.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in emotionPoints"
                :key="item.id || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateEmotionPointsSelection"
                />
                <div class="item-content">
                  <div class="item-header">
                    <el-tag size="small">{{ item.character_name }}</el-tag>
                    <el-tag size="small" :type="getEmotionTypeTag(item.emotion)">
                      {{ item.emotion }}
                    </el-tag>
                    <el-rate
                      v-model="item.intensity"
                      :max="10"
                      disabled
                      show-score
                      text-color="#ff9900"
                    />
                  </div>
                  <div class="emotion-trigger">
                    触发: {{ item.trigger }}
                  </div>
                  <div v-if="item.source_context" class="item-context">
                    <el-text size="small" type="info">原文: {{ item.source_context }}</el-text>
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('emotionPoints', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('emotionPoints', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="emotionPoints.length === 0" description="暂无情感弧线" />
            </el-scrollbar>
          </div>
        </el-tab-pane>

        <el-tab-pane label="时间线" name="timelineEvents">
          <div class="tab-content">
            <div class="tab-header">
              <el-checkbox
                v-model="selectAllTimelineEvents"
                :indeterminate="indeterminateTimelineEvents"
                @change="toggleAllTimelineEvents"
              >
                全选
              </el-checkbox>
              <span class="item-count">{{ timelineEvents.length }} 项待审核</span>
            </div>
            <el-scrollbar height="400px">
              <div
                v-for="(item, index) in timelineEvents"
                :key="item.id || index"
                class="audit-item"
              >
                <el-checkbox
                  v-model="item._accepted"
                  @change="updateTimelineEventsSelection"
                />
                <div class="item-content">
                  <div class="item-header">
                    <el-tag size="small" type="info">{{ item.time_label }}</el-tag>
                    <el-tag size="small" v-if="item.duration">持续: {{ item.duration }}</el-tag>
                  </div>
                  <div class="timeline-events">
                    <div v-for="(event, ei) in item.events" :key="ei" class="timeline-event">
                      {{ event }}
                    </div>
                  </div>
                  <div v-if="item.characters && item.characters.length" class="timeline-characters">
                    涉及角色: {{ item.characters.join(', ') }}
                  </div>
                  <div v-if="item.location" class="timeline-location">
                    地点: {{ item.location }}
                  </div>
                </div>
                <div class="item-actions">
                  <el-button size="small" @click="editItem('timelineEvents', index)">
                    编辑
                  </el-button>
                  <el-button size="small" type="danger" @click="rejectItem('timelineEvents', index)">
                    拒绝
                  </el-button>
                </div>
              </div>
              <el-empty v-if="timelineEvents.length === 0" description="暂无时间线事件" />
            </el-scrollbar>
          </div>
        </el-tab-pane>
      </el-tabs>

      <!-- Summary bar -->
      <div class="audit-summary">
        <div class="summary-stats">
          <el-tag type="info">
            状态变更: {{ acceptedStateChangesCount }}/{{ stateChanges.length }}
          </el-tag>
          <el-tag type="info">
            因果链: {{ acceptedCausalEventsCount }}/{{ causalEvents.length }}
          </el-tag>
          <el-tag type="info">
            伏笔: {{ acceptedForeshadowsCount }}/{{ foreshadows.length }}
          </el-tag>
          <el-tag type="info">
            叙事线程: {{ acceptedThreadUpdatesCount }}/{{ threadUpdates.length }}
          </el-tag>
          <el-tag type="info">
            情感弧线: {{ acceptedEmotionPointsCount }}/{{ emotionPoints.length }}
          </el-tag>
          <el-tag type="info">
            时间线: {{ acceptedTimelineEventsCount }}/{{ timelineEvents.length }}
          </el-tag>
        </div>
        <div class="summary-total">
          <span>总计: {{ totalAcceptedCount }}/{{ totalItemCount }} 项已选中</span>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="audit-empty">
      <el-icon :size="48"><Document /></el-icon>
      <p>未检测到世界状态变更</p>
      <el-text type="info">当前章节没有提取到任何图谱数据变更</el-text>
    </div>

    <!-- Footer -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="danger" @click="rejectAll" :disabled="!hasAcceptedItems">
          拒绝全部
        </el-button>
        <el-button type="primary" @click="applyChanges" :disabled="!hasAcceptedItems" :loading="applying">
          应用变更
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- Edit dialog -->
  <el-dialog v-model="editDialogVisible" title="编辑项" width="500px">
    <el-form v-if="editItemData" label-width="80px">
      <!-- Dynamic form based on item type -->
      <template v-if="editItemType === 'stateChanges'">
        <el-form-item label="类型">
          <el-input v-model="editItemData.type" disabled />
        </el-form-item>
        <el-form-item label="实体">
          <el-input v-model="editItemData.entity" />
        </el-form-item>
        <el-form-item label="字段">
          <el-input v-model="editItemData.field" />
        </el-form-item>
        <el-form-item label="旧值">
          <el-input v-model="editItemData.old_value" />
        </el-form-item>
        <el-form-item label="新值">
          <el-input v-model="editItemData.new_value" />
        </el-form-item>
        <el-form-item label="原因">
          <el-input v-model="editItemData.reason" type="textarea" :rows="2" />
        </el-form-item>
      </template>

      <template v-if="editItemType === 'causalEvents'">
        <el-form-item label="因">
          <el-input v-model="editItemData.cause" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="事">
          <el-input v-model="editItemData.event" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="果">
          <el-input v-model="editItemData.effect" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="决">
          <el-input v-model="editItemData.decision" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="涉及角色">
          <el-select v-model="editItemData.characters" multiple filterable allow-create>
            <el-option v-for="c in availableCharacters" :key="c" :value="c" :label="c" />
          </el-select>
        </el-form-item>
      </template>

      <template v-if="editItemType === 'foreshadows'">
        <el-form-item label="内容">
          <el-input v-model="editItemData.content" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="editItemData.type">
            <el-option label="埋设" value="埋设" />
            <el-option label="回收" value="回收" />
          </el-select>
        </el-form-item>
        <el-form-item label="重要性">
          <el-select v-model="editItemData.importance">
            <el-option label="高" value="高" />
            <el-option label="中" value="中" />
            <el-option label="低" value="低" />
          </el-select>
        </el-form-item>
        <el-form-item label="预期章节">
          <el-input-number v-model="editItemData.target_chapter" :min="0" />
        </el-form-item>
      </template>

      <template v-if="editItemType === 'threadUpdates'">
        <el-form-item label="线程名">
          <el-input v-model="editItemData.thread_name" />
        </el-form-item>
        <el-form-item label="更新类型">
          <el-select v-model="editItemData.update_type">
            <el-option label="新建" value="new" />
            <el-option label="添加章节" value="chapter_add" />
            <el-option label="POV变更" value="pov_change" />
          </el-select>
        </el-form-item>
        <el-form-item label="涉及章节">
          <el-select v-model="editItemData.chapters" multiple filterable allow-create>
            <el-option v-for="ch in availableChapters" :key="ch" :value="ch" :label="`第${ch}章`" />
          </el-select>
        </el-form-item>
        <el-form-item label="POV角色">
          <el-select v-model="editItemData.pov_characters" multiple filterable allow-create>
            <el-option v-for="c in availableCharacters" :key="c" :value="c" :label="c" />
          </el-select>
        </el-form-item>
      </template>

      <template v-if="editItemType === 'emotionPoints'">
        <el-form-item label="角色">
          <el-select v-model="editItemData.character_name" filterable allow-create>
            <el-option v-for="c in availableCharacters" :key="c" :value="c" :label="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="情感">
          <el-input v-model="editItemData.emotion" />
        </el-form-item>
        <el-form-item label="强度">
          <el-rate v-model="editItemData.intensity" :max="10" show-score />
        </el-form-item>
        <el-form-item label="触发">
          <el-input v-model="editItemData.trigger" type="textarea" :rows="2" />
        </el-form-item>
      </template>

      <template v-if="editItemType === 'timelineEvents'">
        <el-form-item label="时间标签">
          <el-input v-model="editItemData.time_label" />
        </el-form-item>
        <el-form-item label="持续时间">
          <el-input v-model="editItemData.duration" />
        </el-form-item>
        <el-form-item label="事件">
          <el-select v-model="editItemData.events" multiple filterable allow-create>
            <el-option v-for="e in editItemData.events" :key="e" :value="e" :label="e" />
          </el-select>
        </el-form-item>
        <el-form-item label="涉及角色">
          <el-select v-model="editItemData.characters" multiple filterable allow-create>
            <el-option v-for="c in availableCharacters" :key="c" :value="c" :label="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="地点">
          <el-input v-model="editItemData.location" />
        </el-form-item>
      </template>
    </el-form>
    <template #footer>
      <el-button @click="editDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="saveEdit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Loading,
  WarningFilled,
  Document,
  ArrowRight,
  ArrowDown
} from '@element-plus/icons-vue'
import { auditApi, settingsApi, chapterApi } from '@/api'

// Props
const props = defineProps({
  bookId: {
    type: String,
    required: true
  },
  chapterId: {
    type: Number,
    required: true
  },
  modelValue: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'applied'])

// Dialog visibility
const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// State
const loading = ref(false)
const loadingProgress = ref(0)
const error = ref('')
const applying = ref(false)
const activeTab = ref('stateChanges')

// Data
const stateChanges = ref([])
const causalEvents = ref([])
const foreshadows = ref([])
const threadUpdates = ref([])
const emotionPoints = ref([])
const timelineEvents = ref([])

// Available options
const availableCharacters = ref([])
const availableChapters = ref([])

// Edit dialog
const editDialogVisible = ref(false)
const editItemType = ref('')
const editItemIndex = ref(-1)
const editItemData = ref(null)

// Selection state for each tab
const selectAllStateChanges = ref(false)
const selectAllCausalEvents = ref(false)
const selectAllForeshadows = ref(false)
const selectAllThreadUpdates = ref(false)
const selectAllEmotionPoints = ref(false)
const selectAllTimelineEvents = ref(false)

// Indeterminate state
const indeterminateStateChanges = computed(() => {
  const accepted = stateChanges.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < stateChanges.value.length
})
const indeterminateCausalEvents = computed(() => {
  const accepted = causalEvents.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < causalEvents.value.length
})
const indeterminateForeshadows = computed(() => {
  const accepted = foreshadows.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < foreshadows.value.length
})
const indeterminateThreadUpdates = computed(() => {
  const accepted = threadUpdates.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < threadUpdates.value.length
})
const indeterminateEmotionPoints = computed(() => {
  const accepted = emotionPoints.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < emotionPoints.value.length
})
const indeterminateTimelineEvents = computed(() => {
  const accepted = timelineEvents.value.filter(i => i._accepted).length
  return accepted > 0 && accepted < timelineEvents.value.length
})

// Accepted counts
const acceptedStateChangesCount = computed(() => stateChanges.value.filter(i => i._accepted).length)
const acceptedCausalEventsCount = computed(() => causalEvents.value.filter(i => i._accepted).length)
const acceptedForeshadowsCount = computed(() => foreshadows.value.filter(i => i._accepted).length)
const acceptedThreadUpdatesCount = computed(() => threadUpdates.value.filter(i => i._accepted).length)
const acceptedEmotionPointsCount = computed(() => emotionPoints.value.filter(i => i._accepted).length)
const acceptedTimelineEventsCount = computed(() => timelineEvents.value.filter(i => i._accepted).length)

const totalAcceptedCount = computed(() =>
  acceptedStateChangesCount.value +
  acceptedCausalEventsCount.value +
  acceptedForeshadowsCount.value +
  acceptedThreadUpdatesCount.value +
  acceptedEmotionPointsCount.value +
  acceptedTimelineEventsCount.value
)

const totalItemCount = computed(() =>
  stateChanges.value.length +
  causalEvents.value.length +
  foreshadows.value.length +
  threadUpdates.value.length +
  emotionPoints.value.length +
  timelineEvents.value.length
)

const hasData = computed(() => totalItemCount.value > 0)
const hasAcceptedItems = computed(() => totalAcceptedCount.value > 0)

// Methods
const extractData = async () => {
  loading.value = true
  loadingProgress.value = 0
  error.value = ''

  try {
    // Simulate progress
    loadingProgress.value = 30

    const res = await auditApi.extractAll(props.bookId, props.chapterId)

    loadingProgress.value = 70

    if (res.data) {
      // Initialize data with _accepted property
      stateChanges.value = (res.data.state_changes || []).map(i => ({ ...i, _accepted: true }))
      causalEvents.value = (res.data.causal_events || []).map(i => ({ ...i, _accepted: true }))
      foreshadows.value = (res.data.foreshadows || []).map(i => ({ ...i, _accepted: true }))
      threadUpdates.value = (res.data.thread_updates || []).map(i => ({ ...i, _accepted: true }))
      emotionPoints.value = (res.data.emotion_points || []).map(i => ({ ...i, _accepted: true }))
      timelineEvents.value = (res.data.timeline_events || []).map(i => ({ ...i, _accepted: true }))
    }

    loadingProgress.value = 100

    // Update select all states
    updateAllSelectionStates()
  } catch (err) {
    error.value = err.response?.data?.error || err.message || '提取失败'
    ElMessage.error(error.value)
  }

  loading.value = false
}

const loadAvailableOptions = async () => {
  try {
    // Load characters
    const charRes = await settingsApi.getCharacters(props.bookId)
    availableCharacters.value = (charRes.data || []).map(c => c.name)

    // Load chapters
    const chapterRes = await chapterApi.list(props.bookId)
    availableChapters.value = (chapterRes.data || []).map(c => c.id)
  } catch (err) {
    console.error('Failed to load options:', err)
  }
}

const retryExtract = () => {
  extractData()
}

const handleClose = () => {
  emit('update:modelValue', false)
}

// Selection methods
const toggleAllStateChanges = (val) => {
  stateChanges.value.forEach(i => i._accepted = val)
}
const toggleAllCausalEvents = (val) => {
  causalEvents.value.forEach(i => i._accepted = val)
}
const toggleAllForeshadows = (val) => {
  foreshadows.value.forEach(i => i._accepted = val)
}
const toggleAllThreadUpdates = (val) => {
  threadUpdates.value.forEach(i => i._accepted = val)
}
const toggleAllEmotionPoints = (val) => {
  emotionPoints.value.forEach(i => i._accepted = val)
}
const toggleAllTimelineEvents = (val) => {
  timelineEvents.value.forEach(i => i._accepted = val)
}

const updateStateChangesSelection = () => {
  selectAllStateChanges.value = stateChanges.value.every(i => i._accepted)
}
const updateCausalEventsSelection = () => {
  selectAllCausalEvents.value = causalEvents.value.every(i => i._accepted)
}
const updateForeshadowsSelection = () => {
  selectAllForeshadows.value = foreshadows.value.every(i => i._accepted)
}
const updateThreadUpdatesSelection = () => {
  selectAllThreadUpdates.value = threadUpdates.value.every(i => i._accepted)
}
const updateEmotionPointsSelection = () => {
  selectAllEmotionPoints.value = emotionPoints.value.every(i => i._accepted)
}
const updateTimelineEventsSelection = () => {
  selectAllTimelineEvents.value = timelineEvents.value.every(i => i._accepted)
}

const updateAllSelectionStates = () => {
  selectAllStateChanges.value = stateChanges.value.length > 0 && stateChanges.value.every(i => i._accepted)
  selectAllCausalEvents.value = causalEvents.value.length > 0 && causalEvents.value.every(i => i._accepted)
  selectAllForeshadows.value = foreshadows.value.length > 0 && foreshadows.value.every(i => i._accepted)
  selectAllThreadUpdates.value = threadUpdates.value.length > 0 && threadUpdates.value.every(i => i._accepted)
  selectAllEmotionPoints.value = emotionPoints.value.length > 0 && emotionPoints.value.every(i => i._accepted)
  selectAllTimelineEvents.value = timelineEvents.value.length > 0 && timelineEvents.value.every(i => i._accepted)
}

// Item actions
const editItem = (type, index) => {
  editItemType.value = type
  editItemIndex.value = index

  const items = getItemsByType(type)
  editItemData.value = { ...items[index] }

  editDialogVisible.value = true
}

const saveEdit = () => {
  const items = getItemsByType(editItemType.value)
  if (editItemIndex.value >= 0 && editItemIndex.value < items.length) {
    items[editItemIndex.value] = { ...editItemData.value }
  }
  editDialogVisible.value = false
  ElMessage.success('已保存修改')
}

const rejectItem = async (type, index) => {
  try {
    await ElMessageBox.confirm('确定拒绝此项目？', '确认拒绝', {
      confirmButtonText: '拒绝',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const items = getItemsByType(type)
    items[index]._accepted = false
    updateSelectionByType(type)
    ElMessage.info('已标记为拒绝')
  } catch {
    // User cancelled
  }
}

const rejectAll = async () => {
  try {
    await ElMessageBox.confirm('确定拒绝所有已选中的项目？', '确认拒绝全部', {
      confirmButtonText: '拒绝全部',
      cancelButtonText: '取消',
      type: 'warning'
    })

    stateChanges.value.forEach(i => i._accepted = false)
    causalEvents.value.forEach(i => i._accepted = false)
    foreshadows.value.forEach(i => i._accepted = false)
    threadUpdates.value.forEach(i => i._accepted = false)
    emotionPoints.value.forEach(i => i._accepted = false)
    timelineEvents.value.forEach(i => i._accepted = false)

    updateAllSelectionStates()
    ElMessage.info('已拒绝所有项目')
  } catch {
    // User cancelled
  }
}

const applyChanges = async () => {
  if (!hasAcceptedItems.value) {
    ElMessage.warning('请至少选择一项')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定应用 ${totalAcceptedCount.value} 项变更？`,
      '确认应用',
      {
        confirmButtonText: '应用',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    applying.value = true

    // Collect accepted item IDs
    const acceptedIds = []

    stateChanges.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.id))
    causalEvents.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.id))
    foreshadows.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.id))
    threadUpdates.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.thread_name))
    emotionPoints.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.id))
    timelineEvents.value.filter(i => i._accepted).forEach(i => acceptedIds.push(i.id))

    // Also include the full data for accepted items
    const acceptedData = {
      state_changes: stateChanges.value.filter(i => i._accepted),
      causal_events: causalEvents.value.filter(i => i._accepted),
      foreshadows: foreshadows.value.filter(i => i._accepted),
      thread_updates: threadUpdates.value.filter(i => i._accepted),
      emotion_points: emotionPoints.value.filter(i => i._accepted),
      timeline_events: timelineEvents.value.filter(i => i._accepted)
    }

    await auditApi.applyGraphs(props.bookId, acceptedIds)

    ElMessage.success(`已应用 ${totalAcceptedCount.value} 项变更`)
    emit('applied', acceptedData)
    handleClose()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('应用失败: ' + (err.response?.data?.error || err.message))
    }
  }

  applying.value = false
}

// Helper methods
const getItemsByType = (type) => {
  const typeMap = {
    'stateChanges': stateChanges.value,
    'causalEvents': causalEvents.value,
    'foreshadows': foreshadows.value,
    'threadUpdates': threadUpdates.value,
    'emotionPoints': emotionPoints.value,
    'timelineEvents': timelineEvents.value
  }
  return typeMap[type] || []
}

const updateSelectionByType = (type) => {
  const updateMap = {
    'stateChanges': updateStateChangesSelection,
    'causalEvents': updateCausalEventsSelection,
    'foreshadows': updateForeshadowsSelection,
    'threadUpdates': updateThreadUpdatesSelection,
    'emotionPoints': updateEmotionPointsSelection,
    'timelineEvents': updateTimelineEventsSelection
  }
  const fn = updateMap[type]
  if (fn) fn()
}

// Tag type helpers
const getStateChangeType = (type) => {
  const typeMap = {
    'character_status': 'primary',
    'item_owner': 'warning',
    'relation': 'success'
  }
  return typeMap[type] || 'info'
}

const getStateChangeTypeLabel = (type) => {
  const labelMap = {
    'character_status': '人物状态',
    'item_owner': '物品持有',
    'relation': '关系变更'
  }
  return labelMap[type] || type
}

const getForeshadowTypeTag = (type) => {
  return type === '埋设' ? 'warning' : 'success'
}

const getImportanceTypeTag = (importance) => {
  const typeMap = {
    '高': 'danger',
    '中': 'warning',
    '低': 'info'
  }
  return typeMap[importance] || 'info'
}

const getThreadUpdateType = (type) => {
  const typeMap = {
    'new': 'success',
    'chapter_add': 'warning',
    'pov_change': 'info'
  }
  return typeMap[type] || 'info'
}

const getThreadUpdateTypeLabel = (type) => {
  const labelMap = {
    'new': '新建',
    'chapter_add': '添加章节',
    'pov_change': 'POV变更'
  }
  return labelMap[type] || type
}

const getEmotionTypeTag = (emotion) => {
  // Basic emotion categorization
  const positiveEmotions = ['喜悦', '满足', '兴奋', '安心', '感激', '爱']
  const negativeEmotions = ['愤怒', '悲伤', '恐惧', '焦虑', '绝望', '仇恨']

  if (positiveEmotions.some(e => emotion?.includes(e))) return 'success'
  if (negativeEmotions.some(e => emotion?.includes(e))) return 'danger'
  return 'warning'
}

// Watch for dialog open
watch(visible, (val) => {
  if (val) {
    loadAvailableOptions()
    extractData()
  }
})
</script>

<style scoped>
.audit-loading,
.audit-error,
.audit-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 48px;
  color: #909399;
}

.audit-error {
  color: #f56c6c;
}

.audit-loading .el-progress {
  width: 300px;
}

.audit-content {
  min-height: 500px;
}

.tab-content {
  padding: 16px;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.item-count {
  color: #909399;
  font-size: 13px;
}

.audit-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  margin-bottom: 8px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.audit-item:hover {
  background: #ecf5ff;
  border-color: #b3d8ff;
}

.audit-item .el-checkbox {
  margin-top: 4px;
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.entity-name,
.thread-name,
.foreshadow-content {
  font-weight: 500;
  color: #303133;
}

.item-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.field-name {
  color: #606266;
}

.old-value {
  color: #909399;
}

.new-value {
  color: #409eff;
  font-weight: 500;
}

.item-reason,
.item-context {
  margin-top: 8px;
}

.item-actions {
  display: flex;
  gap: 8px;
}

/* Causal chain specific styles */
.causal-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.causal-flow {
  display: flex;
  align-items: center;
  gap: 8px;
}

.flow-arrow {
  color: #909399;
  margin-left: 16px;
}

.causal-decision {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e4e7ed;
}

.causal-characters {
  color: #606266;
  font-size: 13px;
  margin-top: 4px;
}

/* Foreshadow specific styles */
.foreshadow-meta {
  margin-top: 8px;
}

/* Thread specific styles */
.thread-meta,
.thread-pov {
  margin-top: 8px;
  color: #606266;
  font-size: 13px;
}

/* Emotion specific styles */
.emotion-trigger {
  margin-top: 8px;
  color: #606266;
}

/* Timeline specific styles */
.timeline-events {
  margin-top: 8px;
}

.timeline-event {
  padding: 4px 0;
  border-bottom: 1px dashed #e4e7ed;
}

.timeline-characters,
.timeline-location {
  margin-top: 8px;
  color: #606266;
  font-size: 13px;
}

/* Summary bar */
.audit-summary {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  margin-top: 16px;
  border-radius: 8px;
}

.summary-stats {
  display: flex;
  gap: 8px;
}

.summary-total {
  font-weight: 500;
  color: #303133;
}

/* Dialog footer */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>