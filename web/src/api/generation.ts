import type {
  GenerateCharactersRequest,
  GenerateShotsRequest,
  GenerateShotsResult,
  ParseScriptRequest,
  ParseScriptResult,
  ParsedEpisode,
  ParsedScene,
} from '../types/generation'
import request from '../utils/request'

function splitScenes(scriptContent: string): ParsedScene[] {
  return scriptContent
    .split(/\n{2,}/)
    .map((chunk) => chunk.trim())
    .filter(Boolean)
    .map((content, index) => ({
      storyboard_number: index + 1,
      title: `场景 ${index + 1}`,
      dialogue: content,
      content,
      duration: 5,
    }))
}

function buildParsedEpisodes(data: ParseScriptRequest): ParsedEpisode[] {
  const scenes = splitScenes(data.script_content)

  return [
    {
      episode_number: 1,
      title: '导入剧本',
      description: '由前端兼容解析生成，请在保存后再人工校验。',
      script_content: data.script_content,
      duration: scenes.reduce((sum, scene) => sum + (scene.duration || 0), 0),
      scenes,
    },
  ]
}

export const generationAPI = {
  generateCharacters(data: GenerateCharactersRequest) {
    return request.post<{ task_id: string; status: string; message: string }>('/generation/characters', data)
  },

  generateStoryboard(episodeId: string, model?: string) {
    return request.post<{ task_id: string; status: string; message: string }>(`/episodes/${episodeId}/storyboards`, { model })
  },

  getTaskStatus(taskId: string) {
    return request.get<{
      id: string
      type: string
      status: string
      progress: number
      message?: string
      error?: string
      result?: string
      created_at: string
      updated_at: string
      completed_at?: string
    }>(`/tasks/${taskId}`)
  },

  async parseScript(data: ParseScriptRequest): Promise<ParseScriptResult> {
    return {
      episodes: buildParsedEpisodes(data),
      characters: [],
      summary: data.script_content.slice(0, 120),
    }
  },

  async generateShots(data: GenerateShotsRequest): Promise<GenerateShotsResult> {
    return {
      shots: splitScenes(data.script_content),
    }
  },
}
