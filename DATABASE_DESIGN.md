# 数据库表设计文档

> WeKnora 知识图谱 RAG 系统数据库设计

## 📊 数据存储架构

```
┌─────────────────────────────────────────────────┐
│              应用层 (Go Backend)               │
└─────────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
┌─────────────┐ ┌──────────┐ ┌─────────────┐
│    MySQL    │ │  Neo4j   │ │   Milvus    │
│  元数据库    │ │  知识图谱 │ │  向量数据库  │
└─────────────┘ └──────────┘ └─────────────┘
     │                │              │
     │                │              │
  ┌──┴────────────────┴───────────────┴──┐
  │            Redis 缓存               │
  └────────────────────────────────────┘
```

## 🗄️ MySQL 表设计

### 1. 用户模块 (User)

#### users - 用户表
```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT '邮箱',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    avatar VARCHAR(500) COMMENT '头像URL',
    status TINYINT DEFAULT 1 COMMENT '状态: 0=禁用, 1=正常',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',

    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
```

#### user_preferences - 用户偏好设置
```sql
CREATE TABLE user_preferences (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '偏好ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    language VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言',
    theme VARCHAR(20) DEFAULT 'light' COMMENT '主题: light/dark',
    notification_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用通知',
    preference_json JSON COMMENT '其他偏好设置(JSON格式)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户偏好设置';
```

### 2. 知识库模块 (Knowledge Base)

#### knowledge_bases - 知识库表
```sql
CREATE TABLE knowledge_bases (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '知识库ID',
    user_id BIGINT NOT NULL COMMENT '所属用户ID',
    name VARCHAR(100) NOT NULL COMMENT '知识库名称',
    description TEXT COMMENT '描述',
    avatar VARCHAR(500) COMMENT '图标/封面',
    embedding_model VARCHAR(50) DEFAULT 'text-embedding-ada-002' COMMENT '向量模型',
    chunk_size INT DEFAULT 1000 COMMENT '分块大小',
    chunk_overlap INT DEFAULT 200 COMMENT '分块重叠',
    status TINYINT DEFAULT 1 COMMENT '状态: 0=禁用, 1=启用',
    is_public BOOLEAN DEFAULT FALSE COMMENT '是否公开',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_public (is_public),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库表';
```

#### kb_settings - 知识库设置
```sql
CREATE TABLE kb_settings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '设置ID',
    kb_id BIGINT NOT NULL COMMENT '知识库ID',
    retrieval_mode VARCHAR(20) DEFAULT 'hybrid' COMMENT '检索模式: vector/bm25/hybrid/graph',
    similarity_threshold DECIMAL(3,2) DEFAULT 0.7 COMMENT '相似度阈值',
    top_k INT DEFAULT 10 COMMENT '返回结果数量',
    rerank_enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用重排序',
    graph_enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用图谱检索',
    settings_json JSON COMMENT '其他设置(JSON格式)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    UNIQUE KEY uk_kb_id (kb_id),
    FOREIGN KEY (kb_id) REFERENCES knowledge_bases(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库设置';
```

### 3. 文档模块 (Document)

#### documents - 文档表
```sql
CREATE TABLE documents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '文档ID',
    kb_id BIGINT NOT NULL COMMENT '所属知识库ID',
    user_id BIGINT NOT NULL COMMENT '上传用户ID',
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    file_type VARCHAR(50) COMMENT '文件类型: pdf/docx/txt/md',
    file_size BIGINT COMMENT '文件大小(字节)',
    file_path VARCHAR(500) COMMENT '存储路径',
    file_hash VARCHAR(64) COMMENT '文件哈希(SHA256)',
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending' COMMENT '处理状态',
    error_message TEXT COMMENT '错误信息',
    chunk_count INT DEFAULT 0 COMMENT '分块数量',
    processed_at TIMESTAMP NULL COMMENT '处理完成时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_kb_id (kb_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_file_hash (file_hash),
    FOREIGN KEY (kb_id) REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档表';
```

#### chunks - 文档分块表
```sql
CREATE TABLE chunks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '分块ID',
    document_id BIGINT NOT NULL COMMENT '文档ID',
    chunk_index INT NOT NULL COMMENT '分块序号',
    content TEXT NOT NULL COMMENT '文本内容',
    token_count INT COMMENT 'Token数量',
    embedding_id VARCHAR(100) COMMENT '向量ID(Milvus)',
    metadata JSON COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_document_id (document_id),
    INDEX idx_embedding_id (embedding_id),
    UNIQUE KEY uk_doc_chunk (document_id, chunk_index),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档分块表';
```

### 4. 对话模块 (Chat)

#### chats - 对话表
```sql
CREATE TABLE chats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '对话ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    kb_id BIGINT COMMENT '关联知识库ID',
    title VARCHAR(255) COMMENT '对话标题',
    message_count INT DEFAULT 0 COMMENT '消息数量',
    model VARCHAR(50) DEFAULT 'gpt-3.5-turbo' COMMENT '使用的模型',
    status TINYINT DEFAULT 1 COMMENT '状态: 0=归档, 1=正常',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_user_id (user_id),
    INDEX idx_kb_id (kb_id),
    INDEX idx_status (status),
    INDEX idx_updated_at (updated_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kb_id) REFERENCES knowledge_bases(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='对话表';
```

#### messages - 消息表
```sql
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '消息ID',
    chat_id BIGINT NOT NULL COMMENT '对话ID',
    role ENUM('system', 'user', 'assistant', 'tool') NOT NULL COMMENT '角色',
    content TEXT NOT NULL COMMENT '消息内容',
    tool_calls JSON COMMENT '工具调用记录',
    token_count INT COMMENT 'Token使用量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_chat_id (chat_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';
```

#### message_feedback - 消息反馈表
```sql
CREATE TABLE message_feedback (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '反馈ID',
    message_id BIGINT NOT NULL COMMENT '消息ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    rating INT CHECK (rating IN (1, 2, 3, 4, 5)) COMMENT '评分: 1-5星',
    comment TEXT COMMENT '评论',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_message_id (message_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息反馈表';
```

### 5. Agent & 工具模块

#### tools - 工具表
```sql
CREATE TABLE tools (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '工具ID',
    name VARCHAR(100) NOT NULL COMMENT '工具名称',
    type VARCHAR(50) NOT NULL COMMENT '工具类型: search/database/http/custom',
    description TEXT COMMENT '描述',
    config JSON NOT NULL COMMENT '配置(JSON格式)',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_by BIGINT COMMENT '创建者',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_type (type),
    INDEX idx_enabled (enabled),
    FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工具表';
```

#### tool_executions - 工具执行记录
```sql
CREATE TABLE tool_executions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '执行ID',
    message_id BIGINT NOT NULL COMMENT '关联消息ID',
    tool_id BIGINT NOT NULL COMMENT '工具ID',
    input_params JSON COMMENT '输入参数',
    output_data JSON COMMENT '输出数据',
    status ENUM('success', 'failed', 'timeout') COMMENT '执行状态',
    duration_ms INT COMMENT '执行时长(毫秒)',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '执行时间',

    INDEX idx_message_id (message_id),
    INDEX idx_tool_id (tool_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工具执行记录';
```

### 6. 检索模块

#### search_history - 搜索历史
```sql
CREATE TABLE search_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '搜索ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    kb_id BIGINT COMMENT '知识库ID',
    query TEXT NOT NULL COMMENT '查询内容',
    retrieval_type VARCHAR(20) COMMENT '检索类型: vector/bm25/hybrid/graph',
    result_count INT COMMENT '结果数量',
    latency_ms INT COMMENT '耗时(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '搜索时间',

    INDEX idx_user_id (user_id),
    INDEX idx_kb_id (kb_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (kb_id) REFERENCES knowledge_bases(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索历史';
```

### 7. 系统模块

#### api_keys - API密钥管理
```sql
CREATE TABLE api_keys (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '密钥ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    name VARCHAR(100) NOT NULL COMMENT '密钥名称',
    key_hash VARCHAR(64) NOT NULL COMMENT '密钥哈希',
    key_prefix VARCHAR(20) NOT NULL COMMENT '密钥前缀(用于显示)',
    scopes JSON COMMENT '权限范围',
    last_used_at TIMESTAMP NULL COMMENT '最后使用时间',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    status TINYINT DEFAULT 1 COMMENT '状态: 0=禁用, 1=启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_user_id (user_id),
    INDEX idx_key_hash (key_hash),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API密钥管理';
```

#### audit_logs - 审计日志
```sql
CREATE TABLE audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    user_id BIGINT COMMENT '用户ID',
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) COMMENT '资源类型: user/kb/document/chat',
    resource_id BIGINT COMMENT '资源ID',
    details JSON COMMENT '详细信息',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT 'User-Agent',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',

    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志';
```

## 🕸️ Neo4j 图数据库设计

### 节点类型 (Node Labels)

```cypher
// 文档节点
(:Document {
  id: "uuid",
  kb_id: "kb_id",
  name: "文档名称",
  type: "pdf",
  created_at: datetime()
})

// 分块节点
(:Chunk {
  id: "uuid",
  doc_id: "doc_id",
  content: "文本内容",
  index: 0,
  token_count: 256
})

// 实体节点
(:Entity {
  name: "实体名称",
  type: "PERSON|ORG|LOC|CONCEPT|EVENT",
  description: "描述",
  aliases: ["别名1", "别名2"],
  properties: {}
})

// 概念节点
(:Concept {
  name: "概念名",
  definition: "定义",
  category: "分类"
})

// 用户节点
(:User {
  id: "user_id",
  username: "用户名",
  email: "邮箱"
})
```

### 关系类型 (Relationship Types)

```cypher
// 文档关系
(:Document)-[:HAS_CHUNK]->(:Chunk)
(:Document)-[:UPLOADED_BY]->(:User)

// 实体关系
(:Chunk)-[:MENTIONS]->(:Entity)
(:Entity)-[:RELATES_TO {weight: 0.8, source: "extracted"}]->(:Entity)
(:Entity)-[:IS_A]->(:Concept)
(:Entity)-[:BELONGS_TO]->(:Document)

// 概念关系
(:Concept)-[:RELATED_TO]->(:Concept)

// 用户关系
(:User)-[:OWNS]->(:Document)
(:User)-[:INTERACTED_WITH]->(:Entity)
```

### 索引优化

```cypher
// 创建索引
CREATE INDEX ON :Document(id);
CREATE INDEX ON :Document(kb_id);
CREATE INDEX ON :Entity(name);
CREATE INDEX ON :Entity(type);
CREATE INDEX ON :Chunk(doc_id);
CREATE INDEX ON :User(id);

// 全文搜索索引
CREATE FULLTEXT INDEX ON :Chunk(content);
CREATE FULLTEXT INDEX ON :Entity(name, description);
```

## 📊 Milvus 向量数据库设计

### Collection: document_chunks

```go
Schema {
  CollectionName: "document_chunks",

  Fields: [
    {
      Name: "id",
      DataType: Int64,
      PrimaryKey: true,
      AutoID: false
    },
    {
      Name: "embedding",
      DataType: FloatVector,
      Dim: 1536  // 根据模型调整
    },
    {
      Name: "chunk_id",
      DataType: VarChar,
      MaxLength: 100
    },
    {
      Name: "doc_id",
      DataType: VarChar,
      MaxLength: 100
    },
    {
      Name: "kb_id",
      DataType: VarChar,
      MaxLength: 50
    }
  ]
}

// 索引配置
Index {
  Type: "IVF_FLAT",
  MetricType: "L2",
  Params: {
    "nlist": 128
  }
}
```

## 🔑 Redis 缓存设计

### Key 设计规范

```
# 用户会话
session:{user_id}:{session_id} -> SessionData (TTL: 24h)

# 查询缓存
search:{kb_id}:{query_hash} -> SearchResults (TTL: 1h)

# 文档状态
doc:status:{doc_id} -> ProcessingStatus (TTL: 10min)

# 限流
rate_limit:{user_id}:{action} -> Count (TTL: 1min)

# 热点数据
kb:info:{kb_id} -> KnowledgeBaseInfo (TTL: 30min)
```

### 数据结构示例

```go
// 会话数据
type SessionData struct {
    UserID    string
    Username  string
    CreatedAt time.Time
    ExpiresAt time.Time
}

// 搜索结果缓存
type SearchResults struct {
    Query       string
    Results     []Chunk
    Total       int
    CachedAt    time.Time
}

// 文档处理状态
type ProcessingStatus struct {
    Status      string  // pending/processing/completed/failed
    Progress    float64 // 0.0 - 1.0
    TotalChunks int
    ProcessedChunks int
    Error       string
}
```

## 📈 数据表关系图

```
┌──────────────┐
│     users     │
└──────┬───────┘
       │
       ├──────────────────┐
       │                  │
       ▼                  ▼
┌──────────────┐  ┌──────────────┐
│knowledge_base│  │  chats       │
└──────┬───────┘  └──────┬───────┘
       │                 │
       ▼                 ▼
┌──────────────┐  ┌──────────────┐
│  documents  │  │  messages    │
└──────┬───────┘  └──────────────┘
       │
       ▼
┌──────────────┐
│   chunks     │
└──────────────┘
```

## 🔧 数据库迁移脚本

### 初始化脚本

**位置**: `migrations/`

```sql
-- 000001_init_schema.up.sql
-- 001_create_users.sql
-- 002_create_knowledge_bases.sql
-- ...
```

## 📊 数据统计查询

```sql
-- 用户统计
SELECT
    COUNT(*) as total_users,
    COUNT(CASE WHEN DATE(created_at) = CURDATE() THEN 1 END) as today_new_users
FROM users;

-- 知识库统计
SELECT
    u.username,
    COUNT(DISTINCT kb.id) as kb_count,
    COUNT(DISTINCT doc.id) as doc_count
FROM users u
LEFT JOIN knowledge_bases kb ON kb.user_id = u.id
LEFT JOIN documents doc ON doc.kb_id = kb.id
GROUP BY u.id
ORDER BY kb_count DESC;

-- 存储空间统计
SELECT
    file_type,
    COUNT(*) as count,
    SUM(file_size) as total_size,
    SUM(file_size) / 1024 / 1024 / 1024 as size_gb
FROM documents
WHERE status = 'completed'
GROUP BY file_type;
```

## 🎯 最佳实践

### 1. 命名规范
- 表名: 小写复数 (`users`, `documents`)
- 字段名: 小写下划线分隔 (`created_at`, `file_name`)
- 索引名: `idx_` 前缀 (`idx_user_id`)

### 2. 字段类型选择
- 主键: `BIGINT AUTO_INCREMENT`
- 外键: `BIGINT`
- 文本大段: `TEXT`
- 固定长度文本: `VARCHAR`
- JSON 数据: `JSON` 类型

### 3. 索引策略
- 为外键创建索引
- 为常用查询条件创建索引
- 复合索引遵循最左前缀原则

### 4. 软删除
```sql
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN deleted_by BIGINT NULL;
-- 查询时添加 WHERE deleted_at IS NULL
```

---

**文档版本**: v1.0.0
**最后更新**: 2026-02-03
