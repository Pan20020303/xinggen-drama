import request from '../utils/request'

export const uploadAPI = {
  uploadImage(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<{
      url: string
      local_path?: string
      filename: string
      size: number
    }>('/upload/image', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}
