import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加认证头
api.interceptors.request.use(config => {
  const authKey = localStorage.getItem('authKey')
  if (authKey) {
    config.headers['X-Auth-Key'] = authKey
  }
  return config
}, error => {
  return Promise.reject(error)
})

// 响应拦截器 - 处理认证错误
api.interceptors.response.use(response => response, error => {
  if (error.response && error.response.status === 401) {
    // 清除本地存储的 authKey
    localStorage.removeItem('authKey')
    console.error('认证失败，请检查 API Key 配置')
  }
  return Promise.reject(error)
})

// 书籍管理
export const bookApi = {
  list: () => api.get('/books'),
  get: (id) => api.get(`/books/${id}`),
  create: (data) => api.post('/books', data),
  update: (id, data) => api.put(`/books/${id}`, data),
  delete: (id) => api.delete(`/books/${id}`),
  init: (data) => api.post('/books/init', data)
}

// 章节管理
export const chapterApi = {
  list: (bookId) => api.get(`/books/${bookId}/chapters`),
  get: (bookId, chapterId) => api.get(`/books/${bookId}/chapters/${chapterId}`),
  create: (bookId, data) => api.post(`/books/${bookId}/chapters`, data),
  update: (bookId, chapterId, data) => api.put(`/books/${bookId}/chapters/${chapterId}`, data),
  delete: (bookId, chapterId) => api.delete(`/books/${bookId}/chapters/${chapterId}`),
  getContent: (bookId, chapterId) => api.get(`/books/${bookId}/chapters/${chapterId}/content`),
  saveContent: (bookId, chapterId, content) => api.put(`/books/${bookId}/chapters/${chapterId}/content`, { content }),
  updateParagraph: (bookId, chapterId, paragraphId, text) => api.put(`/books/${bookId}/chapters/${chapterId}/paragraph`, { paragraph_id: paragraphId, text })
}

// 设定管理
export const settingsApi = {
  getWorldView: (bookId) => api.get(`/books/${bookId}/settings/worldview`),
  updateWorldView: (bookId, data) => api.put(`/books/${bookId}/settings/worldview`, data),
  getCharacters: (bookId) => api.get(`/books/${bookId}/settings/characters`),
  createCharacter: (bookId, data) => api.post(`/books/${bookId}/settings/characters`, data),
  updateCharacter: (bookId, charId, data) => api.put(`/books/${bookId}/settings/characters/${charId}`, data),
  deleteCharacter: (bookId, charId) => api.delete(`/books/${bookId}/settings/characters/${charId}`),
  getItems: (bookId) => api.get(`/books/${bookId}/settings/items`),
  createItem: (bookId, data) => api.post(`/books/${bookId}/settings/items`, data),
  updateItem: (bookId, itemId, data) => api.put(`/books/${bookId}/settings/items/${itemId}`, data),
  deleteItem: (bookId, itemId) => api.delete(`/books/${bookId}/settings/items/${itemId}`),
  getLocations: (bookId) => api.get(`/books/${bookId}/settings/locations`),
  createLocation: (bookId, data) => api.post(`/books/${bookId}/settings/locations`, data),
  updateLocation: (bookId, locId, data) => api.put(`/books/${bookId}/settings/locations/${locId}`, data),
  deleteLocation: (bookId, locId) => api.delete(`/books/${bookId}/settings/locations/${locId}`)
}

// AI 写作
export const aiApi = {
  generate: (data) => api.post('/ai/generate', data),
  generateStream: (data) => {
    // SSE 流式生成
    const eventSource = new EventSource(`/api/ai/generate/stream?${new URLSearchParams(data)}`)
    return eventSource
  },
  review: (data) => api.post('/ai/review', data),
  reviewByParagraph: (data) => api.post('/ai/review-paragraph', data),
  getReview: (bookId, chapterId) => api.get(`/ai/review?book_name=${bookId}&chapter_id=${chapterId}`),
  audit: (data) => api.post('/ai/audit', data),
  rewrite: (data) => api.post('/ai/rewrite', data),
  rewriteParagraph: (data) => api.post('/ai/rewrite-paragraph', data),
  continue: (data) => api.post('/ai/continue', data)
}

// 批量生成
export const batchApi = {
  generate: (data) => {
    // SSE 流式批量生成，返回 EventSource
    const params = new URLSearchParams({
      book_name: data.book_name,
      from: data.from,
      to: data.to,
      stream: data.stream || true,
      retry: data.retry || 2
    })
    // POST 请求无法直接用 EventSource，这里返回 fetch
    return fetch('/api/batch/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
  },
  status: (bookName) => api.get(`/batch/status?book_name=${bookName}`),
  reset: (bookName) => api.delete(`/batch/reset?book_name=${bookName}`)
}

// 导出
export const exportApi = {
  txt: (bookId) => `/api/books/${bookId}/export/txt`,
  markdown: (bookId) => `/api/books/${bookId}/export/markdown`,
  json: (bookId) => `/api/books/${bookId}/export/json`
}

// 状态同步
export const syncApi = {
  extract: (data) => api.post('/sync/extract', data),
  pending: (bookName) => api.get(`/sync/pending?book_name=${bookName}`),
  apply: (data) => api.post('/sync/apply', data),
  reject: (data) => api.post('/sync/reject', data)
}

// 工具箱
export const toolboxApi = {
  naming: (data) => api.post('/toolbox/naming', data),
  character: (data) => api.post('/toolbox/character', data),
  conflict: (data) => api.post('/toolbox/conflict', data),
  scene: (data) => api.post('/toolbox/scene', data),
  goldfinger: (data) => api.post('/toolbox/goldfinger', data),
  title: (data) => api.post('/toolbox/title', data),
  synopsis: (data) => api.post('/toolbox/synopsis', data),
  twist: (data) => api.post('/toolbox/twist', data),
  dialogue: (data) => api.post('/toolbox/dialogue', data)
}

// 架构师
export const architectApi = {
  // 分形写作流程
  generateSynopsis: (data) => api.post('/architect/synopsis', data),
  generateWorldView: (data) => api.post('/architect/worldview', data),
  generateVolumes: (data) => api.post('/architect/volumes', data),
  expandVolume: (data) => api.post('/architect/volume/expand', data),
  expandChapter: (data) => api.post('/architect/chapter/expand', data),
  saveOutline: (data) => api.post('/architect/save', data),
  saveWorldView: (data) => api.post('/architect/save-worldview', data),
  // 兼容旧接口
  generate: (data) => api.post('/architect/generate', data),
  fission: (data) => api.post('/architect/fission', data),
  strategies: () => api.get('/architect/strategies')
}

// 拆书分析
export const analysisApi = {
  parse: (data) => api.post('/analysis/parse', data),
  analyze: (data) => api.post('/analysis/analyze', data)
}

// 时间线和图谱
export const timelineApi = {
  get: (bookId) => api.get(`/books/${bookId}/timeline`),
  getThreads: (bookId) => api.get(`/books/${bookId}/timeline/threads`),
  createThread: (bookId, data) => api.post(`/books/${bookId}/timeline/threads`, data)
}

export const graphApi = {
  get: (bookId) => api.get(`/books/${bookId}/graph`),
  getECharts: (bookId) => api.get(`/books/${bookId}/graph/echarts`)
}

// 伏笔
export const foreshadowApi = {
  list: (bookId) => api.get(`/books/${bookId}/foreshadows`),
  create: (bookId, data) => api.post(`/books/${bookId}/foreshadows`, data),
  update: (bookId, id, data) => api.put(`/books/${bookId}/foreshadows/${id}`, data),
  resolve: (bookId, id, chapterId, resolvedContent = '') => api.post(`/books/${bookId}/foreshadows/${id}/resolve`, {
    chapter_id: chapterId,
    resolved_content: resolvedContent
  }),
  warnings: (bookId) => api.get(`/books/${bookId}/foreshadows/warnings`)
}

// 因果链
export const causalApi = {
  get: (bookId) => api.get(`/books/${bookId}/causal-chain`),
  create: (bookId, data) => api.post(`/books/${bookId}/causal-chain`, data),
  update: (bookId, eventId, data) => api.put(`/books/${bookId}/causal-chain/${eventId}`, data)
}

// 系统
export const systemApi = {
  getConfig: () => api.get('/system/config'),
  updateConfig: (data) => api.put('/system/config', data),
  getPrompts: () => api.get('/system/prompts'),
  updatePrompts: (data) => api.put('/system/prompts', data),
  getBilling: () => api.get('/system/billing'),
  getGoals: () => api.get('/system/goals'),
  updateGoals: (data) => api.put('/system/goals', data)
}

// 向量存储
export const vectorApi = {
  indexBook: (bookName) => api.post('/vector/index', { book_name: bookName }),
  indexChapter: (bookName, chapterId) => api.post('/vector/index/chapter', { book_name: bookName, chapter_id: chapterId }),
  search: (bookName, query, topK = 5) => api.post('/vector/search', { book_name: bookName, query, top_k: topK }),
  getStatus: (bookName) => api.get(`/vector/status?book_name=${bookName}`),
  deleteBook: (bookName) => api.delete(`/vector/index/${bookName}`)
}

// 一致性检查
export const consistencyApi = {
  check: (bookId, from, to) => api.post(`/books/${bookId}/check/consistency?from=${from}&to=${to}`),
  checkChapter: (bookId, chapterId) => api.post(`/books/${bookId}/check/consistency?chapter=${chapterId}`)
}

// 情感弧线
export const emotionApi = {
  track: (bookId, chapterId) => api.post(`/books/${bookId}/emotion/track?chapter=${chapterId}`),
  getArc: (bookId, charName) => api.get(`/books/${bookId}/emotion/${charName}`)
}

// 信息边界
export const infoBoundaryApi = {
  checkLeak: (bookId, chapterId) => api.post(`/books/${bookId}/info-boundary/check?chapter=${chapterId}`),
  extractInfo: (bookId, chapterId) => api.post(`/books/${bookId}/info-boundary/extract?chapter=${chapterId}`)
}

export default api