# 📊 数据库设计概览

## 🗄️ 数据库架构总览

```
┌─────────────────────────────────────────────────┐
│              WeKnora RAG 系统                    │
└─────────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
┌─────────────┐ ┌──────────┐ ┌─────────────┐
│    MySQL    │ │  Neo4j   │ │   Milvus    │
│  元数据库    │ │  知识图谱 │ │  向量数据库  │
│  17 张表     │ │  节点+关系 │ │  Collection │
└─────────────┘ └──────────┘ └─────────────┘
     │                │              │
     └────────────────┴──────────────┘
                    │
            ┌───────┴────────┐
            │    Redis     │
            │    缓存      │
            └───────────────┘
```

## 📋 MySQL 表清单（17张表）

### 1️⃣ 用户模块（2张表）
- `users` - 用户信息表
- `user_preferences` - 用户偏好设置

### 2️⃣ 知识库模块（2张表）
- `knowledge_bases` - 知识库表
- `kb_settings` - 知识库设置

### 3️⃣ 文档模块（2张表）
- `documents` - 文档表
- `chunks` - 文档分块表

### 4️⃣ 对话模块（3张表）
- `chats` - 对话表
- `messages` - 消息表
- `message_feedback` - 消息反馈表

### 5️⃣ Agent工具模块（2张表）
- `tools` - 工具表
- `tool_executions` - 工具执行记录

### 6️⃣ 检索模块（1张表）
- `search_history` - 搜索历史

### 7️⃣ 系统模块（2张表）
- `api_keys` - API密钥管理
- `audit_logs` - 审计日志

## 🕸️ Neo4j 图数据库设计

### 节点类型（5种）
```
Document   - 文档节点
Chunk     - 文档分块
Entity    - 实体节点（人物/组织/地点/概念）
Concept   - 概念节点
User      - 用户节点
```

### 关系类型（8种）
```
HAS_CHUNK      - 文档→分块
MENTIONS       - 分块→实体
RELATES_TO     - 实体→实体（带权重）
IS_A           - 实体→概念
BELONGS_TO     - 实体→文档
UPLOADED_BY    - 文档→用户
OWNS           - 用户→文档
INTERACTED_WITH - 用户→实体
```

## 📊 Milvus 向量数据库设计

### Collection: document_chunks

```
字段：
├── id (Int64, PrimaryKey)
├── embedding (FloatVector, Dim: 1536)
├── chunk_id (VarChar)
├── doc_id (VarChar)
└── kb_id (VarChar)

索引：IVF_FLAT, Metric: L2
```

## 💡 数据关系核心流程

### 文档上传流程
```
用户上传 → documents 表
解析分块 → chunks 表
向量化 → Milvus
提取实体 → Neo4j
更新状态 → documents.status = 'completed'
```

### 智能问答流程
```
用户提问 → 创建 chat
记录消息 → messages 表
检索查询 → search_history
调用工具 → tool_executions
返回结果 → 更新 messages
```

## 🔧 使用指南

### 初始化数据库
```bash
# Windows
mysql -u root -p < migrations/init.sql

# Linux/Mac
mysql -u root -p < migrations/init.sql
```

### 连接数据库
```go
// 连接字符串
user:password@tcp(localhost:3306)/weknora?charset=utf8mb4&parseTime=True
```

### 数据模型使用
```go
import "link/internal/models"

// 创建知识库
kb := &models.KnowledgeBase{
    Name:        "技术文档",
    Description: "项目技术文档库",
    UserID:      userID,
}
db.Create(&kb)
```

## 📈 统计查询示例

```sql
-- 各模块数据统计
SELECT
    (SELECT COUNT(*) FROM users) as user_count,
    (SELECT COUNT(*) FROM knowledge_bases) as kb_count,
    (SELECT COUNT(*) FROM documents) as doc_count,
    (SELECT COUNT(*) FROM chats) as chat_count;

-- 用户活跃度排行
SELECT
    u.username,
    COUNT(DISTINCT c.id) as chat_count,
    COUNT(DISTINCT d.id) as doc_count
FROM users u
LEFT JOIN chats c ON c.user_id = u.id
LEFT JOIN knowledge_bases kb ON kb.user_id = u.id
LEFT JOIN documents d ON d.kb_id = kb.id
GROUP BY u.id
ORDER BY chat_count DESC;

-- 存储空间使用
SELECT
    file_type,
    COUNT(*) as count,
    SUM(file_size) / 1024 / 1024 / 1024 as size_gb
FROM documents
WHERE status = 'completed'
GROUP BY file_type
ORDER BY size_gb DESC;
```

## 📚 相关文档

- **完整设计**: [DATABASE_DESIGN.md](DATABASE_DESIGN.md)
- **项目文档**: [PROJECT_START.md](PROJECT_START.md)
- **开发指南**: [README_NEO4J.md](README_NEO4J.md)

---

**文档版本**: v1.0.0
**最后更新**: 2026-02-03
