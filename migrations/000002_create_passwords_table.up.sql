-- 000002_create_passwords_table.up.sql

-- Создание таблицы с паролями
CREATE TABLE passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    login VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Добавляем комментарии к столбцам
COMMENT ON COLUMN passwords.id IS 'Идентификатор записи';
COMMENT ON COLUMN passwords.user_id IS 'Идентификатор пользователя';
COMMENT ON COLUMN passwords.name IS 'Название';
COMMENT ON COLUMN passwords.login IS 'Логин';
COMMENT ON COLUMN passwords.password IS 'Пароль';
COMMENT ON COLUMN passwords.created_at IS 'Момент создания записи';

-- Индекс уникальности названия для конкретного пользователя
CREATE UNIQUE INDEX idx_passwords_name_user_id_unique ON passwords(user_id, name);

-- Индекс для идентификатора пользователя
CREATE INDEX idx_passwords_user_id ON passwords(user_id);

-- Связь поля с идентификатором пользователя с таблицей пользователей + удаление всех записей при удалении пользователя
ALTER TABLE passwords
    ADD CONSTRAINT fk_passwords_users
        FOREIGN KEY (user_id)
            REFERENCES users(id);