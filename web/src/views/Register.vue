<template>
  <AuthSceneLayout
    title="注册 xinggen 账号"
    subtitle="创建账号后立即开始你的创作旅程"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="auth-form"
    >
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
          @keyup.enter="handleRegister"
        />
      </el-form-item>

      <el-button class="submit-btn" type="primary" :loading="submitting" @click="handleRegister">
        注 册
      </el-button>
    </el-form>

    <template #footer>
      <p class="agreement">登录/注册即同意《用户协议》与《隐私政策》</p>
      <div class="auth-footer">
        <span>已有账号？</span>
        <router-link to="/login">直接登录</router-link>
      </div>
    </template>
  </AuthSceneLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AuthSceneLayout from '@/components/auth/AuthSceneLayout.vue'
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
.submit-btn {
  width: 100%;
  margin-top: 2px;
  height: 46px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 700;
}

.auth-footer {
  margin-top: 10px;
  font-size: 13px;
  color: rgba(222, 236, 255, 0.7);
  display: flex;
  gap: 4px;
  justify-content: center;
}

.auth-footer a {
  color: rgba(255, 255, 255, 0.94);
  text-decoration: none;
}

.agreement {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  text-align: center;
  color: rgba(222, 236, 255, 0.42);
}

:deep(.el-form-item__label) {
  color: rgba(233, 242, 255, 0.9);
}

:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.08) !important;
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.08) inset !important;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.22) inset !important;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(36, 155, 255, 0.7) inset !important;
}

:deep(.el-input__inner) {
  color: #fff;
}

:deep(.el-input__inner::placeholder) {
  color: rgba(233, 242, 255, 0.35);
}
</style>
