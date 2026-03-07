package domain

import (
	"context"
)

type Repository interface {
	// Define repository interfaces
	// This is a placeholder for common repository methods
}

type UseCase interface {
	// Define use case interfaces
	// This is a placeholder for common use case methods
}

type Service interface {
	// Define service interfaces
	// This is a placeholder for common service methods
}

type Validator interface {
	Validate(ctx context.Context, entity interface{}) error
}
