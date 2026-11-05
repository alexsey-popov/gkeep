-- 000005_create_files_table.down.sql
-- Создание таблицы с файлами
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    file VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем комментарии к столбцам
COMMENT ON COLUMN files.id IS 'Идентификатор записи';
COMMENT ON COLUMN files.user_id IS 'Идентификатор пользователя';
COMMENT ON COLUMN files.name IS 'Название';
COMMENT ON COLUMN files.file IS 'Файл';
COMMENT ON COLUMN files.created_at IS 'Момент создания записи';

-- Индекс уникальности названия для конкретного пользователя
CREATE UNIQUE INDEX idx_files_name_user_id_unique ON texts(user_id, name);

-- Индекс для идентификатора пользователя
CREATE INDEX idx_files_user_id ON files(user_id);

-- Связь поля с идентификатором пользователя с таблицей пользователей + удаление всех записей при удалении пользователя
ALTER TABLE files
    ADD CONSTRAINT fk_files_users
        FOREIGN KEY (user_id)
            REFERENCES users(id);