CREATE TABLE users (
    id serial not null unique,
    name varchar(255) not null,
    mail varchar(255) not null unique,
    password varchar(255), -- Может быть NULL для OAuth пользователей
    avatar_url text,
    github_id int unique, -- Уникальный ID из GitHub
    github_login varchar(100),
    auth_provider varchar(20) not null default 'email', -- 'email' или 'github'
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);

-- Индексы для быстрого поиска
CREATE INDEX idx_users_email ON users(mail);
CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = current_timestamp;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();