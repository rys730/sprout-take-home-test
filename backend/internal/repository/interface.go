package repository

type UserRepository interface {
	// Add your repository methods here
	// Example:
	// CreateUser(ctx context.Context, user *User) error
	// GetUserByID(ctx context.Context, id string) (*User, error)
	// GetUserByEmail(ctx context.Context, email string) (*User, error)
	// UpdateUser(ctx context.Context, user *User) error
	// DeleteUser(ctx context.Context, id string) error
}

type ProductRepository interface {
	// Add your repository methods here
	// Example:
	// CreateProduct(ctx context.Context, product *Product) error
	// GetProductByID(ctx context.Context, id string) (*Product, error)
	// ListProducts(ctx context.Context, limit, offset int32) ([]*Product, error)
	// UpdateProduct(ctx context.Context, product *Product) error
	// DeleteProduct(ctx context.Context, id string) error
}
