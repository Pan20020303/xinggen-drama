<template>
  <div class="admin-page">
    <header class="admin-header">
      <div class="header-left">
        <h1>用户管理</h1>
        <p>平台管理员可修改用户状态与角色</p>
      </div>
      <div class="header-right">
        <LanguageSwitcher />
        <ThemeToggle />
        <el-button @click="router.push('/admin/ai-config')">模型配置</el-button>
        <el-button @click="router.push('/admin/billing')">计费管理</el-button>
        <el-button type="danger" @click="handleLogout">退出管理端</el-button>
      </div>
    </header>

    <el-card class="main-card">
      <template #header>
        <div class="card-header">
          <span>用户列表</span>
          <el-button :loading="loading" @click="loadUsers">刷新</el-button>
        </div>
      </template>

      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="用户ID" min-width="88" />
        <el-table-column prop="email" label="邮箱" min-width="220" />
        <el-table-column label="角色" min-width="140">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="credits" label="积分" min-width="100" />
        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openStatusDialog(row)">改状态</el-button>
            <el-button size="small" type="primary" @click="openRoleDialog(row)">改角色</el-button>
          </template>
        </el-table-column>
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

    <el-dialog v-model="statusDialogVisible" title="更新用户状态" width="420px">
      <el-form label-position="top">
        <el-form-item label="用户ID">
          <el-input :model-value="String(statusForm.userId)" disabled />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="statusForm.status" style="width: 100%">
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="statusSubmitting" @click="submitStatusUpdate">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="roleDialogVisible" title="更新用户角色" width="420px">
      <el-form label-position="top">
        <el-form-item label="用户ID">
          <el-input :model-value="String(roleForm.userId)" disabled />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="roleForm.role" style="width: 100%">
            <el-option label="普通用户" value="user" />
            <el-option label="VIP" value="vip" />
            <el-option label="管理员(兼容)" value="admin" />
            <el-option label="平台管理员" value="platform_admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="roleSubmitting" @click="submitRoleUpdate">确认</el-button>
      </template>
    </el-dialog>
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
import type { AdminUser, AdminUserRole, AdminUserStatus } from '@/types/admin'

const router = useRouter()
const adminAuthStore = useAdminAuthStore()

const loading = ref(false)
const users = ref<AdminUser[]>([])
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const statusDialogVisible = ref(false)
const statusSubmitting = ref(false)
const statusForm = reactive({
  userId: 0,
  status: 'active' as AdminUserStatus
})

const roleDialogVisible = ref(false)
const roleSubmitting = ref(false)
const roleForm = reactive({
  userId: 0,
  role: 'user' as AdminUserRole
})

const roleLabel = (role: AdminUserRole) => {
  switch (role) {
    case 'platform_admin':
      return '平台管理员'
    case 'admin':
      return '管理员'
    case 'vip':
      return 'VIP'
    default:
      return '普通用户'
  }
}

const roleTagType = (role: AdminUserRole) => {
  if (role === 'platform_admin' || role === 'admin') return 'danger'
  if (role === 'vip') return 'warning'
  return 'info'
}

const loadUsers = async () => {
  loading.value = true
  try {
    const data = await adminAPI.listUsers({
      page: pagination.page,
      page_size: pagination.page_size
    })
    users.value = data.items
    pagination.total = data.pagination.total
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '加载用户失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadUsers()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadUsers()
}

const openStatusDialog = (row: AdminUser) => {
  statusForm.userId = row.id
  statusForm.status = row.status
  statusDialogVisible.value = true
}

const submitStatusUpdate = async () => {
  statusSubmitting.value = true
  try {
    await adminAPI.updateUserStatus(statusForm.userId, { status: statusForm.status })
    ElMessage.success('状态更新成功')
    statusDialogVisible.value = false
    await loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '状态更新失败')
  } finally {
    statusSubmitting.value = false
  }
}

const openRoleDialog = (row: AdminUser) => {
  roleForm.userId = row.id
  roleForm.role = row.role
  roleDialogVisible.value = true
}

const submitRoleUpdate = async () => {
  roleSubmitting.value = true
  try {
    await adminAPI.updateUserRole(roleForm.userId, { role: roleForm.role })
    ElMessage.success('角色更新成功')
    roleDialogVisible.value = false
    await loadUsers()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '角色更新失败')
  } finally {
    roleSubmitting.value = false
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
  loadUsers()
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

.main-card {
  border-radius: var(--radius-xl);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
