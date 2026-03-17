<template>
  <div class="admin-page">
    <header class="admin-header">
      <div class="header-left">
        <h1>模型 Token 统计</h1>
        <p>从本次改造上线后开始精确统计，各模型的 prompt/completion/total token 消耗。</p>
      </div>
      <div class="header-right">
        <LanguageSwitcher />
        <ThemeToggle />
        <el-button type="danger" @click="handleLogout">退出</el-button>
      </div>
    </header>

    <el-card class="filter-card">
      <div class="filters">
        <el-select v-model="filters.service_type" clearable placeholder="服务类型" style="width: 180px">
          <el-option label="文本" value="text" />
          <el-option label="图片" value="image" />
          <el-option label="视频" value="video" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="YYYY-MM-DD"
        />
        <el-button :loading="loading" @click="loadStats">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </el-card>

    <div class="summary-grid">
      <el-card class="summary-card">
        <div class="summary-label">总 Prompt Tokens</div>
        <div class="summary-value">{{ summary.prompt_tokens }}</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">总 Completion Tokens</div>
        <div class="summary-value">{{ summary.completion_tokens }}</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">总 Total Tokens</div>
        <div class="summary-value">{{ summary.total_tokens }}</div>
      </el-card>
      <el-card class="summary-card">
        <div class="summary-label">模型数</div>
        <div class="summary-value">{{ summary.model_count }}</div>
      </el-card>
    </div>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>按模型汇总</span>
          <span class="muted">仅统计已写入 token usage 的新调用</span>
        </div>
      </template>

      <el-table :data="items" v-loading="loading" stripe>
        <el-table-column prop="model" label="模型" min-width="220" />
        <el-table-column prop="service_type" label="服务类型" min-width="110" />
        <el-table-column prop="calls" label="调用次数" min-width="100" />
        <el-table-column prop="prompt_tokens" label="Prompt Tokens" min-width="140" />
        <el-table-column prop="completion_tokens" label="Completion Tokens" min-width="170" />
        <el-table-column prop="total_tokens" label="Total Tokens" min-width="140" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import ThemeToggle from '@/components/common/ThemeToggle.vue'
import { adminAPI } from '@/api/admin'
import { useAdminAuthStore } from '@/stores/adminAuth'
import type { AdminTokenStatsItem, AdminTokenStatsSummary } from '@/types/admin'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()

const loading = ref(false)
const items = ref<AdminTokenStatsItem[]>([])
const summary = reactive<AdminTokenStatsSummary>({
  prompt_tokens: 0,
  completion_tokens: 0,
  total_tokens: 0,
  model_count: 0
})
const filters = reactive<{
  service_type?: 'text' | 'image' | 'video'
}>({
  service_type: undefined
})
const dateRange = ref<[string, string] | []>([])

const loadStats = async () => {
  loading.value = true
  try {
    const data = await adminAPI.getTokenStats({
      service_type: filters.service_type,
      start_date: dateRange.value[0],
      end_date: dateRange.value[1]
    })
    items.value = data.items || []
    summary.prompt_tokens = data.summary?.prompt_tokens || 0
    summary.completion_tokens = data.summary?.completion_tokens || 0
    summary.total_tokens = data.summary?.total_tokens || 0
    summary.model_count = data.summary?.model_count || 0
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '加载 token 统计失败')
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  filters.service_type = undefined
  dateRange.value = []
  loadStats()
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
  loadStats()
})
</script>

<style scoped>
.admin-page {
  min-height: 100vh;
  background: #F7F6F3;
  padding: 48px 64px;
  font-family: 'Sora', sans-serif;
  color: #2D2D2D;

  --el-color-primary: #3D5A80;
  --el-color-success: #10B981;
  --el-color-danger: #C53D43;
  --el-color-info: #6B7280;
  --el-bg-color: transparent;
  --el-bg-color-page: #F7F6F3;
  --el-border-color: #E8E6E1;
  --el-border-color-light: #E8E6E1;
  --el-border-radius-base: 2px;
  --el-box-shadow-light: none;
  --el-fill-color-blank: transparent;
  --el-text-color-primary: #2D2D2D;
  --el-text-color-regular: #6B7280;
}

.admin-page :deep(.el-card) {
  border: 1px solid #E8E6E1;
  border-radius: 2px;
  background: transparent;
  box-shadow: none;
}

.admin-page :deep(.el-table) {
  background: transparent;
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: transparent;
  --el-table-border-color: #E8E6E1;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
}

.admin-page :deep(.el-table th.el-table__cell) {
  font-family: 'Sora', sans-serif;
  font-weight: 500;
  font-size: 12px;
  color: #6B7280;
}

.admin-page :deep(.el-button) {
  border-radius: 2px;
  font-family: 'Sora', sans-serif;
  font-weight: 500;
}

.admin-header {
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 40px;
}

.header-left h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 400;
  letter-spacing: -0.5px;
  color: #2D2D2D;
}

.header-left p {
  margin: 8px 0 0;
  color: #6B7280;
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.filter-card {
  margin-bottom: 24px;
}

.filters {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.summary-card {
  text-align: left;
  padding: 12px;
}

.summary-label {
  color: #6B7280;
  font-family: 'Sora', sans-serif;
  font-size: 14px;
  font-weight: 500;
}

.summary-value {
  margin-top: 16px;
  font-family: 'Sora', sans-serif;
  font-size: 48px;
  font-weight: 300;
  letter-spacing: -1px;
  color: #2D2D2D;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-family: 'Sora', sans-serif;
  font-weight: 500;
  font-size: 16px;
}

.muted {
  color: #9CA3AF;
  font-size: 12px;
}

@media (max-width: 1024px) {
  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
