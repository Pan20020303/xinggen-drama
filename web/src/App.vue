<template>
  <FloatingQuickNav v-if="showFloatingNav" />
  <div :class="{ 'with-floating-nav': showFloatingNav }">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { FloatingQuickNav } from '@/components/common'

const route = useRoute()
const showFloatingNav = computed(() => {
  const path = route.path
  if (path.startsWith('/admin')) return false
  return path !== '/login' && path !== '/register'
})
</script>

<style>
#app {
  width: 100%;
  height: 100vh;
}

@media (min-width: 1025px) {
  .with-floating-nav .content-wrapper > :not(.app-header-wrapper) {
    margin-left: 96px;
  }

  .with-floating-nav .page-container > :not(.content-wrapper) {
    margin-left: 96px;
  }
}
</style>
