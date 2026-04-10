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

      <!-- 势力 -->
      <el-tab-pane label="势力" name="factions">
        <div class="tab-header">
          <el-button type="primary" @click="showNewFactionDialog">
            <el-icon><Plus /></el-icon>
            新建势力
          </el-button>
        </div>
        <el-table :data="factions">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type || '势力' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="leader" label="首领" />
          <el-table-column label="成员数" width="80">
            <template #default="{ row }">
              {{ row.members?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="领地数" width="80">
            <template #default="{ row }">
              {{ row.territories?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="关系数" width="80">
            <template #default="{ row }">
              {{ row.relations?.length || 0 }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" @click="editFaction(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteFaction(row.id)">删除</el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
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
          <el-table-column prop="role" label="角色" width="80" />
          <el-table-column prop="faction" label="势力" />
          <el-table-column prop="sect" label="宗门" />
          <el-table-column prop="cultivation" label="境界" />
          <el-table-column prop="status" label="状态" width="80" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" @click="editCharacter(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteCharacter(row.id)">删除</el-button>
              </el-button-group>
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
          <el-table-column prop="type" label="类型" width="80" />
          <el-table-column prop="rank" label="品阶" width="80" />
          <el-table-column prop="owner" label="持有者" />
          <el-table-column prop="faction" label="势力" />
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" @click="editItem(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteItem(row.id)">删除</el-button>
              </el-button-group>
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
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <el-button-group>
                <el-button size="small" @click="editLocation(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteLocation(row.id)">删除</el-button>
              </el-button-group>
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
    <el-dialog v-model="newCharacterDialogVisible" :title="newCharacter.id ? '编辑人物' : '新建人物'" width="600px">
      <el-form :model="newCharacter" label-width="100px">
        <el-form-item label="姓名" required>
          <el-input v-model="newCharacter.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="角色" required>
          <el-select v-model="newCharacter.role" allow-create filterable placeholder="选择或输入角色">
            <el-option label="主角" value="主角" />
            <el-option label="配角" value="配角" />
            <el-option label="反派" value="反派" />
            <el-option label="路人" value="路人" />
          </el-select>
        </el-form-item>
        <el-form-item label="性别">
          <el-select v-model="newCharacter.gender" allow-create filterable placeholder="选择或输入性别">
            <el-option label="男" value="男" />
            <el-option label="女" value="女" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属势力">
          <el-input v-model="newCharacter.faction" placeholder="所属势力/组织" />
        </el-form-item>
        <el-form-item label="宗门/门派">
          <el-input v-model="newCharacter.sect" placeholder="宗门或门派" />
        </el-form-item>
        <el-form-item label="职位/身份">
          <el-input v-model="newCharacter.position" placeholder="职位或身份" />
        </el-form-item>
        <el-form-item label="修为境界">
          <el-input v-model="newCharacter.cultivation" placeholder="修为境界" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="newCharacter.status" allow-create filterable placeholder="选择或输入状态">
            <el-option label="存活" value="存活" />
            <el-option label="死亡" value="死亡" />
            <el-option label="失踪" value="失踪" />
          </el-select>
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="newCharacter.bio" type="textarea" :rows="3" />
        </el-form-item>

        <!-- 自定义属性 -->
        <el-divider content-position="left">自定义属性</el-divider>
        <div class="custom-attrs">
          <div v-for="(value, key, index) in newCharacter.custom_attributes" :key="index" class="custom-attr-item">
            <el-input v-model="Object.keys(newCharacter.custom_attributes)[index]" placeholder="属性名" style="width: 120px;" @change="updateCustomAttrKey(newCharacter, index, $event)" />
            <el-input v-model="newCharacter.custom_attributes[key]" placeholder="属性值" style="flex: 1;" />
            <el-button type="danger" size="small" @click="removeCustomAttr(newCharacter, key)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addCustomAttr(newCharacter)">添加属性</el-button>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="newCharacterDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createCharacter">{{ newCharacter.id ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <!-- 新建伏笔对话框 -->
    <el-dialog v-model="newForeshadowDialogVisible" title="新建伏笔" width="500px">
      <el-form :model="newForeshadow" label-width="100px">
        <el-form-item label="类型" required>
          <el-select v-model="newForeshadow.type">
            <el-option label="物品" value="item" />
            <el-option label="人物" value="character" />
            <el-option label="剧情" value="plot" />
            <el-option label="悬念" value="mystery" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容" required>
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
    <el-dialog v-model="newItemDialogVisible" :title="newItem.id ? '编辑物品' : '新建物品'" width="600px">
      <el-form :model="newItem" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="newItem.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="newItem.type" allow-create filterable placeholder="选择或输入类型">
            <el-option label="法宝" value="法宝" />
            <el-option label="武器" value="武器" />
            <el-option label="丹药" value="丹药" />
            <el-option label="材料" value="材料" />
            <el-option label="功法" value="功法" />
            <el-option label="秘籍" value="秘籍" />
          </el-select>
        </el-form-item>
        <el-form-item label="品阶">
          <el-select v-model="newItem.rank" allow-create filterable placeholder="选择或输入品阶">
            <el-option label="天阶" value="天阶" />
            <el-option label="地阶" value="地阶" />
            <el-option label="玄阶" value="玄阶" />
            <el-option label="黄阶" value="黄阶" />
          </el-select>
        </el-form-item>
        <el-form-item label="持有者">
          <el-input v-model="newItem.owner" />
        </el-form-item>
        <el-form-item label="所属势力">
          <el-input v-model="newItem.faction" placeholder="所属势力" />
        </el-form-item>
        <el-form-item label="所属宗门">
          <el-input v-model="newItem.sect" placeholder="所属宗门" />
        </el-form-item>
        <el-form-item label="所在地点">
          <el-select v-model="newItem.location" placeholder="所在地点" clearable filterable>
            <el-option v-for="loc in locations" :key="loc.id" :label="loc.name" :value="loc.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newItem.description" type="textarea" :rows="3" />
        </el-form-item>

        <!-- 自定义属性 -->
        <el-divider content-position="left">自定义属性</el-divider>
        <div class="custom-attrs">
          <div v-for="(value, key, index) in newItem.custom_attributes" :key="index" class="custom-attr-item">
            <el-input :model-value="key" placeholder="属性名" style="width: 120px;" @change="updateCustomAttrKey(newItem, index, $event)" />
            <el-input v-model="newItem.custom_attributes[key]" placeholder="属性值" style="flex: 1;" />
            <el-button type="danger" size="small" @click="removeCustomAttr(newItem, key)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addCustomAttr(newItem)">添加属性</el-button>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="newItemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createItem">{{ newItem.id ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <!-- 新建地点对话框 -->
    <el-dialog v-model="newLocationDialogVisible" :title="newLocation.id ? '编辑地点' : '新建地点'" width="500px">
      <el-form :model="newLocation" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="newLocation.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="所属势力">
          <el-input v-model="newLocation.faction" />
        </el-form-item>
        <el-form-item label="危险等级">
          <el-select v-model="newLocation.danger" allow-create filterable placeholder="选择或输入危险等级">
            <el-option label="安全" value="安全" />
            <el-option label="低危" value="低危" />
            <el-option label="中危" value="中危" />
            <el-option label="高危" value="高危" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newLocation.description" type="textarea" :rows="3" />
        </el-form-item>

        <!-- 自定义属性 -->
        <el-divider content-position="left">自定义属性</el-divider>
        <div class="custom-attrs">
          <div v-for="(value, key, index) in newLocation.custom_attributes" :key="index" class="custom-attr-item">
            <el-input :model-value="key" placeholder="属性名" style="width: 120px;" @change="updateCustomAttrKey(newLocation, index, $event)" />
            <el-input v-model="newLocation.custom_attributes[key]" placeholder="属性值" style="flex: 1;" />
            <el-button type="danger" size="small" @click="removeCustomAttr(newLocation, key)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addCustomAttr(newLocation)">添加属性</el-button>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="newLocationDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createLocation">{{ newLocation.id ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <!-- 新建/编辑势力对话框 -->
    <el-dialog v-model="newFactionDialogVisible" :title="newFaction.id ? '编辑势力' : '新建势力'" width="700px">
      <el-form :model="newFaction" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="newFaction.name" placeholder="必填" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="newFaction.type" allow-create filterable placeholder="选择或输入类型">
            <el-option label="宗门" value="宗门" />
            <el-option label="家族" value="家族" />
            <el-option label="帮派" value="帮派" />
            <el-option label="帝国" value="帝国" />
            <el-option label="商会" value="商会" />
            <el-option label="势力" value="势力" />
          </el-select>
        </el-form-item>
        <el-form-item label="首领">
          <el-select v-model="newFaction.leader" placeholder="选择首领" clearable filterable>
            <el-option v-for="char in characters" :key="char.id" :label="char.name" :value="char.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newFaction.description" type="textarea" :rows="3" />
        </el-form-item>

        <!-- 势力关系 -->
        <el-divider content-position="left">势力关系</el-divider>
        <el-form-item label="关联势力">
          <div class="faction-relations">
            <div v-for="(rel, index) in newFaction.relations" :key="index" class="relation-item">
              <el-select v-model="rel.name" placeholder="势力名称" filterable style="width: 150px;">
                <el-option v-for="f in otherFactions" :key="f.id" :label="f.name" :value="f.name" />
              </el-select>
              <el-select v-model="rel.type" placeholder="关系类型" style="width: 120px;">
                <el-option label="联盟" value="ally" />
                <el-option label="敌对" value="enemy" />
                <el-option label="附属" value="subordinate" />
                <el-option label="中立" value="neutral" />
              </el-select>
              <el-button type="danger" size="small" @click="removeFactionRelation(index)">删除</el-button>
            </div>
            <el-button type="primary" size="small" @click="addFactionRelation">添加关系</el-button>
          </div>
        </el-form-item>

        <!-- 成员列表 -->
        <el-divider content-position="left">成员列表</el-divider>
        <el-form-item label="成员">
          <div class="faction-members">
            <el-select v-model="newFaction.members" multiple placeholder="选择成员" filterable style="width: 100%;">
              <el-option v-for="char in characters" :key="char.id" :label="char.name" :value="char.name" />
            </el-select>
          </div>
        </el-form-item>

        <!-- 领地列表 -->
        <el-divider content-position="left">领地列表</el-divider>
        <el-form-item label="领地">
          <div class="faction-territories">
            <el-select v-model="newFaction.territories" multiple placeholder="选择领地" filterable style="width: 100%;">
              <el-option v-for="loc in locations" :key="loc.id" :label="loc.name" :value="loc.name" />
            </el-select>
          </div>
        </el-form-item>

        <!-- 自定义属性 -->
        <el-divider content-position="left">自定义属性</el-divider>
        <div class="custom-attrs">
          <div v-for="(value, key, index) in newFaction.custom_attributes" :key="index" class="custom-attr-item">
            <el-input :model-value="key" placeholder="属性名" style="width: 120px;" @change="updateCustomAttrKey(newFaction, index, $event)" />
            <el-input v-model="newFaction.custom_attributes[key]" placeholder="属性值" style="flex: 1;" />
            <el-button type="danger" size="small" @click="removeCustomAttr(newFaction, key)">删除</el-button>
          </div>
          <el-button type="primary" size="small" @click="addCustomAttr(newFaction)">添加属性</el-button>
        </div>
      </el-form>
      <template #footer>
        <el-button @click="newFactionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createFaction">{{ newFaction.id ? '保存' : '创建' }}</el-button>
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
const factions = ref([])

const newCharacterDialogVisible = ref(false)
const newCharacter = ref({ name: '', role: '配角', gender: '男', bio: '', status: '存活', faction: '', sect: '', position: '', cultivation: '', custom_attributes: {}, id: null })

const newForeshadowDialogVisible = ref(false)
const newForeshadow = ref({ type: 'plot', content: '', source_chapter: 1, target_chapter: 10 })

const newItemDialogVisible = ref(false)
const newItem = ref({ name: '', type: '法宝', owner: '', description: '', rank: '', faction: '', sect: '', location: '', custom_attributes: {}, id: null })

const newLocationDialogVisible = ref(false)
const newLocation = ref({ name: '', faction: '', danger: '安全', description: '', custom_attributes: {}, id: null })

const newFactionDialogVisible = ref(false)
const newFaction = ref({ name: '', type: '势力', leader: '', description: '', relations: [], members: [], territories: [], custom_attributes: {}, id: null })

// 其他势力（用于关系选择）
const otherFactions = computed(() => {
  if (!newFaction.value.id) return factions.value
  return factions.value.filter(f => f.id !== newFaction.value.id)
})

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
  newCharacter.value = { name: '', role: '配角', gender: '男', bio: '', status: '存活', faction: '', sect: '', position: '', cultivation: '', custom_attributes: {}, id: null }
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
    faction: char.faction || '',
    sect: char.sect || '',
    position: char.position || '',
    cultivation: char.cultivation || '',
    custom_attributes: char.custom_attributes || {},
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
  newItem.value = { name: '', type: '法宝', owner: '', description: '', rank: '', faction: '', sect: '', location: '', id: null }
  newItemDialogVisible.value = true
}

const editItem = (item) => {
  newItem.value = {
    name: item.name,
    type: item.type,
    owner: item.owner,
    description: item.description,
    rank: item.rank || '',
    faction: item.faction || '',
    sect: item.sect || '',
    location: item.location || '',
    custom_attributes: item.custom_attributes || {},
    id: item.id
  }
  newItemDialogVisible.value = true
}

const createItem = async () => {
  try {
    if (newItem.value.id) {
      await settingsApi.updateItem(bookId.value, newItem.value.id, newItem.value)
      ElMessage.success('物品更新成功')
    } else {
      await settingsApi.createItem(bookId.value, newItem.value)
      ElMessage.success('物品创建成功')
    }
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
  newLocation.value = { name: '', faction: '', danger: '安全', description: '', id: null }
  newLocationDialogVisible.value = true
}

const editLocation = (loc) => {
  newLocation.value = {
    name: loc.name,
    faction: loc.faction || '',
    danger: loc.danger || '安全',
    description: loc.description,
    custom_attributes: loc.custom_attributes || {},
    id: loc.id
  }
  newLocationDialogVisible.value = true
}

const createLocation = async () => {
  try {
    if (newLocation.value.id) {
      await settingsApi.updateLocation(bookId.value, newLocation.value.id, newLocation.value)
      ElMessage.success('地点更新成功')
    } else {
      await settingsApi.createLocation(bookId.value, newLocation.value)
      ElMessage.success('地点创建成功')
    }
    newLocationDialogVisible.value = false
    loadLocations()
  } catch (error) {
    ElMessage.error(newLocation.value.id ? '更新失败' : '创建失败')
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

// ========== 势力管理 ==========

const loadFactions = async () => {
  try {
    const res = await settingsApi.getWorldView(bookId.value)
    const data = res.data || {}
    factions.value = data.key_elements?.factions || []
  } catch (error) {
    factions.value = []
  }
}

const showNewFactionDialog = () => {
  newFaction.value = { name: '', type: '势力', leader: '', description: '', relations: [], members: [], territories: [], id: null }
  newFactionDialogVisible.value = true
}

const editFaction = (faction) => {
  newFaction.value = {
    name: faction.name,
    type: faction.type || '势力',
    leader: faction.leader || '',
    description: faction.description || '',
    relations: faction.relations?.map(r => ({
      name: r.name || (typeof r === 'string' ? r : ''),
      type: r.type || 'neutral'
    })) || [],
    members: faction.members || [],
    territories: faction.territories || [],
    custom_attributes: faction.custom_attributes || {},
    id: faction.id
  }
  newFactionDialogVisible.value = true
}

const createFaction = async () => {
  try {
    // 验证必填字段
    if (!newFaction.value.name) {
      ElMessage.error('请填写势力名称')
      return
    }
    if (!newFaction.value.type) {
      ElMessage.error('请选择势力类型')
      return
    }

    // 加载当前世界观数据
    const res = await settingsApi.getWorldView(bookId.value)
    const worldviewData = res.data || {
      basic_info: { genre: '', era: '' },
      core_settings: { power_system: '', social_structure: '' },
      key_elements: { factions: [] }
    }

    if (!worldviewData.key_elements) {
      worldviewData.key_elements = { factions: [] }
    }
    if (!worldviewData.key_elements.factions) {
      worldviewData.key_elements.factions = []
    }

    // 过滤掉空的关系
    const validRelations = newFaction.value.relations.filter(r => r.name && r.type)

    if (newFaction.value.id) {
      // 更新
      const idx = worldviewData.key_elements.factions.findIndex(f => f.id === newFaction.value.id)
      if (idx >= 0) {
        worldviewData.key_elements.factions[idx] = {
          id: newFaction.value.id,
          name: newFaction.value.name,
          type: newFaction.value.type,
          leader: newFaction.value.leader || '',
          description: newFaction.value.description || '',
          relations: validRelations,
          members: newFaction.value.members || [],
          territories: newFaction.value.territories || [],
          custom_attributes: newFaction.value.custom_attributes || {}
        }
      }
      ElMessage.success('势力更新成功')
    } else {
      // 创建
      const newId = 'faction_' + Date.now()
      worldviewData.key_elements.factions.push({
        id: newId,
        name: newFaction.value.name,
        type: newFaction.value.type,
        leader: newFaction.value.leader || '',
        description: newFaction.value.description || '',
        relations: validRelations,
        members: newFaction.value.members || [],
        territories: newFaction.value.territories || [],
        custom_attributes: newFaction.value.custom_attributes || {}
      })
      ElMessage.success('势力创建成功')
    }

    // 保存世界观
    await settingsApi.updateWorldView(bookId.value, worldviewData)
    newFactionDialogVisible.value = false
    loadFactions()
  } catch (error) {
    console.error('创建势力失败:', error)
    ElMessage.error(newFaction.value.id ? '更新失败' : '创建失败')
  }
}

const deleteFaction = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此势力？', '删除确认', { type: 'warning' })

    // 加载当前世界观数据
    const res = await settingsApi.getWorldView(bookId.value)
    const worldviewData = res.data

    if (worldviewData?.key_elements?.factions) {
      worldviewData.key_elements.factions = worldviewData.key_elements.factions.filter(f => f.id !== id)
      await settingsApi.updateWorldView(bookId.value, worldviewData)
    }

    ElMessage.success('删除成功')
    loadFactions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const addFactionRelation = () => {
  newFaction.value.relations.push({ name: '', type: 'neutral' })
}

const removeFactionRelation = (index) => {
  newFaction.value.relations.splice(index, 1)
}

// ========== 自定义属性管理 ==========
const addCustomAttr = (entity) => {
  if (!entity.custom_attributes) {
    entity.custom_attributes = {}
  }
  const key = `属性${Object.keys(entity.custom_attributes).length + 1}`
  entity.custom_attributes[key] = ''
}

const removeCustomAttr = (entity, key) => {
  if (entity.custom_attributes) {
    delete entity.custom_attributes[key]
  }
}

const updateCustomAttrKey = (entity, index, newKey) => {
  const keys = Object.keys(entity.custom_attributes || {})
  const oldKey = keys[index]
  if (oldKey && oldKey !== newKey) {
    const value = entity.custom_attributes[oldKey]
    delete entity.custom_attributes[oldKey]
    entity.custom_attributes[newKey] = value
  }
}

onMounted(() => {
  loadWorldView()
  loadCharacters()
  loadItems()
  loadLocations()
  loadForeshadows()
  loadFactions()
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

.faction-relations {
  width: 100%;
}

.relation-item {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
  align-items: center;
}

.faction-members,
.faction-territories {
  width: 100%;
}

.custom-attrs {
  width: 100%;
}

.custom-attr-item {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
  align-items: center;
}
</style>