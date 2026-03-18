<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>工具箱</h1>
            <span class="subtitle">集中放置常用创作工具和快捷入口</span>
          </div>
        </template>
      </AppHeader>

      <div class="toolbox-grid">
        <el-card class="tool-card" shadow="hover">
          <template #header>
            <div class="card-title">提示词助手</div>
          </template>
          <p>后续可接入提示词优化、风格转换和模板管理。</p>
          <el-button disabled>即将开放</el-button>
        </el-card>

        <el-card class="tool-card" shadow="hover">
          <template #header>
            <div class="card-title">素材处理</div>
          </template>
          <p>后续可放置图片裁剪、批量压缩、格式转换等工具。</p>
          <el-button disabled>即将开放</el-button>
        </el-card>

        <el-card class="tool-card" shadow="hover">
          <template #header>
            <div class="card-title">创作捷径</div>
          </template>
          <p>后续可集中放置常用工作流入口，减少页面跳转成本。</p>
          <el-button disabled>即将开放</el-button>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { AppHeader } from '@/components/common'
</script>

<style scoped>
.toolbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}

.tool-card {
  min-height: 220px;
}

.card-title {
  font-size: 18px;
  font-weight: 700;
}

.tool-card p {
  margin: 0 0 20px;
  color: var(--text-secondary);
  line-height: 1.7;
}
</style>
