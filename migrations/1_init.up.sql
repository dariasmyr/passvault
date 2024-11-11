-- EncryptionKey Table
CREATE TABLE IF NOT EXISTS encryption_key
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    account_id   BIGINT NOT NULL,
    key_part     TEXT NOT NULL
    );

-- Entry Table
CREATE TABLE IF NOT EXISTS entry
(
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    account_id   BIGINT NOT NULL,
    entry_type   TEXT NOT NULL,
    entry_data   TEXT NOT NULL
    );

