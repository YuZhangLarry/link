import { http } from '@/utils/request'
import type {
  Session,
  CreateSessionRequest,
  UpdateSessionRequest,
  SessionDetail,
  SessionListResponse
} from '@/types'

export const sessionApi = {
  /**
   * 创建会话
   */
  create(data: CreateSessionRequest) {
    return http.post<Session>('/sessions', data)
  },

  /**
   * 获取会话详情
   */
  getById(id: string) {
    return http.get<Session>(`/sessions/${id}`)
  },

  /**
   * 获取会话完整详情（包含消息）
   */
  getDetail(id: string) {
    return http.get<SessionDetail>(`/sessions/${id}/detail`)
  },

  /**
   * 获取会话列表
   */
  list(params?: { page?: number; size?: number; status?: number }) {
    return http.get<SessionListResponse>('/sessions', { params })
  },

  /**
   * 更新会话
   */
  update(id: string, data: UpdateSessionRequest) {
    return http.put<Session>(`/sessions/${id}`, data)
  },

  /**
   * 删除会话
   */
  delete(id: string) {
    return http.delete(`/sessions/${id}`)
  },

  /**
   * 归档会话
   */
  archive(id: string) {
    return http.post(`/sessions/${id}/archive`)
  },

  /**
   * 激活会话
   */
  activate(id: string) {
    return http.post(`/sessions/${id}/activate`)
  }
}
