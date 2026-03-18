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
            <div class="card-title">头像设置</div>
          </template>
          <div class="profile-avatar-section">
            <el-avatar :size="96" :src="avatarSrc" class="profile-avatar">
              {{ avatarFallback }}
            </el-avatar>
            <div class="profile-avatar-meta">
              <div class="profile-avatar-title">个人头像</div>
              <div class="profile-avatar-hint">顶部导航仅展示头像，可在这里随时更换。</div>
              <div class="profile-avatar-actions">
                <el-upload
                  :action="uploadAction"
                  :headers="uploadHeaders"
                  :show-file-list="false"
                  accept="image/jpeg,image/png,image/jpg,image/webp"
                  :before-upload="beforeAvatarUpload"
                  :on-success="handleAvatarUploadSuccess"
                  :on-error="handleAvatarUploadError"
                >
                  <el-button type="primary" :loading="avatarSubmitting">上传头像</el-button>
                </el-upload>
                <el-button
                  text
                  :disabled="!user?.avatar_url || avatarSubmitting"
                  @click="handleClearAvatar"
                >
                  清空头像
                </el-button>
              </div>
            </div>
          </div>
        </el-card>

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
import { fixImageUrl } from '@/utils/image'

const router = useRouter()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
const avatarSubmitting = ref(false)
const passwordFormRef = ref<FormInstance>()
const passwordSubmitting = ref(false)
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})
const uploadAction = '/api/v1/upload/image'
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token') || ''}`
}))
const avatarSrc = computed(() => {
  const raw = user.value?.avatar_url
  return raw ? fixImageUrl(raw) : ''
})
const avatarFallback = computed(() => {
  return user.value?.email?.trim()?.[0]?.toUpperCase() || 'U'
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

const beforeAvatarUpload = (file: File) => {
  const isImage = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp'].includes(file.type)
  if (!isImage) {
    ElMessage.error('只支持 jpg、png、webp 格式图片')
    return false
  }

  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    ElMessage.error('头像大小不能超过 5MB')
    return false
  }

  avatarSubmitting.value = true
  return true
}

const handleAvatarUploadSuccess = async (response: any) => {
  try {
    const avatarUrl = response?.data?.url || response?.url
    if (!avatarUrl) {
      throw new Error('未获取到头像地址')
    }

    await authStore.updateProfile({ avatar_url: avatarUrl })
    ElMessage.success('头像更新成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '头像更新失败')
  } finally {
    avatarSubmitting.value = false
  }
}

const handleAvatarUploadError = () => {
  avatarSubmitting.value = false
  ElMessage.error('头像上传失败')
}

const handleClearAvatar = async () => {
  avatarSubmitting.value = true
  try {
    await authStore.updateProfile({ avatar_url: '' })
    ElMessage.success('头像已清空')
  } catch (error: any) {
    ElMessage.error(error?.message || '清空头像失败')
  } finally {
    avatarSubmitting.value = false
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

.profile-avatar-section {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.profile-avatar {
  flex-shrink: 0;
  background: linear-gradient(135deg, var(--accent) 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 32px;
  font-weight: 700;
}

.profile-avatar-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.profile-avatar-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.profile-avatar-hint {
  font-size: 13px;
  color: var(--text-muted);
}

.profile-avatar-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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
