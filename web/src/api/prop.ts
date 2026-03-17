import request from '../utils/request'
import type { Prop, CreatePropRequest, UpdatePropRequest } from '../types/prop'
import type { EntityId } from '../types/drama'

export const propAPI = {
    list(dramaId: string | number) {
        return request.get<Prop[]>('/dramas/' + dramaId + '/props')
    },
    create(data: CreatePropRequest) {
        return request.post<Prop>('/props', data)
    },
    update(id: EntityId, data: UpdatePropRequest) {
        return request.put<void>('/props/' + id, data)
    },
    delete(id: EntityId) {
        return request.delete<void>('/props/' + id)
    },
    extractFromScript(episodeId: EntityId) {
        return request.post<{ task_id: string }>(`/episodes/${episodeId}/props/extract`)
    },
    generateImage(id: EntityId) {
        return request.post<{ task_id: string }>(`/props/${id}/generate`)
    },
    associateWithStoryboard(storyboardId: EntityId, propIds: Array<string | number>) {
        return request.post<void>(`/storyboards/${storyboardId}/props`, { prop_ids: propIds })
    }
}
