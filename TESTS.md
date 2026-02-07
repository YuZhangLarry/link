# 测试目录说明

本文档列出了项目中所有的测试程序，按照标准命名规范：`test-技术栈-功能`

## 📋 测试目录列表

### 1. Neo4j 图数据库测试
```bash
# 插入测试数据
go run cmd/test-neo4j-insert/main.go

# 查询所有数据
go run cmd/test-neo4j-query/main.go

# 查找特定用户
go run cmd/test-neo4j-find/main.go

# 查询用户的朋友
go run cmd/test-neo4j-friends/main.go

# 查询共同好友
go run cmd/test-neo4j-common/main.go

# 显示数据库统计
go run cmd/test-neo4j-stats/main.go

# 删除所有数据
go run cmd/test-neo4j-delete/main.go
```

### 2. MySQL 数据库测试
```bash
# 测试 MySQL 连接
go run cmd/test-mysql-connect/main.go
```

### 3. Milvus 向量数据库测试
```bash
# 测试 Milvus 连接
go run cmd/test-milvus-connect/main.go
```

### 4. OpenAI 聊天测试
```bash
# 测试 OpenAI 对话功能
go run cmd/test-openai-chat/main.go
```

## 🎯 快速开始

### 环境准备
```bash
# 1. 复制环境配置文件
cp .env.example .env

# 2. 编辑 .env 文件，填入你的配置
# - MySQL: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
# - Milvus: MILVUS_HOST, MILVUS_TOKEN
```

### 运行所有测试
```bash
# 1. 测试数据库连接
go run cmd/test-mysql-connect/main.go
go run cmd/test-milvus-connect/main.go

# 2. 测试 Neo4j (需要先启动 Neo4j)
go run cmd/test-neo4j-insert/main.go
go run cmd/test-neo4j-query/main.go

# 3. 测试 OpenAI 聊天
go run cmd/test-openai-chat/main.go
```

## 📦 测试说明

### Neo4j 测试
- **test-neo4j-insert**: 创建测试数据（5个用户，6个关系）
- **test-neo4j-query**: 查询并显示所有节点和关系
- **test-neo4j-find**: 根据用户名查找特定用户
- **test-neo4j-friends**: 查询指定用户的所有朋友
- **test-neo4j-common**: 查询两个用户的共同好友
- **test-neo4j-stats**: 显示数据库统计信息
- **test-neo4j-delete**: 删除所有节点和关系

### MySQL 测试
- **test-mysql-connect**: 测试数据库连接，显示版本和表列表

### Milvus 测试
- **test-milvus-connect**: 测试向量数据库连接，列出所有 Collection

### OpenAI 测试
- **test-openai-chat**: 包含6个测试场景
  1. 简单连接测试
  2. 基础聊天
  3. 带温度参数
  4. 多轮对话
  5. 知识图谱场景
  6. 流式聊天

## 🔧 配置说明

### .env 文件
```bash
# MySQL 配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=link_go

# Milvus 配置
MILVUS_HOST=https://your-milvus-instance.com
MILVUS_TOKEN=your_token_here

# OpenAI 配置 (在测试代码中配置)
# APIKey:  your key
# BaseURL: https://api.gpts.vin/v1
# Model:   gpt-3.5-turbo
```

## 📝 命名规范

所有测试目录遵循：`test-技术栈-功能` 的命名规范

| 技术栈 | 功能 | 目录名 |
|--------|------|--------|
| neo4j | insert | test-neo4j-insert |
| neo4j | query | test-neo4j-query |
| neo4j | find | test-neo4j-find |
| neo4j | friends | test-neo4j-friends |
| neo4j | common | test-neo4j-common |
| neo4j | stats | test-neo4j-stats |
| neo4j | delete | test-neo4j-delete |
| mysql | connect | test-mysql-connect |
| milvus | connect | test-milvus-connect |
| openai | chat | test-openai-chat |

---

**更新时间**: 2026-02-03
**版本**: v1.0.0
