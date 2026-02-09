<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>角色库</h1>
            <span class="subtitle">管理可复用的角色形象素材</span>
          </div>
        </template>
        <template #right>
          <el-button type="primary" @click="openCreateDialog">新增角色素材</el-button>
        </template>
      </AppHeader>

      <div class="filters">
        <el-input v-model="query.keyword" placeholder="按名称搜索" clearable @change="loadItems" />
        <el-input v-model="query.category" placeholder="分类（可选）" clearable @change="loadItems" />
      </div>

      <div class="table-wrapper" v-loading="loading">
        <el-table :data="items" stripe>
          <el-table-column label="预览" width="90">
            <template #default="{ row }">
              <el-image
                :src="row.image_url"
                style="width: 56px; height: 56px; border-radius: 8px"
                fit="cover"
                :preview-src-list="[row.image_url]"
                preview-teleported
              />
            </template>
          </el-table-column>

          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="category" label="分类" min-width="120" />
          <el-table-column prop="source_type" label="来源" min-width="100" />
          <el-table-column prop="updated_at" label="更新时间" min-width="180" />
          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" link @click="deleteItem(row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadItems"
          @size-change="loadItems"
        />
      </div>
    </div>

    <el-dialog v-model="createDialogVisible" title="新增角色素材" width="520px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-position="top">
        <el-form-item label="名称" prop="name">
          <el-input v-model="createForm.name" placeholder="例如：女主角（现代装）" />
        </el-form-item>
        <el-form-item label="图片 URL" prop="image_url">
          <el-input v-model="createForm.image_url" placeholder="请输入可访问图片地址" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="createForm.category" placeholder="可选，如：主角 / 配角" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="createForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createItem">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { AppHeader } from '@/components/common'
import { characterLibraryAPI, type CharacterLibraryItem } from '@/api/character-library'

const loading = ref(false)
const items = ref<CharacterLibraryItem[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  page_size: 10,
  category: '',
  keyword: ''
})

const loadItems = async () => {
  loading.value = true
  try {
    const res = await characterLibraryAPI.list({
      page: query.page,
      page_size: query.page_size,
      category: query.category || undefined,
      keyword: query.keyword || undefined
    })
    items.value = res.items || []
    total.value = res.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error?.message || '加载角色库失败')
  } finally {
    loading.value = false
  }
}

const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = reactive({
  name: '',
  image_url: '',
  category: '',
  description: ''
})

const createRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  image_url: [{ required: true, message: '请输入图片地址', trigger: 'blur' }]
}

const openCreateDialog = () => {
  createDialogVisible.value = true
  createForm.name = ''
  createForm.image_url = ''
  createForm.category = ''
  createForm.description = ''
}

const createItem = async () => {
  if (!createFormRef.value) return
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  creating.value = true
  try {
    await characterLibraryAPI.create({
      name: createForm.name,
      image_url: createForm.image_url,
      category: createForm.category || undefined,
      description: createForm.description || undefined,
      source_type: 'upload'
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    await loadItems()
  } catch (error: any) {
    ElMessage.error(error?.message || '创建失败')
  } finally {
    creating.value = false
  }
}

const deleteItem = async (id: string) => {
  try {
    await ElMessageBox.confirm('确认删除该角色素材吗？', '提示', { type: 'warning' })
    await characterLibraryAPI.delete(id)
    ElMessage.success('删除成功')
    await loadItems()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error?.message || '删除失败')
    }
  }
}

onMounted(loadItems)
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
}

.content-wrapper {
  width: 100%;
}

.page-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-title h1 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.filters {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  padding: 12px;
}

.table-wrapper {
  padding: 0 12px;
}

.pagination {
  padding: 12px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
