<template>
  <nav class="floating-quick-nav" aria-label="快捷导航">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="quick-nav-item"
      :class="{ 'is-active': item.active }"
      @click="handleClick(item)"
    >
      <el-icon class="quick-nav-icon">
        <component :is="item.icon" />
      </el-icon>
      <span class="quick-nav-label">{{ item.label }}</span>
    </button>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import { useRoute, useRouter } from 'vue-router'
import { Box, Collection, Coin, Document, Files, FolderOpened, Picture, Reading, Tools, User } from '@element-plus/icons-vue'

interface QuickNavItem {
  key: string
  label: string
  target: RouteLocationRaw
  icon: any
  active: boolean
}

const router = useRouter()
const route = useRoute()
const projectTabKeys = ['overview', 'episodes', 'characters', 'scenes', 'props', 'assets'] as const
const currentProjectId = computed(() => {
  const raw = route.params.dramaId || route.params.id
  return typeof raw === 'string' ? raw : ''
})
const isProjectContext = computed(() => {
  return Boolean(currentProjectId.value) && route.path.startsWith('/dramas/')
})

const currentProjectSection = computed(() => {
  const queryTab = typeof route.query.tab === 'string' ? route.query.tab : ''
  if (projectTabKeys.includes(queryTab as (typeof projectTabKeys)[number])) {
    return queryTab
  }

  if (route.path.includes('/characters')) return 'characters'
  if (route.path.includes('/episode/')) return 'episodes'
  if (route.path.includes('/settings')) return 'overview'

  return 'overview'
})

const items = computed<QuickNavItem[]>(() => {
  if (isProjectContext.value) {
    const projectPath = `/dramas/${currentProjectId.value}`
    return [
      {
        key: 'overview',
        label: '项目概览',
        target: { path: projectPath, query: { tab: 'overview' } },
        icon: Reading,
        active: currentProjectSection.value === 'overview'
      },
      {
        key: 'episodes',
        label: '章节管理',
        target: { path: projectPath, query: { tab: 'episodes' } },
        icon: Document,
        active: currentProjectSection.value === 'episodes'
      },
      {
        key: 'characters',
        label: '角色管理',
        target: { path: projectPath, query: { tab: 'characters' } },
        icon: User,
        active: currentProjectSection.value === 'characters'
      },
      {
        key: 'scenes',
        label: '场景列表',
        target: { path: projectPath, query: { tab: 'scenes' } },
        icon: Picture,
        active: currentProjectSection.value === 'scenes'
      },
      {
        key: 'props',
        label: '道具列表',
        target: { path: projectPath, query: { tab: 'props' } },
        icon: Box,
        active: currentProjectSection.value === 'props'
      },
      {
        key: 'assets',
        label: '资源列表',
        target: { path: projectPath, query: { tab: 'assets' } },
        icon: Files,
        active: currentProjectSection.value === 'assets'
      }
    ]
  }

  return [
    {
      key: 'home',
      label: '首页',
      target: '/',
      icon: FolderOpened,
      active: route.path === '/'
    },
    {
      key: 'library',
      label: '角色库',
      target: '/character-library',
      icon: Collection,
      active: route.path.startsWith('/character-library')
    },
    {
      key: 'credits',
      label: '积分',
      target: '/billing/purchase',
      icon: Coin,
      active: route.path.startsWith('/billing')
    },
    {
      key: 'tools',
      label: '工具箱',
      target: '/tools',
      icon: Tools,
      active: route.path.startsWith('/tools')
    },
    {
      key: 'account',
      label: '我的',
      target: '/settings/account',
      icon: User,
      active: route.path.startsWith('/settings/account')
    }
  ]
})

const handleClick = (item: QuickNavItem) => {
  router.push(item.target)
}
</script>

<style scoped>
.floating-quick-nav {
  position: fixed;
  top: 50%;
  left: 12px;
  z-index: 220;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 8px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  border-radius: 32px;
  background: rgba(248, 250, 252, 0.56);
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.05);
  backdrop-filter: blur(18px);
  transform: translateY(-50%);
}

.quick-nav-item {
  width: 48px;
  min-height: 74px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 6px;
  border: 0;
  border-radius: 24px;
  background: transparent;
  color: rgba(71, 85, 105, 0.82);
  cursor: pointer;
  transition: transform var(--transition-fast), background var(--transition-fast), color var(--transition-fast);
}

.quick-nav-item:hover {
  transform: translateY(-1px);
  background: rgba(148, 163, 184, 0.12);
  color: var(--text-primary);
}

.quick-nav-item.is-active {
  background: rgba(255, 255, 255, 0.42);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.28);
  color: rgba(30, 41, 59, 0.96);
}

.quick-nav-icon {
  font-size: 20px;
}

.quick-nav-label {
  font-size: 11px;
  font-weight: 600;
  line-height: 1.15;
  text-align: center;
  white-space: normal;
  word-break: break-all;
}

.dark .floating-quick-nav {
  background: rgba(51, 65, 85, 0.34);
  border-color: rgba(226, 232, 240, 0.1);
  box-shadow: 0 14px 30px rgba(0, 0, 0, 0.18);
}

.dark .quick-nav-item {
  color: rgba(226, 232, 240, 0.72);
}

.dark .quick-nav-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.dark .quick-nav-item.is-active {
  color: #fff;
  background: rgba(255, 255, 255, 0.12);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
}

@media (max-width: 1024px) {
  .floating-quick-nav {
    display: none;
  }
}
</style>
