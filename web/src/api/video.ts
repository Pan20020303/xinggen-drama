import type {
  GenerateVideoRequest,
  VideoGeneration,
  VideoGenerationListParams
} from '../types/video'
import type { EntityId } from '../types/drama'
import request from '../utils/request'

export const videoAPI = {
  generateVideo(data: GenerateVideoRequest) {
    return request.post<VideoGeneration>('/videos', data)
  },

  generateFromImage(imageGenId: EntityId) {
    return request.post<VideoGeneration>(`/videos/image/${imageGenId}`)
  },

  batchGenerateForEpisode(episodeId: EntityId) {
    return request.post<VideoGeneration[]>(`/videos/episode/${episodeId}/batch`)
  },

  getVideoGeneration(id: EntityId) {
    return request.get<VideoGeneration>(`/videos/${id}`)
  },
  
  getVideo(id: EntityId) {
    return request.get<VideoGeneration>(`/videos/${id}`)
  },

  listVideos(params: VideoGenerationListParams) {
    return request.get<{
      items: VideoGeneration[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/videos', { params })
  },

  deleteVideo(id: EntityId) {
    return request.delete(`/videos/${id}`)
  }
}
