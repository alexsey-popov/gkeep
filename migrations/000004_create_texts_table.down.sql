-- 000004_create_texts_table.down.sql
-- Откат создания таблицы с текстами
ALTER TABLE texts DROP CONSTRAINT fk_texts_users;
DROP INDEX IF EXISTS idx_texts_user_id;
DROP INDEX IF EXISTS idx_texts_name_user_id_unique;

DROP TABLE IF EXISTS texts;