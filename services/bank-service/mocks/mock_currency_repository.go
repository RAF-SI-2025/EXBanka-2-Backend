package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"banka-backend/services/bank-service/internal/domain"
)

// MockCurrencyRepository is a testify mock for domain.CurrencyRepository.
type MockCurrencyRepository struct {
	mock.Mock
}

func (m *MockCurrencyRepository) GetAll(ctx context.Context) ([]domain.Currency, error) {
	args := m.Called(ctx)
	v, _ := args.Get(0).([]domain.Currency)
	return v, args.Error(1)
}

func (m *MockCurrencyRepository) GetByID(ctx context.Context, id int64) (*domain.Currency, error) {
	args := m.Called(ctx, id)
	v, _ := args.Get(0).(*domain.Currency)
	return v, args.Error(1)
}
