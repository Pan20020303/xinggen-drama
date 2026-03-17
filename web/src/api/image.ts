import type {
  GenerateImageRequest,
  ImageGeneration,
  ImageGenerationListParams
} from '../types/image'
import type { EntityId } from '../types/drama'
import request from '../utils/request'

export const imageAPI = {
  generateImage(data: GenerateImageRequest) {
    return request.post<ImageGeneration>('/images', data)
  },

  generateForScene(sceneId: EntityId) {
    return request.post<ImageGeneration[]>(`/images/scene/${sceneId}`)
  },

  batchGenerateForEpisode(episodeId: EntityId) {
    return request.post<ImageGeneration[]>(`/images/episode/${episodeId}/batch`)
  },

  getImage(id: EntityId) {
    return request.get<ImageGeneration>(`/images/${id}`)
  },

  listImages(params: ImageGenerationListParams) {
    return request.get<{
      items: ImageGeneration[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/images', { params })
  },

  deleteImage(id: EntityId) {
    return request.delete(`/images/${id}`)
  },

  // 上传图片并创建图片生成记录
  uploadImage(data: {
    storyboard_id: EntityId
    drama_id: EntityId
    frame_type: string
    image_url: string
    prompt?: string
  }) {
    return request.post<ImageGeneration>('/images/upload', data)
  }
}
