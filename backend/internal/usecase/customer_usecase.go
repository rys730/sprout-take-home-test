package usecase

import (
	"context"
	"fmt"

	"sprout-backend/internal/domain"
)

type customerUseCase struct {
	repo domain.CustomerRepository
}

func NewCustomerUseCase(repo domain.CustomerRepository) domain.CustomerUseCase {
	return &customerUseCase{repo: repo}
}

func (uc *customerUseCase) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	c, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("customer not found")
	}
	return c, nil
}

func (uc *customerUseCase) List(ctx context.Context, filter domain.CustomerFilter) ([]domain.Customer, int64, error) {
	return uc.repo.List(ctx, filter)
}

func (uc *customerUseCase) Create(ctx context.Context, req domain.CreateCustomerRequest, createdBy string) (*domain.Customer, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	customer := &domain.Customer{
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		CreatedBy: &createdBy,
	}
	return uc.repo.Create(ctx, customer)
}

func (uc *customerUseCase) Update(ctx context.Context, id string, req domain.UpdateCustomerRequest) (*domain.Customer, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("customer not found")
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Email != nil {
		existing.Email = req.Email
	}
	if req.Phone != nil {
		existing.Phone = req.Phone
	}
	if req.Address != nil {
		existing.Address = req.Address
	}
	return uc.repo.Update(ctx, existing)
}

func (uc *customerUseCase) Delete(ctx context.Context, id string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("customer not found")
	}
	return uc.repo.Delete(ctx, id)
}
