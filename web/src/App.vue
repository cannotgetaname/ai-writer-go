<template>
  <el-config-provider :locale="zhCn">
    <div class="app-container">
      <el-container>
        <el-header>
          <div class="header-content">
            <h1 class="logo">
              <el-icon><EditPen /></el-icon>
              AI Writer
            </h1>
            <el-menu mode="horizontal" :default-active="activeMenu" router>
              <el-menu-item index="/books">
                <el-icon><Folder /></el-icon>
                书籍管理
              </el-menu-item>
              <el-menu-item index="/toolbox">
                <el-icon><MagicStick /></el-icon>
                智能工具箱
              </el-menu-item>
              <el-menu-item index="/analysis">
                <el-icon><Reading /></el-icon>
                拆书分析
              </el-menu-item>
              <el-menu-item index="/system">
                <el-icon><Setting /></el-icon>
                系统设置
              </el-menu-item>
            </el-menu>
            <UpdateBadge />
          </div>
        </el-header>
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </div>
  </el-config-provider>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { EditPen, Folder, MagicStick, Reading, Setting } from '@element-plus/icons-vue'
import UpdateBadge from '@/components/UpdateBadge.vue'

const router = useRouter()
const activeMenu = computed(() => {
  const path = router.currentRoute.value.path
  if (path.startsWith('/books')) return '/books'
  return path
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Microsoft YaHei', sans-serif;
  background: #f5f7fa;
}

.app-container {
  min-height: 100vh;
}

.el-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0 20px;
  height: 60px;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  display: flex;
  align-items: center;
  gap: 8px;
}

.el-menu--horizontal {
  background: transparent;
  border: none;
}

.el-menu--horizontal .el-menu-item {
  color: white;
  border-bottom: none;
}

.el-menu--horizontal .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.el-menu--horizontal .el-menu-item.is-active {
  background: rgba(255, 255, 255, 0.2);
  border-bottom: 2px solid white;
}

.el-main {
  padding: 20px;
  background: #f5f7fa;
}
</style>