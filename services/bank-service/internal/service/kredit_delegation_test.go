package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"banka-backend/services/bank-service/internal/domain"
	"banka-backend/services/bank-service/internal/service"
	"banka-backend/services/bank-service/mocks"
)

func newKreditSvc(repo *mocks.MockKreditRepository) domain.KreditService {
	return service.NewKreditService(repo)
}

func TestGetClientCredits_Success(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	want := []domain.Kredit{{ID: 1}, {ID: 2}}
	repo.On("GetKreditsByVlasnik", ctx, int64(5)).Return(want, nil)

	got, err := newKreditSvc(repo).GetClientCredits(ctx, 5)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetClientCredits_Error(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	repo.On("GetKreditsByVlasnik", ctx, int64(5)).Return(nil, errors.New("db error"))

	_, err := newKreditSvc(repo).GetClientCredits(ctx, 5)
	assert.Error(t, err)
}

func TestGetCreditDetails_Success(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	kredit := &domain.Kredit{ID: 1, VlasnikID: 3}
	repo.On("GetKreditByID", ctx, int64(1)).Return(kredit, nil)
	repo.On("GetInstallmentsByKredit", ctx, int64(1)).Return([]domain.Rata{{ID: 1}}, nil)

	svc := newKreditSvc(repo)
	k, rata, err := svc.GetCreditDetails(ctx, 1, 3)
	require.NoError(t, err)
	assert.Equal(t, kredit, k)
	assert.Len(t, rata, 1)
}

func TestGetCreditDetails_NotOwner(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	kredit := &domain.Kredit{ID: 1, VlasnikID: 99} // owned by 99, not 3
	repo.On("GetKreditByID", ctx, int64(1)).Return(kredit, nil)

	_, _, err := newKreditSvc(repo).GetCreditDetails(ctx, 1, 3)
	assert.Error(t, err)
}

func TestGetCreditDetails_NotFound(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	repo.On("GetKreditByID", ctx, int64(99)).Return((*domain.Kredit)(nil), errors.New("not found"))

	_, _, err := newKreditSvc(repo).GetCreditDetails(ctx, 99, 1)
	assert.Error(t, err)
}

func TestGetAllPendingRequests(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	filter := domain.GetPendingRequestsFilter{}
	want := []domain.KreditniZahtev{{ID: 1}}
	repo.On("GetPendingRequests", ctx, filter).Return(want, nil)

	got, err := newKreditSvc(repo).GetAllPendingRequests(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestRejectCredit_Success(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	repo.On("RejectKreditRequest", ctx, int64(1)).Return(nil)

	err := newKreditSvc(repo).RejectCredit(ctx, 1)
	assert.NoError(t, err)
}

func TestRejectCredit_Error(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	repo.On("RejectKreditRequest", ctx, int64(1)).Return(errors.New("db error"))

	err := newKreditSvc(repo).RejectCredit(ctx, 1)
	assert.Error(t, err)
}

func TestGetAllApprovedCredits(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	ctx := context.Background()
	filter := domain.GetAllCreditsFilter{}
	want := []domain.Kredit{{ID: 5}}
	repo.On("GetAllCredits", ctx, filter).Return(want, nil)

	got, err := newKreditSvc(repo).GetAllApprovedCredits(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
