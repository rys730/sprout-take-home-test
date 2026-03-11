-- ============================================================================
-- CUSTOMERS
-- ============================================================================

-- name: GetCustomerByID :one
SELECT id, name, email, phone, address, is_active,
       created_by, created_at, updated_at
FROM customers
WHERE id = $1;

-- name: ListCustomers :many
SELECT id, name, email, phone, address, is_active,
       created_by, created_at, updated_at
FROM customers
WHERE (sqlc.narg('is_active')::BOOLEAN IS NULL OR is_active = sqlc.narg('is_active')::BOOLEAN)
  AND (sqlc.narg('search')::TEXT IS NULL OR name ILIKE '%' || sqlc.narg('search')::TEXT || '%')
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountCustomers :one
SELECT COUNT(*) FROM customers
WHERE (sqlc.narg('is_active')::BOOLEAN IS NULL OR is_active = sqlc.narg('is_active')::BOOLEAN)
  AND (sqlc.narg('search')::TEXT IS NULL OR name ILIKE '%' || sqlc.narg('search')::TEXT || '%');

-- name: CreateCustomer :one
INSERT INTO customers (name, email, phone, address, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email, phone, address, is_active,
          created_by, created_at, updated_at;

-- name: UpdateCustomer :one
UPDATE customers
SET name = $2, email = $3, phone = $4, address = $5, updated_at = NOW()
WHERE id = $1
RETURNING id, name, email, phone, address, is_active,
          created_by, created_at, updated_at;

-- name: DeleteCustomer :exec
DELETE FROM customers WHERE id = $1;

-- name: SetCustomerActive :one
UPDATE customers
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, email, phone, address, is_active,
          created_by, created_at, updated_at;
