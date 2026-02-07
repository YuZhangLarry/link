-- 完整删除所有指向 chats 表的外键约束

USE link_go;

-- 1. 查看所有相关的外键约束
SELECT
    CONSTRAINT_NAME,
    TABLE_NAME,
    REFERENCED_TABLE_NAME
FROM
    information_schema.KEY_COLUMN_USAGE
WHERE
    TABLE_SCHEMA = 'link_go'
    AND TABLE_NAME = 'messages'
    AND REFERENCED_TABLE_NAME = 'chats';

-- 2. 删除所有指向 chats 表的外键（无论名字是什么）
ALTER TABLE messages DROP FOREIGN KEY IF EXISTS messages_ibfk_1;

-- 3. 再次验证（应该返回 0 行）
SELECT
    CONSTRAINT_NAME,
    TABLE_NAME,
    REFERENCED_TABLE_NAME
FROM
    information_schema.KEY_COLUMN_USAGE
WHERE
    TABLE_SCHEMA = 'link_go'
    AND TABLE_NAME = 'messages'
    AND REFERENCED_TABLE_NAME = 'chats';

-- 4. 查看当前 messages 表的所有外键（应该没有了，或指向 sessions）
SELECT
    CONSTRAINT_NAME,
    TABLE_NAME,
    REFERENCED_TABLE_NAME
FROM
    information_schema.KEY_COLUMN_USAGE
WHERE
    TABLE_SCHEMA = 'link_go'
    AND TABLE_NAME = 'messages';
