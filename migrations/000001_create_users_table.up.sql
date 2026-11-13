-- migrations/000001_create_users_table.up.sql
-- Создание таблицы с пользователями
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);

-- Добавляем комментарии к столбцам
COMMENT ON COLUMN users.id IS 'Идентификатор пользователя';
COMMENT ON COLUMN users.login IS 'Логин';
COMMENT ON COLUMN users.password IS 'Пароль';

-- Индекс уникальности для префикса
CREATE UNIQUE INDEX idx_users_login_unique ON users(login);