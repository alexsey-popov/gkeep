-- 000003_create_cards_table.down.sql
-- Откат создания таблицы карт
ALTER TABLE cards DROP CONSTRAINT fk_cards_users;
DROP INDEX IF EXISTS idx_cards_user_id;
DROP INDEX IF EXISTS idx_cards_name_user_id_unique;

DROP TABLE IF EXISTS cards;