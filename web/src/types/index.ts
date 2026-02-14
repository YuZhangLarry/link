/**
 * 全局类型定义
 */

// ============ 认证相关 ============
export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  username: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_at: number
  user: UserInfo
  tenant_id?: number
}

export interface UserInfo {
  id: number
  email: string
  username: string
  avatar?: string
  created_at: string
  updated_at: string
}

// ============ 租户相关 ============
export interface Tenant {
  id: number
  name: string
  description?: string
  api_key?: string
  storage_used?: number
  storage_limit?: number
  created_at: string
  updated_at: string
}

export interface CreateTenantRequest {
  name: string
  description?: string
}

export interface UpdateTenantRequest {
  name?: string
  description?: string
}

// ============ 聊天相关 ============
export interface ChatRequest {
  content: string
  session_id?: string
  stream?: boolean
  model?: string
  temperature?: number
  max_tokens?: number
}

export interface ChatResponse {
  content: string
  role: string
  token_count: number
  tool_calls?: ToolCall[]
}

export interface ToolCall {
  id: string
  type: string
  function: {
    name: string
    arguments: string
  }
}

export interface StreamChatEvent {
  event: 'content' | 'end' | 'error'
  content: string
  message_id?: string
  token_count?: number
  tool_calls?: ToolCall[]
  error?: string
}

// ============ 会话相关 ============
export interface Session {
  id: string
  title: string
  description?: string
  status: number
  max_rounds: number
  created_at: string
  updated_at: string
}

export interface CreateSessionRequest {
  title?: string
  description?: string
  max_rounds?: number
}

export interface UpdateSessionRequest {
  title?: string
  description?: string
  status?: number
}

export interface SessionDetail {
  session: Session
  messages: Message[]
}

export interface SessionListResponse {
  items: Session[]
  total: number
  page: number
  size: number
}

// ============ 消息相关 ============
export interface Message {
  id: string
  session_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  tool_calls?: string
  token_count: number
  created_at: string
}

export interface CreateMessageRequest {
  session_id: string
  role: string
  content: string
  tool_calls?: string
  token_count?: number
}

// ============ 知识库相关 ============
export interface KnowledgeBase {
  id: string
  name: string
  description?: string
  avatar?: string
  is_public?: boolean
  status: number
  created_at: string
  updated_at: string
  // 统计字段
  document_count?: number
  chunk_count?: number
  storage_size?: number
  // 数据处理配置（从 kb_settings 表获取）
  setting?: {
    chunk_size?: number
    chunk_overlap?: number
    graph_enabled?: boolean
    bm25_enabled?: boolean
    image_processing_mode?: string
    extract_mode?: string
  }
  // 检索模式数组（从后端返回）
  retrieval_modes?: string[]
}

export interface CreateKnowledgeBaseRequest {
  name: string
  description?: string
  avatar?: string
  is_public?: boolean
  // 数据处理配置（存储到 kb_settings 表）
  chunk_size?: number
  chunk_overlap?: number
  graph_enabled?: boolean
  bm25_enabled?: boolean
  image_processing_mode?: string  // 'none' | 'ocr' | 'vlm' | 'all'
  extract_mode?: string  // 'none' | 'rule' | 'llm'
}

export interface UpdateKnowledgeBaseRequest {
  name?: string
  description?: string
  avatar?: string
  is_public?: boolean
  status?: number
  // 数据处理配置（存储到 kb_settings 表）
  chunk_size?: number
  chunk_overlap?: number
  graph_enabled?: boolean
  bm25_enabled?: boolean
  image_processing_mode?: string  // 'none' | 'ocr' | 'vlm' | 'all'
  extract_mode?: string  // 'none' | 'rule' | 'llm'
}

export interface KnowledgeBaseStats {
  kb_id: string
  knowledge_count: number
  chunk_count: number
  total_size: number
}

// 知识条目相关
export interface Knowledge {
  id: string
  kb_id: string
  title: string
  type: string
  storage_size: number
  file_path: string
  parse_status: 'unprocessed' | 'pending' | 'processing' | 'completed' | 'failed'
  enable_status: 'enabled' | 'disabled'
  chunk_count: number
  created_at: string
  processed_at?: string
}

export interface UploadKnowledgeFileRequest {
  file: File
  title?: string
  file_type?: string
  chunk_size?: number
  chunk_overlap?: number
}

export interface KnowledgeStatus {
  knowledge_id: string
  parse_status: 'pending' | 'processing' | 'completed' | 'failed'
  enable_status: 'enabled' | 'disabled'
  chunk_count: number
  created_at: string
  processed_at?: string
}

// 文档分块相关
export interface Chunk {
  id: string
  kb_id: string
  knowledge_id: string
  content: string
  chunk_index: number
  token_count?: number
  embedding_id?: string
  created_at: string
}

export interface ChunkListResponse {
  items: Chunk[]
  total: number
  page: number
  size: number
}

// 知识检索相关
export interface SearchRequest {
  query: string
  kb_ids: string[]
  top_k?: number
  score_threshold?: number
  include_graph?: boolean
}

export interface SearchResult {
  chunk_id: string
  knowledge_id: string
  knowledge_title: string
  content: string
  score: number
  metadata?: Record<string, any>
  tags?: string[]
}

export interface SearchResponse {
  query: string
  results: SearchResult[]
  total: number
  graph_entities?: any[]
  graph_relationships?: any[]
}

// ============ FAQ相关 ============
export interface FAQEntry {
  id: string
  kb_id: string
  question: string
  answer: string
  priority: number
  enabled: boolean
  created_at: string
  updated_at: string
}

// ============ Agent相关 ============
export interface Agent {
  id: string
  name: string
  description?: string
  avatar?: string
  system_prompt?: string
  model?: string
  temperature?: number
  enabled: boolean
  created_at: string
  updated_at: string
}

// 导出图谱相关类型
export * from './graph'
