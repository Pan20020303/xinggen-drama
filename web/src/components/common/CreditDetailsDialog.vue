<template>
  <el-dialog
    :model-value="modelValue"
    width="960px"
    :close-on-click-modal="false"
    class="credit-details-dialog"
    @update:model-value="emit('update:modelValue', $event)"
    @open="handleOpen"
  >
    <template #header>
      <div class="dialog-title">积分详情</div>
    </template>

    <div class="credit-summary">
      <div class="summary-card">
        <span class="summary-label">剩余积分</span>
        <strong>{{ currentCredits }}</strong>
      </div>
      <div class="summary-card">
        <span class="summary-label">本页获得</span>
        <strong class="positive">+{{ totalEarned }}</strong>
      </div>
      <div class="summary-card">
        <span class="summary-label">本页消耗</span>
        <strong class="negative">-{{ totalSpent }}</strong>
      </div>
      <div class="summary-actions">
        <el-button type="primary" @click="goPurchase">加购积分</el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="credit-tabs">
      <el-tab-pane label="全部" name="all" />
      <el-tab-pane label="获得" name="earn" />
      <el-tab-pane label="消耗" name="spend" />
    </el-tabs>

    <div v-loading="loading" class="transactions-panel">
      <el-empty v-if="!loading && filteredTransactions.length === 0" description="暂无积分记录" />
      <div v-else class="transactions-list">
        <div
          v-for="item in filteredTransactions"
          :key="item.id"
          class="transaction-item"
        >
          <div class="transaction-main">
            <div class="transaction-title">
              <span>{{ formatTransactionTitle(item) }}</span>
              <el-tag size="small" :type="item.amount >= 0 ? 'success' : 'danger'">
                {{ item.amount >= 0 ? `+${item.amount}` : item.amount }}
              </el-tag>
            </div>
            <div class="transaction-meta">
              <span>{{ formatTransactionTime(item.created_at) }}</span>
              <span v-if="item.service_type">服务：{{ item.service_type }}</span>
              <span v-if="item.model">模型：{{ item.model }}</span>
              <span v-if="item.total_tokens !== undefined && item.total_tokens !== null">
                tokens：{{ item.total_tokens }}
              </span>
            </div>
          </div>
          <div class="transaction-type">{{ formatTransactionType(item.type) }}</div>
        </div>
      </div>
    </div>

    <div class="dialog-footer">
      <el-pagination
        v-if="pagination.total > pagination.page_size"
        background
        layout="total, prev, pager, next"
        :total="pagination.total"
        :current-page="pagination.page"
        :page-size="pagination.page_size"
        @current-change="handlePageChange"
      />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { billingAPI } from '@/api/billing'
import { useAuthStore } from '@/stores/auth'
import type { CreditTransaction } from '@/types/billing'

interface Props {
  modelValue: boolean
}

defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const activeTab = ref<'all' | 'earn' | 'spend'>('all')
const transactions = ref<CreditTransaction[]>([])
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const currentCredits = computed(() => authStore.user?.credits ?? 0)
const totalEarned = computed(() =>
  transactions.value.filter((item) => item.amount > 0).reduce((sum, item) => sum + item.amount, 0)
)
const totalSpent = computed(() =>
  Math.abs(transactions.value.filter((item) => item.amount < 0).reduce((sum, item) => sum + item.amount, 0))
)
const filteredTransactions = computed(() => {
  if (activeTab.value === 'earn') {
    return transactions.value.filter((item) => item.amount > 0)
  }
  if (activeTab.value === 'spend') {
    return transactions.value.filter((item) => item.amount < 0)
  }
  return transactions.value
})

const loadTransactions = async () => {
  loading.value = true
  try {
    const data = await billingAPI.listTransactions({
      page: pagination.page,
      page_size: pagination.page_size
    })
    transactions.value = data.items
    pagination.total = data.pagination.total
  } catch (error: any) {
    ElMessage.error(error?.response?.data?.error?.message || error?.message || '加载积分记录失败')
  } finally {
    loading.value = false
  }
}

const handleOpen = () => {
  pagination.page = 1
  loadTransactions()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadTransactions()
}

const goPurchase = () => {
  emit('update:modelValue', false)
  router.push('/billing/purchase')
}

const formatTransactionTitle = (item: CreditTransaction) => {
  return item.description || formatTransactionType(item.type)
}

const formatTransactionType = (type: string) => {
  const labels: Record<string, string> = {
    RECHARGE: '积分充值',
    GENERATE_FRAME_PROMPT: '生成分镜提示词',
    GENERATE_IMAGE: '生成图片',
    AI_TEXT: '文本生成',
    AI_TEXT_REFUND: '文本退款',
    AI_IMAGE: '图片生成',
    AI_IMAGE_REFUND: '图片退款',
    AI_VIDEO: '视频生成',
    AI_VIDEO_REFUND: '视频退款'
  }
  return labels[type] || type
}

const formatTransactionTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.dialog-title {
  font-size: 24px;
  font-weight: 700;
}

.credit-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
  gap: 16px;
  margin-bottom: 20px;
}

.summary-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 18px 20px;
  border: 1px solid var(--border-primary);
  border-radius: 16px;
  background: var(--bg-card);
}

.summary-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.summary-card strong {
  font-size: 28px;
  line-height: 1;
}

.summary-card strong.positive {
  color: #16a34a;
}

.summary-card strong.negative {
  color: #dc2626;
}

.summary-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.credit-tabs {
  margin-bottom: 12px;
}

.transactions-panel {
  min-height: 420px;
}

.transactions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.transaction-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--bg-card);
}

.transaction-main {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.transaction-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  font-weight: 600;
}

.transaction-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.transaction-type {
  font-size: 13px;
  color: var(--text-tertiary);
  white-space: nowrap;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

@media (max-width: 960px) {
  .credit-summary {
    grid-template-columns: 1fr;
  }

  .summary-actions {
    justify-content: flex-start;
  }

  .transaction-item {
    flex-direction: column;
  }
}
</style>
