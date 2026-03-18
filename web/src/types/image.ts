import type { EntityId } from './drama'

export interface ImageGeneration {
  id: EntityId
  storyboard_id?: EntityId
  scene_id?: EntityId
  drama_id?: EntityId
  character_id?: EntityId
  image_type?: string
  frame_type?: string
  provider: string
  prompt: string
  negative_prompt?: string
  model: string
  size?: string
  quality?: string
  style?: string
  steps?: number
  cfg_scale?: number
  seed?: number
  image_url?: string
  image_generation?: { id: EntityId }
  local_path?: string
  status: ImageStatus
  task_id?: string
  error_msg?: string
  width?: number
  height?: number
  reference_images?: string[]
  created_at: string
  updated_at: string
  completed_at?: string
}

export type ImageStatus = 'pending' | 'processing' | 'completed' | 'failed'

export type ImageProvider = 'openai' | 'dalle' | 'midjourney' | 'stable_diffusion' | 'sd'

export interface GenerateImageRequest {
  scene_id?: EntityId
  storyboard_id?: EntityId
  drama_id?: EntityId
  image_type?: string
  frame_type?: string
  prompt: string
  negative_prompt?: string
  reference_images?: string[]
  provider?: string
  model?: string
  size?: string
  quality?: string
  style?: string
  steps?: number
  cfg_scale?: number
  seed?: number
  width?: number
  height?: number
}

export interface ImageCreatorOptions {
  mode: 'text' | 'image'
  model?: string
  size?: string
  quality?: string
  style?: string
  steps?: number
  cfg_scale?: number
  seed?: number
  imageLocalPath?: string
  prompt?: string
}

export interface ImageGenerationListParams {
  drama_id?: EntityId
  scene_id?: EntityId
  character_id?: EntityId
  storyboard_id?: EntityId
  image_type?: string
  frame_type?: string
  status?: ImageStatus
  page?: number
  page_size?: number
}
