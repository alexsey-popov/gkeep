-- 000004_create_texts_table.down.sql
-- Создание таблицы с текстами
CREATE TABLE texts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем комментарии к столбцам
COMMENT ON COLUMN texts.id IS 'Идентификатор записи';
COMMENT ON COLUMN texts.user_id IS 'Идентификатор пользователя';
COMMENT ON COLUMN texts.name IS 'Название';
COMMENT ON COLUMN texts.text IS 'Текст';
COMMENT ON COLUMN texts.created_at IS 'Момент создания записи';

-- Индекс уникальности названия для конкретного пользователя
CREATE UNIQUE INDEX idx_texts_name_user_id_unique ON texts(user_id, name);

-- Индекс для идентификатора пользователя
CREATE INDEX idx_texts_user_id ON texts(user_id);

-- Связь поля с идентификатором пользователя с таблицей пользователей + удаление всех записей при удалении пользователя
ALTER TABLE texts
    ADD CONSTRAINT fk_texts_users
        FOREIGN KEY (user_id)
            REFERENCES users(id);