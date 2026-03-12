-- ============================================================================
-- JOURNAL ENTRY LINES (Detail Jurnal)
-- ============================================================================

-- name: GetJournalEntryLinesByEntryID :many
SELECT jel.id, jel.journal_entry_id, jel.account_id,
       jel.debit, jel.credit, jel.line_order, jel.created_at,
       a.code AS account_code, a.name AS account_name
FROM journal_entry_lines jel
JOIN accounts a ON a.id = jel.account_id
WHERE jel.journal_entry_id = $1
ORDER BY jel.line_order ASC;

-- name: CreateJournalEntryLine :one
INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, line_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, journal_entry_id, account_id, debit, credit, line_order, created_at;

-- name: DeleteJournalEntryLinesByEntryID :exec
DELETE FROM journal_entry_lines WHERE journal_entry_id = $1;

-- name: GetJournalEntryIDsByAccountID :many
SELECT DISTINCT journal_entry_id FROM journal_entry_lines WHERE account_id = $1;

-- name: DeleteJournalEntryLinesByAccountID :exec
DELETE FROM journal_entry_lines WHERE account_id = $1;

-- name: GetJournalEntryLinesTotals :one
SELECT COALESCE(SUM(debit), 0)::NUMERIC(18,2) AS total_debit,
       COALESCE(SUM(credit), 0)::NUMERIC(18,2) AS total_credit
FROM journal_entry_lines
WHERE journal_entry_id = $1;

-- name: GetAccountLedger :many
SELECT jel.id, jel.journal_entry_id, jel.account_id,
       jel.debit, jel.credit, jel.line_order, jel.created_at,
       je.entry_number, je.date AS entry_date, je.status AS entry_status
FROM journal_entry_lines jel
JOIN journal_entries je ON je.id = jel.journal_entry_id
WHERE jel.account_id = $1
  AND je.status = 'posted'
  AND (sqlc.narg('start_date')::DATE IS NULL OR je.date >= sqlc.narg('start_date')::DATE)
  AND (sqlc.narg('end_date')::DATE IS NULL OR je.date <= sqlc.narg('end_date')::DATE)
ORDER BY je.date ASC, jel.line_order ASC;
