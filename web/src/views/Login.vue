<template>
  <AuthSceneLayout
    title="欢迎来到 xinggen"
    subtitle="xinggen，让想象发生"
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
          placeholder="请输入密码"
          autocomplete="current-password"
          @keyup.enter="handleLogin"
        />
      </el-form-item>

      <el-form-item label="图形验证码" prop="captchaCode">
        <div class="captcha-row">
          <el-input
            v-model="form.captchaCode"
            placeholder="请输入图形验证码"
            autocomplete="off"
            @keyup.enter="handleLogin"
          />
          <button
            type="button"
            class="captcha-image-btn"
            :disabled="captchaLoading"
            @click="loadCaptcha"
          >
            <img v-if="captchaImage" :src="captchaImage" alt="captcha" class="captcha-image" />
            <span v-else>{{ captchaLoading ? '加载中' : '获取验证码' }}</span>
          </button>
        </div>
      </el-form-item>

      <el-button class="submit-btn" type="primary" :loading="submitting" @click="handleLogin">
        登 录
      </el-button>
    </el-form>

    <template #footer>
      <p class="agreement">登录即表示您同意遵守《用户协议》与《隐私政策》</p>
      <div class="auth-footer">
        <span>还没有账号？</span>
        <router-link to="/register">去注册</router-link>
      </div>
      <router-link class="secondary-link" to="/admin/login">切换管理端登录</router-link>
    </template>
  </AuthSceneLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import AuthSceneLayout from '@/components/auth/AuthSceneLayout.vue'
import { authAPI } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const formRef = ref<FormInstance>()
const submitting = ref(false)
const captchaLoading = ref(false)
const captchaId = ref('')
const captchaImage = ref('')
const form = reactive({
  email: '',
  password: '',
  captchaCode: ''
})

const rules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captchaCode: [{ required: true, message: '请输入图形验证码', trigger: 'blur' }]
}

const loadCaptcha = async () => {
  captchaLoading.value = true
  try {
    const resp = await authAPI.captcha()
    captchaId.value = resp.captcha_id
    captchaImage.value = resp.image_data
    form.captchaCode = ''
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '获取图形验证码失败')
  } finally {
    captchaLoading.value = false
  }
}

const handleLogin = async () => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    await authStore.login({
      email: form.email,
      password: form.password,
      captcha_id: captchaId.value,
      captcha_code: form.captchaCode
    })
    ElMessage.success('登录成功')
    const redirect = route.query.redirect as string | undefined
    await router.replace(redirect || '/')
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '登录失败')
    await loadCaptcha()
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadCaptcha()
})
</script>

<style scoped>
.captcha-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 112px;
  gap: 10px;
}

.captcha-image-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  border: 0;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.08);
  overflow: hidden;
  cursor: pointer;
  color: rgba(233, 242, 255, 0.84);
}

.captcha-image {
  display: block;
  width: 100%;
  height: 40px;
  object-fit: cover;
}

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
