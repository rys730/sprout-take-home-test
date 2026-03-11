package repository

import (
	"context"
	"errors"
	"fmt"

	"sprout-backend/db/queries"
	"sprout-backend/internal/domain"
	"sprout-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---------------------------------------------------------------------------
// Customer Repository
// ---------------------------------------------------------------------------

type CustomerRepository struct {
	pool *pgxpool.Pool
	q    *queries.Queries
}

func NewCustomerRepository(pool *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{
		pool: pool,
		q:    queries.New(pool),
	}
}

func customerToDomain(c queries.Customer) domain.Customer {
	return domain.Customer{
		ID:        utils.UUIDToString(c.ID),
		Name:      c.Name,
		Email:     utils.TextToStringPtr(c.Email),
		Phone:     utils.TextToStringPtr(c.Phone),
		Address:   utils.TextToStringPtr(c.Address),
		IsActive:  c.IsActive,
		CreatedBy: utils.UUIDToStringPtr(c.CreatedBy),
		CreatedAt: utils.TimestamptzToTime(c.CreatedAt),
		UpdatedAt: utils.TimestamptzToTime(c.UpdatedAt),
	}
}

func (r *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	row, err := r.q.GetCustomerByID(ctx, utils.ParseUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get customer by id: %w", err)
	}
	c := customerToDomain(row)
	return &c, nil
}

func (r *CustomerRepository) List(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int64, error) {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int32(filter.Offset)
	if offset < 0 {
		offset = 0
	}

	rows, err := r.q.ListCustomers(ctx, queries.ListCustomersParams{
		Limit:    limit,
		Offset:   offset,
		IsActive: pgtype.Bool{Valid: false}, // show all
		Search:   utils.TextFromString(filter.Search),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	total, err := r.q.CountCustomers(ctx, queries.CountCustomersParams{
		IsActive: pgtype.Bool{Valid: false},
		Search:   utils.TextFromString(filter.Search),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	customers := make([]domain.Customer, 0, len(rows))
	for _, row := range rows {
		customers = append(customers, customerToDomain(row))
	}
	return customers, total, nil
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	row, err := r.q.CreateCustomer(ctx, queries.CreateCustomerParams{
		Name:      customer.Name,
		Email:     utils.TextFromString(ptrToString(customer.Email)),
		Phone:     utils.TextFromString(ptrToString(customer.Phone)),
		Address:   utils.TextFromString(ptrToString(customer.Address)),
		CreatedBy: utils.StringPtrToUUID(customer.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create customer: %w", err)
	}
	c := customerToDomain(row)
	return &c, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	row, err := r.q.UpdateCustomer(ctx, queries.UpdateCustomerParams{
		ID:      utils.ParseUUID(customer.ID),
		Name:    customer.Name,
		Email:   utils.TextFromString(ptrToString(customer.Email)),
		Phone:   utils.TextFromString(ptrToString(customer.Phone)),
		Address: utils.TextFromString(ptrToString(customer.Address)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("update customer: %w", err)
	}
	c := customerToDomain(row)
	return &c, nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	err := r.q.DeleteCustomer(ctx, utils.ParseUUID(id))
	if err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

// ptrToString safely dereferences a string pointer.
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
