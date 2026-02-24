-- +goose Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- Users
CREATE TABLE users (
                       id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
                       username      VARCHAR(64)  NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       role          VARCHAR(16)  NOT NULL CHECK (role IN ('admin', 'manager', 'viewer')),
                       created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
                       updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Items
CREATE TABLE items (
                       id         UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
                       name       VARCHAR(255)   NOT NULL,
                       sku        VARCHAR(64)    NOT NULL UNIQUE,
                       quantity   INT            NOT NULL DEFAULT 0 CHECK (quantity >= 0),
                       price      NUMERIC(12,2)  NOT NULL DEFAULT 0 CHECK (price >= 0),
                       location   VARCHAR(128),
                       created_at TIMESTAMPTZ    NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ    NOT NULL DEFAULT now()
);

-- Auto-update updated_at on every UPDATE
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP TRIGGER IF EXISTS trg_items_updated_at ON items;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";