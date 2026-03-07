package domain

import "time"

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserUseCase interface {
	// Define user use case methods
	// Example:
	// CreateUser(ctx context.Context, user *User) error
	// GetUser(ctx context.Context, id string) (*User, error)
	// GetUserByEmail(ctx context.Context, email string) (*User, error)
	// UpdateUser(ctx context.Context, user *User) error
	// DeleteUser(ctx context.Context, id string) error
}

type ProductUseCase interface {
	// Define product use case methods
	// Example:
	// CreateProduct(ctx context.Context, product *Product) error
	// GetProduct(ctx context.Context, id string) (*Product, error)
	// ListProducts(ctx context.Context, limit, offset int) ([]*Product, error)
	// UpdateProduct(ctx context.Context, product *Product) error
	// DeleteProduct(ctx context.Context, id string) error
}
