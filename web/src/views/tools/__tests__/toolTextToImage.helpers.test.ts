import { describe, expect, it } from 'vitest'
import type { ImageGeneration } from '@/types/image'
import {
  collectBulkDeletableImages,
  collectSelectedBulkDeleteImages,
  shouldShowHistoryEmptyState,
  groupGeneratedImages
} from '../toolTextToImage.helpers'

const createImage = (overrides: Partial<ImageGeneration> = {}): ImageGeneration => ({
  id: overrides.id ?? '1',
  provider: overrides.provider ?? 'volcengine',
  prompt: overrides.prompt ?? 'sunset over the lake',
  model: overrides.model ?? 'seedream',
  status: overrides.status ?? 'completed',
  created_at: overrides.created_at ?? '2026-03-19T10:00:00.000Z',
  updated_at: overrides.updated_at ?? '2026-03-19T10:00:00.000Z',
  width: overrides.width ?? 2560,
  height: overrides.height ?? 1440,
  size: overrides.size ?? '2560x1440',
  reference_images: overrides.reference_images ?? [],
  ...overrides
})

describe('groupGeneratedImages', () => {
  it('groups images created in the same request window and keeps standalone items separate', () => {
    const images = [
      createImage({ id: '11', created_at: '2026-03-19T10:00:00.000Z' }),
      createImage({ id: '12', created_at: '2026-03-19T10:00:08.000Z' }),
      createImage({ id: '13', created_at: '2026-03-19T10:00:28.500Z' }),
      createImage({ id: '14', prompt: 'another prompt', created_at: '2026-03-19T10:01:00.000Z' })
    ]

    const grouped = groupGeneratedImages(images)

    expect(grouped).toHaveLength(3)
    expect(grouped[0]).toMatchObject({
      kind: 'generated-group',
      images: [{ id: '11' }, { id: '12' }]
    })
    expect(grouped[1]).toMatchObject({
      kind: 'generated',
      image: { id: '13' }
    })
    expect(grouped[2]).toMatchObject({
      kind: 'generated',
      image: { id: '14' }
    })
  })
})

describe('collectBulkDeletableImages', () => {
  it('returns processing items when the panel is filtered to in-progress tasks', () => {
    const processingImages = [
      createImage({ id: '21', status: 'processing' }),
      createImage({ id: '22', status: 'pending' })
    ]
    const completedImages = [createImage({ id: '31' })]

    const targets = collectBulkDeletableImages({
      activeMediaTab: 'image',
      onlyFavorites: false,
      onlyProcessing: true,
      processingImages,
      completedToolboxImages: completedImages
    })

    expect(targets.map((item) => item.id)).toEqual(['21', '22'])
  })

  it('returns completed toolbox images only for the image tab', () => {
    const completedImages = [
      createImage({ id: '41' }),
      createImage({ id: '42' })
    ]

    expect(collectBulkDeletableImages({
      activeMediaTab: 'image',
      onlyFavorites: false,
      onlyProcessing: false,
      processingImages: [],
      completedToolboxImages: completedImages
    }).map((item) => item.id)).toEqual(['41', '42'])

    expect(collectBulkDeletableImages({
      activeMediaTab: 'video',
      onlyFavorites: false,
      onlyProcessing: false,
      processingImages: [],
      completedToolboxImages: completedImages
    })).toEqual([])

    expect(collectBulkDeletableImages({
      activeMediaTab: 'image',
      onlyFavorites: true,
      onlyProcessing: false,
      processingImages: [],
      completedToolboxImages: completedImages
    })).toEqual([])
  })
})

describe('collectSelectedBulkDeleteImages', () => {
  it('returns only the checked images and ignores stale ids', () => {
    const images = [
      createImage({ id: '51' }),
      createImage({ id: '52' }),
      createImage({ id: '53' })
    ]

    const selected = collectSelectedBulkDeleteImages(images, ['52', '999', '51'])

    expect(selected.map((item) => item.id)).toEqual(['51', '52'])
  })
})

describe('shouldShowHistoryEmptyState', () => {
  it('does not show the empty placeholder when processing items exist', () => {
    expect(shouldShowHistoryEmptyState({
      historyLoading: false,
      onlyProcessing: true,
      displayItemCount: 0,
      processingItemCount: 2
    })).toBe(false)
  })
})
