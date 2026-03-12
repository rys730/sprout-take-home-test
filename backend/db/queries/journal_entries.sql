-- ============================================================================
-- JOURNAL ENTRIES (Jurnal Umum) - Headers
-- ============================================================================

-- name: GetJournalEntryByID :one
SELECT id, entry_number, date, description, status,
       total_debit, total_credit, reversal_of, reversal_reason,
       reversed_by, source, created_by, created_at, updated_at
FROM journal_entries
WHERE id = $1;

-- name: GetJournalEntryByEntryNumber :one
SELECT id, entry_number, date, description, status,
       total_debit, total_credit, reversal_of, reversal_reason,
       reversed_by, source, created_by, created_at, updated_at
FROM journal_entries
WHERE entry_number = $1;

-- name: ListJournalEntries :many
SELECT id, entry_number, date, description, status,
       total_debit, total_credit, reversal_of, reversal_reason,
       reversed_by, source, created_by, created_at, updated_at
FROM journal_entries
WHERE (sqlc.narg('status')::journal_status IS NULL OR status = sqlc.narg('status')::journal_status)
  AND (sqlc.narg('source')::TEXT IS NULL OR source = sqlc.narg('source')::TEXT)
  AND (sqlc.narg('start_date')::DATE IS NULL OR date >= sqlc.narg('start_date')::DATE)
  AND (sqlc.narg('end_date')::DATE IS NULL OR date <= sqlc.narg('end_date')::DATE)
ORDER BY date DESC, entry_number DESC
LIMIT $1 OFFSET $2;

-- name: CountJournalEntries :one
SELECT COUNT(*) FROM journal_entries
WHERE (sqlc.narg('status')::journal_status IS NULL OR status = sqlc.narg('status')::journal_status)
  AND (sqlc.narg('source')::TEXT IS NULL OR source = sqlc.narg('source')::TEXT)
  AND (sqlc.narg('start_date')::DATE IS NULL OR date >= sqlc.narg('start_date')::DATE)
  AND (sqlc.narg('end_date')::DATE IS NULL OR date <= sqlc.narg('end_date')::DATE);

-- name: CreateJournalEntry :one
INSERT INTO journal_entries (entry_number, date, description, source, status, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, entry_number, date, description, status,
          total_debit, total_credit, reversal_of, reversal_reason,
          reversed_by, source, created_by, created_at, updated_at;

-- name: UpdateJournalEntry :one
UPDATE journal_entries
SET description = $2, date = $3, updated_at = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING id, entry_number, date, description, status,
          total_debit, total_credit, reversal_of, reversal_reason,
          reversed_by, source, created_by, created_at, updated_at;

-- name: PostJournalEntry :one
UPDATE journal_entries
SET status = 'posted',
    total_debit = $2,
    total_credit = $3,
    updated_at = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING id, entry_number, date, description, status,
          total_debit, total_credit, reversal_of, reversal_reason,
          reversed_by, source, created_by, created_at, updated_at;

-- name: ReverseJournalEntry :one
UPDATE journal_entries
SET status = 'reversed',
    reversed_by = $2,
    updated_at = NOW()
WHERE id = $1 AND status = 'posted'
RETURNING id, entry_number, date, description, status,
          total_debit, total_credit, reversal_of, reversal_reason,
          reversed_by, source, created_by, created_at, updated_at;

-- name: CreateReversalJournalEntry :one
INSERT INTO journal_entries (entry_number, date, description, source, status, reversal_of, reversal_reason, created_by)
VALUES ($1, $2, $3, $4, 'posted', $5, $6, $7)
RETURNING id, entry_number, date, description, status,
          total_debit, total_credit, reversal_of, reversal_reason,
          reversed_by, source, created_by, created_at, updated_at;

-- name: DeleteJournalEntry :exec
DELETE FROM journal_entries WHERE id = $1 AND status = 'draft';

-- name: ForceDeleteJournalEntry :exec
DELETE FROM journal_entries WHERE id = $1;

-- name: GenerateJournalEntryNumber :one
SELECT 'JU-' || to_char(CURRENT_DATE, 'YYYY') || '-' ||
       LPAD((COALESCE(
           (SELECT COUNT(*) + 1 FROM journal_entries
            WHERE entry_number LIKE 'JU-' || to_char(CURRENT_DATE, 'YYYY') || '-%'),
           1
       ))::TEXT, 3, '0') AS entry_number;
