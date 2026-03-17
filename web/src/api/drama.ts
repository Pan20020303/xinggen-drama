import type {
  CreateDramaRequest,
  Drama,
  DramaListQuery,
  DramaStats,
  UpdateDramaRequest
} from '../types/drama'
import type { EntityId } from '../types/drama'
import request from '../utils/request'

export const dramaAPI = {
  list(params?: DramaListQuery) {
    return request.get<{
      items: Drama[]
      pagination: {
        page: number
        page_size: number
        total: number
        total_pages: number
      }
    }>('/dramas', { params })
  },

  create(data: CreateDramaRequest) {
    return request.post<Drama>('/dramas', data)
  },

  get(id: EntityId) {
    return request.get<Drama>(`/dramas/${id}`)
  },

  update(id: EntityId, data: UpdateDramaRequest) {
    return request.put<Drama>(`/dramas/${id}`, data)
  },

  delete(id: EntityId) {
    return request.delete(`/dramas/${id}`)
  },

  getStats() {
    return request.get<DramaStats>('/dramas/stats')
  },

  saveOutline(id: string, data: { title: string; summary: string; genre?: string; tags?: string[] }) {
    return request.put(`/dramas/${id}/outline`, data)
  },

  getCharacters(dramaId: EntityId) {
    return request.get(`/dramas/${dramaId}/characters`)
  },

  saveCharacters(id: string, data: any[], episodeId?: string) {
    return request.put(`/dramas/${id}/characters`, {
      characters: data,
      episode_id: episodeId ? parseInt(episodeId) : undefined
    })
  },

  updateCharacter(id: number, data: any) {
    return request.put(`/characters/${id}`, data)
  },

  saveEpisodes(id: string, data: any[]) {
    return request.put(`/dramas/${id}/episodes`, { episodes: data })
  },

  saveProgress(id: string, data: { current_step: string; step_data?: any }) {
    return request.put(`/dramas/${id}/progress`, data)
  },

  generateStoryboard(episodeId: string) {
    return request.post(`/episodes/${episodeId}/storyboards`)
  },

  polishEpisodeScript(episodeId: string, data: { content?: string; model?: string; skill_name?: string }) {
    return request.post<{ content: string; skill_name: string }>(`/episodes/${episodeId}/polish-script`, data)
  },

  polishScriptText(data: { content: string; model?: string; skill_name?: string }) {
    return request.post<{ content: string; skill_name: string }>('/generation/script/polish', data)
  },

  getBackgrounds(episodeId: string) {
    return request.get(`/images/episode/${episodeId}/backgrounds`)
  },

  extractBackgrounds(episodeId: string, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>(`/images/episode/${episodeId}/backgrounds/extract`, { model })
  },

  batchGenerateBackgrounds(episodeId: string) {
    return request.post(`/images/episode/${episodeId}/batch`)
  },

  generateSingleBackground(backgroundId: EntityId, dramaId: EntityId, prompt: string) {
    return request.post('/images', {
      background_id: backgroundId,
      drama_id: dramaId,
      prompt: prompt
    })
  },

  getStoryboards(episodeId: EntityId) {
    return request.get(`/episodes/${episodeId}/storyboards`)
  },

  updateStoryboard(storyboardId: EntityId, data: any) {
    return request.put(`/storyboards/${storyboardId}`, data)
  },

  optimizeVideoPrompt(storyboardId: EntityId, data: { prompt?: string; model?: string }) {
    return request.post<{ prompt: string }>(`/storyboards/${storyboardId}/optimize-video-prompt`, data)
  },

  updateScene(sceneId: EntityId, data: {
    background_id?: EntityId;
    characters?: Array<string | number>;
    location?: string;
    time?: string;
    prompt?: string;
    action?: string;
    dialogue?: string;
    description?: string;
    duration?: number;
    image_url?: string;
    local_path?: string;
  }) {
    return request.put(`/scenes/${sceneId}`, data)
  },

  createScene(data: {
    drama_id: EntityId;
    episode_id?: EntityId;
    location: string;
    time?: string;
    prompt?: string;
    description?: string;
    image_url?: string;
    local_path?: string;
  }) {
    return request.post('/scenes', data)
  },

  generateSceneImage(data: { scene_id: EntityId; prompt?: string; model?: string; image_local_path?: string; size?: string; quality?: string; style?: string; steps?: number; cfg_scale?: number; seed?: number }) {
    return request.post<{ image_generation: { id: EntityId } }>('/scenes/generate-image', data)
  },

  updateScenePrompt(sceneId: EntityId, prompt: string) {
    return request.put(`/scenes/${sceneId}/prompt`, { prompt })
  },

  deleteScene(sceneId: EntityId) {
    return request.delete(`/scenes/${sceneId}`)
  },

  // 完成集数制作（触发视频合成）
  finalizeEpisode(episodeId: string, timelineData?: any) {
    return request.post(`/episodes/${episodeId}/finalize`, timelineData || {})
  },

  createStoryboard(data: {
    episode_id: EntityId;
    storyboard_number: number;
    title?: string;
    description?: string;
    action?: string;
    dialogue?: string;
    scene_id?: EntityId;
    duration: number;
  }) {
    return request.post('/storyboards', data)
  },

  deleteStoryboard(storyboardId: EntityId) {
    return request.delete(`/storyboards/${storyboardId}`)
  }
}
