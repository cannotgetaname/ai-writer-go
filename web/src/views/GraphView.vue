<template>
  <div class="graph-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 知识图谱</h2>
      <div class="header-actions">
        <el-button @click="refreshAll">
          <el-icon><Refresh /></el-icon>
          刷新全部
        </el-button>
      </div>
    </div>

    <GraphGrid :bookId="bookId" ref="gridRef" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import GraphGrid from '@/components/GraphGrid.vue'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)
const gridRef = ref(null)

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const refreshAll = () => {
  gridRef.value?.loadAllGraphData()
}
</script>

<style scoped>
.graph-view {
  max-width: 1600px;
  margin: 0 auto;
  padding: 20px;
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
</style>