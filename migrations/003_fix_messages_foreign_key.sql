-- 修改 messages 表，删除外键约束（因为我们使用 sessions 表而不是 chats 表）

USE link_go;

-- 删除指向 chats 表的外键约束
ALTER TABLE messages DROP FOREIGN KEY messages_ibfk_1;

-- 验证外键已删除
SELECT
    TABLE_NAME,
    CONSTRAINT_NAME,
    REFERENCED_TABLE_NAME
FROM
    information_schema.KEY_COLUMN_USAGE
WHERE
    TABLE_SCHEMA = 'link_go'
    AND TABLE_NAME = 'messages';
