package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"banka-backend/services/bank-service/internal/domain"
)

// MockKarticaRepository is a testify mock for domain.KarticaRepository.
type MockKarticaRepository struct {
	mock.Mock
}

func (m *MockKarticaRepository) CreateKartica(ctx context.Context, input domain.CreateKarticaInput) (int64, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKarticaRepository) CountKarticeZaRacun(ctx context.Context, racunID int64) (int64, error) {
	args := m.Called(ctx, racunID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKarticaRepository) HasVlasnikovaKarticaPostoji(ctx context.Context, racunID int64) (bool, error) {
	args := m.Called(ctx, racunID)
	return args.Bool(0), args.Error(1)
}

func (m *MockKarticaRepository) GetKarticaByID(ctx context.Context, karticaID int64) (*domain.Kartica, error) {
	args := m.Called(ctx, karticaID)
	v, _ := args.Get(0).(*domain.Kartica)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) GetKarticeByRacun(ctx context.Context, racunID int64) ([]domain.Kartica, error) {
	args := m.Called(ctx, racunID)
	v, _ := args.Get(0).([]domain.Kartica)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) HasOvlascenoLiceKarticu(ctx context.Context, emailAdresa string) (bool, error) {
	args := m.Called(ctx, emailAdresa)
	return args.Bool(0), args.Error(1)
}

func (m *MockKarticaRepository) GetRacunInfo(ctx context.Context, racunID int64) (*domain.RacunInfo, error) {
	args := m.Called(ctx, racunID)
	v, _ := args.Get(0).(*domain.RacunInfo)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) GetRacunVlasnikInfo(ctx context.Context, racunID int64) (*domain.RacunVlasnikInfo, error) {
	args := m.Called(ctx, racunID)
	v, _ := args.Get(0).(*domain.RacunVlasnikInfo)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) CreateKarticaSaOvlascenoLicem(ctx context.Context, karticaInput domain.CreateKarticaInput, olInput domain.OvlascenoLiceInput) (int64, error) {
	args := m.Called(ctx, karticaInput, olInput)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKarticaRepository) GetKarticeKorisnika(ctx context.Context, korisnikID int64) ([]domain.KarticaSaRacunom, error) {
	args := m.Called(ctx, korisnikID)
	v, _ := args.Get(0).([]domain.KarticaSaRacunom)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) GetKarticaOwnerInfo(ctx context.Context, karticaID int64) (*domain.KarticaOwnerInfo, error) {
	args := m.Called(ctx, karticaID)
	v, _ := args.Get(0).(*domain.KarticaOwnerInfo)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) SetKarticaStatus(ctx context.Context, karticaID int64, noviStatus string) error {
	args := m.Called(ctx, karticaID, noviStatus)
	return args.Error(0)
}

func (m *MockKarticaRepository) GetKarticeZaRacunBroj(ctx context.Context, brojRacuna string) ([]domain.KarticaEmployeeRow, error) {
	args := m.Called(ctx, brojRacuna)
	v, _ := args.Get(0).([]domain.KarticaEmployeeRow)
	return v, args.Error(1)
}

func (m *MockKarticaRepository) GetKarticaZaStatusChange(ctx context.Context, brojKartice string) (*domain.KarticaZaStatusChange, error) {
	args := m.Called(ctx, brojKartice)
	v, _ := args.Get(0).(*domain.KarticaZaStatusChange)
	return v, args.Error(1)
}

// MockCardRequestStore is a testify mock for domain.CardRequestStore.
type MockCardRequestStore struct {
	mock.Mock
}

func (m *MockCardRequestStore) SaveCardRequest(ctx context.Context, ownerID int64, state domain.CardRequestState, ttl time.Duration) error {
	args := m.Called(ctx, ownerID, state, ttl)
	return args.Error(0)
}

func (m *MockCardRequestStore) GetCardRequest(ctx context.Context, ownerID int64) (*domain.CardRequestState, error) {
	args := m.Called(ctx, ownerID)
	v, _ := args.Get(0).(*domain.CardRequestState)
	return v, args.Error(1)
}

func (m *MockCardRequestStore) DeleteCardRequest(ctx context.Context, ownerID int64) error {
	args := m.Called(ctx, ownerID)
	return args.Error(0)
}

// MockNotificationSender is a testify mock for domain.NotificationSender.
type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) SendCardOTP(ctx context.Context, toEmail, otpCode string) error {
	args := m.Called(ctx, toEmail, otpCode)
	return args.Error(0)
}
