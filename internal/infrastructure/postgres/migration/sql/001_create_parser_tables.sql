CREATE TABLE IF NOT EXISTS vk_accounts (
    id BIGSERIAL PRIMARY KEY,
    social_id BIGINT NOT NULL,
    account_type TEXT NOT NULL,
    full_name TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL,
    followers_count INTEGER NOT NULL DEFAULT 0 CHECK (followers_count >= 0),
    avatar_url TEXT NOT NULL DEFAULT '',
    private BOOLEAN NOT NULL DEFAULT FALSE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_type, social_id)
);

CREATE INDEX IF NOT EXISTS idx_vk_accounts_username ON vk_accounts (username);

CREATE TABLE IF NOT EXISTS parsing_orders (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    username TEXT NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    error_text TEXT NOT NULL DEFAULT '',
    account_id BIGINT REFERENCES vk_accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id),
    UNIQUE (order_id)
);

CREATE INDEX IF NOT EXISTS idx_parsing_orders_status ON parsing_orders (status);

CREATE TABLE IF NOT EXISTS processed_kafka_messages (
    message_id TEXT PRIMARY KEY,
    topic TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
