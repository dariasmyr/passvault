-- EncryptionKey Table
CREATE TABLE IF NOT EXISTS encryption_key
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      BIGINT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    key_part     TEXT NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_user_id ON encryption_key (user_id);

-- User Table
CREATE TABLE IF NOT EXISTS user
(
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sso_account_id BIGINT NOT NULL UNIQUE,
    email          TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_sso_account_id ON user (sso_account_id);
CREATE INDEX IF NOT EXISTS idx_email ON user (email);

-- Vault Table
CREATE TABLE IF NOT EXISTS vault
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id      BIGINT NOT NULL REFERENCES user(id) ON DELETE CASCADE,
    entry_type   TEXT NOT NULL,
    entry_data   TEXT NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_user_id ON vault (user_id);
