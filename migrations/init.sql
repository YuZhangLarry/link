

-- 用户表
CREATE TABLE IF NOT EXISTS users (
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

-- 用户偏好设置
CREATE TABLE IF NOT EXISTS user_preferences (
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

-- ============================================
-- 2. 知识库模块
-- ============================================

-- 知识库表
CREATE TABLE IF NOT EXISTS knowledge_bases (
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

-- 知识库设置
CREATE TABLE IF NOT EXISTS kb_settings (
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

-- ============================================
-- 3. 文档模块
-- ============================================

-- 文档表
CREATE TABLE IF NOT EXISTS documents (
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

-- 文档分块表
CREATE TABLE IF NOT EXISTS chunks (
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

-- ============================================
-- 4. 对话模块
-- ============================================

-- 对话表
CREATE TABLE IF NOT EXISTS chats (
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

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
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

-- 消息反馈表
CREATE TABLE IF NOT EXISTS message_feedback (
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

-- ============================================
-- 5. Agent & 工具模块
-- ============================================

-- 工具表
CREATE TABLE IF NOT EXISTS tools (
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

-- 工具执行记录
CREATE TABLE IF NOT EXISTS tool_executions (
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

-- ============================================
-- 6. 检索模块
-- ============================================

-- 搜索历史
CREATE TABLE IF NOT EXISTS search_history (
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

-- ============================================
-- 7. 系统模块
-- ============================================

-- API密钥管理
CREATE TABLE IF NOT EXISTS api_keys (
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

-- 审计日志
CREATE TABLE IF NOT EXISTS audit_logs (
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

-- ============================================
-- 初始化数据
-- ============================================

-- 创建默认管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, status) VALUES
('admin', 'admin@weknora.com', '$2a$10$YourHashedPasswordHere', 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
