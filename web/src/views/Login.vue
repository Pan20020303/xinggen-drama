<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <h1>登录火爆剧</h1>
        <p>登录后可使用多租户隔离与积分能力</p>
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
            placeholder="请输入密码"
            autocomplete="current-password"
          />
        </el-form-item>

        <el-button class="submit-btn" type="primary" :loading="submitting" @click="handleLogin">
          登录
        </el-button>
      </el-form>

      <div class="auth-footer">
        <span>还没有账号？</span>
        <router-link to="/register">去注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const form = reactive({
  email: '',
  password: ''
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await authStore.login({
      email: form.email,
      password: form.password
    })
    ElMessage.success('登录成功')
    const redirect = route.query.redirect as string | undefined
    await router.replace(redirect || '/')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '登录失败')
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
