<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <h1>注册火爆剧</h1>
        <p>新用户注册后自动发放初始积分</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" autocomplete="email" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="至少 6 位密码"
            autocomplete="new-password"
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            show-password
            placeholder="请再次输入密码"
            autocomplete="new-password"
          />
        </el-form-item>

        <el-button class="submit-btn" type="primary" :loading="submitting" @click="handleRegister">
          注册并登录
        </el-button>
      </el-form>

      <div class="auth-footer">
        <span>已有账号？</span>
        <router-link to="/login">去登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const form = reactive({
  email: '',
  password: '',
  confirmPassword: ''
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (_rule: any, value: string, callback: (error?: Error) => void) => {
        if (value !== form.password) {
          callback(new Error('两次输入的密码不一致'))
          return
        }
        callback()
      },
      trigger: 'blur'
    }
  ]
}

const handleRegister = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await authStore.register({
      email: form.email,
      password: form.password
    })
    ElMessage.success('注册成功')
    await router.replace('/')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '注册失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-card);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-card);
  padding: 28px;
}

.auth-header {
  margin-bottom: 20px;
}

.auth-header h1 {
  margin: 0;
  font-size: 24px;
  color: var(--text-primary);
}

.auth-header p {
  margin-top: 8px;
  color: var(--text-muted);
  font-size: 13px;
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
}

.auth-footer {
  margin-top: 16px;
  font-size: 13px;
  color: var(--text-secondary);
  display: flex;
  gap: 4px;
  justify-content: center;
}

.auth-footer a {
  color: var(--accent);
  text-decoration: none;
}
</style>
