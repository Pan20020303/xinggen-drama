<template>
  <div class="admin-page">
    <header class="admin-header">
      <div class="header-left">
        <h1>模型配置</h1>
        <p>平台级 AI 服务配置（仅管理员可编辑，密钥不会明文回显）</p>
      </div>
      <div class="header-right">
        <LanguageSwitcher />
        <ThemeToggle />
        <el-button @click="router.push('/admin/users')">用户管理</el-button>
        <el-button @click="router.push('/admin/billing')">计费管理</el-button>
        <el-button type="danger" @click="handleLogout">退出管理端</el-button>
      </div>
    </header>

    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <el-tabs v-model="activeTab" class="tabs" @tab-change="loadConfigs">
            <el-tab-pane label="文本" name="text" />
            <el-tab-pane label="图片" name="image" />
            <el-tab-pane label="视频" name="video" />
          </el-tabs>
          <div class="header-actions">
            <el-button :loading="loading" @click="loadConfigs">刷新</el-button>
            <el-button type="primary" @click="openCreateDialog">新增配置</el-button>
          </div>
        </div>
      </template>

      <el-table :data="configs" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" min-width="80" />
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column prop="provider" label="Provider" min-width="120" />
        <el-table-column prop="base_url" label="BaseURL" min-width="220" show-overflow-tooltip />
        <el-table-column label="Models" min-width="220">
          <template #default="{ row }">
            <div class="models-cell">
              <el-tag v-for="m in row.model" :key="m" size="small" type="info" class="model-tag">{{ m }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="priority" label="优先级" min-width="90" />
        <el-table-column label="密钥" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.api_key_set ? 'success' : 'warning'">
              {{ row.api_key_set ? '已设置' : '未设置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_default ? 'success' : 'info'">
              {{ row.is_default ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用" min-width="110">
          <template #default="{ row }">
            <el-switch
              v-model="row.is_active"
              :loading="updatingIds.has(row.id)"
              @change="(val: boolean) => toggleActive(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑配置' : '新增配置'" width="640px" :close-on-click-modal="false">
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="服务类型" prop="service_type">
          <el-select v-model="form.service_type" style="width: 100%" :disabled="isEdit">
            <el-option label="文本" value="text" />
            <el-option label="图片" value="image" />
            <el-option label="视频" value="video" />
          </el-select>
        </el-form-item>

        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="例如：OpenAI-Text-01" />
        </el-form-item>

        <el-form-item label="Provider" prop="provider">
          <el-select v-model="form.provider" filterable allow-create default-first-option style="width: 100%">
            <el-option v-for="p in providerOptions" :key="p" :label="p" :value="p" />
          </el-select>
        </el-form-item>

        <el-form-item label="BaseURL" prop="base_url">
          <el-input v-model="form.base_url" placeholder="例如：https://api.openai.com/v1" />
        </el-form-item>

        <el-form-item label="API Key" prop="api_key">
          <el-input
            v-model="form.api_key"
            type="password"
            show-password
            :placeholder="apiKeyPlaceholder"
          />
          <div class="field-tip">
            {{ isEdit ? '留空表示不修改现有密钥；如需测试连接，请输入密钥后点击“测试连接”。' : '创建时必须提供密钥。' }}
          </div>
        </el-form-item>

        <el-form-item label="Models" prop="model">
          <el-select
            v-model="form.model"
            multiple
            filterable
            allow-create
            default-first-option
            collapse-tags
            collapse-tags-tooltip
            style="width: 100%"
          >
            <el-option v-for="m in suggestedModels" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>

        <el-form-item label="优先级" prop="priority">
          <el-input-number v-model="form.priority" :min="0" :max="100" style="width: 100%" />
        </el-form-item>

        <el-form-item label="默认" prop="is_default">
          <el-switch v-model="form.is_default" />
        </el-form-item>

        <el-form-item label="启用" prop="is_active">
          <el-switch v-model="form.is_active" />
        </el-form-item>

        <el-form-item label="Endpoint (可选)">
          <el-input v-model="form.endpoint" placeholder="留空则由后端根据 provider 自动设置" />
        </el-form-item>

        <el-form-item v-if="form.service_type === 'video'" label="Query Endpoint (视频异步查询，允许为空)">
          <el-input v-model="form.query_endpoint" placeholder="例如：/generations/tasks/{taskId}" />
        </el-form-item>

        <el-form-item label="Settings (JSON，可选)">
          <el-input v-model="form.settings" type="textarea" :rows="3" placeholder="可留空" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button :loading="testing" @click="handleTestConnection">测试连接</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import ThemeToggle from '@/components/common/ThemeToggle.vue'
import { adminAPI } from '@/api/admin'
import { useAdminAuthStore } from '@/stores/adminAuth'
import type { AdminAIServiceConfigView, AdminAIServiceType } from '@/types/admin'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()

const activeTab = ref<AdminAIServiceType>('text')
const loading = ref(false)
const configs = ref<AdminAIServiceConfigView[]>([])
const updatingIds = ref<Set<number>>(new Set())

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const testing = ref(false)
const formRef = ref<FormInstance>()

const providerOptions = ['openai', 'gemini', 'google', 'chatfire', 'doubao', 'volcengine', 'volces', 'runway', 'pika', 'minimax']

const suggestedModels = computed(() => {
  if (activeTab.value === 'image') return ['gpt-image-1', 'doubao-vision', 'gemini-2.0-flash-exp']
  if (activeTab.value === 'video') return ['doubao-seedance-1-5-pro-251215', 'runway', 'pika', 'MiniMax-Hailuo-02']
  return ['gpt-4o-mini', 'gpt-4o', 'gemini-2.0-flash']
})

const form = reactive({
  id: 0,
  service_type: 'text' as AdminAIServiceType,
  name: '',
  provider: '',
  base_url: '',
  api_key: '',
  api_key_set: false,
  model: [] as string[],
  endpoint: '',
  query_endpoint: '',
  priority: 0,
  is_default: false,
  is_active: true,
  settings: ''
})

const apiKeyPlaceholder = computed(() => {
  if (!isEdit.value) return '必填'
  return form.api_key_set ? '已设置(不回显)，留空不修改' : '未设置，请输入密钥'
})

const rules: FormRules = {
  service_type: [{ required: true, message: '请选择服务类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  provider: [{ required: true, message: '请输入 provider', trigger: 'change' }],
  base_url: [{ required: true, message: '请输入 base_url', trigger: 'blur' }],
  model: [{ required: true, message: '请至少填写一个模型', trigger: 'change' }]
}

const loadConfigs = async () => {
  loading.value = true
  try {
    const list = await adminAPI.listAIConfigs({ service_type: activeTab.value })
    configs.value = list || []
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '加载配置失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.id = 0
  form.service_type = activeTab.value
  form.name = ''
  form.provider = ''
  form.base_url = ''
  form.api_key = ''
  form.api_key_set = false
  form.model = []
  form.endpoint = ''
  form.query_endpoint = ''
  form.priority = 0
  form.is_default = false
  form.is_active = true
  form.settings = ''
}

const openCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row: AdminAIServiceConfigView) => {
  isEdit.value = true
  form.id = row.id
  form.service_type = row.service_type
  form.name = row.name
  form.provider = row.provider
  form.base_url = row.base_url
  form.api_key = ''
  form.api_key_set = Boolean(row.api_key_set)
  form.model = Array.isArray(row.model) ? row.model.slice() : []
  form.endpoint = row.endpoint || ''
  form.query_endpoint = row.query_endpoint || ''
  form.priority = row.priority || 0
  form.is_default = Boolean(row.is_default)
  form.is_active = Boolean(row.is_active)
  form.settings = row.settings || ''
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // 创建时必须有密钥
  if (!isEdit.value && !form.api_key.trim()) {
    ElMessage.warning('创建配置时必须填写 API Key')
    return
  }

  submitting.value = true
  try {
    if (!isEdit.value) {
      await adminAPI.createAIConfig({
        service_type: form.service_type,
        name: form.name,
        provider: form.provider,
        base_url: form.base_url,
        api_key: form.api_key,
        model: form.model,
        endpoint: form.endpoint || undefined,
        query_endpoint: form.query_endpoint || undefined,
        priority: form.priority,
        is_default: form.is_default,
        settings: form.settings || undefined
      })
      ElMessage.success('创建成功')
    } else {
      await adminAPI.updateAIConfig(form.id, {
        name: form.name,
        provider: form.provider,
        base_url: form.base_url,
        api_key: form.api_key.trim() ? form.api_key : undefined,
        model: form.model,
        endpoint: form.endpoint || undefined,
        query_endpoint: form.query_endpoint || '',
        priority: form.priority,
        is_default: form.is_default,
        is_active: form.is_active,
        settings: form.settings || undefined
      })
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    await loadConfigs()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '保存失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row: AdminAIServiceConfigView) => {
  try {
    await ElMessageBox.confirm(`确认删除配置「${row.name}」吗？`, '提示', { type: 'warning' })
    await adminAPI.deleteAIConfig(row.id)
    ElMessage.success('删除成功')
    await loadConfigs()
  } catch {
    return
  }
}

const toggleActive = async (row: AdminAIServiceConfigView, val: boolean) => {
  updatingIds.value.add(row.id)
  try {
    await adminAPI.updateAIConfig(row.id, {
      name: row.name,
      provider: row.provider,
      base_url: row.base_url,
      model: Array.isArray(row.model) ? row.model : [],
      endpoint: row.endpoint || undefined,
      query_endpoint: row.query_endpoint || '',
      priority: row.priority || 0,
      is_default: Boolean(row.is_default),
      is_active: Boolean(val),
      settings: row.settings || undefined
    })
    ElMessage.success('已更新')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '更新失败')
    row.is_active = !val
  } finally {
    updatingIds.value.delete(row.id)
  }
}

const handleTestConnection = async () => {
  // 只能用用户输入的密钥测试（后端不会回显密钥）
  if (!form.api_key.trim()) {
    ElMessage.warning('请先输入 API Key 再测试')
    return
  }

  if (!form.model || form.model.length === 0) {
    ElMessage.warning('请先填写至少一个模型再测试')
    return
  }

  testing.value = true
  try {
    await adminAPI.testAIConfig({
      base_url: form.base_url,
      api_key: form.api_key,
      model: form.model,
      provider: form.provider,
      endpoint: form.endpoint || undefined
    })
    ElMessage.success('连接测试成功')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '连接测试失败')
  } finally {
    testing.value = false
  }
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确认退出管理端账号吗？', '提示', { type: 'warning' })
    adminAuthStore.logout()
    await router.replace('/admin/login')
  } catch {
    return
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
.admin-page {
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 16px;
}

.admin-header {
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  padding: 14px 16px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.header-left h1 {
  margin: 0;
  font-size: 22px;
}

.header-left p {
  margin: 4px 0 0;
  color: var(--text-muted);
  font-size: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.main-card {
  border-radius: var(--radius-xl);
  border: 1px solid var(--border-primary);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tabs {
  flex: 1;
  min-width: 320px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.models-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.model-tag {
  max-width: 220px;
}

.field-tip {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-muted);
}
</style>

