-- Создание таблицы черного списка
CREATE TABLE IF NOT EXISTS user_blacklists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    blocked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Ограничения
    CONSTRAINT check_different_users_blacklist CHECK (user_id != blocked_user_id),
    CONSTRAINT unique_blacklist_entry UNIQUE(user_id, blocked_user_id)
);

-- Индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_user_blacklists_user_id ON user_blacklists(user_id);
CREATE INDEX IF NOT EXISTS idx_user_blacklists_blocked_user_id ON user_blacklists(blocked_user_id);
CREATE INDEX IF NOT EXISTS idx_user_blacklists_blocked_at ON user_blacklists(blocked_at);
