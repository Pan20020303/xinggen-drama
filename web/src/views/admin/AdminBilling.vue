<template>
  <div class="admin-page">
    <header class="admin-header">
      <div class="header-left">
        <h1>计费管理</h1>
        <p>管理员充值与积分流水查询</p>
      </div>
      <div class="header-right">
        <LanguageSwitcher />
        <ThemeToggle />
        <el-button @click="router.push('/admin/users')">用户管理</el-button>
        <el-button @click="router.push('/admin/ai-config')">模型配置</el-button>
        <el-button type="danger" @click="handleLogout">退出管理端</el-button>
      </div>
    </header>

    <div class="content-grid">
      <el-card class="recharge-card">
        <template #header>
          <span>手动充值</span>
        </template>
        <el-form ref="rechargeFormRef" :model="rechargeForm" :rules="rechargeRules" label-position="top">
          <el-form-item label="用户ID" prop="user_id">
            <el-input-number v-model="rechargeForm.user_id" :min="1" style="width: 100%" />
          </el-form-item>
          <el-form-item label="充值积分" prop="amount">
            <el-input-number v-model="rechargeForm.amount" :min="1" style="width: 100%" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="rechargeForm.note" placeholder="如：manual recharge" />
          </el-form-item>
          <el-button type="primary" :loading="rechargeSubmitting" @click="handleRecharge">执行充值</el-button>
        </el-form>
      </el-card>

      <el-card class="transactions-card">
        <template #header>
          <div class="card-header">
            <span>积分流水</span>
            <div class="filters">
              <el-input-number
                v-model="filters.user_id"
                :min="1"
                :controls="false"
                placeholder="按用户ID过滤"
                style="width: 180px"
              />
              <el-button @click="handleSearch">查询</el-button>
              <el-button @click="handleReset">重置</el-button>
            </div>
          </div>
        </template>

        <el-table :data="transactions" v-loading="loading" stripe>
          <el-table-column prop="id" label="流水ID" min-width="88" />
          <el-table-column prop="user_id" label="用户ID" min-width="88" />
          <el-table-column label="变动积分" min-width="100">
            <template #default="{ row }">
              <el-tag :type="row.amount >= 0 ? 'success' : 'danger'">
                {{ row.amount >= 0 ? `+${row.amount}` : row.amount }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" min-width="160" />
          <el-table-column prop="description" label="说明" min-width="220" />
          <el-table-column prop="created_at" label="时间" min-width="180" />
        </el-table>

        <div class="pagination-wrap">
          <el-pagination
            background
            layout="total, prev, pager, next, sizes"
            :total="pagination.total"
            :current-page="pagination.page"
            :page-size="pagination.page_size"
            :page-sizes="[10, 20, 50, 100]"
            @current-change="handlePageChange"
            @size-change="handlePageSizeChange"
          />
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import ThemeToggle from '@/components/common/ThemeToggle.vue'
import { adminAPI } from '@/api/admin'
import { useAdminAuthStore } from '@/stores/adminAuth'
import type { CreditTransaction } from '@/types/admin'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()

const loading = ref(false)
const transactions = ref<CreditTransaction[]>([])
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})
const filters = reactive({
  user_id: undefined as number | undefined
})

const rechargeFormRef = ref<FormInstance>()
const rechargeSubmitting = ref(false)
const rechargeForm = reactive({
  user_id: undefined as number | undefined,
  amount: 100,
  note: ''
})

const rechargeRules: FormRules = {
  user_id: [{ required: true, message: '请输入用户ID', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入充值积分', trigger: 'blur' }]
}

const loadTransactions = async () => {
  loading.value = true
  try {
    const data = await adminAPI.listTransactions({
      user_id: filters.user_id,
      page: pagination.page,
      page_size: pagination.page_size
    })
    transactions.value = data.items
    pagination.total = data.pagination.total
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '加载流水失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadTransactions()
}

const handleReset = () => {
  filters.user_id = undefined
  pagination.page = 1
  loadTransactions()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadTransactions()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadTransactions()
}

const handleRecharge = async () => {
  if (!rechargeFormRef.value) return
  const valid = await rechargeFormRef.value.validate().catch(() => false)
  if (!valid || !rechargeForm.user_id) return

  rechargeSubmitting.value = true
  try {
    await adminAPI.recharge({
      user_id: rechargeForm.user_id,
      amount: rechargeForm.amount,
      note: rechargeForm.note || undefined
    })
    ElMessage.success('充值成功')
    filters.user_id = rechargeForm.user_id
    pagination.page = 1
    await loadTransactions()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '充值失败')
  } finally {
    rechargeSubmitting.value = false
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
  loadTransactions()
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
  font-size: 13px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.content-grid {
  display: grid;
  grid-template-columns: 340px minmax(0, 1fr);
  gap: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.filters {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 1024px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
