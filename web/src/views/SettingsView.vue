<template>
  <div class="settings-view">
    <div class="page-header">
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ bookId }} - 设定管理</h2>
    </div>

    <el-tabs v-model="activeTab">
      <!-- 世界观 -->
      <el-tab-pane label="世界观" name="worldview">
        <el-card>
          <el-form label-width="120px">
            <el-form-item label="题材类型">
              <el-input v-model="worldview.basic_info.genre" />
            </el-form-item>
            <el-form-item label="时代背景">
              <el-input v-model="worldview.basic_info.era" />
            </el-form-item>
            <el-form-item label="力量体系">
              <el-input v-model="worldview.core_settings.power_system" type="textarea" :rows="3" />
            </el-form-item>
            <el-form-item label="社会结构">
              <el-input v-model="worldview.core_settings.social_structure" type="textarea" :rows="3" />
            </el-form-item>
          </el-form>
          <el-button type="primary" @click="saveWorldView">保存世界观</el-button>
        </el-card>
      </el-tab-pane>

      <!-- 人物 -->
      <el-tab-pane label="人物" name="characters">
        <div class="tab-header">
          <el-button type="primary" @click="showNewCharacterDialog">
            <el-icon><Plus /></el-icon>
            新建人物
          </el-button>
        </div>
        <el-table :data="characters">
          <el-table-column prop="name" label="姓名" />
          <el-table-column prop="role" label="角色" />
          <el-table-column prop="gender" label="性别" />
          <el-table-column prop="bio" label="简介" />
          <el-table-column prop="status" label="状态" />
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button size="small" @click="editCharacter(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="deleteCharacter(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 物品 -->
      <el-tab-pane label="物品" name="items">
        <div class="tab-header">
          <el-button type="primary" @click="showNewItemDialog">
            <el-icon><Plus /></el-icon>
            新建物品
          </el-button>
        </div>
        <el-table :data="items">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="type" label="类型" />
          <el-table-column prop="owner" label="持有者" />
          <el-table-column prop="description" label="描述" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="deleteItem(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 地点 -->
      <el-tab-pane label="地点" name="locations">
        <div class="tab-header">
          <el-button type="primary" @click="showNewLocationDialog">
            <el-icon><Plus /></el-icon>
            新建地点
          </el-button>
        </div>
        <el-table :data="locations">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="faction" label="势力" />
          <el-table-column prop="danger" label="危险等级" />
          <el-table-column prop="description" label="描述" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button size="small" type="danger" @click="deleteLocation(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 伏笔 -->
      <el-tab-pane label="伏笔" name="foreshadows">
        <div class="tab-header">
          <el-button type="primary" @click="showNewForeshadowDialog">
            <el-icon><Plus /></el-icon>
            新建伏笔
          </el-button>
        </div>
        <el-table :data="foreshadows">
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="content" label="内容" />
          <el-table-column prop="source_chapter" label="埋设章节" width="80" />
          <el-table-column prop="target_chapter" label="目标章节" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'resolved' ? 'success' : 'warning'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作">
            <template #default="{ row }">
              <el-button size="small" type="success" @click="resolveForeshadow(row)"
                v-if="row.status !== 'resolved'">
                回收
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 新建/编辑人物对话框 -->
    <el-dialog v-model="newCharacterDialogVisible" :title="newCharacter.id ? '编辑人物' : '新建人物'" width="500px">
      <el-form :model="newCharacter" label-width="100px">
        <el-form-item label="姓名">
          <el-input v-model="newCharacter.name" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="newCharacter.role">
            <el-option label="主角" value="主角" />
            <el-option label="配角" value="配角" />
            <el-option label="反派" value="反派" />
            <el-option label="路人" value="路人" />
          </el-select>
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="newCharacter.gender">
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="newCharacter.status">
            <el-option label="存活" value="存活" />
            <el-option label="死亡" value="死亡" />
            <el-option label="失踪" value="失踪" />
          </el-select>
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="newCharacter.bio" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newCharacterDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createCharacter">{{ newCharacter.id ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <!-- 新建伏笔对话框 -->
    <el-dialog v-model="newForeshadowDialogVisible" title="新建伏笔" width="500px">
      <el-form :model="newForeshadow" label-width="100px">
        <el-form-item label="类型">
          <el-select v-model="newForeshadow.type">
            <el-option label="物品" value="item" />
            <el-option label="人物" value="character" />
            <el-option label="剧情" value="plot" />
            <el-option label="悬念" value="mystery" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="newForeshadow.content" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="埋设章节">
          <el-input-number v-model="newForeshadow.source_chapter" :min="1" />
        </el-form-item>
        <el-form-item label="目标章节">
          <el-input-number v-model="newForeshadow.target_chapter" :min="1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newForeshadowDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createForeshadow">创建</el-button>
      </template>
    </el-dialog>

    <!-- 新建物品对话框 -->
    <el-dialog v-model="newItemDialogVisible" title="新建物品" width="500px">
      <el-form :model="newItem" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="newItem.name" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="newItem.type">
            <el-option label="法宝" value="法宝" />
            <el-option label="武器" value="武器" />
            <el-option label="丹药" value="丹药" />
            <el-option label="材料" value="材料" />
          </el-select>
        </el-form-item>
        <el-form-item label="持有者">
          <el-input v-model="newItem.owner" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newItem.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newItemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createItem">创建</el-button>
      </template>
    </el-dialog>

    <!-- 新建地点对话框 -->
    <el-dialog v-model="newLocationDialogVisible" title="新建地点" width="500px">
      <el-form :model="newLocation" label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="newLocation.name" />
        </el-form-item>
        <el-form-item label="所属势力">
          <el-input v-model="newLocation.faction" />
        </el-form-item>
        <el-form-item label="危险等级">
          <el-select v-model="newLocation.danger">
            <el-option label="安全" value="安全" />
            <el-option label="低危" value="低危" />
            <el-option label="中危" value="中危" />
            <el-option label="高危" value="高危" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newLocation.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="newLocationDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createLocation">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { settingsApi, foreshadowApi } from '@/api'

const router = useRouter()
const route = useRoute()
const bookId = computed(() => route.params.id)

const activeTab = ref('worldview')
const worldview = ref({
  basic_info: { genre: '', era: '' },
  core_settings: { power_system: '', social_structure: '' }
})
const characters = ref([])
const items = ref([])
const locations = ref([])
const foreshadows = ref([])

const newCharacterDialogVisible = ref(false)
const newCharacter = ref({ name: '', role: '配角', gender: '男', bio: '', status: '存活', id: null })

const newForeshadowDialogVisible = ref(false)
const newForeshadow = ref({ type: 'plot', content: '', source_chapter: 1, target_chapter: 10 })

const newItemDialogVisible = ref(false)
const newItem = ref({ name: '', type: '法宝', owner: '', description: '' })

const newLocationDialogVisible = ref(false)
const newLocation = ref({ name: '', faction: '', danger: '安全', description: '' })

const goBack = () => {
  router.push(`/books/${bookId.value}`)
}

const loadWorldView = async () => {
  try {
    const res = await settingsApi.getWorldView(bookId.value)
    // Merge with default structure
    const data = res.data || {}
    worldview.value = {
      basic_info: data.basic_info || { genre: '', era: '' },
      core_settings: data.core_settings || { power_system: '', social_structure: '' }
    }
  } catch (error) {
    worldview.value = {
      basic_info: { genre: '', era: '' },
      core_settings: { power_system: '', social_structure: '' }
    }
  }
}

const saveWorldView = async () => {
  try {
    await settingsApi.updateWorldView(bookId.value, worldview.value)
    ElMessage.success('世界观保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const loadCharacters = async () => {
  try {
    const res = await settingsApi.getCharacters(bookId.value)
    characters.value = res.data || []
  } catch (error) {
    characters.value = []
  }
}

const showNewCharacterDialog = () => {
  newCharacter.value = { name: '', role: '配角', gender: '男', bio: '', status: '存活', id: null }
  newCharacterDialogVisible.value = true
}

const editCharacter = async (char) => {
  // 弹出编辑对话框
  newCharacter.value = {
    name: char.name,
    role: char.role,
    gender: char.gender,
    bio: char.bio,
    status: char.status,
    id: char.id
  }
  newCharacterDialogVisible.value = true
}

const deleteCharacter = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此人物？', '删除确认', { type: 'warning' })
    await settingsApi.deleteCharacter(bookId.value, id)
    ElMessage.success('删除成功')
    loadCharacters()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const createCharacter = async () => {
  try {
    if (newCharacter.value.id) {
      // 更新
      await settingsApi.updateCharacter(bookId.value, newCharacter.value.id, newCharacter.value)
      ElMessage.success('人物更新成功')
    } else {
      // 创建
      await settingsApi.createCharacter(bookId.value, newCharacter.value)
      ElMessage.success('人物创建成功')
    }
    newCharacterDialogVisible.value = false
    loadCharacters()
  } catch (error) {
    ElMessage.error(newCharacter.value.id ? '更新失败' : '创建失败')
  }
}

const loadItems = async () => {
  try {
    const res = await settingsApi.getItems(bookId.value)
    items.value = res.data || []
  } catch (error) {
    items.value = []
  }
}

const showNewItemDialog = () => {
  newItem.value = { name: '', type: '法宝', owner: '', description: '' }
  newItemDialogVisible.value = true
}

const createItem = async () => {
  try {
    await settingsApi.createItem(bookId.value, newItem.value)
    ElMessage.success('物品创建成功')
    newItemDialogVisible.value = false
    loadItems()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const deleteItem = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此物品？', '删除确认', { type: 'warning' })
    await settingsApi.deleteItem(bookId.value, id)
    ElMessage.success('删除成功')
    loadItems()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const loadLocations = async () => {
  try {
    const res = await settingsApi.getLocations(bookId.value)
    locations.value = res.data || []
  } catch (error) {
    locations.value = []
  }
}

const showNewLocationDialog = () => {
  newLocation.value = { name: '', faction: '', danger: '安全', description: '' }
  newLocationDialogVisible.value = true
}

const createLocation = async () => {
  try {
    await settingsApi.createLocation(bookId.value, newLocation.value)
    ElMessage.success('地点创建成功')
    newLocationDialogVisible.value = false
    loadLocations()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const deleteLocation = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此地点？', '删除确认', { type: 'warning' })
    await settingsApi.deleteLocation(bookId.value, id)
    ElMessage.success('删除成功')
    loadLocations()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const loadForeshadows = async () => {
  try {
    const res = await foreshadowApi.list(bookId.value)
    foreshadows.value = res.data || []
  } catch (error) {
    foreshadows.value = []
  }
}

const showNewForeshadowDialog = () => {
  newForeshadow.value = { type: 'plot', content: '', source_chapter: 1, target_chapter: 10 }
  newForeshadowDialogVisible.value = true
}

const createForeshadow = async () => {
  try {
    await foreshadowApi.create(bookId.value, newForeshadow.value)
    ElMessage.success('伏笔创建成功')
    newForeshadowDialogVisible.value = false
    loadForeshadows()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const resolveForeshadow = async (fs) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入回收章节号', '回收伏笔', {
      inputType: 'number',
      inputValue: fs.target_chapter?.toString() || '10'
    })
    await foreshadowApi.resolve(bookId.value, fs.id, parseInt(value))
    ElMessage.success('伏笔已回收')
    loadForeshadows()
  } catch (error) {}
}

onMounted(() => {
  loadWorldView()
  loadCharacters()
  loadItems()
  loadLocations()
  loadForeshadows()
})
</script>

<style scoped>
.settings-view {
  max-width: 1200px;
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

.tab-header {
  margin-bottom: 15px;
}
</style>