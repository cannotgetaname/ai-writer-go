import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/books'
  },
  {
    path: '/books',
    name: 'Books',
    component: () => import('@/views/BooksView.vue')
  },
  {
    path: '/books/:id',
    name: 'BookDetail',
    component: () => import('@/views/BookDetailView.vue')
  },
  {
    path: '/books/:id/write',
    name: 'Writing',
    component: () => import('@/views/WritingView.vue')
  },
  {
    path: '/books/:id/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue')
  },
  {
    path: '/books/:id/timeline',
    name: 'Timeline',
    component: () => import('@/views/TimelineView.vue')
  },
  {
    path: '/books/:id/graph',
    name: 'Graph',
    component: () => import('@/views/GraphView.vue')
  },
  {
    path: '/books/:id/architect',
    name: 'Architect',
    component: () => import('@/views/ArchitectView.vue')
  },
  {
    path: '/books/:id/batch',
    name: 'Batch',
    component: () => import('@/views/BatchView.vue')
  },
  {
    path: '/books/:id/sync',
    name: 'Sync',
    component: () => import('@/views/SyncView.vue')
  },
  {
    path: '/books/:id/analysis',
    name: 'BookAnalysis',
    component: () => import('@/views/AnalysisBookView.vue')
  },
  {
    path: '/toolbox',
    name: 'Toolbox',
    component: () => import('@/views/ToolboxView.vue')
  },
  {
    path: '/analysis',
    name: 'Analysis',
    component: () => import('@/views/AnalysisView.vue')
  },
  {
    path: '/system',
    name: 'System',
    component: () => import('@/views/SystemView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router