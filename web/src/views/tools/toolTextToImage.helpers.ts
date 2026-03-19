import type { Asset, AssetType } from '@/types/asset'
import type { ImageGeneration } from '@/types/image'

export type DisplayItem =
  | { key: string; kind: 'asset'; asset: Asset }
  | { key: string; kind: 'generated'; image: ImageGeneration }
  | { key: string; kind: 'generated-group'; images: ImageGeneration[] }

export const GROUP_WINDOW_MS = 20 * 1000

export interface CollectBulkDeletableImagesOptions {
  activeMediaTab: AssetType
  onlyFavorites: boolean
  onlyProcessing: boolean
  processingImages: ImageGeneration[]
  completedToolboxImages: ImageGeneration[]
}

export interface HistoryEmptyStateOptions {
  historyLoading: boolean
  onlyProcessing: boolean
  displayItemCount: number
  processingItemCount: number
}

export const getGeneratedGroupKey = (image: ImageGeneration) => {
  const references = Array.isArray(image.reference_images) ? image.reference_images.filter(Boolean).join('|') : ''
  return [
    image.prompt || '',
    image.model || '',
    image.size || '',
    image.width || '',
    image.height || '',
    references
  ].join('::')
}

export const groupGeneratedImages = (images: ImageGeneration[]): DisplayItem[] => {
  const grouped: DisplayItem[] = []
  let currentGroup: ImageGeneration[] = []
  let currentKey = ''
  let currentStartAt = 0

  const flush = () => {
    if (currentGroup.length === 0) return

    if (currentGroup.length === 1) {
      grouped.push({
        key: `generated-${currentGroup[0].id}`,
        kind: 'generated',
        image: currentGroup[0]
      })
    } else {
      grouped.push({
        key: `generated-group-${currentGroup.map((item) => item.id).join('-')}`,
        kind: 'generated-group',
        images: [...currentGroup]
      })
    }

    currentGroup = []
    currentKey = ''
    currentStartAt = 0
  }

  for (const image of images) {
    const imageKey = getGeneratedGroupKey(image)
    const createdAt = new Date(image.created_at).getTime()
    const canJoin =
      currentGroup.length > 0 &&
      imageKey === currentKey &&
      Number.isFinite(createdAt) &&
      Math.abs(createdAt - currentStartAt) <= GROUP_WINDOW_MS &&
      currentGroup.length < 4

    if (!canJoin) {
      flush()
      currentGroup = [image]
      currentKey = imageKey
      currentStartAt = createdAt
      continue
    }

    currentGroup.push(image)
  }

  flush()
  return grouped
}

export const collectBulkDeletableImages = ({
  activeMediaTab,
  onlyFavorites,
  onlyProcessing,
  processingImages,
  completedToolboxImages
}: CollectBulkDeletableImagesOptions) => {
  if (activeMediaTab !== 'image') {
    return []
  }

  if (onlyProcessing) {
    return [...processingImages]
  }

  if (onlyFavorites) {
    return []
  }

  return [...completedToolboxImages]
}

export const collectSelectedBulkDeleteImages = (
  images: ImageGeneration[],
  selectedIds: Array<string | number>
) => {
  const selectedIdSet = new Set(selectedIds.map((id) => String(id)))
  return images.filter((image) => selectedIdSet.has(String(image.id)))
}

export const shouldShowHistoryEmptyState = ({
  historyLoading,
  onlyProcessing,
  displayItemCount,
  processingItemCount
}: HistoryEmptyStateOptions) => {
  if (historyLoading) {
    return false
  }

  if (onlyProcessing) {
    return processingItemCount === 0
  }

  return displayItemCount === 0
}
