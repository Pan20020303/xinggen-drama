<template>
  <div class="admin-layout-wrapper" :class="{ 'is-collapsed': isCollapsed }">
    <!-- Sidebar -->
    <aside class="admin-sidebar" @mouseenter="isHover = true" @mouseleave="isHover = false">
      <div class="sidebar-top">
        <div class="logo-box">
          <LucideLayoutGrid class="icon logo" :size="24" color="#2D2D2D" />
          <span v-if="!isCollapsed || isHover" class="logo-text">管理面板</span>
        </div>

        <div class="nav-links">
          <router-link
            to="/admin/users"
            custom
            v-slot="{ navigate, isActive }"
          >
            <div
              class="nav-item"
              :class="{ active: isActive }"
              @click="navigate"
              title="用户管理"
            >
              <LucideUsers class="icon" :size="18" :color="isActive ? '#F7F6F3' : '#6B7280'" />
              <span v-show="!isCollapsed || isHover" class="nav-text">用户管理</span>
            </div>
          </router-link>

          <router-link
            to="/admin/ai-config"
            custom
            v-slot="{ navigate, isActive }"
          >
            <div
              class="nav-item"
              :class="{ active: isActive }"
              @click="navigate"
              title="模型配置"
            >
              <LucideBox class="icon" :size="18" :color="isActive ? '#F7F6F3' : '#6B7280'" />
              <span v-show="!isCollapsed || isHover" class="nav-text">模型配置</span>
            </div>
          </router-link>

          <router-link
            to="/admin/billing"
            custom
            v-slot="{ navigate, isActive }"
          >
            <div
              class="nav-item"
              :class="{ active: isActive }"
              @click="navigate"
              title="积分流水"
            >
              <LucideCreditCard class="icon" :size="18" :color="isActive ? '#F7F6F3' : '#6B7280'" />
              <span v-show="!isCollapsed || isHover" class="nav-text">积分流动</span>
            </div>
          </router-link>

          <router-link
            to="/admin/token-stats"
            custom
            v-slot="{ navigate, isActive }"
          >
            <div
              class="nav-item"
              :class="{ active: isActive }"
              @click="navigate"
              title="Token统计"
            >
              <LucideTrendingUp class="icon" :size="18" :color="isActive ? '#F7F6F3' : '#6B7280'" />
              <span v-show="!isCollapsed || isHover" class="nav-text">数据统计</span>
            </div>
          </router-link>
        </div>
      </div>

      <div class="sidebar-bottom">
        <div class="nav-item toggle-btn" @click="toggleCollapse">
          <LucidePanelLeftClose v-if="!isCollapsed" class="icon" :size="18" color="#6B7280" />
          <LucidePanelLeftOpen v-else class="icon" :size="18" color="#6B7280" />
          <span v-show="!isCollapsed || isHover" class="nav-text">收起侧栏</span>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="admin-main">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  LayoutGrid as LucideLayoutGrid,
  Users as LucideUsers,
  Box as LucideBox,
  CreditCard as LucideCreditCard,
  TrendingUp as LucideTrendingUp,
  PanelLeftClose as LucidePanelLeftClose,
  PanelLeftOpen as LucidePanelLeftOpen
} from 'lucide-vue-next'

const isCollapsed = ref(true)
const isHover = ref(false)

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
}
</script>

<style scoped>
.admin-layout-wrapper {
  display: flex;
  min-height: 100vh;
  background: #F7F6F3;
  font-family: 'Sora', sans-serif;
  color: #2D2D2D;
}

.admin-sidebar {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 200px;
  background: transparent;
  border-right: 1px solid #E8E6E1;
  padding: 24px 12px;
  transition: width 0.3s ease;
  overflow: hidden;
  z-index: 10;
  white-space: nowrap;
}

.admin-layout-wrapper.is-collapsed .admin-sidebar:not(:hover) {
  width: 64px;
}

.admin-layout-wrapper.is-collapsed .admin-sidebar:hover {
  width: 200px;
  position: absolute;
  height: 100vh;
  background: #F7F6F3; /* Fill background when hovering over collapsed side */
  box-shadow: 4px 0 12px rgba(0, 0, 0, 0.05); /* Optional shadow feeling */
}

.logo-box {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 48px;
  padding: 0 8px;
  margin-bottom: 32px;
}

.logo-text {
  font-size: 16px;
  font-weight: 500;
  color: #2D2D2D;
}

.nav-links {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 40px;
  padding: 0 10px;
  border-radius: 2px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: rgba(61, 90, 128, 0.05);
}

.nav-item.active {
  background: #3D5A80;
  color: #F7F6F3;
}

.nav-text {
  font-size: 13px;
  font-weight: 500;
  opacity: 1;
  transition: opacity 0.2s;
}

.admin-layout-wrapper.is-collapsed .admin-sidebar:not(:hover) .nav-text {
  opacity: 0;
  display: none;
}

.sidebar-bottom {
  margin-top: auto;
}

.toggle-btn {
  margin-top: 24px;
  color: #6B7280;
}

.toggle-btn:hover {
  color: #2D2D2D;
}

/* Main Area */
.admin-main {
  flex: 1;
  min-width: 0; /* Important for flex child to not push layout */
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
