<template>
  <div class="graph-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 知识图谱</h2>
      <div class="header-actions">
        <el-button @click="refreshGraph">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-select v-model="filterType" placeholder="筛选类型" style="width: 120px;" clearable>
          <el-option label="全部" value="" />
          <el-option label="人物" value="character" />
          <el-option label="物品" value="item" />
          <el-option label="地点" value="location" />
        </el-select>
      </div>
    </div>

    <el-row :gutter="20">
      <!-- 图谱主体 -->
      <el-col :span="18">
        <el-card>
          <div ref="chartContainer" class="chart-container"></div>
        </el-card>
      </el-col>

      <!-- 右侧面板 -->
      <el-col :span="6">
        <!-- 统计信息 -->
        <el-card>
          <template #header>
            <span>图谱统计</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="节点总数">{{ stats.totalNodes }}</el-descriptions-item>
            <el-descriptions-item label="人物数">{{ stats.characters }}</el-descriptions-item>
            <el-descriptions-item label="物品数">{{ stats.items }}</el-descriptions-item>
            <el-descriptions-item label="地点数">{{ stats.locations }}</el-descriptions-item>
            <el-descriptions-item label="关系数">{{ stats.links }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 节点详情 -->
        <el-card v-if="selectedNode" style="margin-top: 20px;">
          <template #header>
            <div class="card-header">
              <span>节点详情</span>
              <el-tag :type="getNodeTypeColor(selectedNode.category)">{{ getNodeTypeName(selectedNode.category) }}</el-tag>
            </div>
          </template>
          <h3>{{ selectedNode.name }}</h3>
          <p class="node-desc">{{ selectedNode.value }}</p>
          <div class="related-nodes" v-if="relatedNodes.length">
            <h4>关联节点</h4>
            <el-tag v-for="node in relatedNodes" :key="node.name"
                    :type="getNodeTypeColor(node.category)"
                    style="margin: 2px; cursor: pointer;"
                    @click="selectNodeByName(node.name)">
              {{ node.name }} ({{ node.relation }})
            </el-tag>
          </div>
        </el-card>

        <!-- 快速操作 -->
        <el-card style="margin-top: 20px;">
          <template #header>
            <span>快速操作</span>
          </template>
          <el-button type="primary" @click="goToSettings" style="width: 100%;">
            <el-icon><Plus /></el-icon>
            添加人物/物品/地点
          </el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { graphApi } from '@/api'
import * as echarts from 'echarts'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const chartContainer = ref(null)
const hasData = ref(false)
const loading = ref(false)
const filterType = ref('')
const selectedNode = ref(null)
const relatedNodes = ref([])
const stats = ref({
  totalNodes: 0,
  characters: 0,
  items: 0,
  locations: 0,
  links: 0
})

let chartInstance = null
let graphData = { nodes: [], links: [], categories: [] }

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const goToSettings = () => {
  router.push(`/books/${bookId.value}/settings`)
}

const getNodeTypeName = (type) => {
  const names = {
    'character': '人物',
    'item': '物品',
    'location': '地点'
  }
  return names[type] || type
}

const getNodeTypeColor = (type) => {
  const colors = {
    'character': 'primary',
    'item': 'warning',
    'location': 'success'
  }
  return colors[type] || 'info'
}

const loadGraph = async () => {
  loading.value = true
  try {
    const res = await graphApi.getECharts(bookId.value)
    console.log('API response:', res)
    graphData = res.data || { nodes: [], links: [], categories: [] }
    console.log('graphData:', graphData, 'nodes count:', graphData.nodes?.length)

    if (graphData.nodes && graphData.nodes.length > 0) {
      hasData.value = true
      updateStats()
      // 等待 DOM 渲染完成
      await nextTick()
      setTimeout(() => {
        renderGraph()
      }, 100)
    } else {
      hasData.value = false
    }
  } catch (error) {
    console.error('加载图谱失败:', error)
    hasData.value = false
  }
  loading.value = false
}

const updateStats = () => {
  const nodes = graphData.nodes || []
  const links = graphData.links || []
  stats.value = {
    totalNodes: nodes.length,
    characters: nodes.filter(n => n.category === 'character').length,
    items: nodes.filter(n => n.category === 'item').length,
    locations: nodes.filter(n => n.category === 'location').length,
    links: links.length
  }
}

const renderGraph = () => {
  console.log('renderGraph called, chartContainer:', chartContainer.value)
  if (!chartContainer.value) {
    console.log('chartContainer not ready, retrying...')
    setTimeout(renderGraph, 200)
    return
  }

  // 销毁旧实例，重新创建
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }

  chartInstance = echarts.init(chartContainer.value)
  console.log('echarts instance created:', chartInstance)
  console.log('container size:', chartContainer.value.offsetWidth, 'x', chartContainer.value.offsetHeight)

  // 根据筛选条件过滤
  let nodes = graphData.nodes || []
  let links = graphData.links || []
  console.log('nodes to render:', nodes.length, 'links:', links.length)

  if (nodes.length === 0) {
    chartInstance.setOption({
      title: {
        text: '暂无数据',
        left: 'center',
        top: 'center'
      }
    })
    return
  }

  if (filterType.value) {
    nodes = nodes.filter(n => n.category === filterType.value)
    const nodeNames = new Set(nodes.map(n => n.name))
    links = links.filter(l => nodeNames.has(l.source) && nodeNames.has(l.target))
  }

  const option = {
    title: {
      text: '知识关系图谱',
      left: 'center',
      top: 10
    },
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        if (params.dataType === 'node') {
          return `${params.data.name}<br/>类型: ${getNodeTypeName(params.data.category)}<br/>${params.data.value || ''}`
        } else {
          return `${params.data.source} → ${params.data.target}<br/>关系: ${params.data.value}`
        }
      }
    },
    legend: [{
      data: ['character', 'item', 'location'],
      top: 40,
      formatter: (name) => getNodeTypeName(name)
    }],
    series: [{
      name: '知识图谱',
      type: 'graph',
      layout: 'force',
      data: nodes.map(n => ({
        name: n.name,
        category: n.category,
        symbolSize: n.symbolSize || 30,
        value: n.value,
        itemStyle: {
          color: n.itemStyle?.color || '#5470c6'
        }
      })),
      links: links.map(l => ({
        source: l.source,
        target: l.target,
        value: l.value
      })),
      categories: [
        { name: 'character', itemStyle: { color: '#5470c6' } },
        { name: 'item', itemStyle: { color: '#fac858' } },
        { name: 'location', itemStyle: { color: '#91cc75' } }
      ],
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
      emphasis: {
        focus: 'adjacency'
      },
      lineStyle: {
        curveness: 0.2
      }
    }]
  }

  console.log('option:', JSON.stringify(option, null, 2))
  chartInstance.setOption(option, true)
  console.log('chart rendered')

  // 点击事件
  chartInstance.off('click')
  chartInstance.on('click', (params) => {
    if (params.dataType === 'node') {
      selectNode(params.data)
    }
  })
}

const selectNode = (node) => {
  selectedNode.value = node

  // 查找关联节点
  relatedNodes.value = []
  const links = graphData.links || []
  const nodes = graphData.nodes || []

  links.forEach(link => {
    if (link.source === node.name) {
      const targetNode = nodes.find(n => n.name === link.target)
      if (targetNode) {
        relatedNodes.value.push({ ...targetNode, relation: link.value })
      }
    } else if (link.target === node.name) {
      const sourceNode = nodes.find(n => n.name === link.source)
      if (sourceNode) {
        relatedNodes.value.push({ ...sourceNode, relation: link.value })
      }
    }
  })
}

const selectNodeByName = (name) => {
  const nodes = graphData.nodes || []
  const node = nodes.find(n => n.name === name)
  if (node) {
    selectNode(node)
    if (chartInstance) {
      const idx = nodes.findIndex(n => n.name === name)
      chartInstance.dispatchAction({
        type: 'focusNodeAdjacency',
        seriesIndex: 0,
        dataIndex: idx
      })
    }
  }
}

const refreshGraph = () => {
  loadGraph()
}

// 监听筛选变化
const unwatchFilter = ref(null)

onMounted(() => {
  loadGraph()

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)

  // 监听筛选变化
  unwatchFilter.value = watch(filterType, () => {
    if (hasData.value) {
      renderGraph()
    }
  })
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.dispose()
  }
  window.removeEventListener('resize', handleResize)
  if (unwatchFilter.value) {
    unwatchFilter.value()
  }
})

const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 导入 watch
import { watch } from 'vue'
</script>

<style scoped>
.graph-view {
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

.header-actions {
  display: flex;
  gap: 10px;
}

.chart-container {
  height: 600px;
  background: #fafafa;
  border-radius: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.node-desc {
  color: #666;
  font-size: 13px;
  margin: 10px 0;
}

.related-nodes h4 {
  margin: 10px 0 5px 0;
  font-size: 13px;
  color: #666;
}
</style>