<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-header">
        <h1>平台管理端登录</h1>
        <p>仅平台管理员可登录，登录后访问用户与计费管理能力</p>
      </div>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入管理员邮箱" autocomplete="email" />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            placeholder="请输入管理员密码"
            autocomplete="current-password"
          />
        </el-form-item>

        <el-button class="submit-btn" type="primary" :loading="submitting" @click="handleLogin">
          登录管理端
        </el-button>
      </el-form>

      <div class="auth-footer">
        <router-link to="/login">返回普通用户登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { useAdminAuthStore } from '@/stores/adminAuth'

const router = useRouter()
const route = useRoute()
const adminAuthStore = useAdminAuthStore()

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
    await adminAuthStore.login({
      email: form.email,
      password: form.password
    })
    ElMessage.success('管理端登录成功')
    const redirect = route.query.redirect as string | undefined
    await router.replace(redirect || '/admin/users')
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
  max-width: 440px;
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
  display: flex;
  justify-content: center;
}

.auth-footer a {
  color: var(--accent);
  text-decoration: none;
}
</style>
