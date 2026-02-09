/*
 Navicat Premium Data Transfer

 Source Server         : root
 Source Server Type    : MySQL
 Source Server Version : 80033 (8.0.33)
 Source Host           : localhost:3306
 Source Schema         : link_go

 Target Server Type    : MySQL
 Target Server Version : 80033 (8.0.33)
 File Encoding         : 65001

 Date: 09/02/2026 04:31:31
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for api_keys
-- ----------------------------
DROP TABLE IF EXISTS `api_keys`;
CREATE TABLE `api_keys`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '密钥ID',
  `user_id` bigint NOT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密钥名称',
  `key_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密钥哈希',
  `key_prefix` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密钥前缀(用于显示)',
  `scopes` json NULL COMMENT '权限范围',
  `last_used_at` timestamp NULL DEFAULT NULL COMMENT '最后使用时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `status` tinyint NULL DEFAULT 1 COMMENT '状态: 0=禁用, 1=启用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_key_hash`(`key_hash` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = 'API密钥管理' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of api_keys
-- ----------------------------

-- ----------------------------
-- Table structure for audit_logs
-- ----------------------------
DROP TABLE IF EXISTS `audit_logs`;
CREATE TABLE `audit_logs`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `tenant_id` bigint NULL DEFAULT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `user_id` bigint NULL DEFAULT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `action` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '资源类型: tenant/user/kb/document/chat',
  `resource_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '资源ID',
  `details` json NULL COMMENT '详细信息',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT 'User-Agent',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_action`(`action` ASC) USING BTREE,
  INDEX `idx_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_created_at`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '审计日志' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of audit_logs
-- ----------------------------

-- ----------------------------
-- Table structure for chunks
-- ----------------------------
DROP TABLE IF EXISTS `chunks`;
CREATE TABLE `chunks`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '分块ID (UUID)',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `kb_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识库ID [逻辑外键 -> knowledge_bases.id]',
  `knowledge_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识条目ID [逻辑外键 -> knowledges.id]',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '内容',
  `chunk_index` int NOT NULL COMMENT '分块序号',
  `is_enabled` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  `start_at` int NOT NULL COMMENT '起始位置',
  `end_at` int NOT NULL COMMENT '结束位置',
  `pre_chunk_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '前置分块ID',
  `next_chunk_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '后置分块ID',
  `chunk_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'text' COMMENT '类型: text/image/table',
  `parent_chunk_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '父分块ID',
  `image_info` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '图片信息',
  `relation_chunks` json NULL COMMENT '相关分块',
  `indirect_relation_chunks` json NULL COMMENT '间接相关分块',
  `embedding_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '向量ID(Milvus)',
  `token_count` int NULL DEFAULT NULL COMMENT 'Token数量',
  `metadata` json NULL COMMENT '元数据',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_kb`(`tenant_id` ASC, `kb_id` ASC) USING BTREE,
  INDEX `idx_knowledge_id`(`knowledge_id` ASC) USING BTREE,
  INDEX `idx_parent_id`(`parent_chunk_id` ASC) USING BTREE,
  INDEX `idx_chunk_type`(`chunk_type` ASC) USING BTREE,
  INDEX `idx_embedding_id`(`embedding_id` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE,
  FULLTEXT INDEX `ft_content`(`content`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '分块表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of chunks
-- ----------------------------

-- ----------------------------
-- Table structure for kb_settings
-- ----------------------------
DROP TABLE IF EXISTS `kb_settings`;
CREATE TABLE `kb_settings`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '设置ID',
  `kb_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识库ID [逻辑外键 -> knowledge_bases.id]',
  `retrieval_mode` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'hybrid' COMMENT '检索模式: vector/bm25/hybrid/graph',
  `similarity_threshold` decimal(3, 2) NULL DEFAULT 0.70 COMMENT '相似度阈值',
  `top_k` int NULL DEFAULT 10 COMMENT '返回结果数量',
  `rerank_enabled` tinyint(1) NULL DEFAULT 0 COMMENT '是否启用重排序',
  `graph_enabled` tinyint(1) NULL DEFAULT 0 COMMENT '是否启用图谱检索',
  `settings_json` json NULL COMMENT '其他设置(JSON格式)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_kb_id`(`kb_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '知识库设置' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of kb_settings
-- ----------------------------

-- ----------------------------
-- Table structure for knowledge_bases
-- ----------------------------
DROP TABLE IF EXISTS `knowledge_bases`;
CREATE TABLE `knowledge_bases`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识库ID (UUID)',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `user_id` bigint NOT NULL COMMENT '创建用户ID [逻辑外键 -> users.id]',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识库名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '描述',
  `avatar` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '图标/封面',
  `embedding_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '向量模型ID [逻辑外键 -> models.id]',
  `chunking_config` json NOT NULL COMMENT '分块配置 {\"chunk_size\": 1000, \"chunk_overlap\": 200}',
  `image_processing_config` json NOT NULL COMMENT '图片处理配置',
  `summary_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '摘要模型ID [逻辑外键 -> models.id]',
  `rerank_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '重排序模型ID [逻辑外键 -> models.id]',
  `cos_config` json NOT NULL COMMENT 'COS相似度配置',
  `vlm_config` json NOT NULL COMMENT 'VLM多模态配置',
  `extract_config` json NULL COMMENT '抽取配置',
  `status` tinyint NULL DEFAULT 1 COMMENT '状态: 0=禁用, 1=启用',
  `is_public` tinyint(1) NULL DEFAULT 0 COMMENT '是否公开',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_tenant_name`(`tenant_id` ASC, `name` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `idx_embedding_model`(`embedding_model_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '知识库表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of knowledge_bases
-- ----------------------------

-- ----------------------------
-- Table structure for knowledges
-- ----------------------------
DROP TABLE IF EXISTS `knowledges`;
CREATE TABLE `knowledges`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识ID (UUID)',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `kb_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '知识库ID [逻辑外键 -> knowledge_bases.id]',
  `user_id` bigint NOT NULL COMMENT '创建用户ID [逻辑外键 -> users.id]',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '类型: document/file/url',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '描述',
  `source` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '来源: upload/crawler/api',
  `parse_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'unprocessed' COMMENT '解析状态',
  `enable_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'enabled' COMMENT '启用状态',
  `embedding_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '向量模型ID [逻辑外键 -> models.id]',
  `file_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '文件名',
  `file_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '文件类型',
  `file_size` bigint NULL DEFAULT NULL COMMENT '文件大小',
  `file_path` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '文件路径',
  `file_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '文件哈希',
  `storage_size` bigint NOT NULL DEFAULT 0 COMMENT '存储大小',
  `metadata` json NULL COMMENT '元数据',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  `processed_at` timestamp NULL DEFAULT NULL COMMENT '处理完成时间',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '错误信息',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_kb`(`tenant_id` ASC, `kb_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_kb_id`(`kb_id` ASC) USING BTREE,
  INDEX `idx_status`(`parse_status` ASC, `enable_status` ASC) USING BTREE,
  INDEX `idx_source`(`source` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `idx_file_hash`(`file_hash` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '知识条目表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of knowledges
-- ----------------------------

-- ----------------------------
-- Table structure for message_feedback
-- ----------------------------
DROP TABLE IF EXISTS `message_feedback`;
CREATE TABLE `message_feedback`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '反馈ID',
  `message_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息ID [逻辑外键 -> messages.id]',
  `user_id` bigint NOT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `rating` int NULL DEFAULT NULL COMMENT '评分: 1-5星',
  `comment` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '评论',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_message_id`(`message_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  CONSTRAINT `message_feedback_chk_1` CHECK (`rating` in (1,2,3,4,5))
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '消息反馈表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of message_feedback
-- ----------------------------

-- ----------------------------
-- Table structure for messages
-- ----------------------------
DROP TABLE IF EXISTS `messages`;
CREATE TABLE `messages`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息ID (UUID)',
  `request_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '请求ID (UUID)',
  `session_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话ID [逻辑外键 -> sessions.id]',
  `role` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色: system/user/assistant/tool',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '消息内容',
  `knowledge_references` json NULL COMMENT '知识引用',
  `agent_steps` json NULL COMMENT 'Agent执行步骤',
  `tool_calls` json NULL COMMENT '工具调用记录',
  `is_completed` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否完成',
  `token_count` int NULL DEFAULT NULL COMMENT 'Token使用量',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_session_id`(`session_id` ASC) USING BTREE,
  INDEX `idx_request_id`(`request_id` ASC) USING BTREE,
  INDEX `idx_role`(`role` ASC) USING BTREE,
  INDEX `idx_created_at`(`created_at` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '消息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of messages
-- ----------------------------
INSERT INTO `messages` VALUES ('018d1d89-67ca-41c1-a776-df225beef3a5', 'c5572032-e35d-4bfe-b066-cc35f32ccb8b', '1b5deb47-8fbe-485d-b132-b1bdad97bb96', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:09:50', '2026-02-09 04:09:50', NULL);
INSERT INTO `messages` VALUES ('01d6afd8-b901-475c-ab5f-9fd859da472c', '0ebf400a-f3fe-4dfc-b3b7-fe3f73b52e75', '6faa0137-6988-464a-8411-6a8a2c576451', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:07:49', '2026-02-09 04:07:49', NULL);
INSERT INTO `messages` VALUES ('1b7fff29-bd05-4f20-9b45-3ca3bd8b73a5', 'de251754-57c3-449a-b1ce-4a790b8b35be', 'd5fce0b7-f0e7-488c-96c0-84ed297b7bc0', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:10:37', '2026-02-09 04:10:37', NULL);
INSERT INTO `messages` VALUES ('532b2e61-2ca1-480e-881e-b97310b30132', '709e7206-d4ec-4b0f-8dc4-ffc2c6b33f66', 'e605d570-afad-44e9-97bd-1c88f2438101', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:06:17', '2026-02-09 04:06:17', NULL);
INSERT INTO `messages` VALUES ('9793840c-534b-4a3d-ae34-5b4d5b5a2e1f', '69326409-9dbe-4222-b3e7-1df3e6503a2e', '8c37e3e0-2a9e-4e31-9a1b-d54a05ca99db', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:08:38', '2026-02-09 04:08:38', NULL);
INSERT INTO `messages` VALUES ('c6c3613a-9498-4f6c-82cf-ff747adde292', '80a3688a-c88e-4b95-955c-cda19be4d9fc', 'e4d4162b-911f-4aea-b9db-e6d43362b42d', 'user', '这是一条测试消息', '{}', '{}', '{}', 0, 0, '2026-02-09 04:12:42', '2026-02-09 04:12:42', NULL);

-- ----------------------------
-- Table structure for models
-- ----------------------------
DROP TABLE IF EXISTS `models`;
CREATE TABLE `models`  (
  `id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型ID (UUID)',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型名称',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型类型: embedding/chat/rerank/vlm/summary',
  `source` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模型来源: openai/azure/dashscope/custom',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '模型描述',
  `parameters` json NOT NULL COMMENT '模型参数配置 {\"model\": \"xxx\", \"dim\": 1536}',
  `is_default` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为默认模型',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_tenant_source_type`(`tenant_id` ASC, `source` ASC, `type` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '模型表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of models
-- ----------------------------
INSERT INTO `models` VALUES ('model-chat-250bb9247a682a42', 1, 'qwen-turbo', 'chat', 'dashscope', NULL, '{\"model\": \"qwen-turbo\", \"temperature\": 0.7}', 1, 'active', '2026-02-09 01:39:46', '2026-02-09 01:39:46', NULL);
INSERT INTO `models` VALUES ('model-embed-b3cb1a1b02660b0c', 1, 'text-embedding-v4', 'embedding', 'dashscope', NULL, '{\"dim\": 1536, \"model\": \"text-embedding-v4\"}', 1, 'active', '2026-02-09 01:39:46', '2026-02-09 01:39:46', NULL);
INSERT INTO `models` VALUES ('model-rerank-248b27b7016572ce', 1, 'gte-rerank-v2', 'rerank', 'dashscope', NULL, '{\"model\": \"gte-rerank-v2\"}', 1, 'active', '2026-02-09 01:39:46', '2026-02-09 01:39:46', NULL);

-- ----------------------------
-- Table structure for permission_audit_logs
-- ----------------------------
DROP TABLE IF EXISTS `permission_audit_logs`;
CREATE TABLE `permission_audit_logs`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `tenant_id` bigint NULL DEFAULT NULL COMMENT '租户ID',
  `user_id` bigint NULL DEFAULT NULL COMMENT '目标用户ID',
  `operator_id` bigint NOT NULL COMMENT '操作人ID',
  `operation_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `target_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标类型: role/resource',
  `target_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标ID',
  `before_value` json NULL COMMENT '变更前值',
  `after_value` json NULL COMMENT '变更后值',
  `reason` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '变更原因',
  `ip_address` varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT 'IP地址',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_operator_id`(`operator_id` ASC) USING BTREE,
  INDEX `idx_operation_type`(`operation_type` ASC) USING BTREE,
  INDEX `idx_created_at`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '权限变更审计日志' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of permission_audit_logs
-- ----------------------------

-- ----------------------------
-- Table structure for permissions
-- ----------------------------
DROP TABLE IF EXISTS `permissions`;
CREATE TABLE `permissions`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `resource_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型: kb/session/document/user/role/tenant',
  `action` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作: create/read/update/delete/assign',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '权限描述',
  `is_system` tinyint(1) NULL DEFAULT 0 COMMENT '是否为系统预设权限',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_resource_action`(`resource_type` ASC, `action` ASC) USING BTREE,
  INDEX `idx_resource_type`(`resource_type` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 25 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '权限表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of permissions
-- ----------------------------
INSERT INTO `permissions` VALUES (1, 'kb', 'create', '创建知识库', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (2, 'kb', 'read', '查看知识库', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (3, 'kb', 'update', '更新知识库', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (4, 'kb', 'delete', '删除知识库', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (5, 'document', 'create', '上传文档', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (6, 'document', 'read', '查看文档', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (7, 'document', 'update', '更新文档', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (8, 'document', 'delete', '删除文档', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (9, 'session', 'create', '创建会话', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (10, 'session', 'read', '查看会话', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (11, 'session', 'update', '更新会话', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (12, 'session', 'delete', '删除会话', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (13, 'user', 'create', '创建用户', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (14, 'user', 'read', '查看用户', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (15, 'user', 'update', '更新用户', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (16, 'user', 'delete', '删除用户', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (17, 'user', 'assign_role', '分配角色', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (18, 'role', 'create', '创建角色', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (19, 'role', 'read', '查看角色', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (20, 'role', 'update', '更新角色', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (21, 'role', 'delete', '删除角色', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (22, 'role', 'assign_permission', '分配权限', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (23, 'tenant', 'update', '更新租户设置', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');
INSERT INTO `permissions` VALUES (24, 'tenant', 'delete', '删除租户', 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20');

-- ----------------------------
-- Table structure for refresh_tokens
-- ----------------------------
DROP TABLE IF EXISTS `refresh_tokens`;
CREATE TABLE `refresh_tokens`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '刷新Token ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `token_hash` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Token哈希值',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_token_hash`(`token_hash` ASC) USING BTREE,
  INDEX `idx_expires_at`(`expires_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 22 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '刷新Token表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of refresh_tokens
-- ----------------------------
INSERT INTO `refresh_tokens` VALUES (1, 6, 'f9f986f3bc7b706d8c752f96cb89b1ebe9e4fcf73ee46802a07ae57827160782', '2026-02-16 03:59:56', '2026-02-09 03:59:56');
INSERT INTO `refresh_tokens` VALUES (2, 6, '53c5f1c0fc425bea0da74560671febaf9a27f450b5de05a4fb1835374e6347fe', '2026-02-16 04:01:22', '2026-02-09 04:01:21');
INSERT INTO `refresh_tokens` VALUES (3, 7, 'ee5beba60b5fc60683f2a037e0009d597f7b5771d021821cc45c138bd4be395c', '2026-02-16 04:01:22', '2026-02-09 04:01:21');
INSERT INTO `refresh_tokens` VALUES (4, 6, 'a0cc6d8df9b98c52ed051cce26965726e79cbf607c517cb7889ff5e83a291091', '2026-02-16 04:02:27', '2026-02-09 04:02:26');
INSERT INTO `refresh_tokens` VALUES (5, 7, '06f7ab0cbc37c9a116effc5afb485840cccd79cce29f6685143b66b95d67ba2f', '2026-02-16 04:02:27', '2026-02-09 04:02:26');
INSERT INTO `refresh_tokens` VALUES (6, 6, '5ca371f3ca64497014b8077c67894086eabc07afafa84123df9edb0d7b3c62cf', '2026-02-16 04:02:45', '2026-02-09 04:02:44');
INSERT INTO `refresh_tokens` VALUES (7, 7, '35835a88bd3eaf9dec5648106f5223ded478ada690e9c9b859bbaa39a7fd808d', '2026-02-16 04:02:45', '2026-02-09 04:02:44');
INSERT INTO `refresh_tokens` VALUES (8, 6, '2e396c1c30b3df1032e68258a947fc2e6c29f0dc6624c540bbb55290d95b2838', '2026-02-16 04:03:20', '2026-02-09 04:03:19');
INSERT INTO `refresh_tokens` VALUES (9, 7, 'aa350f8b5f0ffac353f2b8067e536a69f7e415ab91864420b589a0f149c3d0e5', '2026-02-16 04:03:20', '2026-02-09 04:03:19');
INSERT INTO `refresh_tokens` VALUES (10, 6, 'c51c10838e83ff34afd68a6419a6498828b562eefa8ebbcee0d62f2b41f8ea26', '2026-02-16 04:06:17', '2026-02-09 04:06:16');
INSERT INTO `refresh_tokens` VALUES (11, 7, '537f8fdcecb0ffeed2533eb8b3d65f285ce2e136542bef7511de1b50c94119f1', '2026-02-16 04:06:17', '2026-02-09 04:06:16');
INSERT INTO `refresh_tokens` VALUES (12, 6, 'f4a0faabf07127c3d41f5958c491dfa560113012f3e9be386518bf49343ebffa', '2026-02-16 04:07:49', '2026-02-09 04:07:49');
INSERT INTO `refresh_tokens` VALUES (13, 7, '9be40819d9ca699543204a461e458ae398bf514b15981dd866c898b821b5cd27', '2026-02-16 04:07:49', '2026-02-09 04:07:49');
INSERT INTO `refresh_tokens` VALUES (14, 6, '68648008b119122dc0aead2b8ba2f9f86406248495a404da1f89e251a75a9620', '2026-02-16 04:08:38', '2026-02-09 04:08:38');
INSERT INTO `refresh_tokens` VALUES (15, 7, '3f5db69dad074f302322a772895ef40ec9840dce3011bb94ace042a2a23238e9', '2026-02-16 04:08:38', '2026-02-09 04:08:38');
INSERT INTO `refresh_tokens` VALUES (16, 6, 'add9a9fe8eca95dce31753ff0e2449340591165d469d4fa59a8157e714563fe8', '2026-02-16 04:09:50', '2026-02-09 04:09:49');
INSERT INTO `refresh_tokens` VALUES (17, 7, 'a7b79a2f0a9c5c2d00a14fc0f7e6a22361049f9129897e719f462885e82fc387', '2026-02-16 04:09:50', '2026-02-09 04:09:49');
INSERT INTO `refresh_tokens` VALUES (18, 6, '897ae9e05fbb4933a7d2c4ba15904ff1b62cb774adeaa0abab79bd4ca5beb09b', '2026-02-16 04:10:37', '2026-02-09 04:10:37');
INSERT INTO `refresh_tokens` VALUES (19, 7, '6e8239b716c55a6c67dfd568b18757a330e8786a69c574df86a350f6ecd1c7e3', '2026-02-16 04:10:38', '2026-02-09 04:10:37');
INSERT INTO `refresh_tokens` VALUES (20, 6, '6528e377fb61b88f687a28dc6808b5fafe1623c0f1a3baece4c0c246f312fee1', '2026-02-16 04:12:42', '2026-02-09 04:12:41');
INSERT INTO `refresh_tokens` VALUES (21, 7, 'd818ae5dc70975aba363cc79e2f7446829404c8faf14ef1f21b5353f94b30cd5', '2026-02-16 04:12:42', '2026-02-09 04:12:41');

-- ----------------------------
-- Table structure for resource_permissions
-- ----------------------------
DROP TABLE IF EXISTS `resource_permissions`;
CREATE TABLE `resource_permissions`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `resource_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型: kb/session/document',
  `resource_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源ID',
  `permission_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '权限类型: read/write/delete/admin',
  `granted_by` bigint NULL DEFAULT NULL COMMENT '授权人ID',
  `granted_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_resource`(`tenant_id` ASC, `user_id` ASC, `resource_type` ASC, `resource_id` ASC, `permission_type` ASC) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_resource`(`resource_type` ASC, `resource_id` ASC) USING BTREE,
  INDEX `idx_expires_at`(`expires_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '资源级权限表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of resource_permissions
-- ----------------------------

-- ----------------------------
-- Table structure for role_permissions
-- ----------------------------
DROP TABLE IF EXISTS `role_permissions`;
CREATE TABLE `role_permissions`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_role_permission`(`role_id` ASC, `permission_id` ASC) USING BTREE,
  INDEX `idx_role_id`(`role_id` ASC) USING BTREE,
  INDEX `idx_permission_id`(`permission_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 286 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色权限关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of role_permissions
-- ----------------------------
INSERT INTO `role_permissions` VALUES (1, 3, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (2, 2, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (3, 1, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (4, 3, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (5, 2, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (6, 1, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (7, 3, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (8, 2, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (9, 1, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (10, 3, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (11, 2, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (12, 1, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (13, 3, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (14, 2, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (15, 1, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (16, 3, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (17, 2, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (18, 1, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (19, 3, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (20, 2, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (21, 1, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (22, 3, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (23, 2, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (24, 1, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (25, 3, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (26, 2, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (27, 1, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (28, 3, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (29, 2, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (30, 1, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (31, 3, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (32, 2, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (33, 1, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (34, 3, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (35, 2, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (36, 1, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (37, 3, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (38, 2, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (39, 1, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (40, 3, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (41, 2, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (42, 1, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (43, 3, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (44, 2, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (45, 1, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (46, 3, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (47, 2, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (48, 1, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (49, 3, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (50, 2, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (51, 1, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (52, 3, 23, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (53, 2, 23, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (54, 1, 23, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (55, 3, 24, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (56, 2, 24, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (57, 1, 24, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (58, 3, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (59, 2, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (60, 1, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (61, 3, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (62, 2, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (63, 1, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (64, 3, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (65, 2, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (66, 1, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (67, 3, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (68, 2, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (69, 1, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (70, 3, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (71, 2, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (72, 1, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (128, 6, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (129, 5, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (130, 4, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (131, 6, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (132, 5, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (133, 4, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (134, 6, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (135, 5, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (136, 4, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (137, 6, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (138, 5, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (139, 4, 8, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (140, 6, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (141, 5, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (142, 4, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (143, 6, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (144, 5, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (145, 4, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (146, 6, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (147, 5, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (148, 4, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (149, 6, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (150, 5, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (151, 4, 4, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (152, 6, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (153, 5, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (154, 4, 18, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (155, 6, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (156, 5, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (157, 4, 19, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (158, 6, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (159, 5, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (160, 4, 20, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (161, 6, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (162, 5, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (163, 4, 21, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (164, 6, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (165, 5, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (166, 4, 22, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (167, 6, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (168, 5, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (169, 4, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (170, 6, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (171, 5, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (172, 4, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (173, 6, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (174, 5, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (175, 4, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (176, 6, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (177, 5, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (178, 4, 12, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (179, 6, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (180, 5, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (181, 4, 13, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (182, 6, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (183, 5, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (184, 4, 14, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (185, 6, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (186, 5, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (187, 4, 15, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (188, 6, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (189, 5, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (190, 4, 16, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (191, 6, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (192, 5, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (193, 4, 17, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (255, 9, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (256, 8, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (257, 7, 5, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (258, 9, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (259, 8, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (260, 7, 6, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (261, 9, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (262, 8, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (263, 7, 7, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (264, 9, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (265, 8, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (266, 7, 1, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (267, 9, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (268, 8, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (269, 7, 2, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (270, 9, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (271, 8, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (272, 7, 3, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (273, 9, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (274, 8, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (275, 7, 9, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (276, 9, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (277, 8, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (278, 7, 10, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (279, 9, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (280, 8, 11, '2026-02-09 04:27:20');
INSERT INTO `role_permissions` VALUES (281, 7, 11, '2026-02-09 04:27:20');

-- ----------------------------
-- Table structure for roles
-- ----------------------------
DROP TABLE IF EXISTS `roles`;
CREATE TABLE `roles`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色编码',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '角色描述',
  `is_system` tinyint(1) NULL DEFAULT 0 COMMENT '是否为系统预设角色',
  `is_default` tinyint(1) NULL DEFAULT 0 COMMENT '是否为新用户默认角色',
  `level` int NULL DEFAULT 0 COMMENT '角色层级（数字越大权限越高）',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'active' COMMENT '状态: active/inactive',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_tenant_code`(`tenant_id` ASC, `code` ASC) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '角色表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of roles
-- ----------------------------
INSERT INTO `roles` VALUES (1, 1, '所有者', 'owner', '租户所有者，拥有所有权限', 1, 0, 100, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (2, 2, '所有者', 'owner', '租户所有者，拥有所有权限', 1, 0, 100, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (3, 3, '所有者', 'owner', '租户所有者，拥有所有权限', 1, 0, 100, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (4, 1, '管理员', 'admin', '管理员，可以管理资源和用户', 1, 0, 80, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (5, 2, '管理员', 'admin', '管理员，可以管理资源和用户', 1, 0, 80, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (6, 3, '管理员', 'admin', '管理员，可以管理资源和用户', 1, 0, 80, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (7, 1, '普通用户', 'user', '普通用户，基本使用权限', 1, 1, 50, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (8, 2, '普通用户', 'user', '普通用户，基本使用权限', 1, 1, 50, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);
INSERT INTO `roles` VALUES (9, 3, '普通用户', 'user', '普通用户，基本使用权限', 1, 1, 50, 'active', '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL);

-- ----------------------------
-- Table structure for search_history
-- ----------------------------
DROP TABLE IF EXISTS `search_history`;
CREATE TABLE `search_history`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '搜索ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `user_id` bigint NOT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `kb_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '知识库ID [逻辑外键 -> knowledge_bases.id]',
  `query` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '查询内容',
  `retrieval_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '检索类型: vector/bm25/hybrid/graph',
  `result_count` int NULL DEFAULT NULL COMMENT '结果数量',
  `latency_ms` int NULL DEFAULT NULL COMMENT '耗时(毫秒)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '搜索时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_kb_id`(`kb_id` ASC) USING BTREE,
  INDEX `idx_created_at`(`created_at` ASC) USING BTREE,
  FULLTEXT INDEX `ft_query`(`query`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '搜索历史' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of search_history
-- ----------------------------

-- ----------------------------
-- Table structure for sessions
-- ----------------------------
DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions`  (
  `id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '会话ID (UUID)',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `user_id` bigint NOT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '会话标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '会话描述',
  `kb_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '关联知识库ID [逻辑外键 -> knowledge_bases.id]',
  `max_rounds` int NOT NULL DEFAULT 5 COMMENT '最大轮次',
  `enable_rewrite` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用改写',
  `fallback_strategy` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'fixed' COMMENT '降级策略',
  `fallback_response` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '很抱歉，我暂时无法回答这个问题。' COMMENT '降级回复',
  `keyword_threshold` float NOT NULL DEFAULT 0.5 COMMENT '关键词阈值',
  `vector_threshold` float NOT NULL DEFAULT 0.5 COMMENT '向量阈值',
  `rerank_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '重排序模型 [逻辑外键 -> models.id]',
  `embedding_top_k` int NOT NULL DEFAULT 10 COMMENT '向量TopK',
  `rerank_top_k` int NOT NULL DEFAULT 10 COMMENT '重排序TopK',
  `rerank_threshold` float NOT NULL DEFAULT 0.65 COMMENT '重排序阈值',
  `summary_model_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '摘要模型 [逻辑外键 -> models.id]',
  `summary_parameters` json NOT NULL COMMENT '摘要参数',
  `agent_config` json NULL COMMENT '会话级Agent配置',
  `context_config` json NULL COMMENT '上下文配置',
  `status` tinyint NULL DEFAULT 1 COMMENT '状态: 0=归档, 1=正常',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_kb_id`(`kb_id` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `idx_updated_at`(`updated_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '会话表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of sessions
-- ----------------------------
INSERT INTO `sessions` VALUES ('1b5deb47-8fbe-485d-b132-b1bdad97bb96', 0, 6, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:09:50', '2026-02-09 04:09:50', NULL);
INSERT INTO `sessions` VALUES ('6faa0137-6988-464a-8411-6a8a2c576451', 0, 1, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:07:49', '2026-02-09 04:07:49', NULL);
INSERT INTO `sessions` VALUES ('8c37e3e0-2a9e-4e31-9a1b-d54a05ca99db', 0, 6, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:08:38', '2026-02-09 04:08:38', NULL);
INSERT INTO `sessions` VALUES ('a4a4aaec-1fd7-48dd-bb01-eacf907e7145', 0, 1, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:03:20', '2026-02-09 04:03:20', NULL);
INSERT INTO `sessions` VALUES ('d5fce0b7-f0e7-488c-96c0-84ed297b7bc0', 0, 6, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:10:37', '2026-02-09 04:10:37', NULL);
INSERT INTO `sessions` VALUES ('e4d4162b-911f-4aea-b9db-e6d43362b42d', 0, 6, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:12:42', '2026-02-09 04:12:42', NULL);
INSERT INTO `sessions` VALUES ('e605d570-afad-44e9-97bd-1c88f2438101', 0, 1, '用户A的新会话', '', '', 5, 1, 'fixed', '很抱歉，我暂时无法回答这个问题。', 0.5, 0.5, '', 10, 10, 0.65, '', '{}', '{}', '{}', 1, '2026-02-09 04:06:17', '2026-02-09 04:06:17', NULL);

-- ----------------------------
-- Table structure for system_config
-- ----------------------------
DROP TABLE IF EXISTS `system_config`;
CREATE TABLE `system_config`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `config_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置键',
  `config_value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '配置值',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '描述',
  `is_public` tinyint(1) NULL DEFAULT 0 COMMENT '是否公开',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `config_key`(`config_key` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 8 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '系统配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of system_config
-- ----------------------------
INSERT INTO `system_config` VALUES (1, 'max_file_size', '104857600', '最大文件上传大小(字节) 默认100MB', 0, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (2, 'allowed_file_types', '[\"pdf\",\"docx\",\"txt\",\"md\",\"csv\",\"json\"]', '允许的文件类型', 1, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (3, 'max_chunk_size', '2000', '最大分块大小', 0, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (4, 'max_chunks_per_file', '10000', '单个文件最大分块数', 0, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (5, 'vector_dimension', '1536', '向量维度', 1, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (6, 'enable_multi_modal', 'true', '是否启用多模态功能', 1, '2026-02-09 01:39:46', '2026-02-09 01:39:46');
INSERT INTO `system_config` VALUES (7, 'system_version', '2.0.0', '系统版本', 1, '2026-02-09 01:39:46', '2026-02-09 01:39:46');

-- ----------------------------
-- Table structure for tenants
-- ----------------------------
DROP TABLE IF EXISTS `tenants`;
CREATE TABLE `tenants`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '租户ID',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '租户名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '租户描述',
  `api_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'API密钥',
  `retriever_engines` json NOT NULL COMMENT '检索引擎配置 {\"vector\": \"milvus\", \"graph\": \"neo4j\", \"bm25\": \"redis\"}',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'active' COMMENT '状态: active/suspended/deleted',
  `business` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '业务类型',
  `storage_quota` bigint NOT NULL DEFAULT 10737418240 COMMENT '存储配额(字节) 默认10GB',
  `storage_used` bigint NOT NULL DEFAULT 0 COMMENT '已使用存储(字节)',
  `agent_config` json NULL COMMENT '租户级Agent配置',
  `settings` json NULL COMMENT '租户配置 {embedding_model, rerank_model, summary_model}',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_business`(`business` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `idx_api_key`(`api_key` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '租户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tenants
-- ----------------------------
INSERT INTO `tenants` VALUES (1, '默认租户', '系统默认租户', 'sk-default-8cfd2944455e88ad', '{\"bm25\": \"redis\", \"graph\": \"neo4j\", \"vector\": \"milvus\"}', 'active', 'enterprise', 107374182400, 0, NULL, '{\"rerank_model\": \"gte-rerank-v2\", \"summary_model\": \"qwen-turbo\", \"embedding_model\": \"text-embedding-v4\"}', '2026-02-09 01:39:46', '2026-02-09 01:39:46', NULL);
INSERT INTO `tenants` VALUES (2, '测试租户A', '这是一个测试租户，用于开发测试', 'sk-test-a-e5e70a38bbaca1ee', '{\"graph\": \"neo4j\", \"vector\": \"milvus\"}', 'active', 'technology', 10737418240, 0, NULL, '{\"embedding_model\": \"text-embedding-v4\"}', '2026-02-09 03:50:54', '2026-02-09 03:50:54', NULL);
INSERT INTO `tenants` VALUES (3, '演示租户B', '用于演示的租户', 'sk-demo-b-fab0846126c62b57', '{\"vector\": \"milvus\"}', 'active', 'education', 21474836480, 0, NULL, '{\"embedding_model\": \"text-embedding-v4\"}', '2026-02-09 03:50:54', '2026-02-09 03:50:54', NULL);

-- ----------------------------
-- Table structure for tool_executions
-- ----------------------------
DROP TABLE IF EXISTS `tool_executions`;
CREATE TABLE `tool_executions`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '执行ID',
  `message_id` varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联消息ID [逻辑外键 -> messages.id]',
  `tool_id` bigint NOT NULL COMMENT '工具ID [逻辑外键 -> tools.id]',
  `input_params` json NULL COMMENT '输入参数',
  `output_data` json NULL COMMENT '输出数据',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '执行状态',
  `duration_ms` int NULL DEFAULT NULL COMMENT '执行时长(毫秒)',
  `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '错误信息',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '执行时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_message_id`(`message_id` ASC) USING BTREE,
  INDEX `idx_tool_id`(`tool_id` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_created_at`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '工具执行记录' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tool_executions
-- ----------------------------

-- ----------------------------
-- Table structure for tools
-- ----------------------------
DROP TABLE IF EXISTS `tools`;
CREATE TABLE `tools`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '工具ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID [逻辑外键 -> tenants.id]',
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工具名称',
  `type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工具类型: search/database/http/custom',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL COMMENT '描述',
  `config` json NOT NULL COMMENT '配置',
  `enabled` tinyint(1) NULL DEFAULT 1 COMMENT '是否启用',
  `created_by` bigint NULL DEFAULT NULL COMMENT '创建者 [逻辑外键 -> users.id]',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_type`(`type` ASC) USING BTREE,
  INDEX `idx_enabled`(`enabled` ASC) USING BTREE,
  INDEX `idx_created_by`(`created_by` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '工具表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tools
-- ----------------------------

-- ----------------------------
-- Table structure for user_preferences
-- ----------------------------
DROP TABLE IF EXISTS `user_preferences`;
CREATE TABLE `user_preferences`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '偏好ID',
  `user_id` bigint NOT NULL COMMENT '用户ID [逻辑外键 -> users.id]',
  `language` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'zh-CN' COMMENT '语言',
  `theme` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT 'light' COMMENT '主题: light/dark',
  `notification_enabled` tinyint(1) NULL DEFAULT 1 COMMENT '是否启用通知',
  `preference_json` json NULL COMMENT '其他偏好设置(JSON格式)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_id`(`user_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户偏好设置' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_preferences
-- ----------------------------

-- ----------------------------
-- Table structure for user_roles
-- ----------------------------
DROP TABLE IF EXISTS `user_roles`;
CREATE TABLE `user_roles`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `assigned_by` bigint NULL DEFAULT NULL COMMENT '分配人ID',
  `assigned_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分配时间',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_tenant_user`(`tenant_id` ASC, `user_id` ASC) USING BTREE,
  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
  INDEX `idx_role_id`(`role_id` ASC) USING BTREE,
  INDEX `idx_assigned_by`(`assigned_by` ASC) USING BTREE,
  INDEX `idx_expires_at`(`expires_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户角色关联表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user_roles
-- ----------------------------
INSERT INTO `user_roles` VALUES (1, 1, 1, 1, NULL, '2026-02-09 04:27:20', NULL);

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `tenant_id` bigint NOT NULL COMMENT '租户ID',
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
  `email` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱',
  `password_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码哈希',
  `avatar` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL COMMENT '头像URL',
  `status` tinyint NULL DEFAULT 1 COMMENT '状态: 0=禁用, 1=正常',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `last_login_at` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `deleted_at` timestamp NULL DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_tenant_username`(`tenant_id` ASC, `username` ASC) USING BTREE,
  UNIQUE INDEX `uk_tenant_email`(`tenant_id` ASC, `email` ASC) USING BTREE,
  INDEX `idx_tenant_id`(`tenant_id` ASC) USING BTREE,
  INDEX `idx_username`(`username` ASC) USING BTREE,
  INDEX `idx_email`(`email` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE,
  INDEX `idx_deleted_at`(`deleted_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO `users` VALUES (1, 1, 'admin', 'admin@link.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NULL, 1, '2026-02-09 04:27:20', '2026-02-09 04:27:20', NULL, NULL);

-- ----------------------------
-- View structure for v_kb_stats
-- ----------------------------
DROP VIEW IF EXISTS `v_kb_stats`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `v_kb_stats` AS select `kb`.`id` AS `kb_id`,`kb`.`name` AS `kb_name`,`kb`.`tenant_id` AS `tenant_id`,`kb`.`user_id` AS `user_id`,`u`.`username` AS `creator`,count(distinct `k`.`id`) AS `knowledge_count`,count(distinct `ch`.`id`) AS `chunk_count`,sum(`k`.`storage_size`) AS `total_storage`,`kb`.`status` AS `status`,`kb`.`created_at` AS `created_at` from (((`knowledge_bases` `kb` left join `users` `u` on((`kb`.`user_id` = `u`.`id`))) left join `knowledges` `k` on(((`kb`.`id` = `k`.`kb_id`) and (`k`.`deleted_at` is null)))) left join `chunks` `ch` on(((`k`.`id` = `ch`.`knowledge_id`) and (`ch`.`deleted_at` is null)))) where (`kb`.`deleted_at` is null) group by `kb`.`id`;

-- ----------------------------
-- View structure for v_tenant_stats
-- ----------------------------
DROP VIEW IF EXISTS `v_tenant_stats`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `v_tenant_stats` AS select `t`.`id` AS `tenant_id`,`t`.`name` AS `tenant_name`,`t`.`status` AS `status`,`t`.`storage_quota` AS `storage_quota`,`t`.`storage_used` AS `storage_used`,round(((`t`.`storage_used` / `t`.`storage_quota`) * 100),2) AS `storage_usage_percent`,count(distinct `tu`.`user_id`) AS `user_count`,count(distinct `kb`.`id`) AS `kb_count`,count(distinct `k`.`id`) AS `knowledge_count`,count(distinct `s`.`id`) AS `session_count` from ((((`tenants` `t` left join `tenant_users` `tu` on((`t`.`id` = `tu`.`tenant_id`))) left join `knowledge_bases` `kb` on(((`t`.`id` = `kb`.`tenant_id`) and (`kb`.`deleted_at` is null)))) left join `knowledges` `k` on(((`t`.`id` = `k`.`tenant_id`) and (`k`.`deleted_at` is null)))) left join `sessions` `s` on(((`t`.`id` = `s`.`tenant_id`) and (`s`.`deleted_at` is null)))) where (`t`.`deleted_at` is null) group by `t`.`id`;

-- ----------------------------
-- View structure for v_user_permissions
-- ----------------------------
DROP VIEW IF EXISTS `v_user_permissions`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `v_user_permissions` AS select distinct `u`.`id` AS `user_id`,`u`.`tenant_id` AS `tenant_id`,`u`.`username` AS `username`,`p`.`resource_type` AS `resource_type`,`p`.`action` AS `action`,`r`.`code` AS `role_code`,`r`.`level` AS `role_level` from ((((`users` `u` join `user_roles` `ur` on(((`u`.`id` = `ur`.`user_id`) and (`u`.`tenant_id` = `ur`.`tenant_id`)))) join `role_permissions` `rp` on((`ur`.`role_id` = `rp`.`role_id`))) join `permissions` `p` on((`rp`.`permission_id` = `p`.`id`))) join `roles` `r` on((`ur`.`role_id` = `r`.`id`))) where ((`u`.`deleted_at` is null) and ((`ur`.`expires_at` is null) or (`ur`.`expires_at` > now())));

-- ----------------------------
-- View structure for v_user_roles
-- ----------------------------
DROP VIEW IF EXISTS `v_user_roles`;
CREATE ALGORITHM = UNDEFINED SQL SECURITY DEFINER VIEW `v_user_roles` AS select `u`.`id` AS `user_id`,`u`.`tenant_id` AS `tenant_id`,`u`.`username` AS `username`,`u`.`email` AS `email`,`u`.`status` AS `status`,`r`.`id` AS `role_id`,`r`.`name` AS `role_name`,`r`.`code` AS `role_code`,`r`.`level` AS `role_level`,`ur`.`assigned_at` AS `assigned_at`,`ur`.`expires_at` AS `expires_at` from ((`users` `u` left join `user_roles` `ur` on(((`u`.`id` = `ur`.`user_id`) and (`u`.`tenant_id` = `ur`.`tenant_id`)))) left join `roles` `r` on((`ur`.`role_id` = `r`.`id`))) where (`u`.`deleted_at` is null);

SET FOREIGN_KEY_CHECKS = 1;
