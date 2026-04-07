<template>
  <div
    class="graph-card"
    :class="{
      'graph-card--minimized': isMinimized,
      'graph-card--maximized': isMaximized,
      'graph-card--dragging': isDragging
    }"
    :style="cardStyle"
  >
    <!-- Title Bar (Drag Handle) -->
    <div
      class="graph-card__header"
      @mousedown="startDrag"
    >
      <div class="graph-card__title">
        <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
        <span>{{ title }}</span>
        <el-tag v-if="graphType" size="small" type="info">{{ graphTypeLabel }}</el-tag>
      </div>
      <div class="graph-card__controls">
        <el-button
          size="small"
          circle
          @click.stop="toggleMinimize"
        >
          <el-icon><component :is="isMinimized ? Plus : Minus" /></el-icon>
        </el-button>
        <el-button
          size="small"
          circle
          @click.stop="toggleMaximize"
        >
          <el-icon><component :is="isMaximized ? Close : FullScreen" /></el-icon>
        </el-button>
        <el-button
          v-if="!isMaximized"
          size="small"
          circle
          @click.stop="handleClose"
        >
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

    <!-- Content Area -->
    <div class="graph-card__body" v-show="!isMinimized">
      <!-- Loading State -->
      <div v-if="loading" class="graph-card__loading">
        <el-icon class="is-loading" :size="32"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="graph-card__error">
        <el-icon :size="32"><Warning /></el-icon>
        <span>{{ error }}</span>
        <el-button size="small" @click="retryLoad">重试</el-button>
      </div>

      <!-- Empty State -->
      <div v-else-if="isEmpty" class="graph-card__empty">
        <el-icon :size="32"><Document /></el-icon>
        <span>暂无数据</span>
      </div>

      <!-- Chart Container -->
      <div v-else ref="chartContainer" class="graph-card__chart"></div>
    </div>

    <!-- Minimized Summary -->
    <div v-if="isMinimized" class="graph-card__summary">
      <span>{{ summaryText }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { Loading, Warning, Document, Plus, Minus, FullScreen, Close } from '@element-plus/icons-vue'

const props = defineProps({
  title: {
    type: String,
    default: '知识图谱'
  },
  graphType: {
    type: String,
    default: 'relationship'
  },
  bookId: {
    type: String,
    required: true
  },
  graphData: {
    type: Object,
    default: () => ({ nodes: [], links: [], categories: [] })
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close', 'dragStart', 'minimize', 'maximize'])

// State
const chartContainer = ref(null)
const isMinimized = ref(false)
const isMaximized = ref(false)
const isDragging = ref(false)

// Drag position
const position = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })

// Chart instance
let chartInstance = null

// Computed
const graphTypeLabel = computed(() => {
  const labels = {
    'relationship': '关系图谱',
    'causal': '因果链',
    'foreshadow': '伏笔图',
    'thread': '叙事线',
    'emotion': '情感弧线',
    'timeline': '时间线'
  }
  return labels[props.graphType] || props.graphType
})

const isEmpty = computed(() => {
  const data = props.graphData
  if (!data) return true
  if (props.graphType === 'emotion') {
    return !data.series || data.series.length === 0
  }
  return !data.nodes || data.nodes.length === 0
})

const summaryText = computed(() => {
  const data = props.graphData
  if (!data) return '暂无数据'

  if (props.graphType === 'emotion') {
    const seriesCount = data.series?.length || 0
    return `${seriesCount} 条情感线`
  }

  const nodeCount = data.nodes?.length || 0
  const linkCount = data.links?.length || 0
  return `${nodeCount} 节点，${linkCount} 条关系`
})

const cardStyle = computed(() => {
  if (isMaximized.value) {
    return {
      position: 'fixed',
      top: '0',
      left: '0',
      right: '0',
      bottom: '0',
      width: '100%',
      height: '100%',
      zIndex: 1000
    }
  }
  if (position.value.x !== 0 || position.value.y !== 0) {
    return {
      transform: `translate(${position.value.x}px, ${position.value.y}px)`
    }
  }
  return {}
})

// Methods
const toggleMinimize = () => {
  isMinimized.value = !isMinimized.value
  emit('minimize', isMinimized.value)
  if (!isMinimized.value) {
    nextTick(() => {
      renderChart()
    })
  }
}

const toggleMaximize = () => {
  isMaximized.value = !isMaximized.value
  emit('maximize', isMaximized.value)
  nextTick(() => {
    renderChart()
    if (chartContainer.value && chartInstance) {
      chartInstance.resize()
    }
  })
}

const handleClose = () => {
  emit('close')
}

const retryLoad = () => {
  // Emit an event to parent to retry loading
  emit('reload')
}

// Drag functionality
const startDrag = (e) => {
  if (isMaximized.value) return
  if (e.target.closest('.graph-card__controls')) return

  isDragging.value = true
  dragOffset.value = {
    x: e.clientX - position.value.x,
    y: e.clientY - position.value.y
  }
  emit('dragStart', { x: position.value.x, y: position.value.y })

  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', endDrag)
}

const onDrag = (e) => {
  if (!isDragging.value) return
  position.value = {
    x: e.clientX - dragOffset.value.x,
    y: e.clientY - dragOffset.value.y
  }
}

const endDrag = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', endDrag)
}

// Chart rendering
const renderChart = () => {
  if (!chartContainer.value || isEmpty.value || props.loading || props.error) {
    return
  }

  // Dispose old instance
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }

  chartInstance = echarts.init(chartContainer.value)

  // Render different chart types based on graphType
  if (props.graphType === 'emotion') {
    renderEmotionChart()
  } else {
    renderForceGraph()
  }
}

const renderForceGraph = () => {
  const data = props.graphData
  const nodes = data.nodes || []
  const links = data.links || []
  const categories = data.categories || []

  // 从数据中提取实际使用的category类型
  const categoryNames = categories.map(c => c.name)

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (params.dataType === 'node') {
          const categoryLabel = getNodeTypeName(params.data.category)
          return `<strong>${params.data.name}</strong><br/>
                  类型: ${categoryLabel}<br/>
                  ${params.data.value || ''}`
        } else {
          return `${params.data.source} -> ${params.data.target}<br/>
                  关系: ${params.data.value || ''}`
        }
      }
    },
    legend: {
      data: categoryNames,
      top: 10,
      formatter: (name) => getNodeTypeName(name)
    },
    series: [{
      name: props.title,
      type: 'graph',
      layout: 'force',
      data: nodes.map(n => ({
        name: n.name,
        category: n.category,
        symbolSize: n.symbolSize || 30,
        value: n.value,
        itemStyle: {
          color: n.itemStyle?.color || getNodeColor(n.category)
        }
      })),
      links: links.map(l => ({
        source: l.source,
        target: l.target,
        value: l.value,
        lineStyle: l.lineStyle || { type: 'solid', color: '#aaa' },
        symbol: l.symbol || 'none'
      })),
      categories: categories.map(c => ({
        name: c.name,
        itemStyle: { color: getNodeColor(c.name) }
      })),
      roam: true,
      draggable: true,
      label: {
        show: true,
        position: 'right',
        fontSize: 12
      },
      force: {
        repulsion: 200,
        edgeLength: 120
      },
      lineStyle: {
        curveness: 0.2
      }
    }]
  }

  chartInstance.setOption(option, true)

  // Click event for node selection
  chartInstance.on('click', (params) => {
    if (params.dataType === 'node') {
      emit('nodeClick', params.data)
    }
  })
}

const renderEmotionChart = () => {
  const data = props.graphData
  const series = data.series || []
  const xAxisData = data.xAxis || []

  const option = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: series.map(s => s.name),
      top: 10
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: xAxisData
    },
    yAxis: {
      type: 'value',
      min: -10,
      max: 10
    },
    series: series.map(s => ({
      name: s.name,
      type: 'line',
      data: s.data,
      smooth: true,
      lineStyle: { width: 2 },
      areaStyle: { opacity: 0.1 }
    }))
  }

  chartInstance.setOption(option, true)
}

const getNodeTypeName = (type) => {
  const names = {
    'character': '人物',
    'item': '物品',
    'location': '地点',
    'faction': '势力',
    'event': '事件',
    'chapter': '章节',
    'foreshadow': '伏笔',
    'thread': '叙事线程'
  }
  return names[type] || type
}

const getNodeColor = (type) => {
  const colors = {
    'character': '#5470c6',
    'item': '#fac858',
    'location': '#91cc75',
    'faction': '#ee6666',
    'event': '#5470c6',
    'chapter': '#91cc75',
    'foreshadow': '#5470c6',
    'thread': '#ee6666'
  }
  return colors[type] || '#91cc75'
}

const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// Watch for data changes
watch(
  () => props.graphData,
  () => {
    nextTick(() => {
      renderChart()
    })
  },
  { deep: true }
)

watch(
  () => props.loading,
  (newVal) => {
    if (!newVal && !props.error) {
      nextTick(() => {
        renderChart()
      })
    }
  }
)

// Lifecycle
onMounted(() => {
  window.addEventListener('resize', handleResize)
  if (!props.loading && !props.error && !isEmpty.value) {
    nextTick(() => {
      renderChart()
    })
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})

// Expose methods for parent component
defineExpose({
  renderChart,
  resize: handleResize,
  toggleMinimize,
  toggleMaximize
})
</script>

<style scoped>
.graph-card {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: box-shadow 0.2s ease;
}

.graph-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.graph-card--dragging {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  cursor: grabbing;
}

.graph-card--minimized {
  min-height: auto;
}

.graph-card--maximized {
  border-radius: 0;
  box-shadow: none;
}

.graph-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  cursor: grab;
  user-select: none;
}

.graph-card__header:active {
  cursor: grabbing;
}

.graph-card__title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  font-size: 14px;
}

.graph-card__title .el-tag {
  margin-left: 8px;
}

.graph-card__controls {
  display: flex;
  gap: 8px;
}

.graph-card__controls .el-button {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: #fff;
}

.graph-card__controls .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
}

.graph-card__body {
  flex: 1;
  min-height: 300px;
  position: relative;
}

.graph-card__chart {
  width: 100%;
  height: 100%;
  min-height: 300px;
}

.graph-card__loading,
.graph-card__error,
.graph-card__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 100%;
  min-height: 300px;
  color: #909399;
}

.graph-card__error {
  color: #f56c6c;
}

.graph-card__summary {
  padding: 12px 16px;
  background: #f5f7fa;
  color: #606266;
  font-size: 13px;
}

/* Maximized state adjustments */
.graph-card--maximized .graph-card__body {
  height: calc(100vh - 52px);
}

.graph-card--maximized .graph-card__chart {
  height: 100%;
  min-height: auto;
}
</style>