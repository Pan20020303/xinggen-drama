<template>
  <AuthSceneLayout
    kicker="ADMIN PORTAL"
    title="平台管理端登录"
    subtitle="仅平台管理员可访问用户、计费与配置管理"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      class="auth-form"
    >
      <el-form-item label="管理员邮箱" prop="email">
        <el-input v-model="form.email" placeholder="请输入管理员邮箱" autocomplete="email" />
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="请输入管理员密码"
          autocomplete="current-password"
          @keyup.enter="handleLogin"
        />
      </el-form-item>

      <el-button class="submit-btn admin-btn" type="primary" :loading="submitting" @click="handleLogin">
        登录管理端
      </el-button>
    </el-form>

    <template #footer>
      <p class="agreement">管理员登录记录将用于审计与安全追踪</p>
      <router-link class="secondary-link" to="/login">返回普通用户登录</router-link>
    </template>
  </AuthSceneLayout>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AuthSceneLayout from '@/components/auth/AuthSceneLayout.vue'
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
.submit-btn {
  width: 100%;
  margin-top: 2px;
  height: 46px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 700;
}

.admin-btn {
  background: linear-gradient(135deg, #1774ff 0%, #0b51c8 100%);
  border-color: transparent;
  box-shadow: 0 12px 28px rgba(17, 102, 240, 0.32);
}

.admin-btn:hover {
  background: linear-gradient(135deg, #2a82ff 0%, #0f5fd8 100%);
}

.agreement {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  text-align: center;
  color: rgba(222, 236, 255, 0.42);
}

.secondary-link {
  margin-top: 12px;
  display: block;
  width: 100%;
  height: 42px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.07);
  color: rgba(255, 255, 255, 0.84);
  font-size: 14px;
  line-height: 40px;
  text-align: center;
  text-decoration: none;
  transition: all 0.18s ease;
}

.secondary-link:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
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
