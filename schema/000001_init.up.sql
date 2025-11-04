CREATE TABLE users (
    id serial PRIMARY KEY,  -- Добавлен PRIMARY KEY
    name varchar(255) not null,
    mail varchar(255) not null unique,
    password varchar(255), -- Может быть NULL для OAuth пользователей
    avatar_url text,
    github_id int unique, -- Уникальный ID из GitHub
    github_login varchar(100),
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp,
    last_login_at timestamp with time zone,
    deleted_at timestamp with time zone
);

-- Индексы для быстрого поиска
CREATE INDEX idx_users_email ON users(mail);
CREATE INDEX idx_users_github_id ON users(github_id);

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

-- Коллекции пользователя
CREATE TABLE collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id int NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('books', 'anime', 'series', 'movies')),
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    cover_image TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Элементы коллекции - все хранятся в одном месте
-- Могут быть публичные(предоставленные приложением) и приватными(создаными пользователем)
CREATE TABLE collection_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) CHECK (type IN ('books', 'anime', 'series', 'movies')),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image TEXT,
    is_public BOOLEAN DEFAULT false,
    is_custom BOOLEAN DEFAULT false,
    -- Если custom элемент, то проставляется создатель и этот элемент будет показываться в ленте только создателю 
    creator_id int DEFAULT NULL, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE collections_items_assignment (
    id serial PRIMARY KEY,
    collection_id UUID not null REFERENCES collections(id),
    item_id UUID not null REFERENCES collection_items(id),
    added_at TIMESTAMP default NOW(),
    user_review TEXT, -- заметки для элемента в этой коллекции

    -- Уникальное ограничение: один элемент не может быть дважды в одной коллекции
    UNIQUE(collection_id, item_id)
);

-- Индексы для производительности
CREATE INDEX idx_assignment_collection_id ON collections_items_assignment(collection_id);
CREATE INDEX idx_assignment_item_id ON collections_items_assignment(item_id);