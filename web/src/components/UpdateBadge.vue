<template>
  <div class="update-badge" v-if="hasUpdate" @click="showUpdateDialog">
    <el-badge :value="newVersion" type="primary">
      <el-button circle size="small">
        <el-icon><Download /></el-icon>
      </el-button>
    </el-badge>
  </div>

  <!-- 更新对话框 -->
  <el-dialog v-model="dialogVisible" title="发现新版本" width="500px" custom-class="apple-dialog">
    <div class="update-dialog-content">
      <p>当前版本: v{{ currentVersion }}</p>
      <p>新版本: v{{ newVersion }}</p>
      <div class="changelog-preview">
        <h4>更新内容</h4>
        <div>{{ changelogPreview }}</div>
      </div>
    </div>
    <template #footer>
      <el-button @click="dialogVisible = false">稍后更新</el-button>
      <el-button type="primary" @click="goToUpdate">前往更新</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Download } from '@element-plus/icons-vue'
import { systemApi } from '@/api'

const router = useRouter()

const hasUpdate = ref(false)
const currentVersion = ref('')
const newVersion = ref('')
const changelog = ref('')
const dialogVisible = ref(false)

// 检查更新（启动时自动）
const checkUpdate = async () => {
  try {
    const res = await systemApi.checkUpdate('auto')
    hasUpdate.value = res.data.has_update
    if (hasUpdate.value) {
      newVersion.value = res.data.latest_version
      changelog.value = res.data.changelog || ''
      currentVersion.value = res.data.current_version
    }
  } catch (e) {
    // 静默失败，不影响用户体验
    console.warn('Auto update check failed:', e)
  }
}

const changelogPreview = computed(() => {
  if (!changelog.value) return '暂无更新说明'
  // 只显示前 200 字
  const preview = changelog.value.slice(0, 200)
  return preview + (changelog.value.length > 200 ? '...' : '')
})

const showUpdateDialog = () => {
  dialogVisible.value = true
}

const goToUpdate = () => {
  dialogVisible.value = false
  router.push('/system?tab=version') // 跳转到系统设置页面
}

onMounted(() => {
  checkUpdate()
})
</script>

<style scoped>
.update-badge {
  display: inline-block;
  cursor: pointer;
}

.update-dialog-content {
  line-height: 1.6;
}

.changelog-preview {
  margin-top: 16px;
  padding: 12px;
  background: #f5f5f7;
  border-radius: 8px;
}

.changelog-preview h4 {
  margin-bottom: 8px;
}
</style>