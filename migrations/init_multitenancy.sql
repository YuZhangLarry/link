-- ============================================
-- WeKnora 知识图谱 RAG 系统
-- 多租户数据库初始化脚本
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS weknora
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE weknora;

SET FOREIGN_KEY_CHECKS = 0;
SET UNIQUE_CHECKS = 0;
SET AUTOCOMMIT = 0;

-- ============================================
-- 1. 租户模块
-- ============================================

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '租户ID',
    name VARCHAR(255) NOT NULL COMMENT '租户名称',
    description TEXT COMMENT '租户描述',
    api_key VARCHAR(64) NOT NULL COMMENT 'API密钥',
    retriever_engines JSON NOT NULL COMMENT '检索引擎配置',
    status VARCHAR(50) DEFAULT 'active' COMMENT '状态: active/suspended/deleted',
    business VARCHAR(255) NOT NULL COMMENT '业务类型',
    storage_quota BIGINT NOT NULL DEFAULT 10737418240 COMMENT '存储配额(字节)',
    storage_used BIGINT NOT NULL DEFAULT 0 COMMENT '已使用存储(字节)',
    agent_config JSON COMMENT '租户级Agent配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_status (status),
    INDEX idx_business (business)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- 租户用户关联表
CREATE TABLE IF NOT EXISTS tenant_users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '关联ID',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    role VARCHAR(50) NOT NULL DEFAULT 'member' COMMENT '角色: owner/admin/member',
    status VARCHAR(50) DEFAULT 'active' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    UNIQUE KEY uk_tenant_user (tenant_id, user_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role (role),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户用户关联';

-- 用户表（简化版，实际使用时可能需要完整的用户表）
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
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

-- ============================================
-- 2. 模型管理模块
-- ============================================

CREATE TABLE IF NOT EXISTS models (
    id VARCHAR(64) PRIMARY KEY COMMENT '模型ID (UUID)',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    name VARCHAR(255) NOT NULL COMMENT '模型名称',
    type VARCHAR(50) NOT NULL COMMENT '模型类型: embedding/chat/rerank/vlm/summary',
    source VARCHAR(50) NOT NULL COMMENT '模型来源: openai/azure/custom',
    description TEXT COMMENT '模型描述',
    parameters JSON NOT NULL COMMENT '模型参数配置',
    is_default BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为默认模型',
    status VARCHAR(50) NOT NULL DEFAULT 'active' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_models_tenant_source_type (tenant_id, source, type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型表';

-- ============================================
-- 3. 知识库模块
-- ============================================

CREATE TABLE IF NOT EXISTS knowledge_bases (
    id VARCHAR(36) PRIMARY KEY COMMENT '知识库ID (UUID)',
    name VARCHAR(255) NOT NULL COMMENT '知识库名称',
    description TEXT COMMENT '描述',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    chunking_config JSON NOT NULL COMMENT '分块配置',
    image_processing_config JSON NOT NULL COMMENT '图片处理配置',
    embedding_model_id VARCHAR(64) NOT NULL COMMENT '向量模型ID',
    summary_model_id VARCHAR(64) NOT NULL COMMENT '摘要模型ID',
    rerank_model_id VARCHAR(64) NOT NULL COMMENT '重排序模型ID',
    cos_config JSON NOT NULL COMMENT 'COS相似度配置',
    vlm_config JSON NOT NULL COMMENT 'VLM多模态配置',
    extract_config JSON COMMENT '抽取配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_knowledge_bases_tenant_name (tenant_id, name),
    INDEX idx_tenant_id (tenant_id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识库表';

-- ============================================
-- 4. 知识内容模块
-- ============================================

CREATE TABLE IF NOT EXISTS knowledges (
    id VARCHAR(36) PRIMARY KEY COMMENT '知识ID (UUID)',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    knowledge_base_id VARCHAR(36) NOT NULL COMMENT '知识库ID',
    type VARCHAR(50) NOT NULL COMMENT '类型: document/file/url',
    title VARCHAR(255) NOT NULL COMMENT '标题',
    description TEXT COMMENT '描述',
    source VARCHAR(128) NOT NULL COMMENT '来源: upload/crawler/api',
    parse_status VARCHAR(50) NOT NULL DEFAULT 'unprocessed' COMMENT '解析状态',
    enable_status VARCHAR(50) NOT NULL DEFAULT 'enabled' COMMENT '启用状态',
    embedding_model_id VARCHAR(64) COMMENT '向量模型ID',
    file_name VARCHAR(255) COMMENT '文件名',
    file_type VARCHAR(50) COMMENT '文件类型',
    file_size BIGINT COMMENT '文件大小',
    file_path TEXT COMMENT '文件路径',
    file_hash VARCHAR(64) COMMENT '文件哈希',
    storage_size BIGINT NOT NULL DEFAULT 0 COMMENT '存储大小',
    metadata JSON COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    processed_at TIMESTAMP COMMENT '处理完成时间',
    error_message TEXT COMMENT '错误信息',

    INDEX idx_knowledges_tenant_kb (tenant_id, knowledge_base_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (parse_status, enable_status),
    INDEX idx_source (source),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (knowledge_base_id) REFERENCES knowledge_bases(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='知识条目表';

-- ============================================
-- 5. 会话模块
-- ============================================

CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(36) PRIMARY KEY COMMENT '会话ID (UUID)',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    title VARCHAR(255) COMMENT '会话标题',
    description TEXT COMMENT '会话描述',
    knowledge_base_id VARCHAR(36) COMMENT '关联知识库ID',
    max_rounds INT NOT NULL DEFAULT 5 COMMENT '最大轮次',
    enable_rewrite BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用改写',
    fallback_strategy VARCHAR(255) NOT NULL DEFAULT 'fixed' COMMENT '降级策略',
    fallback_response VARCHAR(255) NOT NULL DEFAULT '很抱歉，我暂时无法回答这个问题。' COMMENT '降级回复',
    keyword_threshold FLOAT NOT NULL DEFAULT 0.5 COMMENT '关键词阈值',
    vector_threshold FLOAT NOT NULL DEFAULT 0.5 COMMENT '向量阈值',
    rerank_model_id VARCHAR(64) COMMENT '重排序模型',
    embedding_top_k INT NOT NULL DEFAULT 10 COMMENT '向量TopK',
    rerank_top_k INT NOT NULL DEFAULT 10 COMMENT '重排序TopK',
    rerank_threshold FLOAT NOT NULL DEFAULT 0.65 COMMENT '重排序阈值',
    summary_model_id VARCHAR(64) COMMENT '摘要模型',
    summary_parameters JSON NOT NULL COMMENT '摘要参数',
    agent_config JSON COMMENT '会话级Agent配置',
    context_config JSON COMMENT '上下文配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_sessions_tenant_id (tenant_id),
    INDEX idx_kb_id (knowledge_base_id),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (knowledge_base_id) REFERENCES knowledge_bases(id) ON DELETE SET NULL,
    FOREIGN KEY (rerank_model_id) REFERENCES models(id),
    FOREIGN KEY (summary_model_id) REFERENCES models(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4 COLLATE=utf8_unicode_ci COMMENT='会话表';

-- ============================================
-- 6. 消息模块
-- ============================================

CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(36) PRIMARY KEY COMMENT '消息ID (UUID)',
    request_id VARCHAR(36) NOT NULL COMMENT '请求ID (UUID)',
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    role VARCHAR(50) NOT NULL COMMENT '角色: system/user/assistant/tool',
    content TEXT NOT NULL COMMENT '消息内容',
    knowledge_references JSON COMMENT '知识引用',
    agent_steps JSON COMMENT 'Agent执行步骤',
    is_completed BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否完成',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_messages_session_role (session_id, role),
    INDEX idx_request_id (request_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8_unicode_ci COMMENT='消息表';

-- ============================================
-- 7. 分块模块
-- ============================================

CREATE TABLE IF NOT EXISTS chunks (
    id VARCHAR(36) PRIMARY KEY COMMENT '分块ID (UUID)',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    knowledge_base_id VARCHAR(36) NOT NULL COMMENT '知识库ID',
    knowledge_id VARCHAR(36) NOT NULL COMMENT '知识条目ID',
    content TEXT NOT NULL COMMENT '内容',
    chunk_index INT NOT NULL COMMENT '分块序号',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    start_at INT NOT NULL COMMENT '起始位置',
    end_at INT NOT NULL COMMENT '结束位置',
    pre_chunk_id VARCHAR(36) COMMENT '前置分块ID',
    next_chunk_id VARCHAR(36) COMMENT '后置分块ID',
    chunk_type VARCHAR(20) NOT NULL DEFAULT 'text' COMMENT '类型: text/image/table',
    parent_chunk_id VARCHAR(36) COMMENT '父分块ID',
    image_info TEXT COMMENT '图片信息',
    relation_chunks JSON COMMENT '相关分块',
    indirect_relation_chunks JSON COMMENT '间接相关分块',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',

    INDEX idx_chunks_tenant_kb (tenant_id, knowledge_base_id),
    INDEX idx_knowledge_id (knowledge_id),
    INDEX idx_parent_id (parent_chunk_id),
    INDEX idx_chunk_type (chunk_type),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (knowledge_base_id) REFERENCES knowledge_bases(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分块表';

-- ============================================
-- 8. 工具和审计模块
-- ============================================

CREATE TABLE IF NOT EXISTS tools (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '工具ID',
    tenant_id BIGINT NOT NULL COMMENT '租户ID',
    name VARCHAR(100) NOT NULL COMMENT '工具名称',
    type VARCHAR(50) NOT NULL COMMENT '工具类型',
    description TEXT COMMENT '描述',
    config JSON NOT NULL COMMENT '配置',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_by BIGINT COMMENT '创建者',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_type (type),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=DEFAULT CHARSET=utf8mb4 COLLATE=utf8_unicode_ci COMMENT='工具表';

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    tenant_id BIGINT COMMENT '租户ID',
    user_id BIGINT COMMENT '用户ID',
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) COMMENT '资源类型',
    resource_id BIGINT COMMENT '资源ID',
    details JSON COMMENT '详细信息',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT 'User-Agent',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8_unicode_ci COMMENT='审计日志';

-- ============================================
-- 初始化数据
-- ============================================

-- 创建默认租户
INSERT INTO tenants (name, description, api_key, retriever_engines, business, storage_quota)
VALUES (
    '默认租户',
    '系统默认租户',
    'sk-default-key-placeholder',
    '{"vector": "milvus", "graph": "neo4j", "bm25": "redis"}',
    'enterprise',
    107374182400
) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id);

-- 创建默认模型
INSERT INTO models (id, tenant_id, name, type, source, parameters, is_default)
VALUES
    ('model-uuid-1', LAST_INSERT_ID(), 'text-embedding-ada-002', 'embedding', 'openai', '{"model": "text-embedding-ada-002", "dim": 1536}', TRUE),
    ('model-uuid-2', LAST_INSERT_ID(), 'gpt-3.5-turbo', 'chat', 'openai', '{"model": "gpt-3.5-turbo", "temperature": 0.7}', TRUE),
    ('model-uuid-3', LAST_INSERT_ID(), 'gpt-4', 'chat', 'openai', '{"model": "gpt-4", "temperature": 0.7}', FALSE)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- 创建默认用户
INSERT INTO users (username, email, password_hash, status)
VALUES
    ('admin', 'admin@weknora.com', '$2a$10$placeholder', 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- 关联用户到租户
INSERT INTO tenant_users (tenant_id, user_id, role)
VALUES
    (1, 1, 'owner')
ON DUPLICATE KEY UPDATE role = 'owner';

SET FOREIGN_KEY_CHECKS = 1;
SET UNIQUE_CHECKS = 1;
SET AUTOCOMMIT = 0;
