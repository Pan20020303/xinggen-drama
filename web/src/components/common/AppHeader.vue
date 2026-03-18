<template>
  <div class="app-header-wrapper">
    <header class="app-header" :class="{ 'header-fixed': fixed }">
      <div class="header-content">
        <!-- Left section: Logo + Left slot -->
        <div class="header-left">
          <router-link v-if="showLogo" to="/" class="logo">
            <span class="logo-text">🎬 星亘 Drama</span>
          </router-link>
          <!-- Left slot for business content | 左侧插槽用于业务内容 -->
          <slot name="left" />
        </div>

        <!-- Center section: Center slot -->
        <div class="header-center">
          <slot name="center" />
        </div>

        <!-- Right section: Actions + Right slot -->
        <div class="header-right">
          <template v-if="showUserActions">
            <!-- Language Switcher | 语言切换 -->
            <LanguageSwitcher v-if="showLanguage" />

            <!-- Theme Toggle | 主题切换 -->
            <ThemeToggle v-if="showTheme" />

            <el-button v-if="showNavButtons && isAuthenticated" class="header-btn" @click="openCreditsDialog">
              <el-icon><Coin /></el-icon>
              <span class="btn-text">积分 {{ authStore.user?.credits ?? 0 }}</span>
            </el-button>

            <el-dropdown v-if="showNavButtons && isAuthenticated" trigger="click">
              <el-button class="header-btn avatar-trigger">
                <el-avatar
                  :size="34"
                  :src="avatarSrc"
                  class="header-avatar"
                >
                  <el-icon><UserFilled /></el-icon>
                </el-avatar>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="goAccountCenter">账户中心</el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>

            <el-button v-if="showNavButtons && !isAuthenticated" type="primary" class="header-btn" @click="goLogin">
              登录
            </el-button>
          </template>

          <!-- Right slot for business content (before actions) | 右侧插槽（在操作按钮前） -->
          <slot name="right" />
        </div>
      </div>
    </header>

    <CreditDetailsDialog v-model="creditsDialogVisible" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Coin, UserFilled } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import ThemeToggle from './ThemeToggle.vue'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import CreditDetailsDialog from './CreditDetailsDialog.vue'
import { useAuthStore } from '@/stores/auth'
import { fixImageUrl } from '@/utils/image'

/**
 * AppHeader - Global application header component
 * 应用顶部头组件
 * 
 * Features | 功能:
 * - Fixed position at top | 固定在顶部
 * - Model/Theme/Language switch | 模型/主题/语言切换
 * - Slots support for business content | 支持插槽放置业务内容
 * 
 * Slots | 插槽:
 * - left: Content after logo | logo 右侧内容
 * - center: Center content | 中间内容
 * - right: Content before actions | 操作按钮左侧内容
 */

interface Props {
  /** Fixed position at top | 是否固定在顶部 */
  fixed?: boolean
  /** Show logo | 是否显示 logo */
  showLogo?: boolean
  /** Show language switcher | 是否显示语言切换 */
  showLanguage?: boolean
  /** Show theme toggle | 是否显示主题切换 */
  showTheme?: boolean
  /** Show built-in right-side actions | 是否显示默认右侧操作 */
  showUserActions?: boolean
}

withDefaults(defineProps<Props>(), {
  fixed: true,
  showLogo: true,
  showLanguage: true,
  showTheme: true,
  showUserActions: true
})

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const isAuthenticated = computed(() => authStore.isAuthenticated)
const showNavButtons = computed(() => route.path !== '/login' && route.path !== '/register')
const creditsDialogVisible = ref(false)
const avatarSrc = computed(() => {
  const raw = authStore.user?.avatar_url
  return raw ? fixImageUrl(raw) : ''
})

const goAccountCenter = () => {
  router.push('/settings/account')
}

const openCreditsDialog = () => {
  creditsDialogVisible.value = true
}

const goLogin = () => {
  router.push('/login')
}

const handleLogout = async () => {
  authStore.logout()
  await router.replace('/login')
}
</script>

<style scoped>
.app-header {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-primary);
  backdrop-filter: blur(8px);
  z-index: 1000;
}

.app-header.header-fixed {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-4);
  height: 70px;
  max-width: 100%;
  margin: 0 auto;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-shrink: 0;
}

.header-center {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  text-decoration: none;
  color: var(--text-primary);
  font-weight: 700;
  font-size: 1.125rem;
  transition: opacity var(--transition-fast);
}

.logo:hover {
  opacity: 0.8;
}

.logo-text {
  background: linear-gradient(135deg, var(--accent) 0%, #06b6d4 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-btn {
  border-radius: var(--radius-lg);
  font-weight: 500;
}

.header-btn .btn-text {
  margin-left: 4px;
}

.avatar-trigger {
  padding: 4px;
  border-radius: 999px;
}

.header-avatar {
  background: linear-gradient(135deg, var(--accent) 0%, #8b5cf6 100%);
  color: #fff;
}

/* Dark mode adjustments | 深色模式适配 */
.dark .app-header {
  background: rgba(26, 33, 41, 0.95);
}

/* ========================================
   Common Slot Styles / 插槽通用样式
   ======================================== */

/* Back Button | 返回按钮 */
:deep(.back-btn) {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

:deep(.back-btn:hover) {
  color: var(--text-primary);
  background: var(--bg-hover);
}

:deep(.back-btn .el-icon) {
  font-size: 1rem;
}

/* Page Title | 页面标题 */
:deep(.page-title) {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

:deep(.page-title h1),
:deep(.header-title),
:deep(.drama-title) {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
}

:deep(.page-title .subtitle) {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

/* Episode Title | 章节标题 */
:deep(.episode-title) {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

/* Responsive | 响应式 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 var(--space-3);
  }
  
  .btn-text {
    display: none;
  }
  
  .header-btn {
    padding: 8px;
  }

  :deep(.page-title h1),
  :deep(.header-title),
  :deep(.drama-title) {
    font-size: 1rem;
  }

  :deep(.back-btn span) {
    display: none;
  }
}
</style>
