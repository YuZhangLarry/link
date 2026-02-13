import { http } from '@/utils/request'
import type {
  GraphData,
  GraphSearchRequest,
  AddNodeRequest,
  AddRelationRequest,
  UpdateNodeRequest,
  UpdateRelationRequest,
  NodeDetailResponse,
  RelationTypeOption
} from '@/types'

/**
 * 图谱相关API
 */
export const graphApi = {
  /**
   * 获取知识库图谱数据
   */
  getGraph(kbId: string) {
    return http.get<GraphData>(`/knowledge-bases/${kbId}/graph`)
  },

  /**
   * 搜索节点
   */
  searchNode(kbId: string, data: GraphSearchRequest) {
    return http.post<GraphData>(`/knowledge-bases/${kbId}/graph/search`, data)
  },

  /**
   * 获取节点详情
   */
  getNodeDetail(kbId: string, nodeId: string) {
    return http.get<NodeDetailResponse>(`/knowledge-bases/${kbId}/graph/nodes/${nodeId}`)
  },

  /**
   * 添加节点
   */
  addNode(kbId: string, data: AddNodeRequest) {
    return http.post<{ message: string; data: any }>(
      `/knowledge-bases/${kbId}/graph/nodes`,
      data
    )
  },

  /**
   * 添加关系
   */
  addRelation(kbId: string, data: AddRelationRequest) {
    return http.post<{ message: string; data: any }>(
      `/knowledge-bases/${kbId}/graph/relations`,
      data
    )
  },

  /**
   * 更新节点
   */
  updateNode(kbId: string, nodeId: string, data: UpdateNodeRequest) {
    return http.put<{ message: string; data: any }>(
      `/knowledge-bases/${kbId}/graph/nodes/${nodeId}`,
      data
    )
  },

  /**
   * 更新关系
   */
  updateRelation(kbId: string, relationId: string, data: UpdateRelationRequest) {
    return http.put<{ message: string; data: any }>(
      `/knowledge-bases/${kbId}/graph/relations/${relationId}`,
      data
    )
  },

  /**
   * 删除图谱
   */
  deleteGraph(kbId: string) {
    return http.delete<{ message: string }>(`/knowledge-bases/${kbId}/graph`)
  },

  /**
   * 删除节点
   */
  deleteNode(kbId: string, nodeId: string) {
    return http.delete<{ message: string }>(
      `/knowledge-bases/${kbId}/graph/nodes/${nodeId}`
    )
  },

  /**
   * 删除关系
   */
  deleteRelation(kbId: string, relationId: string) {
    return http.delete<{ message: string }>(
      `/knowledge-bases/${kbId}/graph/relations/${relationId}`
    )
  },

  /**
   * 获取关系类型选项
   */
  getRelationTypes(kbId: string) {
    return http.get<{ message: string; data: RelationTypeOption[] }>(
      `/knowledge-bases/${kbId}/graph/relation-types`
    )
  }
}

export default graphApi
