import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { public: true, guestOnly: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { public: true, guestOnly: true }
  },
  {
    path: '/',
    name: 'DramaList',
    component: () => import('../views/drama/DramaList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/create',
    name: 'DramaCreate',
    component: () => import('../views/drama/DramaCreate.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:id',
    name: 'DramaManagement',
    component: () => import('../views/drama/DramaManagement.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:id/episode/:episodeNumber',
    name: 'EpisodeWorkflowNew',
    component: () => import('../views/drama/EpisodeWorkflow.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:id/characters',
    name: 'CharacterExtraction',
    component: () => import('../views/workflow/CharacterExtraction.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:id/images/characters',
    name: 'CharacterImages',
    component: () => import('../views/workflow/CharacterImages.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:id/settings',
    name: 'DramaSettings',
    component: () => import('../views/workflow/DramaSettings.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/episodes/:id/edit',
    name: 'ScriptEdit',
    component: () => import('../views/script/ScriptEdit.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/episodes/:id/storyboard',
    name: 'StoryboardEdit',
    component: () => import('../views/storyboard/StoryboardEdit.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/episodes/:id/generate',
    name: 'Generation',
    component: () => import('../views/generation/ImageGeneration.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/timeline/:id',
    name: 'TimelineEditor',
    component: () => import('../views/editor/TimelineEditor.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/dramas/:dramaId/episode/:episodeNumber/professional',
    name: 'ProfessionalEditor',
    component: () => import('../views/drama/ProfessionalEditor.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/character-library',
    name: 'CharacterLibraryCenter',
    component: () => import('../views/workflow/CharacterLibraryCenter.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/settings/ai-config',
    name: 'AIConfig',
    component: () => import('../views/settings/AIConfig.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/settings/account',
    name: 'AccountCenter',
    component: () => import('../views/settings/AccountCenter.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

router.beforeEach((to) => {
  const token = localStorage.getItem('token')
  const isPublic = Boolean(to.meta.public)
  const isGuestOnly = Boolean(to.meta.guestOnly)

  if (!token && !isPublic) {
    return {
      path: '/login',
      query: { redirect: to.fullPath }
    }
  }

  if (token && isGuestOnly) {
    return '/'
  }

  return true
})

export default router
