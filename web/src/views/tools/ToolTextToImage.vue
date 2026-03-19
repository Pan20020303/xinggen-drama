<template>
  <div class="page-container">
    <div class="content-wrapper animate-fade-in">
      <AppHeader :fixed="false">
        <template #left>
          <div class="tool-header-left">
            <el-button text class="back-btn" @click="router.push('/tools')">
              <el-icon><ArrowLeft /></el-icon>
              <span>返回工具箱</span>
            </el-button>
            <div class="page-title">
              <h1>图片生成</h1>
              <span class="subtitle">支持文生图和参考生图，生成结果可加入素材库管理</span>
            </div>
          </div>
        </template>
      </AppHeader>

      <div class="tool-layout">
        <el-card class="tool-form-card" shadow="never">
          <div class="mode-switch">
            <button
              type="button"
              class="mode-button"
              :class="{ 'is-active': generationMode === 'text' }"
              @click="generationMode = 'text'"
            >
              文生图
            </button>
            <button
              type="button"
              class="mode-button"
              :class="{ 'is-active': generationMode === 'reference' }"
              @click="generationMode = 'reference'"
            >
              参考生图
            </button>
          </div>

          <el-form label-position="top" @submit.prevent>
            <template v-if="generationMode === 'reference'">
              <el-form-item label="参考图" required>
                <div class="reference-panel">
                  <div
                    class="reference-dropzone"
                    :class="{ 'has-image': !!selectedReference }"
                    @click="openReferencePicker"
                  >
                    <template v-if="selectedReference">
                      <img :src="selectedReference.url" class="reference-preview" />
                      <div class="reference-overlay">
                        <span>{{ selectedReference.name }}</span>
                      </div>
                    </template>
                    <template v-else>
                      <el-icon class="reference-icon"><PictureFilled /></el-icon>
                      <span>点击上传参考图</span>
                    </template>
                  </div>

                  <div class="reference-actions">
                    <el-button :loading="uploadingReference" @click="openReferencePicker">上传参考图</el-button>
                    <el-button @click="showReferenceDialog = true">从素材库选择</el-button>
                    <el-button v-if="selectedReference" text type="danger" @click="clearReferenceImage">清空参考图</el-button>
                  </div>

                  <input
                    ref="referenceInputRef"
                    type="file"
                    accept="image/png,image/jpeg,image/webp,image/gif"
                    class="hidden-file-input"
                    @change="handleReferenceFileChange"
                  />
                </div>
              </el-form-item>
            </template>

            <el-form-item label="提示词" required>
              <el-input
                v-model="form.prompt"
                type="textarea"
                :rows="7"
                maxlength="2000"
                show-word-limit
                :placeholder="generationMode === 'reference' ? '描述希望如何参考已有图片生成新图' : '描述你想生成的图片内容'"
              />
            </el-form-item>

            <el-form-item label="模型">
              <el-select v-model="form.model" class="full-width" placeholder="请选择模型">
                <el-option
                  v-for="option in modelOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="图片比例">
              <div class="ratio-grid">
                <button
                  v-for="ratio in ratioOptions"
                  :key="ratio.key"
                  type="button"
                  class="ratio-option"
                  :class="{ 'is-active': form.ratio === ratio.key }"
                  @click="applyRatio(ratio.key)"
                >
                  <span class="ratio-label">{{ ratio.key }}</span>
                  <span class="ratio-size">{{ ratio.width }} x {{ ratio.height }}</span>
                </button>
              </div>
            </el-form-item>

            <el-form-item label="图片尺寸">
              <div class="dimension-row">
                <el-input-number v-model="form.width" :min="256" :max="4096" :step="64" controls-position="right" />
                <span class="dimension-separator">×</span>
                <el-input-number v-model="form.height" :min="256" :max="4096" :step="64" controls-position="right" />
              </div>
            </el-form-item>

            <el-form-item label="图片数量">
              <el-radio-group v-model="form.count" class="count-group">
                <el-radio-button :label="1">1</el-radio-button>
                <el-radio-button :label="2">2</el-radio-button>
                <el-radio-button :label="3">3</el-radio-button>
                <el-radio-button :label="4">4</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <div class="tool-footer">
              <div class="cost-hint">预计消耗：{{ estimatedCredits }} 积分</div>
              <el-button type="primary" size="large" :loading="submitting" class="submit-btn" @click="handleGenerate">
                立即生成
              </el-button>
            </div>
          </el-form>
        </el-card>

        <el-card class="tool-history-card" shadow="never">
          <template #header>
            <div class="history-header">
              <div class="history-filter-row">
                <el-radio-group v-model="activeMediaTab" size="small" @change="loadRightPanel">
                  <el-radio-button label="image">图片</el-radio-button>
                  <el-radio-button label="video">视频</el-radio-button>
                  <el-radio-button label="audio">音频</el-radio-button>
                </el-radio-group>

                <div class="history-toggles">
                  <el-checkbox v-model="onlyFavorites" @change="handleFavoriteFilterChange">我的收藏</el-checkbox>
                  <el-checkbox v-model="onlyProcessing" @change="handleProcessingFilterChange">进行中</el-checkbox>
                </div>
              </div>

              <div class="history-header-actions">
                <el-button
                  v-if="bulkDeleteTargets.length > 0 && bulkDeleteSelectionMode"
                  text
                  @click="cancelBulkDeleteSelection"
                >
                  取消选择
                </el-button>
                <el-button
                  v-if="bulkDeleteTargets.length > 0"
                  text
                  type="danger"
                  :icon="Delete"
                  :disabled="bulkDeleteSelectionMode && selectedBulkDeleteTargets.length === 0"
                  :loading="bulkDeleting"
                  @click="handleBulkDelete"
                >
                  {{ bulkDeleteSelectionMode ? `删除已选 (${selectedBulkDeleteTargets.length})` : '批量删除' }}
                </el-button>
                <el-button text @click="loadRightPanel">刷新</el-button>
              </div>
            </div>
          </template>

          <div v-loading="showHistoryLoading" class="history-panel">
            <el-empty
              v-if="showHistoryEmptyState"
              :description="onlyProcessing ? '还没有进行中的任务' : '还没有素材记录'"
            />

            <div v-else-if="onlyProcessing" class="history-grid">
              <article
                v-for="image in processingImages"
                :key="image.id"
                class="history-item"
                :class="{ 'is-selectable': bulkDeleteSelectionMode, 'is-selected': isGeneratedImageSelected(image) }"
                @click="bulkDeleteSelectionMode && toggleGeneratedImageSelection(image)"
              >
                <div class="history-image-wrap">
                  <button
                    v-if="bulkDeleteSelectionMode"
                    type="button"
                    class="selection-check"
                    :class="{ 'is-selected': isGeneratedImageSelected(image) }"
                    @click.stop="toggleGeneratedImageSelection(image)"
                  >
                    <el-icon v-if="isGeneratedImageSelected(image)"><Check /></el-icon>
                  </button>
                  <div class="history-image-placeholder" :class="`is-${image.status}`">
                    <el-icon v-if="image.status === 'processing'" class="loading-icon"><Loading /></el-icon>
                    <el-icon v-else><Picture /></el-icon>
                    <span>{{ statusTextMap[image.status] }}</span>
                  </div>
                </div>
                <div class="history-item-body">
                  <div class="history-item-top">
                    <el-tag size="small" :type="statusTagMap[image.status]">{{ statusTextMap[image.status] }}</el-tag>
                    <span class="history-time">{{ formatTime(image.created_at) }}</span>
                  </div>
                  <div class="history-prompt">{{ image.prompt }}</div>
                  <div class="history-meta">
                    <span>{{ image.model || '默认模型' }}</span>
                    <span v-if="image.width && image.height">{{ image.width }}×{{ image.height }}</span>
                  </div>
                </div>
              </article>
            </div>

            <div v-else class="history-grid history-masonry">
              <article
                v-for="item in displayItems"
                :key="item.key"
                class="history-item"
                :class="{ 'is-image-card': isImageDisplayItem(item) }"
              >
                <div v-if="item.kind === 'generated-group'" class="history-generated-group">
                  <div class="group-badge">{{ item.images.length }} 张</div>
                  <div v-if="!bulkDeleteSelectionMode" class="group-toolbar">
                    <button
                      type="button"
                      class="overlay-chip-btn"
                      @click.stop="importGroupToAssetLibrary(item.images)"
                    >
                      整组加入素材库
                    </button>
                  </div>
                  <div
                    class="history-generated-group-grid"
                    :class="`is-count-${item.images.length}`"
                  >
                    <div
                        v-for="image in item.images"
                        :key="image.id"
                        class="history-generated-group-cell"
                        :class="{ 'is-selectable': bulkDeleteSelectionMode, 'is-selected': isGeneratedImageSelected(image) }"
                        :style="{ aspectRatio: getGeneratedAspectRatio(image) }"
                        @click="bulkDeleteSelectionMode ? toggleGeneratedImageSelection(image) : openGeneratedGroupPreview(item.images, image.id)"
                      >
                      <button
                        v-if="bulkDeleteSelectionMode"
                        type="button"
                        class="selection-check"
                        :class="{ 'is-selected': isGeneratedImageSelected(image) }"
                        @click.stop="toggleGeneratedImageSelection(image)"
                      >
                        <el-icon v-if="isGeneratedImageSelected(image)"><Check /></el-icon>
                      </button>
                      <el-image
                        :src="getGeneratedImageSrc(image)"
                        fit="cover"
                        class="history-image"
                      />
                      <div v-if="!bulkDeleteSelectionMode" class="history-image-toolbar">
                        <button
                          type="button"
                          class="overlay-chip-btn"
                          @click.stop="importToAssetLibrary(image)"
                        >
                          加入素材库
                        </button>
                        <button
                          type="button"
                          class="overlay-chip-btn"
                          @click.stop="reuseTemplate(image)"
                        >
                          模板复用
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                <div
                  v-else-if="isImageDisplayItem(item)"
                  class="history-image-card"
                  :class="{ 'is-selectable': bulkDeleteSelectionMode && item.kind === 'generated', 'is-selected': item.kind === 'generated' && isGeneratedImageSelected(item.image) }"
                  :style="{ aspectRatio: getItemAspectRatio(item) }"
                  @click="bulkDeleteSelectionMode && item.kind === 'generated' ? toggleGeneratedImageSelection(item.image) : openPreview(item)"
                  @keydown.enter="bulkDeleteSelectionMode && item.kind === 'generated' ? toggleGeneratedImageSelection(item.image) : openPreview(item)"
                  @keydown.space.prevent="bulkDeleteSelectionMode && item.kind === 'generated' ? toggleGeneratedImageSelection(item.image) : openPreview(item)"
                  tabindex="0"
                >
                  <button
                    v-if="bulkDeleteSelectionMode && item.kind === 'generated'"
                    type="button"
                    class="selection-check"
                    :class="{ 'is-selected': isGeneratedImageSelected(item.image) }"
                    @click.stop="toggleGeneratedImageSelection(item.image)"
                  >
                    <el-icon v-if="isGeneratedImageSelected(item.image)"><Check /></el-icon>
                  </button>
                  <el-image
                    :src="getItemImageSrc(item)"
                    fit="cover"
                    class="history-image"
                  />
                  <div v-if="!bulkDeleteSelectionMode" class="history-image-toolbar">
                    <button
                      v-if="item.kind === 'asset'"
                      type="button"
                      class="overlay-icon-btn"
                      @click.stop="toggleFavorite(item.asset)"
                    >
                      <el-icon :class="{ 'is-favorite': item.asset.is_favorite }">
                        <StarFilled v-if="item.asset.is_favorite" />
                        <Star v-else />
                      </el-icon>
                    </button>
                    <button
                      v-if="item.kind === 'generated'"
                      type="button"
                      class="overlay-chip-btn"
                      @click.stop="importToAssetLibrary(item.image)"
                    >
                      加入素材库
                    </button>
                    <button
                      v-if="item.kind === 'generated'"
                      type="button"
                      class="overlay-chip-btn"
                      @click.stop="reuseTemplate(item.image)"
                    >
                      模板复用
                    </button>
                  </div>
                </div>

                <template v-else>
                  <div class="history-image-wrap">
                    <video
                      v-if="isVideoAsset(item)"
                      :src="getAssetMediaUrl(item)"
                      class="history-video"
                      controls
                      preload="metadata"
                    />
                    <div v-else class="history-image-placeholder is-audio">
                      <el-icon><Headset /></el-icon>
                      <span>音频素材</span>
                    </div>
                  </div>
                  <div class="history-item-body">
                    <div class="history-item-top">
                      <el-tag
                        size="small"
                        :type="getDisplayItemTagType(item)"
                      >
                        {{ getDisplayItemLabel(item) }}
                      </el-tag>
                    </div>
                    <div class="history-prompt">{{ getDisplayItemTitle(item) }}</div>
                    <div class="history-meta">
                      <span>{{ getDisplayItemCategory(item) }}</span>
                      <span>{{ formatTime(getItemCreatedAt(item)) }}</span>
                    </div>
                  </div>
                </template>
              </article>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <el-dialog v-model="showReferenceDialog" title="从素材库选择参考图" width="880px">
      <div v-loading="referenceDialogLoading" class="reference-library-grid">
        <el-empty v-if="!referenceDialogLoading && referenceCandidates.length === 0" description="素材库暂无图片素材" />
        <button
          v-for="item in referenceCandidates"
          v-else
          :key="item.id"
          type="button"
          class="reference-library-item"
          :class="{ 'is-selected': selectedReference?.url === item.url }"
          @click="selectReferenceAsset(item)"
        >
          <img :src="item.url" :alt="item.name" />
          <span>{{ item.name }}</span>
        </button>
      </div>
    </el-dialog>

    <el-dialog
      v-model="showPreviewDialog"
      width="1200px"
      class="image-preview-dialog"
      destroy-on-close
      align-center
      >
      <template v-if="previewContent">
        <div class="preview-dialog-layout">
          <div class="preview-dialog-image">
            <img :src="previewContent.imageSrc" alt="preview" />
            <div v-if="previewItem?.kind === 'generated-group'" class="preview-group-nav">
              <button type="button" class="preview-nav-btn" @click="showPreviousPreviewImage">
                上一张
              </button>
              <span class="preview-nav-count">{{ previewCurrentIndex + 1 }} / {{ previewItem.images.length }}</span>
              <button type="button" class="preview-nav-btn" @click="showNextPreviewImage">
                下一张
              </button>
            </div>
          </div>
          <aside class="preview-dialog-sidebar">
            <div class="preview-meta-line">
              <span>{{ formatDateOnly(previewContent.createdAt) }}</span>
              <span>{{ previewContent.sourceLabel }}</span>
            </div>

            <div class="preview-section">
              <div class="preview-section-title">图片提示词</div>
              <div class="preview-prompt-text">
                {{ previewContent.prompt }}
              </div>
            </div>

            <div class="preview-submeta">
              <span v-if="previewContent.model">{{ previewContent.model }}</span>
              <span>{{ previewContent.ratio }}</span>
            </div>

            <div class="preview-action-row">
              <el-button
                v-if="previewContent.mode === 'generated'"
                @click="reuseTemplate(previewContent.image); showPreviewDialog = false"
              >
                做同款
              </el-button>
              <el-button
                v-if="previewContent.mode === 'generated'"
                type="primary"
                @click="setGeneratedAsReference(previewContent.image); showPreviewDialog = false"
              >
                用作参考图
              </el-button>
              <el-button
                v-if="previewContent.mode === 'asset'"
                @click="setAssetAsReference(previewContent.asset); showPreviewDialog = false"
              >
                用作参考图
              </el-button>
              <el-button
                v-if="previewContent.mode === 'generated'"
                type="primary"
                plain
                @click="importToAssetLibrary(previewContent.image)"
              >
                加入素材库
              </el-button>
              <el-button
                v-if="previewItem?.kind === 'generated-group'"
                plain
                @click="importGroupToAssetLibrary(previewItem.images)"
              >
                整组加入素材库
              </el-button>
              <el-button
                v-if="previewContent.mode === 'asset'"
                plain
                @click="toggleFavorite(previewContent.asset)"
              >
                {{ previewContent.asset.is_favorite ? '取消收藏' : '收藏' }}
              </el-button>
            </div>
          </aside>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowLeft,
  Check,
  Delete,
  Headset,
  Loading,
  Picture,
  PictureFilled,
  Star,
  StarFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { AppHeader } from '@/components/common'
import { assetAPI } from '@/api/asset'
import { imageAPI } from '@/api/image'
import { uploadAPI } from '@/api/upload'
import { useAuthStore } from '@/stores/auth'
import { usePricingStore } from '@/stores/pricing'
import type { Asset, AssetType } from '@/types/asset'
import type { ImageGeneration, ImageStatus } from '@/types/image'
import { getImageUrl } from '@/utils/image'
import {
  collectBulkDeletableImages,
  collectSelectedBulkDeleteImages,
  shouldShowHistoryEmptyState,
  groupGeneratedImages
} from './toolTextToImage.helpers'

type RatioKey = '16:9' | '9:16' | '4:3' | '3:4' | '1:1'
type GenerationMode = 'text' | 'reference'

interface ModelOption {
  label: string
  value: string
  provider: string
  cost: number
}

interface ReferenceSource {
  url: string
  name: string
  localPath?: string
}

type DisplayItem =
  | { key: string; kind: 'asset'; asset: Asset }
  | { key: string; kind: 'generated'; image: ImageGeneration }
  | { key: string; kind: 'generated-group'; images: ImageGeneration[] }

type SingleImageDisplayItem = Extract<DisplayItem, { kind: 'asset' | 'generated' }>
type PreviewSourceItem = Extract<PreviewItem, { kind: 'asset' | 'generated' }>

type PreviewItem =
  | { kind: 'asset'; asset: Asset }
  | { kind: 'generated'; image: ImageGeneration }
  | { kind: 'generated-group'; images: ImageGeneration[]; currentIndex: number }

const router = useRouter()
const authStore = useAuthStore()
const pricingStore = usePricingStore()

const submitting = ref(false)
const historyLoading = ref(false)
const historyRefreshing = ref(false)
const bulkDeleting = ref(false)
const bulkDeleteSelectionMode = ref(false)
const uploadingReference = ref(false)
const showReferenceDialog = ref(false)
const referenceDialogLoading = ref(false)
const showPreviewDialog = ref(false)

const generationMode = ref<GenerationMode>('text')
const activeMediaTab = ref<AssetType>('image')
const onlyFavorites = ref(false)
const onlyProcessing = ref(false)

const materialItems = ref<Asset[]>([])
const processingImages = ref<ImageGeneration[]>([])
const completedToolboxImages = ref<ImageGeneration[]>([])
const referenceCandidates = ref<Asset[]>([])
const importedImageIds = ref<number[]>([])
const selectedReference = ref<ReferenceSource | null>(null)
const referenceInputRef = ref<HTMLInputElement | null>(null)
const previewItem = ref<PreviewItem | null>(null)
const selectedBulkDeleteIds = ref<string[]>([])
const previewCurrentIndex = computed(() => {
  return previewItem.value?.kind === 'generated-group' ? previewItem.value.currentIndex : 0
})

const ratioOptions: Array<{ key: RatioKey; width: number; height: number }> = [
  { key: '16:9', width: 2560, height: 1440 },
  { key: '9:16', width: 1440, height: 2560 },
  { key: '4:3', width: 2218, height: 1664 },
  { key: '3:4', width: 1664, height: 2218 },
  { key: '1:1', width: 1920, height: 1920 }
]

const form = reactive({
  prompt: '',
  model: '',
  ratio: '16:9' as RatioKey,
  width: 2560,
  height: 1440,
  count: 1
})

const statusTextMap: Record<ImageStatus, string> = {
  pending: '等待中',
  processing: '生成中',
  completed: '已完成',
  failed: '失败'
}

const statusTagMap: Record<ImageStatus, 'info' | 'warning' | 'success' | 'danger'> = {
  pending: 'info',
  processing: 'warning',
  completed: 'success',
  failed: 'danger'
}

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

const modelOptions = computed<ModelOption[]>(() => {
  const configs = pricingStore.pricing?.platform_configs ?? []
  const imageConfigs = configs.filter((config) => config.service_type === 'image' && config.is_active)
  const options = imageConfigs.flatMap((config) => {
    const models = Array.isArray(config.model) ? config.model : []
    return models.map((modelName) => ({
      label: `${modelName} (${config.provider})`,
      value: modelName,
      provider: config.provider,
      cost: config.credit_cost ?? 0
    }))
  })

  if (options.length > 0) {
    return options
  }

  const fallbackModel = pricingStore.getDefaultModel('image') || '默认模型'
  return [{
    label: fallbackModel,
    value: fallbackModel,
    provider: 'openai',
    cost: pricingStore.getDefaultCost('image')
  }]
})

const selectedModel = computed(() => {
  return modelOptions.value.find((option) => option.value === form.model) || modelOptions.value[0]
})

const estimatedCredits = computed(() => {
  return (selectedModel.value?.cost ?? pricingStore.getDefaultCost('image')) * form.count
})

const requiresLargeImageSize = computed(() => {
  const modelName = selectedModel.value?.value?.toLowerCase?.() || ''
  return modelName.includes('seedream-4-5') || modelName.includes('seedream-4.5')
})

const displayItems = computed<DisplayItem[]>(() => {
  const items: DisplayItem[] = materialItems.value.map((asset) => ({
    key: `asset-${asset.id}`,
    kind: 'asset',
    asset
  }))

  if (!onlyFavorites.value && activeMediaTab.value === 'image') {
    return [
      ...groupGeneratedImages(completedToolboxImages.value),
      ...items
    ]
  }

  return items
})

const bulkDeleteTargets = computed(() => {
  return collectBulkDeletableImages({
    activeMediaTab: activeMediaTab.value,
    onlyFavorites: onlyFavorites.value,
    onlyProcessing: onlyProcessing.value,
    processingImages: processingImages.value,
    completedToolboxImages: completedToolboxImages.value
  })
})

const selectedBulkDeleteTargets = computed(() => {
  return collectSelectedBulkDeleteImages(bulkDeleteTargets.value, selectedBulkDeleteIds.value)
})

const showHistoryLoading = computed(() => {
  return historyLoading.value && !historyRefreshing.value
})

const showHistoryEmptyState = computed(() => {
  return shouldShowHistoryEmptyState({
    historyLoading: historyLoading.value,
    onlyProcessing: onlyProcessing.value,
    displayItemCount: displayItems.value.length,
    processingItemCount: processingImages.value.length
  })
})

const isImageDisplayItem = (item: DisplayItem): item is SingleImageDisplayItem => {
  return (item.kind === 'asset' && item.asset.type === 'image') || item.kind === 'generated'
}

const parseSize = (size?: string) => {
  if (!size) return null
  const match = size.match(/^(\d+)x(\d+)$/i)
  if (!match) return null
  return {
    width: Number(match[1]),
    height: Number(match[2])
  }
}

const getItemImageSrc = (item: SingleImageDisplayItem) => {
  if (item.kind === 'asset') {
    return item.asset.local_path ? getImageUrl(item.asset) : item.asset.url
  }

  return getImageUrl(item.image) || item.image.image_url || ''
}

const getGeneratedImageSrc = (image: ImageGeneration) => {
  return getImageUrl(image) || image.image_url || ''
}

const getItemAspectRatio = (item: SingleImageDisplayItem) => {
  const width = item.kind === 'asset' ? item.asset.width : item.image.width
  const height = item.kind === 'asset' ? item.asset.height : item.image.height
  const parsed = item.kind === 'generated' ? parseSize(item.image.size) : null
  const safeWidth = width || parsed?.width || 3
  const safeHeight = height || parsed?.height || 4
  return `${safeWidth} / ${safeHeight}`
}

const getGeneratedAspectRatio = (image: ImageGeneration) => {
  const width = image.width
  const height = image.height
  const parsed = parseSize(image.size)
  const safeWidth = width || parsed?.width || 3
  const safeHeight = height || parsed?.height || 4
  return `${safeWidth} / ${safeHeight}`
}

const isVideoAsset = (item: DisplayItem): item is Extract<DisplayItem, { kind: 'asset' }> => {
  return item.kind === 'asset' && item.asset.type === 'video'
}

const isGeneratedImageSelected = (image: ImageGeneration) => {
  return selectedBulkDeleteIds.value.includes(String(image.id))
}

const toggleGeneratedImageSelection = (image: ImageGeneration) => {
  const imageId = String(image.id)
  if (selectedBulkDeleteIds.value.includes(imageId)) {
    selectedBulkDeleteIds.value = selectedBulkDeleteIds.value.filter((id) => id !== imageId)
    return
  }

  selectedBulkDeleteIds.value = [...selectedBulkDeleteIds.value, imageId]
}

const enterBulkDeleteSelectionMode = () => {
  bulkDeleteSelectionMode.value = true
}

const cancelBulkDeleteSelection = () => {
  bulkDeleteSelectionMode.value = false
  selectedBulkDeleteIds.value = []
}

const getDisplayItemTagType = (item: DisplayItem) => {
  return item.kind === 'asset' ? assetTagMap[item.asset.type] : 'info'
}

const getDisplayItemLabel = (item: DisplayItem) => {
  return item.kind === 'asset' ? assetTypeLabelMap[item.asset.type] : '工具图片'
}

const getAssetMediaUrl = (item: DisplayItem) => {
  return item.kind === 'asset' ? item.asset.url : ''
}

const getDisplayItemTitle = (item: DisplayItem) => {
  if (item.kind === 'asset') {
    return item.asset.name
  }
  if (item.kind === 'generated') {
    return item.image.prompt
  }
  return item.images[0]?.prompt || '暂无提示词'
}

const getDisplayItemCategory = (item: DisplayItem) => {
  if (item.kind === 'asset') {
    return item.asset.category || '未分类'
  }
  if (item.kind === 'generated') {
    return item.image.model || '默认模型'
  }
  return item.images[0]?.model || '默认模型'
}

const getItemCreatedAt = (item: DisplayItem) => {
  if (item.kind === 'asset') {
    return item.asset.created_at
  }
  if (item.kind === 'generated') {
    return item.image.created_at
  }
  return item.images[0]?.created_at
}

const previewContent = computed(() => {
  if (!previewItem.value) return null

  if (previewItem.value.kind === 'asset') {
    return {
      mode: 'asset' as const,
      asset: previewItem.value.asset,
      imageSrc: previewItem.value.asset.local_path ? getImageUrl(previewItem.value.asset) : previewItem.value.asset.url,
      createdAt: previewItem.value.asset.created_at,
      sourceLabel: '图片素材',
      prompt: previewItem.value.asset.description || previewItem.value.asset.name || '暂无描述',
      model: '',
      ratio: getPreviewRatio(previewItem.value)
    }
  }

  const image = previewItem.value.kind === 'generated-group'
    ? previewItem.value.images[previewItem.value.currentIndex]
    : previewItem.value.image

  return {
    mode: 'generated' as const,
    image,
    imageSrc: getGeneratedImageSrc(image),
    createdAt: image.created_at,
    sourceLabel: '内容由 AI 生成',
    prompt: image.prompt || '暂无提示词',
    model: image.model || '默认模型',
    ratio: getGeneratedAspectRatioLabel(image)
  }
})

const getPreviewPrompt = (item: PreviewSourceItem) => {
  if (item.kind === 'generated') {
    return item.image.prompt || '暂无提示词'
  }

  return item.asset.description || item.asset.name || '暂无描述'
}

const getPreviewRatio = (item: PreviewSourceItem) => {
  const width = item.kind === 'asset' ? item.asset.width : item.image.width
  const height = item.kind === 'asset' ? item.asset.height : item.image.height
  const parsed = item.kind === 'generated' ? parseSize(item.image.size) : null
  const safeWidth = width || parsed?.width
  const safeHeight = height || parsed?.height

  if (!safeWidth || !safeHeight) {
    return '比例未知'
  }

  return `${safeWidth}:${safeHeight}`
}

const getGeneratedAspectRatioLabel = (image: ImageGeneration) => {
  const width = image.width
  const height = image.height
  const parsed = parseSize(image.size)
  const safeWidth = width || parsed?.width
  const safeHeight = height || parsed?.height

  if (!safeWidth || !safeHeight) {
    return '比例未知'
  }

  return `${safeWidth}:${safeHeight}`
}

const formatDateOnly = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString('zh-CN')
}

const openPreview = (item: SingleImageDisplayItem) => {
  previewItem.value = item.kind === 'asset'
    ? { kind: 'asset', asset: item.asset }
    : { kind: 'generated', image: item.image }
  showPreviewDialog.value = true
}

const openGeneratedGroupPreview = (images: ImageGeneration[], activeId: number | string) => {
  const currentIndex = Math.max(images.findIndex((image) => String(image.id) === String(activeId)), 0)
  previewItem.value = {
    kind: 'generated-group',
    images,
    currentIndex
  }
  showPreviewDialog.value = true
}

const showPreviousPreviewImage = () => {
  if (!previewItem.value || previewItem.value.kind !== 'generated-group') return
  const total = previewItem.value.images.length
  previewItem.value = {
    ...previewItem.value,
    currentIndex: (previewItem.value.currentIndex - 1 + total) % total
  }
}

const showNextPreviewImage = () => {
  if (!previewItem.value || previewItem.value.kind !== 'generated-group') return
  const total = previewItem.value.images.length
  previewItem.value = {
    ...previewItem.value,
    currentIndex: (previewItem.value.currentIndex + 1) % total
  }
}

const applyRatio = (ratioKey: RatioKey) => {
  const matched = ratioOptions.find((item) => item.key === ratioKey)
  if (!matched) return
  form.ratio = matched.key
  form.width = matched.width
  form.height = matched.height
}

const resetToolForm = () => {
  generationMode.value = 'text'
  form.prompt = ''
  form.count = 1
  clearReferenceImage()
  applyRatio('16:9')
  form.model = modelOptions.value[0]?.value || ''
}

const ensureDefaultSelections = () => {
  if (!form.model && modelOptions.value.length > 0) {
    form.model = modelOptions.value[0].value
  }

  if (requiresLargeImageSize.value && form.width * form.height < 3686400) {
    applyRatio(form.ratio)
  }
}

const openReferencePicker = () => {
  referenceInputRef.value?.click()
}

const clearReferenceImage = () => {
  selectedReference.value = null
  if (referenceInputRef.value) {
    referenceInputRef.value.value = ''
  }
}

const handleReferenceFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  uploadingReference.value = true
  try {
    const result = await uploadAPI.uploadImage(file)
    selectedReference.value = {
      url: result.url,
      name: result.filename || file.name,
      localPath: result.local_path
    }
    ElMessage.success('参考图上传成功')
  } catch (error: any) {
    ElMessage.error(error?.message || '参考图上传失败')
  } finally {
    uploadingReference.value = false
    input.value = ''
  }
}

const loadReferenceCandidates = async () => {
  referenceDialogLoading.value = true
  try {
    const result = await assetAPI.listAssets({
      type: 'image',
      page: 1,
      page_size: 60
    })
    referenceCandidates.value = result.items || []
  } catch (error: any) {
    ElMessage.error(error?.message || '加载素材库失败')
  } finally {
    referenceDialogLoading.value = false
  }
}

const selectReferenceAsset = (item: Asset) => {
  selectedReference.value = {
    url: item.local_path ? getImageUrl(item) : item.url,
    name: item.name,
    localPath: item.local_path
  }
  showReferenceDialog.value = false
}

const resolveReferenceSource = (reference: string) => {
  if (!reference) return null

  if (reference.startsWith('http') || reference.startsWith('data:')) {
    return {
      url: reference,
      name: '参考图'
    } satisfies ReferenceSource
  }

  const normalizedLocalPath = reference.replace(/^\/?static\//, '').replace(/^\/+/, '')

  return {
    url: `/static/${normalizedLocalPath}`,
    name: '参考图',
    localPath: normalizedLocalPath
  } satisfies ReferenceSource
}

const reuseTemplate = (image: ImageGeneration) => {
  form.prompt = image.prompt || ''

  const references = Array.isArray(image.reference_images) ? image.reference_images.filter(Boolean) : []
  const firstReference = references.length > 0 ? resolveReferenceSource(references[0]) : null

  if (firstReference) {
    generationMode.value = 'reference'
    selectedReference.value = firstReference
  } else {
    generationMode.value = 'text'
    selectedReference.value = null
  }

  ElMessage.success(firstReference ? '已复用提示词和参考图' : '已复用提示词')
}

const setGeneratedAsReference = (image: ImageGeneration) => {
  const source = getImageUrl(image) || image.image_url
  if (!source) {
    ElMessage.warning('当前图片没有可用的参考图地址')
    return
  }

  generationMode.value = 'reference'
  selectedReference.value = {
    url: source,
    name: '生成结果',
    localPath: image.local_path
  }
  ElMessage.success('已设置为参考图')
}

const setAssetAsReference = (asset: Asset) => {
  generationMode.value = 'reference'
  selectedReference.value = {
    url: asset.local_path ? getImageUrl(asset) : asset.url,
    name: asset.name,
    localPath: asset.local_path
  }
  ElMessage.success('已设置为参考图')
}

const loadMaterialAssets = async () => {
  const result = await assetAPI.listAssets({
    type: activeMediaTab.value,
    is_favorite: onlyFavorites.value || undefined,
    page: 1,
    page_size: 60
  })
  materialItems.value = result.items || []
  importedImageIds.value = activeMediaTab.value === 'image'
    ? materialItems.value.map((item) => Number(item.image_gen_id)).filter((id) => !Number.isNaN(id))
    : []

  if (activeMediaTab.value === 'image' && !onlyFavorites.value) {
    const imagesResult = await imageAPI.listImages({
      page: 1,
      page_size: 60,
      image_type: 'toolbox',
      status: 'completed'
    })
    completedToolboxImages.value = (imagesResult.items || []).filter((item) => !importedImageIds.value.includes(Number(item.id)))
  } else {
    completedToolboxImages.value = []
  }
}

const loadProcessingItems = async () => {
  if (activeMediaTab.value !== 'image') {
    processingImages.value = []
    return
  }

  const result = await imageAPI.listImages({
    page: 1,
    page_size: 60,
    image_type: 'toolbox',
    status: 'processing'
  })

  const pendingResult = await imageAPI.listImages({
    page: 1,
    page_size: 60,
    image_type: 'toolbox',
    status: 'pending'
  })

  const merged = [...(result.items || []), ...(pendingResult.items || [])]
  const seen = new Set<string>()
  processingImages.value = merged.filter((item) => {
    const key = String(item.id)
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

const loadRightPanel = async (options: { silent?: boolean } = {}) => {
  historyRefreshing.value = !!options.silent
  historyLoading.value = true
  try {
    if (onlyProcessing.value) {
      await loadProcessingItems()
      materialItems.value = []
      completedToolboxImages.value = []
      return
    }

    processingImages.value = []
    await loadMaterialAssets()
  } catch (error: any) {
    ElMessage.error(error?.message || '加载记录失败')
  } finally {
    historyLoading.value = false
    historyRefreshing.value = false
  }
}

const handleFavoriteFilterChange = () => {
  loadRightPanel()
}

const handleProcessingFilterChange = () => {
  loadRightPanel()
}

const handleGenerate = async () => {
  if (form.prompt.trim().length < 5) {
    ElMessage.warning('提示词至少需要 5 个字符')
    return
  }

  if (generationMode.value === 'reference' && !selectedReference.value) {
    ElMessage.warning('请先选择一张参考图')
    return
  }

  const model = selectedModel.value
  if (!model) {
    ElMessage.warning('当前没有可用图片模型')
    return
  }

  if (requiresLargeImageSize.value && form.width * form.height < 3686400) {
    ElMessage.warning('当前模型要求图片尺寸至少为 3686400 像素，已请使用更大的比例预设或尺寸')
    return
  }

  const submittedCount = form.count
  submitting.value = true
  try {
    for (let i = 0; i < submittedCount; i += 1) {
      await imageAPI.generateImage({
        image_type: 'toolbox',
        prompt: form.prompt.trim(),
        provider: model.provider,
        model: model.value,
        size: `${form.width}x${form.height}`,
        width: form.width,
        height: form.height,
        reference_images: selectedReference.value
          ? [selectedReference.value.localPath || selectedReference.value.url]
          : undefined
      })
    }

    await authStore.refreshMe()
    resetToolForm()
    onlyProcessing.value = true
    await loadRightPanel()
    ElMessage.success(`已提交 ${submittedCount} 张图片生成任务`)
  } catch (error: any) {
    ElMessage.error(error?.message || '提交生成任务失败')
  } finally {
    submitting.value = false
  }
}

const toggleFavorite = async (item: Asset) => {
  try {
    const next = !item.is_favorite
    await assetAPI.updateAsset(item.id, { is_favorite: next })
    item.is_favorite = next
    ElMessage.success(next ? '已加入收藏' : '已取消收藏')
    if (onlyFavorites.value && !next) {
      materialItems.value = materialItems.value.filter((asset) => asset.id !== item.id)
    }
  } catch (error: any) {
    ElMessage.error(error?.message || '更新收藏失败')
  }
}

const importToAssetLibrary = async (image: ImageGeneration) => {
  try {
    await assetAPI.importFromImage(image.id)
    ElMessage.success('已加入素材库')
    await loadRightPanel()
  } catch (error: any) {
    ElMessage.error(error?.message || '加入素材库失败')
  }
}

const importGroupToAssetLibrary = async (images: ImageGeneration[]) => {
  try {
    const results = await Promise.allSettled(images.map((image) => assetAPI.importFromImage(image.id)))
    const successCount = results.filter((result) => result.status === 'fulfilled').length
    const failedCount = results.length - successCount

    if (successCount > 0) {
      await loadRightPanel()
    }

    if (failedCount === 0) {
      ElMessage.success(`已将 ${successCount} 张图片加入素材库`)
      return
    }

    ElMessage.warning(`已加入 ${successCount} 张图片，${failedCount} 张失败`)
  } catch (error: any) {
    ElMessage.error(error?.message || '整组加入素材库失败')
  }
}

const handleBulkDelete = async () => {
  if (!bulkDeleteSelectionMode.value) {
    enterBulkDeleteSelectionMode()
    return
  }

  const targets = selectedBulkDeleteTargets.value
  if (targets.length === 0 || bulkDeleting.value) {
    return
  }

  const message = `是否删除选中的 ${targets.length} 张图片？`

  try {
    await ElMessageBox.confirm(message, '批量删除生成图片', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger'
    })
  } catch {
    return
  }

  bulkDeleting.value = true
  try {
    const results = await Promise.allSettled(targets.map((image) => imageAPI.deleteImage(image.id)))
    const successCount = results.filter((result) => result.status === 'fulfilled').length
    const failedCount = results.length - successCount

    if (successCount > 0) {
      await loadRightPanel({ silent: true })
    }

    if (failedCount === 0) {
      ElMessage.success(`已删除 ${successCount} 张图片`)
      cancelBulkDeleteSelection()
      return
    }

    ElMessage.warning(`已删除 ${successCount} 张图片，${failedCount} 张删除失败`)
  } catch (error: any) {
    ElMessage.error(error?.message || '批量删除失败')
  } finally {
    bulkDeleting.value = false
  }
}

const formatTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

watch(modelOptions, () => {
  ensureDefaultSelections()
})

watch(() => form.model, () => {
  ensureDefaultSelections()
})

watch(bulkDeleteTargets, (targets) => {
  const validIds = new Set(targets.map((item) => String(item.id)))
  selectedBulkDeleteIds.value = selectedBulkDeleteIds.value.filter((id) => validIds.has(id))

  if (targets.length === 0) {
    bulkDeleteSelectionMode.value = false
  }
})

watch(showReferenceDialog, (visible) => {
  if (visible) {
    loadReferenceCandidates()
  }
})

let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await pricingStore.loadPricing()
  ensureDefaultSelections()
  await loadRightPanel()

  pollTimer = setInterval(() => {
    if (onlyProcessing.value) {
      loadRightPanel({ silent: true })
    }
  }, 5000)
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>

<style scoped>
.tool-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.back-btn {
  padding-left: 0;
}

.tool-layout {
  display: grid;
  grid-template-columns: minmax(360px, 430px) minmax(0, 1fr);
  gap: 20px;
}

.tool-form-card,
.tool-history-card {
  border-radius: 20px;
}

.mode-switch {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-bottom: 20px;
  padding: 6px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--bg-muted);
}

.mode-button {
  height: 40px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.mode-button.is-active {
  background: var(--color-primary);
  color: #fff;
}

.reference-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.reference-dropzone {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 220px;
  border: 1px dashed var(--border-primary);
  border-radius: 16px;
  background: var(--bg-muted);
  color: var(--text-secondary);
  cursor: pointer;
  overflow: hidden;
}

.reference-dropzone.has-image {
  border-style: solid;
}

.reference-icon {
  margin-right: 8px;
  font-size: 22px;
}

.reference-preview {
  width: 100%;
  height: 220px;
  object-fit: cover;
}

.reference-overlay {
  position: absolute;
  inset: auto 0 0;
  padding: 10px 12px;
  background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.65));
  color: #fff;
  font-size: 12px;
}

.reference-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.hidden-file-input {
  display: none;
}

.full-width {
  width: 100%;
}

.ratio-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 10px;
  width: 100%;
}

.ratio-option {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 10px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--bg-card);
  cursor: pointer;
}

.ratio-option.is-active {
  border-color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--bg-card));
}

.ratio-label {
  font-weight: 700;
}

.ratio-size {
  font-size: 12px;
  color: var(--text-secondary);
}

.dimension-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dimension-row :deep(.el-input-number) {
  flex: 1;
}

.dimension-separator {
  color: var(--text-secondary);
}

.count-group {
  width: 100%;
}

.count-group :deep(.el-radio-button) {
  flex: 1;
}

.count-group :deep(.el-radio-button__inner) {
  width: 100%;
}

.tool-footer {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 8px;
}

.cost-hint {
  font-size: 14px;
  color: var(--text-secondary);
}

.submit-btn {
  width: 100%;
}

.history-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.history-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.history-filter-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.history-toggles {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.history-panel {
  min-height: 560px;
}

.history-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
  align-items: start;
}

.history-masonry {
  display: block;
  column-width: 240px;
  column-gap: 16px;
}

.history-item {
  overflow: hidden;
  border: 1px solid var(--border-primary);
  border-radius: 18px;
  background: var(--bg-card);
}

.history-item.is-selectable,
.history-generated-group-cell.is-selectable,
.history-image-card.is-selectable {
  cursor: pointer;
}

.history-item.is-selected,
.history-generated-group-cell.is-selected,
.history-image-card.is-selected {
  box-shadow: inset 0 0 0 2px var(--color-primary);
}

.history-masonry .history-item {
  display: inline-block;
  width: 100%;
  margin-bottom: 16px;
  break-inside: avoid;
}

.history-item.is-image-card {
  border: none;
  background: transparent;
}

.history-generated-group {
  position: relative;
  border-radius: 18px;
  overflow: hidden;
  background: var(--bg-muted);
}

.group-badge {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 2;
  padding: 4px 8px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.68);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.group-toolbar {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 2;
  opacity: 0;
  transform: translateY(-6px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.history-generated-group:hover .group-toolbar {
  opacity: 1;
  transform: translateY(0);
}

.history-generated-group-grid {
  display: grid;
  gap: 6px;
}

.history-generated-group-grid.is-count-2,
.history-generated-group-grid.is-count-4 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.history-generated-group-grid.is-count-3 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.history-generated-group-cell {
  position: relative;
  overflow: hidden;
  background: var(--bg-card);
  cursor: pointer;
}

.history-generated-group-grid.is-count-3 .history-generated-group-cell:first-child {
  grid-column: 1 / -1;
}

.history-image-card {
  position: relative;
  width: 100%;
  padding: 0;
  border: 0;
  border-radius: 18px;
  background: var(--bg-muted);
  overflow: hidden;
  cursor: pointer;
}

.selection-check {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 3;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid rgba(255, 255, 255, 0.88);
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.36);
  color: #fff;
  cursor: pointer;
  backdrop-filter: blur(6px);
}

.selection-check.is-selected {
  border-color: var(--color-primary);
  background: var(--color-primary);
}

.history-image-wrap {
  position: relative;
  aspect-ratio: 1 / 1;
  background: var(--bg-muted);
}

.history-image,
.history-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.history-image-card .history-image {
  display: block;
}

.history-image-toolbar {
  position: absolute;
  right: 12px;
  bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0;
  transform: translateY(8px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.history-image-card::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 55%, rgba(15, 23, 42, 0.56) 100%);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.history-image-card:hover::after,
.history-image-card:hover .history-image-toolbar {
  opacity: 1;
  transform: translateY(0);
}

.overlay-icon-btn,
.overlay-chip-btn {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 34px;
  border: 0;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.94);
  color: #111827;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.18);
  cursor: pointer;
}

.overlay-icon-btn {
  width: 34px;
}

.overlay-chip-btn {
  padding: 0 12px;
  font-size: 12px;
  font-weight: 600;
}

.overlay-icon-btn .el-icon.is-favorite {
  color: #f59e0b;
}

.history-image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--text-secondary);
}

.history-image-placeholder.is-processing {
  color: #e6a23c;
}

.history-image-placeholder.is-pending {
  color: #909399;
}

.history-image-placeholder.is-audio {
  color: var(--color-primary);
}

.loading-icon {
  animation: spin 1s linear infinite;
}

.history-item-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
}

.history-item-top,
.history-meta,
.history-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.history-time,
.history-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.history-prompt {
  display: -webkit-box;
  overflow: hidden;
  color: var(--text-primary);
  line-height: 1.6;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
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

.image-preview-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}

.preview-dialog-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 24px;
  align-items: start;
}

.preview-dialog-image {
  position: relative;
  border-radius: 20px;
  overflow: hidden;
  background: var(--bg-muted);
}

.preview-dialog-image img {
  display: block;
  width: 100%;
  max-height: 78vh;
  object-fit: contain;
  margin: 0 auto;
}

.preview-group-nav {
  position: absolute;
  right: 16px;
  bottom: 16px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.72);
  color: #fff;
}

.preview-nav-btn {
  border: 0;
  background: rgba(255, 255, 255, 0.18);
  color: #fff;
  border-radius: 999px;
  padding: 6px 10px;
  cursor: pointer;
}

.preview-nav-count {
  font-size: 12px;
  font-weight: 600;
}

.preview-dialog-sidebar {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 8px 0;
}

.preview-meta-line,
.preview-submeta {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  font-size: 13px;
  color: var(--text-secondary);
}

.preview-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-section-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.preview-prompt-text {
  white-space: pre-wrap;
  line-height: 1.85;
  color: var(--text-primary);
}

.preview-action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.reference-library-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 14px;
  min-height: 180px;
}

.reference-library-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--bg-card);
  cursor: pointer;
}

.reference-library-item.is-selected {
  border-color: var(--color-primary);
}

.reference-library-item img {
  width: 100%;
  aspect-ratio: 1 / 1;
  object-fit: cover;
  border-radius: 10px;
}

.reference-library-item span {
  font-size: 12px;
  color: var(--text-secondary);
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1280px) {
  .tool-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .ratio-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .history-header {
    flex-direction: column;
  }

  .history-masonry {
    column-width: auto;
    column-count: 1;
  }

  .preview-dialog-layout {
    grid-template-columns: 1fr;
  }
}
</style>
