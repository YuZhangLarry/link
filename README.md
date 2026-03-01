# Link

基于知识图谱和 RAG（检索增强生成）的智能知识管理系统，集成了向量检索、图检索、联网搜索等功能，提供全方位的知识管理和智能问答服务。

## 项目简介

Link 是一个企业级知识管理平台，通过知识图谱技术构建实体之间的关系网络，结合大语言模型的能力，实现智能问答、知识推理和深度研究等功能。

### 核心特性

- **知识图谱管理**：基于 Neo4j 构建实体关系网络，支持图谱可视化
- **智能检索**：混合检索（向量 + 全文 + 图谱），支持重排序优化
- **RAG 对话**：基于知识库的增强对话，流式响应
- **多代理系统**：Planner、Retriever、Analyzer、Synthesizer、Critic 协作完成复杂任务
- **多租户架构**：支持多组织独立使用，细粒度权限控制
- **模型评估**：内置模型质量评估系统

## 技术栈

### 后端
- **语言**：Go 1.25.6
- **框架**：Gin（Web 服务）、GORM（ORM）
- **AI 框架**：CloudWeGo Eino
- **数据库**：
  - MySQL 8.0+（元数据存储）
  - Neo4j 5.15+（知识图谱）
  - Milvus 2.3+（向量检索）
  - Redis 7.0+（缓存）

### 前端
- **框架**：Vue 3 + TypeScript
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **图谱可视化**：vis-network
- **构建工具**：Vite

## 项目结构

```
link/
├── cmd/                    # 主程序入口
├── internal/               # 核心业务逻辑
│   ├── agent/             # 多代理系统
│   ├── application/       # 应用服务层
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP 处理器
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   ├── router/            # 路由定义
│   └── types/             # 类型定义
├── web/                   # 前端代码
│   └── src/
│       ├── views/         # 页面组件
│       ├── components/    # 公共组件
│       ├── router/        # 路由配置
│       └── stores/        # 状态管理
├── config/                # 配置文件
├── migrations/            # 数据库迁移
└── uploads/               # 文件上传目录
```

## 快速开始

### 环境要求

- Go 1.25.6+
- Node.js 18+
- MySQL 8.0+
- Neo4j 5.15+
- Milvus 2.3+（可选）
- Redis 7.0+

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/link.git
cd link
```

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：
```env
# 数据库
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=link_go

# Neo4j
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=your_neo4j_password

# JWT
JWT_SECRET=your-secret-key

# AI 模型
CHAT_PROVIDER=openai
CHAT_API_KEY=your-api-key
CHAT_BASE_URL=https://api.openai.com/v1
CHAT_MODEL_NAME=gpt-3.5-turbo

# Embedding
EMBEDDING_PROVIDER=dashscope
EMBEDDING_API_KEY=your-dashscope-key
```

### 3. 启动后端

```bash
# 安装依赖
go mod download

# 运行数据库迁移
go run cmd/server/main.go migrate

# 启动服务
go run cmd/server/main.go
```

服务将在 `http://localhost:8080` 启动。

### 4. 启动前端

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动。

## 核心功能

### 1. 知识库管理

- 创建和管理知识库
- 上传文档（PDF、DOCX、TXT、MD）
- 自动分块和向量化
- 实体关系提取

### 2. 智能对话

- 基于 RAG 的知识问答
- 流式响应（SSE）
- 会话历史管理
- 多轮对话支持

### 3. 知识图谱

- 实体节点管理
- 关系类型定义
- 图谱可视化
- Cypher 查询支持

### 4. 智能代理（Agent）

多代理协作系统：
- **Planner**：研究规划
- **Retriever**：信息检索（rag_query、web_search）
- **Analyzer**：深度分析
- **Synthesizer**：报告合成
- **Critic**：质量评审

### 5. 模型评估

- 数据集管理
- 评估指标收集
- 性能监控

## API 文档

### 认证相关

```
POST /api/v1/auth/register       # 用户注册
POST /api/v1/auth/login          # 用户登录
POST /api/v1/auth/refresh        # 刷新 Token
```

### 知识库

```
GET  /api/v1/knowledge-bases     # 获取知识库列表
POST /api/v1/knowledge-bases     # 创建知识库
GET  /api/v1/knowledge-bases/:id # 获取知识库详情
PUT  /api/v1/knowledge-bases/:id # 更新知识库
DELETE /api/v1/knowledge-bases/:id # 删除知识库
```

### 聊天

```
POST /api/v1/chat                # 发送消息
POST /api/v1/chat/stream        # 流式对话
GET  /api/v1/chat/sessions      # 获取会话列表
```

### 代理

```
POST /api/v1/agent/deep-research # 深度研究
GET  /api/v1/agent/status       # 代理状态
```

更多 API 详情请参考 [API 文档](./docs/api.md)。

## 配置说明

### 检索配置

可在系统设置中调整检索策略：
- BM25 权重：关键词匹配权重
- 向量权重：语义相似度权重
- 图谱权重：关系推理权重
- Top-K：返回结果数量
- Rerank：是否启用重排序

### 模型配置

支持多种模型提供商：
- OpenAI
- 阿里云 DashScope
- 自定义兼容 OpenAI API 的服务

## 部署

### Docker 部署

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d
```

### 生产环境配置

1. 修改 `.env` 中的敏感配置
2. 设置 `GIN_MODE=release`
3. 配置反向代理（Nginx）
4. 启用 HTTPS

## 开发指南

### 代码规范

- Go：遵循 [Effective Go](https://go.dev/doc/effective_go) 规范
- Vue：使用 Composition API 和 TypeScript
- 提交信息：遵循 [Conventional Commits](https://www.conventionalcommits.org/)

### 测试

```bash
# 后端测试
go test ./...

# 前端测试
cd web && npm run test
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 联系方式

- 项目主页：[GitHub](https://github.com/yourusername/link)
- 问题反馈：[Issues](https://github.com/yourusername/link/issues)
