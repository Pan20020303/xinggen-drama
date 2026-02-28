import request from '../utils/request'

export interface AsyncTask {
    id: string
    type: string
    status: 'pending' | 'processing' | 'completed' | 'failed'
    progress: number
    message: string
    result?: any
    error?: string
    resource_id?: string
    updated_at?: string
    completed_at?: string
    created_at: string
}

export const taskAPI = {
    getStatus(taskId: string) {
        return request.get<AsyncTask>(`/tasks/${taskId}`)
    },
    listByResource(resourceId: string) {
        return request.get<AsyncTask[]>(`/tasks`, {
            params: {
                resource_id: resourceId
            }
        })
    }
}
