<template>
  <div class="graph-view-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="search-box">
        <el-input
          v-model="searchText"
          placeholder="搜索实体名称..."
          clearable
          @keyup.enter="handleSearch"
        >
          <template #append>
            <el-button :icon="Search" @click="handleSearch">搜索</el-button>
          </template>
        </el-input>
      </div>

      <el-button-group class="action-buttons">
        <el-button :icon="Plus" @click="showAddNodeDialog">添加节点</el-button>
        <el-button :icon="Connection" @click="showAddRelationDialog">添加关系</el-button>
      </el-button-group>

      <el-button :icon="Download" @click="handleExport">导出</el-button>
      <el-button :icon="FullScreen" @click="toggleFullscreen">全屏</el-button>
    </div>

    <!-- 图谱可视化区域 -->
    <div ref="graphContainer" class="graph-container"></div>

    <!-- 底部状态栏 -->
    <div class="status-bar">
      <span>节点: {{ nodeCount }}</span>
      <el-divider direction="vertical" />
      <span>关系: {{ edgeCount }}</span>
      <el-divider direction="vertical" />
      <span>实体类型: {{ entityTypeCount }}</span>
    </div>

    <!-- 节点详情抽屉 -->
    <el-drawer v-model="detailDrawerVisible" title="节点详情" size="500px">
      <div v-if="selectedNode">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="名称">{{ selectedNode.label }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selectedNodeType }}</el-descriptions-item>
          <el-descriptions-item label="属性">
            <template v-if="selectedNodeAttributes.length > 0">
              <el-tag v-for="attr in selectedNodeAttributes" :key="attr" size="small">
                {{ attr }}
              </el-tag>
            </template>
            <span v-else>-</span>
          </el-descriptions-item>
          <el-descriptions-item label="关联分块数">
            {{ selectedNodeChunks.length }}
          </el-descriptions-item>
        </el-descriptions>

        <div style="margin-top: 16px; display: flex; gap: 8px;">
          <el-button type="primary" @click="showEditNodeDialog" style="flex: 1;">
            编辑节点
          </el-button>
          <el-button type="danger" @click="handleDeleteNode" :loading="deleting" style="flex: 1;">
            删除节点
          </el-button>
        </div>

        <el-divider content-position="left">关联关系</el-divider>

        <div class="relations-list">
          <div
            v-for="rel in nodeRelations"
            :key="rel.id"
            class="relation-item"
          >
            <div class="relation-header">
              <el-tag size="small">{{ rel.label || rel.type || '-' }}</el-tag>
              <el-tag size="small" type="info">强度: {{ rel.strength || 0 }}</el-tag>
            </div>
            <div class="relation-content">
              <span class="relation-source">{{ rel.source || '-' }}</span>
              <el-icon class="relation-arrow"><Right /></el-icon>
              <span class="relation-target">{{ rel.target || '-' }}</span>
            </div>
            <div class="relation-description">{{ rel.description || '-' }}</div>
            <div style="display: flex; gap: 8px;">
              <el-button link type="primary" size="small" @click="showEditRelationDialog(rel)">
                编辑关系
              </el-button>
              <el-button link type="danger" size="small" @click="handleDeleteRelation(rel)">
                删除关系
              </el-button>
            </div>
          </div>

          <el-empty v-if="nodeRelations.length === 0" description="暂无关联关系" size="small" />
        </div>
      </div>
    </el-drawer>

    <!-- 添加节点对话框 -->
    <el-dialog v-model="addNodeDialogVisible" title="添加节点" width="500px">
      <el-form :model="addNodeForm" label-width="100px">
        <el-form-item label="节点名称" required>
          <el-input v-model="addNodeForm.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="实体类型">
          <el-select v-model="addNodeForm.entity_type" placeholder="选择类型">
            <el-option label="人物" value="person" />
            <el-option label="组织" value="organization" />
            <el-option label="地点" value="location" />
            <el-option label="概念" value="concept" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="属性">
          <el-input
            v-model="addNodeForm.attributesStr"
            type="textarea"
            :rows="3"
            placeholder="多个属性用逗号分隔"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addNodeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddNode" :loading="adding">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 添加关系对话框 -->
    <el-dialog v-model="addRelationDialogVisible" title="添加关系" width="500px">
      <el-form :model="addRelationForm" label-width="100px">
        <el-form-item label="源节点" required>
          <el-select
            v-model="addRelationForm.source"
            placeholder="选择源节点"
            filterable
          >
            <el-option
              v-for="node in graphData.nodes"
              :key="node.id"
              :label="node.label"
              :value="node.label"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标节点" required>
          <el-select
            v-model="addRelationForm.target"
            placeholder="选择目标节点"
            filterable
          >
            <el-option
              v-for="node in graphData.nodes"
              :key="node.id"
              :label="node.label"
              :value="node.label"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="关系类型" required>
          <el-select v-model="addRelationForm.type" placeholder="选择关系类型">
            <el-option
              v-for="opt in relationTypeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="强度">
          <el-slider v-model="addRelationForm.strength" :min="1" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addRelationDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddRelation" :loading="adding">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑节点对话框 -->
    <el-dialog v-model="editNodeDialogVisible" title="编辑节点" width="500px">
      <el-form :model="editNodeForm" label-width="100px">
        <el-form-item label="节点名称" required>
          <el-input v-model="editNodeForm.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="实体类型">
          <el-select v-model="editNodeForm.entity_type" placeholder="选择类型">
            <el-option label="人物" value="person" />
            <el-option label="组织" value="organization" />
            <el-option label="地点" value="location" />
            <el-option label="概念" value="concept" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="属性">
          <el-input
            v-model="editNodeForm.attributesStr"
            type="textarea"
            :rows="3"
            placeholder="多个属性用逗号分隔"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editNodeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditNode" :loading="editing">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 编辑关系对话框 -->
    <el-dialog v-model="editRelationDialogVisible" title="编辑关系" width="500px">
      <el-form :model="editRelationForm" label-width="100px">
        <el-form-item label="关系类型" required>
          <el-select v-model="editRelationForm.type" placeholder="选择关系类型">
            <el-option
              v-for="opt in relationTypeOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="强度">
          <el-slider v-model="editRelationForm.strength" :min="1" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editRelationDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditRelation" :loading="editing">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Connection, Download, FullScreen, Right } from '@element-plus/icons-vue'
import { Network } from 'vis-network'
import { DataSet } from 'vis-data'
import 'vis-network/styles/vis-network.css'
import { graphApi } from '@/api/graph'
import type { GraphData, AddNodeRequest, AddRelationRequest, UpdateNodeRequest, UpdateRelationRequest, RelationTypeOption } from '@/types'

// 关系列表类型（适配后端返回格式）
interface RelationDisplay {
  id: string
  label?: string
  type?: string
  description?: string
  strength?: number
  from?: string
  to?: string
  source?: string
  target?: string
}

// vis-network 数据类型
interface VisNode {
  id: string
  label: string
  color?: string
  size?: number
  attributes?: string[]
  chunks?: string[]
  entity_type?: string
}

interface VisEdge {
  id: string
  from: string
  to: string
  label?: string
  color?: string
  width?: number
  description?: string
  strength?: number
  type?: string
  source?: string
  target?: string
}

interface VisGraphData {
  nodes: VisNode[]
  edges: VisEdge[]
}

const route = useRoute()
const kbId = ref<string>(route.params.kbId as string)

// 引用
const graphContainer = ref<HTMLElement>()

// 数据状态
const graphData = ref<VisGraphData>({ nodes: [], edges: [] })
const loading = ref(false)
const searchText = ref('')
const selectedNode = ref<VisNode | null>(null)
const nodeRelations = ref<RelationDisplay[]>([])

// 详情抽屉
const detailDrawerVisible = ref(false)

// 添加节点对话框
const addNodeDialogVisible = ref(false)
const addNodeForm = ref({
  name: '',
  entity_type: 'concept',
  attributesStr: ''
})

// 添加关系对话框
const addRelationDialogVisible = ref(false)
const addRelationForm = ref({
  source: '',
  target: '',
  type: 'relates',
  strength: 5
})

const adding = ref(false)
const deleting = ref(false)

// 关系类型选项
const relationTypeOptions = ref<RelationTypeOption[]>([])

// 编辑节点对话框
const editNodeDialogVisible = ref(false)
const editNodeForm = ref({
  id: '',
  name: '',
  entity_type: 'concept',
  attributesStr: ''
})

// 编辑关系对话框
const editRelationDialogVisible = ref(false)
const editRelationForm = ref({
  id: '',
  type: '关联',
  strength: 5
})

const editing = ref(false)

// vis-network 实例
let network: any = null

// 统计数据
const nodeCount = computed(() => graphData.value.nodes?.length || 0)
const edgeCount = computed(() => graphData.value.edges?.length || 0)
const entityTypeCount = computed(() => {
  const types = new Set(graphData.value.nodes?.map(n => n.entity_type) || [])
  return types.size
})

// 节点详情计算属性
const selectedNodeType = computed(() => selectedNode.value?.entity_type || '-')
const selectedNodeAttributes = computed(() => selectedNode.value?.attributes || [])
const selectedNodeChunks = computed(() => selectedNode.value?.chunks || [])

// 颜色映射（根据实体类型）
const typeColorMap: Record<string, string> = {
  // 后端返回的实体类型（首字母大写）
  Department: '#5B8FF9',
  Module: '#F4664A',
  Concept: '#722ED1',
  Technology: '#06D177',
  Product: '#FADB14',
  Company: '#E6A23C',
  Other: '#909399',
  // 兼容小写
  person: '#5B8FF9',
  organization: '#F4664A',
  location: '#06D177',
  concept: '#722ED1',
  other: '#909399'
}

// 初始化图谱
function initGraph() {
  console.log('[initGraph] 开始初始化 vis-network 图谱')
  if (!graphContainer.value) {
    console.error('[initGraph] graphContainer 为空')
    ElMessage.error('图谱容器未找到')
    return
  }

  const container = graphContainer.value as HTMLElement

  // vis-network 配置
  const options = {
    nodes: {
      shape: 'dot',
      size: 16,
      font: {
        size: 14,
        color: '#303133'
      },
      borderWidth: 2,
      shadow: true
    },
    edges: {
      width: 2,
      color: {
        color: '#848484',
        highlight: '#409eff',
        hover: '#409eff'
      },
      arrows: {
        to: {
          enabled: true,
          scaleFactor: 1,
          type: 'arrow'
        }
      },
      smooth: {
        enabled: true,
        type: 'continuous',
        roundness: 0.5
      },
      font: {
        size: 14,
        color: '#409eff',  // 使用蓝色，与节点名称的深色区分
        background: 'rgba(255, 255, 255, 0.9)',
        strokeWidth: 0,
        align: 'middle'
      },
      labelHighlightBold: true
    },
    physics: {
      forceAtlas2Based: {
        gravitationalConstant: -50,
        centralGravity: 0.01,
        springLength: 100,
        springConstant: 0.08
      },
      maxVelocity: 50,
      solver: 'forceAtlas2Based',
      timestep: 0.35,
      stabilization: {
        iterations: 150
      }
    },
    interaction: {
      hover: true,
      tooltipDelay: 200,
      zoomView: true,
      dragView: true
    }
  }

  try {
    // 创建节点和边的数据集
    const nodesDataSet = new DataSet(graphData.value.nodes)
    const edgesDataSet = new DataSet(graphData.value.edges)

    // 创建网络实例
    network = new Network(container, { nodes: nodesDataSet, edges: edgesDataSet }, options)

    console.log('[initGraph] vis-network 创建成功')

    // 监听节点点击事件
    network.on('click', (params: any) => {
      if (params.nodes.length > 0) {
        const nodeId = params.nodes[0]
        const node = graphData.value.nodes.find(n => n.id === nodeId)
        if (node) {
          selectedNode.value = node
          detailDrawerVisible.value = true
          loadNodeRelations(node.id)  // 传递节点 ID 而不是名称
        }
      }
    })

    // 监听节点悬停事件
    network.on('hoverNode', () => {
      network.canvas.body.container.style.cursor = 'pointer'
    })

    network.on('blurNode', () => {
      network.canvas.body.container.style.cursor = 'default'
    })

  } catch (error) {
    console.error('[initGraph] vis-network 初始化异常:', error)
    ElMessage.error(`图谱初始化失败: ${error}`)
  }
}

// 加载图谱数据
async function loadGraph() {
  console.log('[loadGraph] 开始加载图谱数据')
  loading.value = true
  try {
    const res = await graphApi.getGraph(kbId.value)
    console.log('[loadGraph] API 响应:', res)
    console.log('[loadGraph] 响应 data:', res.data)
    console.log('[loadGraph] data.Node:', res.data?.Node)
    console.log('[loadGraph] data.nodes:', res.data?.nodes)
    console.log('[loadGraph] data.Relation:', res.data?.Relation)
    console.log('[loadGraph] data.relations:', res.data?.relations)

    if (res.data) {
      // 转换后端数据到 vis-network 格式
      graphData.value = convertToVisData(res.data)
      console.log('[loadGraph] 转换后的 vis-network 数据:', graphData.value)
      console.log('[loadGraph] 节点数:', graphData.value.nodes.length)
      console.log('[loadGraph] 边数:', graphData.value.edges.length)

      // 重新初始化网络以显示新数据
      if (network) {
        network.destroy()
        network = null
      }
      initGraph()
    }
  } catch (error: any) {
    console.error('[loadGraph] 异常:', error)
    ElMessage.error(`加载图谱失败: ${error?.message || error}`)
  } finally {
    loading.value = false
  }
}

// 转换后端数据到 vis-network 格式
function convertToVisData(data: GraphData): VisGraphData {
  console.log('[convertToVisData] 输入数据:', data)

  const rawNodes = data.Node || data.nodes || []
  const rawEdges = data.Relation || data.relations || []

  console.log('[convertToVisData] 原始节点数:', rawNodes.length, '边数:', rawEdges.length)
  console.log('[convertToVisData] 原始关系数据样本:', rawEdges[0])

  const nodes: VisNode[] = rawNodes.map(node => {
    const entityType = node.entity_type || 'Other'
    return {
      id: node.id,
      label: node.name,
      color: typeColorMap[entityType] || typeColorMap['Other'],
      size: 20,
      attributes: node.attributes || [],
      chunks: node.chunks || [],
      entity_type: entityType
    }
  })

  // 创建节点名称到ID的映射（用于处理 source/target 可能是名称的情况）
  const nodeNameToId = new Map<string, string>()
  nodes.forEach(node => {
    nodeNameToId.set(node.label, node.id)
    // 如果节点有 name 字段，也映射它
    if (node.attributes) {
      nodeNameToId.set(node.label, node.id)
    }
  })

  console.log('[convertToVisData] 节点名称到ID的映射:', Array.from(nodeNameToId.entries()).slice(0, 10))

  console.log('[convertToVisData] 原始关系数据样本:', rawEdges[0])

  const edges: VisEdge[] = rawEdges.map(rel => {
    // 尝试通过 source 查找对应的节点 ID
    let fromId = rel.source
    let toId = rel.target

    // 如果 source 不是有效的节点 ID，尝试通过名称查找
    const fromNode = nodes.find(n => n.id === rel.source || n.label === rel.source)
    const toNode = nodes.find(n => n.id === rel.target || n.label === rel.target)

    if (fromNode) fromId = fromNode.id
    if (toNode) toId = toNode.id

    // 移除循环内的 console.log 以避免 DevTools 打开时浏览器崩溃
    // console.log('[convertToVisData] 边映射:', rel.source, '->', fromId, '|', rel.target, '->', toId)

    return {
      id: rel.id,
      from: fromId,
      to: toId,
      label: rel.type,
      color: getEdgeColor(rel.weight || 5),
      width: Math.max(2, (rel.weight || 5) / 2),
      description: rel.description || '',
      strength: rel.strength || 0,
      type: rel.type,
      // 保留原始的 source 和 target（节点名称），用于显示
      source: rel.source,
      target: rel.target
    }
  })

  console.log('[convertToVisData] 转换后节点数:', nodes.length, '边数:', edges.length)
  console.log('[convertToVisData] 边样本:', edges[0])
  console.log('[convertToVisData] 节点ID列表:', nodes.map(n => n.id).slice(0, 10))

  return { nodes, edges }
}

// 获取边的颜色（根据权重）
function getEdgeColor(weight: number): string {
  if (weight >= 8) return '#F4664A'
  if (weight >= 5) return '#E6A23C'
  if (weight >= 3) return '#FADB14'
  return '#91d5ff'
}

// 加载节点关系
async function loadNodeRelations(nodeId: string) {
  try {
    console.log('[loadNodeRelations] 加载节点关系, nodeId:', nodeId)

    // 从已加载的图谱数据中筛选与该节点相关的关系
    const relations = graphData.value.edges.filter((edge: VisEdge) => {
      return edge.from === nodeId || edge.to === nodeId
    })

    console.log('[loadNodeRelations] 找到关系数量:', relations.length)

    // 辅助函数：根据节点 ID 获取节点名称
    const getNodeName = (nodeId: string) => {
      const node = graphData.value.nodes.find(n => n.id === nodeId)
      return node?.label || nodeId
    }

    // 转换为显示格式
    nodeRelations.value = relations.map((rel: VisEdge) => ({
      id: rel.id,
      label: rel.label || rel.type,
      type: rel.type,
      description: rel.description || '-',
      strength: rel.strength || 0,
      from: getNodeName(rel.from),
      to: getNodeName(rel.to),
      // 使用 preserved source/target (节点名称)，如果没有则回退到查找节点名称
      source: rel.source || getNodeName(rel.from),
      target: rel.target || getNodeName(rel.to)
    }))

    console.log('[loadNodeRelations] 加载关系成功，数量:', nodeRelations.value.length)
  } catch (error: any) {
    console.error('[loadNodeRelations] 加载失败:', error)
    ElMessage.error('加载节点关系失败')
  }
}

// 搜索节点
async function handleSearch() {
  if (!searchText.value.trim()) {
    loadGraph()
    return
  }

  loading.value = true
  try {
    const res = await graphApi.searchNode(kbId.value, { nodes: [searchText.value] })
    if (res.data) {
      graphData.value = convertToVisData(res.data)
      // 重新初始化网络
      if (network) {
        network.destroy()
        network = null
      }
      initGraph()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '搜索失败')
  } finally {
    loading.value = false
  }
}

// 显示添加节点对话框
function showAddNodeDialog() {
  addNodeForm.value = {
    name: '',
    entity_type: 'concept',
    attributesStr: ''
  }
  addNodeDialogVisible.value = true
}

// 添加节点
async function handleAddNode() {
  if (!addNodeForm.value.name) {
    ElMessage.warning('请输入节点名称')
    return
  }

  adding.value = true
  try {
    const attributes = addNodeForm.value.attributesStr
      ? addNodeForm.value.attributesStr.split(',').map(s => s.trim())
      : []

    const data: AddNodeRequest = {
      name: addNodeForm.value.name,
      entity_type: addNodeForm.value.entity_type,
      attributes
    }

    const res = await graphApi.addNode(kbId.value, data)
    if (res.data) {
      ElMessage.success('节点添加成功')
      addNodeDialogVisible.value = false
      await loadGraph()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '添加节点失败')
  } finally {
    adding.value = false
  }
}

// 显示添加关系对话框
function showAddRelationDialog() {
  addRelationForm.value = {
    source: '',
    target: '',
    type: 'relates',
    strength: 5
  }
  addRelationDialogVisible.value = true
}

// 添加关系
async function handleAddRelation() {
  if (!addRelationForm.value.source || !addRelationForm.value.target) {
    ElMessage.warning('请选择源节点和目标节点')
    return
  }

  adding.value = true
  try {
    const data: AddRelationRequest = {
      source: addRelationForm.value.source,
      target: addRelationForm.value.target,
      type: addRelationForm.value.type,
      strength: addRelationForm.value.strength
    }

    const res = await graphApi.addRelation(kbId.value, data)
    if (res.data) {
      ElMessage.success('关系添加成功')
      addRelationDialogVisible.value = false
      await loadGraph()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '添加关系失败')
  } finally {
    adding.value = false
  }
}

// 显示编辑节点对话框
function showEditNodeDialog() {
  if (!selectedNode.value) return

  editNodeForm.value = {
    id: selectedNode.value.id,
    name: selectedNode.value.label,
    entity_type: selectedNode.value.entity_type || 'concept',
    attributesStr: selectedNode.value.attributes?.join(', ') || ''
  }
  editNodeDialogVisible.value = true
}

// 编辑节点
async function handleEditNode() {
  if (!editNodeForm.value.name) {
    ElMessage.warning('请输入节点名称')
    return
  }

  editing.value = true
  try {
    const attributes = editNodeForm.value.attributesStr
      ? editNodeForm.value.attributesStr.split(',').map(s => s.trim())
      : []

    const data: UpdateNodeRequest = {
      name: editNodeForm.value.name,
      title: editNodeForm.value.name,
      entity_type: editNodeForm.value.entity_type,
      attributes
    }

    const res = await graphApi.updateNode(kbId.value, editNodeForm.value.id, data)
    if (res.data) {
      ElMessage.success('节点更新成功')
      editNodeDialogVisible.value = false
      detailDrawerVisible.value = false
      await loadGraph()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新节点失败')
  } finally {
    editing.value = false
  }
}

// 显示编辑关系对话框
function showEditRelationDialog(rel: RelationDisplay) {
  editRelationForm.value = {
    id: rel.id,
    type: rel.type || rel.label || '关联',
    strength: rel.strength || 5
  }
  editRelationDialogVisible.value = true
}

// 编辑关系
async function handleEditRelation() {
  if (!editRelationForm.value.type) {
    ElMessage.warning('请选择关系类型')
    return
  }

  editing.value = true
  try {
    // 从原关系中获取 description，保持不变（但排除占位符 "-"）
    const originalRelation = nodeRelations.value.find(r => r.id === editRelationForm.value.id)
    const originalDescription = originalRelation?.description && originalRelation.description !== '-'
      ? originalRelation.description
      : ''

    const data: UpdateRelationRequest = {
      type: editRelationForm.value.type,
      description: originalDescription,
      strength: editRelationForm.value.strength
    }

    console.log('[handleEditRelation] 更新关系数据:', data)

    const res = await graphApi.updateRelation(kbId.value, editRelationForm.value.id, data)
    if (res.data) {
      ElMessage.success('关系更新成功')
      editRelationDialogVisible.value = false
      detailDrawerVisible.value = false
      await loadGraph()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '更新关系失败')
  } finally {
    editing.value = false
  }
}

// 导出图谱
function handleExport() {
  if (!graphData.value.nodes?.length) {
    ElMessage.warning('没有可导出的数据')
    return
  }

  const dataStr = JSON.stringify(graphData.value, null, 2)
  const blob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `graph-${kbId.value}-${Date.now()}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

// 加载关系类型选项
async function loadRelationTypes() {
  try {
    const res = await graphApi.getRelationTypes(kbId.value)
    if (res.data) {
      relationTypeOptions.value = res.data
    }
  } catch (error: any) {
    console.error('[loadRelationTypes] 加载失败:', error)
    // 使用默认选项
    relationTypeOptions.value = [
      { value: 'contains', label: '包含' },
      { value: 'relates', label: '关联' },
      { value: 'depends', label: '依赖' },
      { value: 'belongs', label: '属于' },
      { value: 'owns', label: '拥有' },
      { value: 'author', label: '作者' },
      { value: 'alias', label: '别名' },
      { value: 'other', label: '其他' }
    ]
  }
}

// 删除节点
async function handleDeleteNode() {
  if (!selectedNode.value) return

  try {
    await ElMessageBox.confirm(
      `确定要删除节点"${selectedNode.value.label}"吗？删除节点将同时删除所有相关关系。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    return
  }

  deleting.value = true
  try {
    await graphApi.deleteNode(kbId.value, selectedNode.value.id)
    ElMessage.success('节点删除成功')
    detailDrawerVisible.value = false
    await loadGraph()
  } catch (error: any) {
    ElMessage.error(error.message || '删除节点失败')
  } finally {
    deleting.value = false
  }
}

// 删除关系
async function handleDeleteRelation(rel: RelationDisplay) {
  try {
    await ElMessageBox.confirm(
      `确定要删除关系吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    return
  }

  deleting.value = true
  try {
    await graphApi.deleteRelation(kbId.value, rel.id)
    ElMessage.success('关系删除成功')
    detailDrawerVisible.value = false
    await loadGraph()
  } catch (error: any) {
    ElMessage.error(error.message || '删除关系失败')
  } finally {
    deleting.value = false
  }
}

// 全屏切换
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    graphContainer.value?.requestFullscreen()
  } else {
    document.exitFullscreen()
  }
}

onMounted(async () => {
  console.log('[GraphView] 组件已挂载')
  console.log('[GraphView] kbId:', kbId.value)

  // 等待 DOM 更新后再初始化
  await nextTick()

  console.log('[GraphView] graphContainer:', graphContainer.value)
  if (!graphContainer.value) {
    console.error('[GraphView] graphContainer 未找到')
    return
  }

  // 加载数据（会自动初始化图谱）
  await loadGraph()
  console.log('[GraphView] 数据加载完成')

  // 加载关系类型选项
  await loadRelationTypes()
})

onUnmounted(() => {
  // 销毁图谱实例
  if (network) {
    network.destroy()
    network = null
  }
})
</script>

<style scoped>
.graph-view-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background: white;
  border-radius: 8px;
  padding: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.search-box {
  width: 300px;
}

.action-buttons {
  margin-right: auto;
}

.graph-container {
  flex: 1;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
  background: #fafafa;
  min-height: 500px;
}

.status-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-top: 16px;
  font-size: 14px;
  color: #606266;
}

.relations-list {
  margin-top: 16px;
}

.relation-item {
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 12px;
}

.relation-header {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.relation-content {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
}

.relation-arrow {
  color: #409eff;
}

.relation-description {
  color: #909399;
  font-size: 13px;
}
</style>
