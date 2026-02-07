-- 简化版 sessions 表（移除外键约束，便于测试）

DROP TABLE IF EXISTS sessions;

CREATE TABLE sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '会话ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    kb_id BIGINT DEFAULT NULL COMMENT '关联知识库ID',
    title VARCHAR(255) NOT NULL COMMENT '会话标题',
    description TEXT COMMENT '会话描述',
    model VARCHAR(50) DEFAULT 'gpt-3.5-turbo' COMMENT '使用的模型',
    max_rounds INT DEFAULT 50 COMMENT '最大轮次',
    enable_rewrite BOOLEAN DEFAULT TRUE COMMENT '是否启用改写',
    keyword_threshold FLOAT DEFAULT 0.5 COMMENT '关键词阈值',
    vector_threshold FLOAT DEFAULT 0.5 COMMENT '向量阈值',
    message_count INT DEFAULT 0 COMMENT '消息数量',
    status TINYINT DEFAULT 1 COMMENT '状态: 0=归档, 1=正常',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间',

    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- 插入测试数据
INSERT INTO sessions (user_id, title, description, model, message_count, status)
VALUES
(1, '测试会话1', '这是一个测试会话', 'gpt-3.5-turbo', 2, 1),
(1, '欢迎对话', '欢迎来到AI对话系统', 'gpt-3.5-turbo', 5, 1);
