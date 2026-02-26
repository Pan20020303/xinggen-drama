<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>账户中心</h1>
            <span class="subtitle">查看登录信息与积分余额</span>
          </div>
        </template>
      </AppHeader>

      <div class="cards">
        <el-card>
          <template #header>
            <div class="card-title">基础信息</div>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户 ID">{{ user?.id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ user?.email || '-' }}</el-descriptions-item>
            <el-descriptions-item label="角色">{{ user?.role || '-' }}</el-descriptions-item>
            <el-descriptions-item label="当前积分">{{ user?.credits ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ user?.created_at || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card>
          <template #header>
            <div class="card-title">修改密码</div>
          </template>
          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-position="top"
            class="password-form"
          >
            <el-form-item label="旧密码" prop="old_password">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                show-password
                placeholder="请输入当前密码"
                autocomplete="current-password"
              />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                show-password
                placeholder="请输入新密码（至少 6 位）"
                autocomplete="new-password"
              />
            </el-form-item>
            <el-form-item label="确认新密码" prop="confirm_password">
              <el-input
                v-model="passwordForm.confirm_password"
                type="password"
                show-password
                placeholder="请再次输入新密码"
                autocomplete="new-password"
              />
            </el-form-item>
            <el-button
              type="primary"
              :loading="passwordSubmitting"
              @click="handleChangePassword"
            >
              保存新密码
            </el-button>
          </el-form>
        </el-card>

        <el-card>
          <template #header>
            <div class="card-title">快捷入口</div>
          </template>
          <div class="actions">
            <el-button @click="router.push('/')">进入项目列表</el-button>
            <el-button type="danger" @click="handleLogout">退出登录</el-button>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { AppHeader } from '@/components/common'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
const passwordFormRef = ref<FormInstance>()
const passwordSubmitting = ref(false)
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const passwordRules: FormRules = {
  old_password: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '新密码至少 6 位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (!value) {
          callback(new Error('请确认新密码'))
          return
        }
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入的新密码不一致'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确认退出当前账号吗？', '提示', {
      type: 'warning'
    })
    authStore.logout()
    await router.replace('/login')
  } catch {
    return
  }
}

const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  const valid = await passwordFormRef.value.validate().catch(() => false)
  if (!valid) return

  passwordSubmitting.value = true
  try {
    await authStore.changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    ElMessage.success('密码修改成功')
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    passwordFormRef.value.clearValidate()
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '修改密码失败')
  } finally {
    passwordSubmitting.value = false
  }
}
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

.cards {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  padding: 12px;
}

.card-title {
  font-weight: 600;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.password-form {
  max-width: 520px;
}
</style>
