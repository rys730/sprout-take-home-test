-- ============================================================================
-- ACCOUNTS (Chart of Accounts)
-- ============================================================================

-- name: GetAllAccounts :many
SELECT id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at
FROM accounts
WHERE is_active = true
ORDER BY code ASC;

-- name: GetAllAccountsWithBalances :many
SELECT
    a.id, a.code, a.name, a.type, a.parent_id, a.level,
    a.is_system, a.is_control, a.is_active, a.created_by,
    a.created_at, a.updated_at,
    COALESCE(
        CASE
            WHEN a.type IN ('asset', 'expense')
                THEN SUM(COALESCE(jel.debit, 0)) - SUM(COALESCE(jel.credit, 0))
            ELSE
                SUM(COALESCE(jel.credit, 0)) - SUM(COALESCE(jel.debit, 0))
        END,
    0)::NUMERIC(18,2) AS balance
FROM accounts a
LEFT JOIN journal_entry_lines jel ON jel.account_id = a.id
LEFT JOIN journal_entries je ON je.id = jel.journal_entry_id AND je.status = 'posted'
WHERE a.is_active = true
GROUP BY a.id, a.code, a.name, a.type, a.parent_id, a.level,
    a.is_system, a.is_control, a.is_active, a.created_by,
    a.created_at, a.updated_at
ORDER BY a.code ASC;

-- name: GetAccountByID :one
SELECT id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at
FROM accounts
WHERE id = $1;

-- name: GetAccountByCode :one
SELECT id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at
FROM accounts
WHERE code = $1;

-- name: GetAccountChildren :many
SELECT id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at
FROM accounts
WHERE parent_id = $1 AND is_active = true
ORDER BY code ASC;

-- name: HasAccountChildren :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE parent_id = $1 AND is_active = true) AS has_children;

-- name: IsAccountReferencedInJournalLines :one
SELECT EXISTS(SELECT 1 FROM journal_entry_lines WHERE account_id = $1) AS is_referenced;

-- name: CreateAccount :one
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control, is_active, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at;

-- name: UpdateAccount :one
UPDATE accounts
SET code = $2, name = $3, parent_id = $4, level = $5, type = $6, updated_at = NOW()
WHERE id = $1
RETURNING id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;

-- name: SearchAccounts :many
SELECT id, code, name, type, parent_id, level, is_system, is_control, is_active, created_by, created_at, updated_at
FROM accounts
WHERE is_active = true
  AND (
    sqlc.narg('search')::TEXT IS NULL
    OR code ILIKE '%' || sqlc.narg('search')::TEXT || '%'
    OR name ILIKE '%' || sqlc.narg('search')::TEXT || '%'
  )
  AND (
    sqlc.narg('account_type')::account_type IS NULL
    OR type = sqlc.narg('account_type')::account_type
  )
  AND (
    sqlc.narg('parent_id')::UUID IS NULL
    OR parent_id = sqlc.narg('parent_id')::UUID
  )
ORDER BY code ASC;

-- name: GetAccountIDByCode :one
SELECT id FROM accounts WHERE code = $1 LIMIT 1;
