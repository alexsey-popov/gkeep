-- 000005_create_files_table.down.sql
-- Откат создания таблицы с файлами
ALTER TABLE files DROP CONSTRAINT fk_files_users;
DROP INDEX IF EXISTS idx_files_user_id;
DROP INDEX IF EXISTS idx_files_name_user_id_unique;

DROP TABLE IF EXISTS files;