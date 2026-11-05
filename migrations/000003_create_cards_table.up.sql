-- 000003_create_cards_table.down.sql
-- Создание таблицы с банковскими картами
CREATE TABLE cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    number VARCHAR(255) NOT NULL,
    expired VARCHAR(255) NOT NULL,
    code VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем комментарии к столбцам
COMMENT ON COLUMN cards.id IS 'Идентификатор записи';
COMMENT ON COLUMN cards.user_id IS 'Идентификатор пользователя';
COMMENT ON COLUMN cards.name IS 'Название';
COMMENT ON COLUMN cards.number IS 'Номер';
COMMENT ON COLUMN cards.expired IS 'Срок';
COMMENT ON COLUMN cards.code IS 'Код';
COMMENT ON COLUMN cards.created_at IS 'Момент создания записи';

-- Индекс уникальности названия для конкретного пользователя
CREATE UNIQUE INDEX idx_cards_name_user_id_unique ON cards(user_id, name);

-- Индекс для идентификатора пользователя
CREATE INDEX idx_cards_user_id ON cards(user_id);

-- Связь поля с идентификатором пользователя с таблицей пользователей + удаление всех записей при удалении пользователя
ALTER TABLE cards
    ADD CONSTRAINT fk_cards_users
        FOREIGN KEY (user_id)
            REFERENCES users(id) ON DELETE CASCADE;