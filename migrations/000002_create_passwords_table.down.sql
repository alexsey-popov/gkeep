-- 000002_create_passwords_table.down.sql
-- Откат создания таблицы пользователей
ALTER TABLE passwords DROP CONSTRAINT fk_passwords_users;
DROP INDEX IF EXISTS idx_passwords_user_id;
DROP INDEX IF EXISTS idx_passwords_name_user_id_unique;

DROP TABLE IF EXISTS passwords;