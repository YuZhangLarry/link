import { http } from '@/utils/request'
import type {
  EvaluationTask,
  EvaluationDetail,
  CreateEvaluationRequest,
  CreateDatasetRequest
} from '@/types'

export const evaluationApi = {
  /**
   * 创建测评任务
   */
  create(data: CreateEvaluationRequest) {
    return http.post<EvaluationDetail>('/evaluation', data)
  },

  /**
   * 获取测评结果
   */
  getResult(taskId: string) {
    return http.get<EvaluationDetail>(`/evaluation?task_id=${taskId}`)
  },

  /**
   * 列出测评任务
   */
  list(params?: { page?: number; page_size?: number }) {
    return http.get<{ tasks: EvaluationTask[]; total: number; page: number; page_size: number }>(
      '/evaluations',
      { params }
    )
  },

  /**
   * 获取单个测评任务
   */
  getById(id: string) {
    return http.get<EvaluationTask>(`/evaluations/${id}`)
  },

  /**
   * 删除测评任务
   */
  delete(id: string) {
    return http.delete(`/evaluations/${id}`)
  },

  /**
   * 创建数据集
   */
  createDataset(data: CreateDatasetRequest) {
    return http.post('/datasets', data)
  },

  /**
   * 列出数据集
   */
  listDatasets() {
    return http.get<string[]>('/datasets')
  }
}
