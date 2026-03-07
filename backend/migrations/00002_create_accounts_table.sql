-- +goose Up
-- +goose StatementBegin
CREATE TYPE account_type AS ENUM ('asset', 'liability', 'equity', 'revenue', 'expense');

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,              -- e.g., "111.000", "112.000"
    name VARCHAR(255) NOT NULL,                     -- e.g., "Kas", "Piutang Usaha"
    type account_type NOT NULL,                     -- inherited from parent on creation
    parent_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    level INT NOT NULL DEFAULT 0,                   -- hierarchy depth (parent level + 1)
    is_system BOOLEAN NOT NULL DEFAULT false,       -- core system accounts (e.g., ASET, KEWAJIBAN)
    is_control BOOLEAN NOT NULL DEFAULT false,      -- control accounts (not directly postable)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_accounts_parent_id ON accounts(parent_id);
CREATE INDEX idx_accounts_code ON accounts(code);
CREATE INDEX idx_accounts_type ON accounts(type);
CREATE INDEX idx_accounts_is_active ON accounts(is_active);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_accounts_is_active;
DROP INDEX IF EXISTS idx_accounts_type;
DROP INDEX IF EXISTS idx_accounts_code;
DROP INDEX IF EXISTS idx_accounts_parent_id;

DROP TABLE IF EXISTS accounts;

DROP TYPE IF EXISTS account_type;
-- +goose StatementEnd
