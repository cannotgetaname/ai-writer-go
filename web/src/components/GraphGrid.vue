<template>
  <div class="graph-grid">
    <!-- Toggle Panel -->
    <div class="graph-grid__panel">
      <div class="panel-header">
        <span>图谱开关</span>
        <el-button
          text
          size="small"
          @click="toggleAll"
        >
          {{ allEnabled ? '全部关闭' : '全部开启' }}
        </el-button>
      </div>
      <div class="panel-content">
        <div
          v-for="graph in allGraphTypes"
          :key="graph.type"
          class="panel-item"
        >
          <el-checkbox
            v-model="graphEnabled[graph.type]"
            @change="handleToggle(graph.type)"
          >
            {{ graph.title }}
          </el-checkbox>
          <div class="panel-item-stats">
            <span v-if="loadingMap[graph.type]" class="stats-loading">
              <el-icon class="is-loading"><Loading /></el-icon>
            </span>
            <span v-else-if="errorMap[graph.type]" class="stats-error">
              <el-icon><Warning /></el-icon>
            </span>
            <span v-else class="stats-count">
              {{ getStatsText(graph.type) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Draggable Grid -->
    <VueDraggable
      v-model="cards"
      :animation="150"
      handle=".graph-card__header"
      ghostClass="graph-grid__ghost"
      chosenClass="graph-grid__chosen"
      @end="saveLayout"
      class="graph-grid__container"
      :class="gridClass"
    >
      <GraphCard
        v-for="card in cards"
        :key="card.type"
        :book-id="bookId"
        :title="getGraphTitle(card.type)"
        :graph-type="card.type"
        :graph-data="graphDataMap[card.type]"
        :loading="loadingMap[card.type]"
        :error="errorMap[card.type]"
        @close="handleCloseCard(card.type)"
        @reload="loadGraphData(card.type)"
      />
    </VueDraggable>

    <!-- Empty State -->
    <div v-if="cards.length === 0" class="graph-grid__empty">
      <el-empty description="请从右侧面板选择要显示的图谱" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { Loading, Warning } from '@element-plus/icons-vue'
import { graphApi } from '@/api'
import GraphCard from './GraphCard.vue'

const props = defineProps({
  bookId: {
    type: String,
    required: true
  }
})

// Graph types definition
const allGraphTypes = [
  { type: 'relationship', title: '基础关系图' },
  { type: 'causal', title: '剧情因果图' },
  { type: 'foreshadow', title: '伏笔追踪图' },
  { type: 'thread', title: '叙事线程图' },
  { type: 'emotion', title: '情感弧线图' },
  { type: 'timeline', title: '时间线图' }
]

// State
const cards = ref([])
const graphDataMap = reactive({})
const loadingMap = reactive({})
const errorMap = reactive({})
const graphEnabled = reactive({})

// Initialize enabled state
allGraphTypes.forEach(g => {
  graphEnabled[g.type] = false
})

// Computed
const gridClass = computed(() => {
  const count = cards.value.length
  if (count <= 1) return 'graph-grid--single'
  return 'graph-grid--multi'
})

const allEnabled = computed(() => {
  return allGraphTypes.every(g => graphEnabled[g.type])
})

// Methods
const getStorageKey = () => `graph-layout-${props.bookId}`

const loadLayout = () => {
  try {
    const saved = localStorage.getItem(getStorageKey())
    if (saved) {
      const types = JSON.parse(saved)
      cards.value = types.map(type => {
        const graph = allGraphTypes.find(g => g.type === type)
        return graph ? { type: graph.type, title: graph.title } : null
      }).filter(Boolean)

      // Update enabled state
      cards.value.forEach(card => {
        graphEnabled[card.type] = true
      })
    }
  } catch (e) {
    console.error('Failed to load layout:', e)
  }
}

const saveLayout = () => {
  try {
    const types = cards.value.map(c => c.type)
    localStorage.setItem(getStorageKey(), JSON.stringify(types))
  } catch (e) {
    console.error('Failed to save layout:', e)
  }
}

const getGraphTitle = (type) => {
  const graph = allGraphTypes.find(g => g.type === type)
  return graph ? graph.title : type
}

const handleToggle = (type) => {
  if (graphEnabled[type]) {
    // Enable - add card
    if (!cards.value.find(c => c.type === type)) {
      const graph = allGraphTypes.find(g => g.type === type)
      cards.value.push({ type: graph.type, title: graph.title })
      loadGraphData(type)
    }
  } else {
    // Disable - remove card
    const index = cards.value.findIndex(c => c.type === type)
    if (index !== -1) {
      cards.value.splice(index, 1)
    }
  }
  saveLayout()
}

const toggleAll = () => {
  if (allEnabled.value) {
    // Disable all
    allGraphTypes.forEach(g => {
      graphEnabled[g.type] = false
    })
    cards.value = []
  } else {
    // Enable all
    allGraphTypes.forEach(g => {
      graphEnabled[g.type] = true
      if (!cards.value.find(c => c.type === g.type)) {
        cards.value.push({ type: g.type, title: g.title })
        loadGraphData(g.type)
      }
    })
  }
  saveLayout()
}

const handleCloseCard = (type) => {
  const index = cards.value.findIndex(c => c.type === type)
  if (index !== -1) {
    cards.value.splice(index, 1)
    graphEnabled[type] = false
    saveLayout()
  }
}

const loadGraphData = async (type) => {
  loadingMap[type] = true
  errorMap[type] = ''

  try {
    const response = await graphApi.getECharts(props.bookId, type)
    graphDataMap[type] = response.data || { nodes: [], links: [], categories: [] }
  } catch (e) {
    console.error(`Failed to load ${type} graph:`, e)
    errorMap[type] = e.response?.data?.error || e.message || '加载失败'
  } finally {
    loadingMap[type] = false
  }
}

const getStatsText = (type) => {
  const data = graphDataMap[type]
  if (!data) return ''

  if (type === 'emotion') {
    const seriesCount = data.series?.length || 0
    return `${seriesCount} 条线`
  }

  if (type === 'timeline') {
    const events = data.nodes?.length || 0
    return `${events} 事件`
  }

  const nodeCount = data.nodes?.length || 0
  const linkCount = data.links?.length || 0
  return `${nodeCount}/${linkCount}`
}

// Watch for bookId changes
watch(() => props.bookId, (newId, oldId) => {
  if (newId !== oldId) {
    // Reset state
    cards.value = []
    allGraphTypes.forEach(g => {
      graphEnabled[g.type] = false
      delete graphDataMap[g.type]
      delete loadingMap[g.type]
      delete errorMap[g.type]
    })
    // Load new layout
    loadLayout()
    // Load data for current cards
    cards.value.forEach(card => {
      loadGraphData(card.type)
    })
  }
})

// Lifecycle
onMounted(() => {
  loadLayout()
  // Load data for initial cards
  cards.value.forEach(card => {
    loadGraphData(card.type)
  })
})
</script>

<style scoped>
.graph-grid {
  position: relative;
  padding: 16px;
  min-height: 100%;
}

.graph-grid__panel {
  position: fixed;
  right: 16px;
  top: 80px;
  width: 200px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 100;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-weight: 500;
  font-size: 14px;
}

.panel-header .el-button {
  color: #fff;
  font-size: 12px;
}

.panel-content {
  padding: 8px 0;
}

.panel-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  transition: background-color 0.2s;
}

.panel-item:hover {
  background-color: #f5f7fa;
}

.panel-item :deep(.el-checkbox__label) {
  font-size: 13px;
  color: #606266;
}

.panel-item-stats {
  display: flex;
  align-items: center;
  min-width: 40px;
  justify-content: flex-end;
}

.stats-loading {
  color: #409eff;
}

.stats-error {
  color: #f56c6c;
}

.stats-count {
  font-size: 12px;
  color: #909399;
}

.graph-grid__container {
  display: grid;
  gap: 16px;
  margin-right: 220px;
  min-height: 400px;
}

/* Responsive grid columns */
.graph-grid--multi {
  grid-template-columns: repeat(3, 1fr);
}

.graph-grid--single {
  grid-template-columns: 1fr;
}

/* Medium screens: 2 columns */
@media (max-width: 1200px) {
  .graph-grid--multi {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Narrow screens: 1 column */
@media (max-width: 800px) {
  .graph-grid__container {
    margin-right: 0;
  }

  .graph-grid--multi {
    grid-template-columns: 1fr;
  }

  .graph-grid__panel {
    position: relative;
    right: auto;
    top: auto;
    width: 100%;
    margin-bottom: 16px;
  }
}

/* Drag styles */
.graph-grid__ghost {
  opacity: 0.5;
  background: #c8ebfb;
  border-radius: 8px;
}

.graph-grid__chosen {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}

/* Empty state */
.graph-grid__empty {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  margin-right: 220px;
}

@media (max-width: 800px) {
  .graph-grid__empty {
    margin-right: 0;
  }
}
</style>