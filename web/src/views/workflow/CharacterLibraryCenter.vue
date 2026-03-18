<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="page-title">
            <h1>素材库</h1>
            <span class="subtitle">统一管理工具和项目生成的图片、视频素材</span>
          </div>
        </template>
      </AppHeader>

      <div class="asset-toolbar">
        <el-tabs v-model="activeTab" class="asset-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="图片素材" name="image" />
          <el-tab-pane label="视频素材" name="video" />
          <el-tab-pane label="音频素材" name="audio" />
        </el-tabs>
        <div class="asset-search">
          <el-input v-model="query.keyword" placeholder="按名称搜索素材" clearable @change="loadItems" />
        </div>
      </div>

      <div v-loading="loading" class="asset-grid">
        <el-empty v-if="!loading && items.length === 0" description="暂无素材" />
        <article v-for="item in items" v-else :key="item.id" class="asset-card">
          <div class="asset-preview" @click="openAsset(item.url)">
            <el-image
              v-if="item.type === 'image'"
              :src="item.url"
              fit="cover"
              class="asset-image"
              :preview-src-list="[item.url]"
              preview-teleported
            />
            <video v-else-if="item.type === 'video'" :src="item.url" class="asset-video" controls preload="metadata" />
            <div v-else class="asset-audio-placeholder">音频素材</div>
          </div>
          <div class="asset-body">
            <div class="asset-top">
              <h3>{{ item.name }}</h3>
              <div class="asset-top-actions">
                <el-tag size="small" :type="assetTagMap[item.type]">
                  {{ assetTypeLabelMap[item.type] }}
                </el-tag>
                <el-button text class="favorite-btn" @click="toggleFavorite(item)">
                  <el-icon :class="{ 'is-favorite': item.is_favorite }">
                    <StarFilled v-if="item.is_favorite" />
                    <Star v-else />
                  </el-icon>
                </el-button>
              </div>
            </div>
            <div class="asset-meta">
              <span v-if="item.category">{{ item.category }}</span>
              <span>{{ formatTime(item.created_at) }}</span>
            </div>
            <div class="asset-actions">
              <el-button size="small" @click="openAsset(item.url)">查看</el-button>
              <el-button size="small" type="danger" plain @click="deleteItem(item.id)">删除</el-button>
            </div>
          </div>
        </article>
      </div>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[12, 24, 48]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadItems"
          @size-change="loadItems"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Star, StarFilled } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AppHeader } from '@/components/common'
import { assetAPI } from '@/api/asset'
import type { Asset, AssetType } from '@/types/asset'

const loading = ref(false)
const items = ref<Asset[]>([])
const total = ref(0)
const activeTab = ref<AssetType>('image')

const assetTypeLabelMap: Record<AssetType, string> = {
  image: '图片',
  video: '视频',
  audio: '音频'
}

const assetTagMap: Record<AssetType, 'info' | 'success' | 'warning'> = {
  image: 'info',
  video: 'success',
  audio: 'warning'
}

const query = reactive({
  page: 1,
  page_size: 12,
  keyword: ''
})

const loadItems = async () => {
  loading.value = true
  try {
    const res = await assetAPI.listAssets({
      page: query.page,
      page_size: query.page_size,
      type: activeTab.value,
      search: query.keyword || undefined
    })
    items.value = res.items || []
    total.value = res.pagination?.total || 0
  } catch (error: any) {
    ElMessage.error(error?.message || '加载素材库失败')
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  query.page = 1
  loadItems()
}

const openAsset = (url: string) => {
  window.open(url, '_blank')
}

const deleteItem = async (id: string) => {
  try {
    await ElMessageBox.confirm('确认删除该素材吗？', '提示', { type: 'warning' })
    await assetAPI.deleteAsset(id)
    ElMessage.success('删除成功')
    await loadItems()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error?.message || '删除失败')
    }
  }
}

const toggleFavorite = async (item: Asset) => {
  try {
    const next = !item.is_favorite
    await assetAPI.updateAsset(item.id, { is_favorite: next })
    item.is_favorite = next
    ElMessage.success(next ? '已加入收藏' : '已取消收藏')
  } catch (error: any) {
    ElMessage.error(error?.message || '更新收藏失败')
  }
}

const formatTime = (value?: string) => {
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

onMounted(loadItems)
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

.asset-toolbar {
  display: grid;
  grid-template-columns: minmax(240px, 360px) minmax(280px, 420px);
  justify-content: space-between;
  gap: 12px;
  padding: 12px;
}

.asset-tabs {
  padding: 0 12px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--bg-card);
}

.asset-search {
  display: flex;
  align-items: center;
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
  padding: 0 12px 12px;
}

.asset-card {
  overflow: hidden;
  border: 1px solid var(--border-primary);
  border-radius: 18px;
  background: var(--bg-card);
}

.asset-preview {
  aspect-ratio: 1 / 1;
  background: var(--bg-muted);
  cursor: pointer;
}

.asset-image,
.asset-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.asset-audio-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 14px;
}

.asset-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
}

.asset-top,
.asset-meta,
.asset-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.asset-top h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
}

.asset-top-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.favorite-btn {
  padding: 0;
}

.favorite-btn :deep(.el-icon) {
  font-size: 18px;
}

.favorite-btn :deep(.el-icon.is-favorite) {
  color: #f59e0b;
}

.asset-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.pagination {
  padding: 12px;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 768px) {
  .asset-toolbar {
    grid-template-columns: 1fr;
  }
}
</style>
