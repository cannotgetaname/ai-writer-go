<template>
  <div class="architect-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 架构师</h2>
      <el-tag type="info">分形写作</el-tag>
    </div>

    <!-- 步骤条 -->
    <el-card class="steps-card">
      <el-steps :active="currentStep" align-center>
        <el-step title="全书总纲" description="故事梗概与核心设定" />
        <el-step title="世界观" description="力量体系与社会结构" />
        <el-step title="分卷大纲" description="各卷主要内容" />
        <el-step title="章节大纲" description="逐章规划" />
        <el-step title="章节细纲" description="场景与对话设计" />
      </el-steps>
    </el-card>

    <el-row :gutter="20">
      <!-- 左侧：当前步骤表单 -->
      <el-col :span="12">
        <!-- Step 1: 全书总纲 -->
        <el-card v-show="currentStep === 0" class="form-card">
          <template #header>
            <span>第一步：生成全书总纲</span>
          </template>
          <el-form :model="synopsisParams" label-width="100px">
            <el-form-item label="题材类型">
              <el-select v-model="synopsisParams.genre">
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
              <el-input-number v-model="synopsisParams.target_words" :min="100000" :max="5000000" :step="100000" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateSynopsis" :loading="loading">
            生成全书总纲
          </el-button>
        </el-card>

        <!-- Step 2: 世界观 -->
        <el-card v-show="currentStep === 1" class="form-card">
          <template #header>
            <div class="card-header">
              <span>第二步：生成世界观</span>
              <el-button size="small" @click="currentStep = 0">返回修改总纲</el-button>
            </div>
          </template>
          <div class="context-preview">
            <el-text type="info">基于总纲：{{ synopsisResult?.synopsis?.substring(0, 100) }}...</el-text>
          </div>
          <el-button type="primary" @click="generateWorldView" :loading="loading">
            生成世界观设定
          </el-button>
        </el-card>

        <!-- Step 3: 分卷大纲 -->
        <el-card v-show="currentStep === 2" class="form-card">
          <template #header>
            <div class="card-header">
              <span>第三步：生成分卷大纲</span>
              <el-button size="small" @click="currentStep = 1">返回修改世界观</el-button>
            </div>
          </template>
          <el-form label-width="100px">
            <el-form-item label="分卷数量">
              <el-input-number v-model="synopsisResult.volume_count" :min="1" :max="20" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="generateVolumes" :loading="loading">
            生成分卷大纲
          </el-button>
        </el-card>

        <!-- Step 4: 章节大纲 -->
        <el-card v-show="currentStep === 3" class="form-card">
          <template #header>
            <div class="card-header">
              <span>第四步：展开章节大纲</span>
              <el-button size="small" @click="currentStep = 2">返回修改分卷</el-button>
            </div>
          </template>
          <div class="volume-select">
            <el-text>选择要展开的分卷：</el-text>
            <el-select v-model="selectedVolumeIndex" placeholder="选择分卷">
              <el-option
                v-for="(vol, idx) in volumeResults"
                :key="idx"
                :label="vol.title"
                :value="idx"
              />
            </el-select>
          </div>
          <el-button type="primary" @click="expandVolume" :loading="loading" :disabled="selectedVolumeIndex < 0">
            展开选中分卷
          </el-button>
          <el-button type="success" @click="expandAllVolumes" :loading="loading">
            展开全部分卷
          </el-button>
        </el-card>

        <!-- Step 5: 章节细纲 -->
        <el-card v-show="currentStep === 4" class="form-card">
          <template #header>
            <div class="card-header">
              <span>第五步：生成章节细纲</span>
              <el-button size="small" @click="currentStep = 3">返回修改章节</el-button>
            </div>
          </template>
          <div class="chapter-select">
            <el-text>选择要细化章节：</el-text>
            <el-cascader
              v-model="selectedChapterPath"
              :options="chapterOptions"
              :props="{ value: 'id', label: 'title', children: 'chapters' }"
              placeholder="选择章节"
              clearable
            />
          </div>
          <el-button type="primary" @click="expandChapter" :loading="loading" :disabled="!selectedChapterPath">
            生成章节细纲
          </el-button>
        </el-card>
      </el-col>

      <!-- 右侧：结果展示 -->
      <el-col :span="12">
        <!-- 总纲结果 -->
        <el-card v-if="synopsisResult" v-show="currentStep >= 0" class="result-card">
          <template #header>
            <div class="card-header">
              <span>全书总纲</span>
              <el-tag v-if="synopsisResult.title">{{ synopsisResult.title }}</el-tag>
            </div>
          </template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="题材">{{ synopsisResult.genre }}</el-descriptions-item>
            <el-descriptions-item label="主题">{{ synopsisResult.theme }}</el-descriptions-item>
            <el-descriptions-item label="预估字数">{{ synopsisResult.word_count }}</el-descriptions-item>
            <el-descriptions-item label="结局类型">{{ synopsisResult.ending_type }}</el-descriptions-item>
          </el-descriptions>
          <div class="synopsis-text">
            <strong>故事梗概：</strong>
            <p>{{ synopsisResult.synopsis }}</p>
          </div>
          <div v-if="synopsisResult.main_plot">
            <strong>主线剧情：</strong>
            <p>{{ synopsisResult.main_plot }}</p>
          </div>
        </el-card>

        <!-- 世界观结果 -->
        <el-card v-if="worldViewResult" v-show="currentStep >= 1" class="result-card">
          <template #header>
            <div class="card-header">
              <span>世界观设定</span>
              <el-button size="small" type="success" @click="saveWorldView">保存世界观</el-button>
            </div>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="时代背景">{{ worldViewResult.era }}</el-descriptions-item>
            <el-descriptions-item label="科技水平">{{ worldViewResult.tech_level }}</el-descriptions-item>
            <el-descriptions-item label="力量体系">
              <div class="long-text">{{ worldViewResult.power_system }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="社会结构">
              <div class="long-text">{{ worldViewResult.social_structure }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="特殊规则">
              <div class="long-text">{{ worldViewResult.special_rules }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="主要势力">{{ worldViewResult.organizations }}</el-descriptions-item>
            <el-descriptions-item label="主要地点">{{ worldViewResult.locations }}</el-descriptions-item>
            <el-descriptions-item label="历史背景">
              <div class="long-text">{{ worldViewResult.history }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="主要矛盾">{{ worldViewResult.main_conflict }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 分卷大纲结果 -->
        <el-card v-if="volumeResults.length > 0" v-show="currentStep >= 2" class="result-card">
          <template #header>
            <div class="card-header">
              <span>分卷大纲 ({{ volumeResults.length }}卷)</span>
              <el-button size="small" type="success" @click="saveOutline">保存为章节</el-button>
            </div>
          </template>
          <el-collapse>
            <el-collapse-item
              v-for="(vol, idx) in volumeResults"
              :key="idx"
              :name="idx"
            >
              <template #title>
                <div class="volume-title">
                  <el-tag>{{ vol.title }}</el-tag>
                  <el-text type="info" size="small">共{{ vol.chapter_count || vol.chapters?.length || 0 }}章</el-text>
                </div>
              </template>
              <div class="volume-content">
                <p><strong>梗概：</strong>{{ vol.synopsis }}</p>
                <p><strong>核心事件：</strong>{{ vol.main_event }}</p>
                <p><strong>情感弧线：</strong>{{ vol.emotion_arc }}</p>

                <!-- 章节列表 -->
                <div v-if="vol.chapters && vol.chapters.length > 0" class="chapters-list">
                  <el-divider content-position="left">章节列表</el-divider>
                  <el-table :data="vol.chapters" size="small" max-height="300">
                    <el-table-column prop="index" label="#" width="50" />
                    <el-table-column prop="title" label="章节名" />
                    <el-table-column prop="synopsis" label="梗概" show-overflow-tooltip />
                  </el-table>
                </div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </el-card>

        <!-- 章节细纲结果 -->
        <el-card v-if="chapterDetail" v-show="currentStep >= 4" class="result-card">
          <template #header>
            <span>章节细纲</span>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="目标字数">{{ chapterDetail.word_target }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="chapterDetail.scenes && chapterDetail.scenes.length > 0" class="scenes-section">
            <h4>场景设计</h4>
            <el-timeline>
              <el-timeline-item
                v-for="(scene, idx) in chapterDetail.scenes"
                :key="idx"
                :timestamp="'场景' + (idx + 1)"
              >
                <el-card>
                  <p><strong>地点：</strong>{{ scene.location }}</p>
                  <p><strong>人物：</strong>{{ scene.characters }}</p>
                  <p><strong>事件：</strong>{{ scene.event }}</p>
                  <p><strong>氛围：</strong>{{ scene.mood }}</p>
                </el-card>
              </el-timeline-item>
            </el-timeline>
          </div>

          <div v-if="chapterDetail.dialogues && chapterDetail.dialogues.length > 0">
            <h4>关键对话</h4>
            <ul class="detail-list">
              <li v-for="(d, idx) in chapterDetail.dialogues" :key="idx">{{ d }}</li>
            </ul>
          </div>

          <div v-if="chapterDetail.foreshadows && chapterDetail.foreshadows.length > 0">
            <h4>伏笔设置</h4>
            <ul class="detail-list">
              <li v-for="(f, idx) in chapterDetail.foreshadows" :key="idx">{{ f }}</li>
            </ul>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { architectApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const currentStep = ref(0)
const loading = ref(false)
const selectedVolumeIndex = ref(-1)
const selectedChapterPath = ref(null)

// 数据
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

// 章节选项（用于级联选择）
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

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

// Step 1: 生成总纲
const generateSynopsis = async () => {
  loading.value = true
  try {
    const res = await architectApi.generateSynopsis(synopsisParams.value)
    if (res.data?.synopsis) {
      synopsisResult.value = res.data
      currentStep.value = 1
      ElMessage.success('总纲生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 2: 生成世界观
const generateWorldView = async () => {
  loading.value = true
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
      ElMessage.success('世界观生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 3: 生成分卷大纲
const generateVolumes = async () => {
  loading.value = true
  try {
    const res = await architectApi.generateVolumes({
      book_name: bookId.value,
      synopsis: synopsisResult.value,
      world_view: {
        power_system: worldViewResult.value.power_system,
        social_structure: worldViewResult.value.social_structure
      }
    })
    if (res.data?.volumes) {
      volumeResults.value = res.data.volumes
      currentStep.value = 3
      ElMessage.success('分卷大纲生成成功')
    } else {
      ElMessage.warning(res.data?.message || '生成失败')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 4: 展开单个分卷
const expandVolume = async () => {
  if (selectedVolumeIndex.value < 0) {
    ElMessage.warning('请选择要展开的分卷')
    return
  }
  loading.value = true
  try {
    const vol = volumeResults.value[selectedVolumeIndex.value]
    const res = await architectApi.expandVolume({
      book_name: bookId.value,
      volume: vol,
      synopsis: { synopsis: synopsisResult.value.synopsis },
      world_view: {
        power_system: worldViewResult.value.power_system,
        social_structure: worldViewResult.value.social_structure
      }
    })
    if (res.data?.chapters) {
      volumeResults.value[selectedVolumeIndex.value].chapters = res.data.chapters
      ElMessage.success('章节展开成功')
    }
  } catch (error) {
    ElMessage.error('展开失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// 展开所有分卷
const expandAllVolumes = async () => {
  loading.value = true
  try {
    for (let i = 0; i < volumeResults.value.length; i++) {
      const vol = volumeResults.value[i]
      if (!vol.chapters || vol.chapters.length === 0) {
        const res = await architectApi.expandVolume({
          book_name: bookId.value,
          volume: vol,
          synopsis: { synopsis: synopsisResult.value.synopsis },
          world_view: {
            power_system: worldViewResult.value.power_system,
            social_structure: worldViewResult.value.social_structure
          }
        })
        if (res.data?.chapters) {
          volumeResults.value[i].chapters = res.data.chapters
        }
      }
    }
    ElMessage.success('全部分卷展开完成')
  } catch (error) {
    ElMessage.error('展开失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// Step 5: 展开章节细纲
const expandChapter = async () => {
  if (!selectedChapterPath.value || selectedChapterPath.value.length < 2) {
    ElMessage.warning('请选择章节')
    return
  }
  loading.value = true
  try {
    const [volId, chapId] = selectedChapterPath.value
    const vIdx = parseInt(volId.split('_')[1])
    const cIdx = parseInt(chapId.split('_')[1])
    const chapter = volumeResults.value[vIdx]?.chapters?.[cIdx]

    if (!chapter) {
      ElMessage.error('找不到章节')
      return
    }

    const res = await architectApi.expandChapter({
      book_name: bookId.value,
      chapter: chapter,
      world_view: {
        power_system: worldViewResult.value.power_system
      }
    })
    if (res.data?.scenes) {
      chapterDetail.value = res.data
      currentStep.value = 4
      ElMessage.success('细纲生成成功')
    }
  } catch (error) {
    ElMessage.error('生成失败: ' + (error.response?.data?.error || error.message))
  }
  loading.value = false
}

// 保存世界观
const saveWorldView = async () => {
  try {
    const res = await architectApi.saveWorldView({
      book_name: bookId.value,
      world_view: worldViewResult.value
    })
    ElMessage.success('世界观已保存')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}

// 保存大纲为章节
const saveOutline = async () => {
  try {
    const res = await architectApi.saveOutline({
      book_name: bookId.value,
      volumes: volumeResults.value
    })
    ElMessage.success('大纲已保存为章节')
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.response?.data?.error || error.message))
  }
}
</script>

<style scoped>
.architect-view {
  max-width: 1600px;
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
}

.steps-card {
  margin-bottom: 20px;
}

.form-card {
  min-height: 300px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.context-preview {
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 15px;
}

.volume-select,
.chapter-select {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.result-card {
  margin-bottom: 20px;
}

.synopsis-text {
  margin-top: 15px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.synopsis-text p {
  margin: 5px 0 0 0;
  line-height: 1.6;
}

.long-text {
  white-space: pre-wrap;
  max-height: 100px;
  overflow-y: auto;
}

.volume-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.volume-content {
  padding: 10px;
}

.chapters-list {
  margin-top: 10px;
}

.scenes-section {
  margin-top: 15px;
}

.scenes-section h4,
.detail-list {
  margin: 10px 0;
}

.detail-list {
  padding-left: 20px;
}

.detail-list li {
  margin: 5px 0;
  color: #606266;
}
</style>