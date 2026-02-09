<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>账户中心</h1>
            <span class="subtitle">查看登录信息与积分余额</span>
          </div>
        </template>
      </AppHeader>

      <div class="cards">
        <el-card>
          <template #header>
            <div class="card-title">基础信息</div>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户 ID">{{ user?.id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ user?.email || '-' }}</el-descriptions-item>
            <el-descriptions-item label="角色">{{ user?.role || '-' }}</el-descriptions-item>
            <el-descriptions-item label="当前积分">{{ user?.credits ?? 0 }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ user?.created_at || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card>
          <template #header>
            <div class="card-title">快捷入口</div>
          </template>
          <div class="actions">
            <el-button @click="router.push('/')">进入项目列表</el-button>
            <el-button type="primary" @click="router.push('/settings/ai-config')">管理 AI 配置</el-button>
            <el-button type="danger" @click="handleLogout">退出登录</el-button>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { AppHeader } from '@/components/common'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const user = computed(() => authStore.user)

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确认退出当前账号吗？', '提示', {
      type: 'warning'
    })
    authStore.logout()
    await router.replace('/login')
  } catch {
    return
  }
}
</script>

<style scoped>
.page-container {
  min-height: 100vh;
  background: var(--bg-primary);
}

.content-wrapper {
  width: 100%;
}

.page-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-title h1 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--text-muted);
}

.cards {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  padding: 12px;
}

.card-title {
  font-weight: 600;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
