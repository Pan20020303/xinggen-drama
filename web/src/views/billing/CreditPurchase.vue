<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>积分加购</h1>
            <span class="subtitle">选择适合你的积分方案</span>
          </div>
        </template>
      </AppHeader>

      <el-card class="purchase-shell">
        <el-tabs v-model="activeTab" class="purchase-tabs">
          <el-tab-pane label="积分加量" name="purchase">
            <div class="hero">
              <h2>AI 驱动的一站式精品漫剧创作平台</h2>
              <el-button text @click="router.push('/settings/account')">查看账户信息</el-button>
            </div>

            <div class="plans-grid">
              <section class="plan-card">
                <div class="plan-title">体验版</div>
                <div class="plan-points">1,000 积分</div>
                <div class="plan-price">¥ 100</div>
                <el-button type="primary" size="large" class="plan-action" @click="showContactTip">
                  立即购买
                </el-button>
                <ul class="plan-features">
                  <li>适合轻量体验和单项目试用</li>
                  <li>支持主流文生图与图生视频</li>
                  <li>适合个人创作者</li>
                </ul>
              </section>

              <section class="plan-card plan-card-featured">
                <div class="plan-badge">最受欢迎</div>
                <div class="plan-title">旗舰版</div>
                <el-select v-model="selectedFlagship" class="plan-select">
                  <el-option
                    v-for="option in flagshipOptions"
                    :key="option.value"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
                <div class="plan-price">¥ {{ selectedFlagshipPrice }}</div>
                <el-button type="primary" size="large" class="plan-action" @click="showContactTip">
                  立即购买
                </el-button>
                <ul class="plan-features">
                  <li>适合高频创作团队</li>
                  <li>支持更多批量生成任务</li>
                  <li>优先客服响应</li>
                  <li>更高并发创作体验</li>
                </ul>
              </section>

              <section class="plan-card plan-card-business">
                <div class="plan-title">企业版</div>
                <div class="plan-business-title">大客户专属客服</div>
                <div class="plan-business-subtitle">为您定制最佳方案</div>
                <el-button size="large" class="plan-action business" @click="showContactTip">
                  联系商务
                </el-button>
                <ul class="plan-features">
                  <li>定制 Agent 工作流</li>
                  <li>团队扩容与权限支持</li>
                  <li>专属技术服务</li>
                  <li>不限团队人数</li>
                </ul>
              </section>
            </div>

            <div class="purchase-notice">
              购买须知：当前页面先承接加购入口，支付能力未接入时将引导联系商务。积分有效期 1 年，具体以平台规则为准。
            </div>
          </el-tab-pane>

          <el-tab-pane label="兑换码" name="redeem">
            <div class="redeem-panel">
              <el-form class="redeem-form" @submit.prevent>
                <el-form-item label="兑换码">
                  <el-input v-model="redeemCode" placeholder="请输入兑换码" size="large" />
                </el-form-item>
                <el-button type="primary" size="large" @click="showRedeemTip">立即兑换</el-button>
              </el-form>
              <p class="redeem-hint">兑换能力暂未接入，先保留页面入口和交互位置。</p>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { AppHeader } from '@/components/common'

const router = useRouter()
const activeTab = ref('purchase')
const redeemCode = ref('')
const selectedFlagship = ref('10000')
const flagshipOptions = [
  { value: '10000', label: '10,000 积分', price: 1000 },
  { value: '20000', label: '20,000 积分', price: 1800 },
  { value: '50000', label: '50,000 积分', price: 4200 }
]

const selectedFlagshipPrice = computed(() => {
  return flagshipOptions.find((option) => option.value === selectedFlagship.value)?.price ?? 1000
})

const showContactTip = () => {
  ElMessage.info('支付能力暂未接入，请联系商务完成加购。')
}

const showRedeemTip = () => {
  if (!redeemCode.value.trim()) {
    ElMessage.warning('请输入兑换码')
    return
  }
  ElMessage.info('兑换能力暂未接入，请联系管理员处理。')
}
</script>

<style scoped>
.purchase-shell {
  border-radius: 24px;
}

.purchase-tabs {
  min-height: 720px;
}

.hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin: 12px 0 28px;
  text-align: center;
}

.hero h2 {
  margin: 0;
  font-size: 48px;
  line-height: 1.1;
  font-weight: 800;
}

.plans-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 24px;
}

.plan-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 520px;
  padding: 32px;
  border: 1px solid var(--border-primary);
  border-radius: 28px;
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.08), rgba(15, 23, 42, 0.02));
}

.plan-card-featured {
  border-color: #8b5cf6;
  box-shadow: 0 18px 50px rgba(139, 92, 246, 0.18);
}

.plan-card-business {
  background: linear-gradient(180deg, rgba(245, 158, 11, 0.08), rgba(15, 23, 42, 0.02));
}

.plan-badge {
  position: absolute;
  top: 18px;
  right: 18px;
  padding: 8px 14px;
  border-radius: 999px;
  background: linear-gradient(135deg, #7c3aed, #fb7185, #f97316);
  color: #fff;
  font-size: 13px;
  font-weight: 700;
}

.plan-title {
  font-size: 20px;
  font-weight: 700;
}

.plan-points,
.plan-price,
.plan-business-title,
.plan-business-subtitle {
  font-size: 18px;
  font-weight: 600;
}

.plan-points {
  font-size: 42px;
  color: #8b5cf6;
}

.plan-price {
  font-size: 36px;
}

.plan-business-title {
  font-size: 34px;
}

.plan-business-subtitle {
  font-size: 28px;
}

.plan-action {
  width: 100%;
  height: 56px;
  font-size: 22px;
  border-radius: 18px;
}

.plan-action.business {
  background: linear-gradient(135deg, #f8e7bf, #efe1c3);
  border: none;
  color: #111827;
}

.plan-select {
  width: 100%;
}

.plan-features {
  margin: 0;
  padding-left: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  color: var(--text-secondary);
}

.purchase-notice {
  margin-top: 28px;
  color: var(--text-secondary);
  font-size: 14px;
}

.redeem-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 18px;
  padding: 72px 24px;
}

.redeem-form {
  width: min(420px, 100%);
}

.redeem-hint {
  color: var(--text-secondary);
}

@media (max-width: 1200px) {
  .plans-grid {
    grid-template-columns: 1fr;
  }

  .hero h2 {
    font-size: 36px;
  }
}
</style>
