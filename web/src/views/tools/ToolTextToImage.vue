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

              <el-button text @click="loadRightPanel">刷新</el-button>
            </div>
          </template>

          <div v-loading="historyLoading" class="history-panel">
            <el-empty
              v-if="!historyLoading && displayItems.length === 0"
              :description="onlyProcessing ? '还没有进行中的任务' : '还没有素材记录'"
            />

            <div v-else-if="onlyProcessing" class="history-grid">
              <article v-for="image in processingImages" :key="image.id" class="history-item">
                <div class="history-image-wrap">
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

            <div v-else class="history-grid">
              <article v-for="item in displayItems" :key="item.key" class="history-item">
                <div class="history-image-wrap">
                  <el-image
                    v-if="item.kind === 'asset' && item.asset.type === 'image'"
                    :src="item.asset.url"
                    fit="cover"
                    class="history-image"
                    :preview-src-list="[item.asset.url]"
                    preview-teleported
                  />
                  <video
                    v-else-if="item.kind === 'asset' && item.asset.type === 'video'"
                    :src="item.asset.url"
                    class="history-video"
                    controls
                    preload="metadata"
                  />
                  <el-image
                    v-else-if="item.kind === 'generated' && item.image.image_url"
                    :src="item.image.image_url"
                    fit="cover"
                    class="history-image"
                    :preview-src-list="[item.image.image_url]"
                    preview-teleported
                  />
                  <div v-else class="history-image-placeholder" :class="item.kind === 'generated' ? 'is-completed' : 'is-audio'">
                    <el-icon><Headset /></el-icon>
                    <span>{{ item.kind === 'generated' ? '待入库图片' : '音频素材' }}</span>
                  </div>
                </div>
                <div class="history-item-body">
                  <div class="history-item-top">
                    <el-tag
                      size="small"
                      :type="item.kind === 'asset' ? assetTagMap[item.asset.type] : 'info'"
                    >
                      {{ item.kind === 'asset' ? assetTypeLabelMap[item.asset.type] : '工具图片' }}
                    </el-tag>
                    <el-button v-if="item.kind === 'asset'" text class="favorite-btn" @click="toggleFavorite(item.asset)">
                      <el-icon :class="{ 'is-favorite': item.asset.is_favorite }">
                        <StarFilled v-if="item.asset.is_favorite" />
                        <Star v-else />
                      </el-icon>
                    </el-button>
                  </div>
                  <div class="history-prompt">{{ item.kind === 'asset' ? item.asset.name : item.image.prompt }}</div>
                  <div class="history-meta">
                    <span>
                      {{ item.kind === 'asset' ? (item.asset.category || '未分类') : (item.image.model || '默认模型') }}
                    </span>
                    <span>{{ formatTime(item.kind === 'asset' ? item.asset.created_at : item.image.created_at) }}</span>
                  </div>
                  <div class="history-actions">
                    <el-button
                      v-if="item.kind === 'asset'"
                      size="small"
                      @click="openAsset(item.asset.url)"
                    >
                      查看
                    </el-button>
                    <el-button
                      v-else
                      size="small"
                      type="primary"
                      @click="importToAssetLibrary(item.image)"
                    >
                      加入素材库
                    </el-button>
                  </div>
                </div>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  ArrowLeft,
  Headset,
  Loading,
  Picture,
  PictureFilled,
  Star,
  StarFilled
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { AppHeader } from '@/components/common'
import { assetAPI } from '@/api/asset'
import { imageAPI } from '@/api/image'
import { uploadAPI } from '@/api/upload'
import { useAuthStore } from '@/stores/auth'
import { usePricingStore } from '@/stores/pricing'
import type { Asset, AssetType } from '@/types/asset'
import type { ImageGeneration, ImageStatus } from '@/types/image'

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

const router = useRouter()
const authStore = useAuthStore()
const pricingStore = usePricingStore()

const submitting = ref(false)
const historyLoading = ref(false)
const uploadingReference = ref(false)
const showReferenceDialog = ref(false)
const referenceDialogLoading = ref(false)

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
      ...completedToolboxImages.value.map((image) => ({
        key: `generated-${image.id}`,
        kind: 'generated' as const,
        image
      })),
      ...items
    ]
  }

  return items
})

const applyRatio = (ratioKey: RatioKey) => {
  const matched = ratioOptions.find((item) => item.key === ratioKey)
  if (!matched) return
  form.ratio = matched.key
  form.width = matched.width
  form.height = matched.height
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
    url: item.url,
    name: item.name,
    localPath: item.local_path
  }
  showReferenceDialog.value = false
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

const loadRightPanel = async () => {
  historyLoading.value = true
  try {
    if (onlyProcessing.value) {
      await loadProcessingItems()
      materialItems.value = []
      return
    }

    processingImages.value = []
    await loadMaterialAssets()
  } catch (error: any) {
    ElMessage.error(error?.message || '加载记录失败')
  } finally {
    historyLoading.value = false
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

  submitting.value = true
  try {
    for (let i = 0; i < form.count; i += 1) {
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
    onlyProcessing.value = true
    await loadRightPanel()
    ElMessage.success(`已提交 ${form.count} 张图片生成任务`)
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

const openAsset = (url: string) => {
  window.open(url, '_blank')
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
      loadRightPanel()
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
}

.history-item {
  overflow: hidden;
  border: 1px solid var(--border-primary);
  border-radius: 18px;
  background: var(--bg-card);
}

.history-image-wrap {
  aspect-ratio: 1 / 1;
  background: var(--bg-muted);
}

.history-image,
.history-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
}
</style>
