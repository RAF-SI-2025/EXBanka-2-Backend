package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"banka-backend/services/bank-service/internal/domain"
	"banka-backend/services/bank-service/internal/service"
	"banka-backend/services/bank-service/mocks"
)

func newKarticaService(repo *mocks.MockKarticaRepository, store *mocks.MockCardRequestStore, notif *mocks.MockNotificationSender) domain.KarticaService {
	return service.NewKarticaService(repo, "test-pepper", store, notif)
}

// ─── GetMojeKartice ───────────────────────────────────────────────────────────

func TestGetMojeKartice_Success(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	want := []domain.KarticaSaRacunom{{BrojRacuna: "123456"}}
	repo.On("GetKarticeKorisnika", ctx, int64(1)).Return(want, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	got, err := svc.GetMojeKartice(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetMojeKartice_Error(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticeKorisnika", ctx, int64(99)).Return(nil, errors.New("db error"))

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.GetMojeKartice(ctx, 99)
	assert.Error(t, err)
}

// ─── GetKarticeZaPortalZaposlenih ─────────────────────────────────────────────

func TestGetKarticeZaPortalZaposlenih_Success(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	want := []domain.KarticaEmployeeRow{{BrojKartice: "4666661234567890"}}
	repo.On("GetKarticeZaRacunBroj", ctx, "123456789012345678").Return(want, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	got, err := svc.GetKarticeZaPortalZaposlenih(ctx, "123456789012345678")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

// ─── BlokirajKarticu ──────────────────────────────────────────────────────────

func TestBlokirajKarticu_NotOwner(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticaOwnerInfo", ctx, int64(1)).
		Return(&domain.KarticaOwnerInfo{VlasnikID: 99, Status: "AKTIVNA"}, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	err := svc.BlokirajKarticu(ctx, 1, 5) // requesterID=5, owner=99
	assert.ErrorIs(t, err, domain.ErrKarticaNijeTvoja)
}

func TestBlokirajKarticu_AlreadyBlocked(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticaOwnerInfo", ctx, int64(1)).
		Return(&domain.KarticaOwnerInfo{VlasnikID: 5, Status: "BLOKIRANA"}, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	err := svc.BlokirajKarticu(ctx, 1, 5)
	assert.ErrorIs(t, err, domain.ErrKarticaVecBlokirana)
}

func TestBlokirajKarticu_Success(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticaOwnerInfo", ctx, int64(1)).
		Return(&domain.KarticaOwnerInfo{VlasnikID: 5, Status: "AKTIVNA"}, nil)
	repo.On("SetKarticaStatus", ctx, int64(1), "BLOKIRANA").Return(nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	err := svc.BlokirajKarticu(ctx, 1, 5)
	assert.NoError(t, err)
}

// ─── ChangeEmployeeCardStatus ─────────────────────────────────────────────────

func TestChangeEmployeeCardStatus_CardNotFound(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticaZaStatusChange", ctx, "4666661234567890").
		Return((*domain.KarticaZaStatusChange)(nil), errors.New("not found"))

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.ChangeEmployeeCardStatus(ctx, "4666661234567890", "BLOKIRANA")
	assert.Error(t, err)
}

func TestChangeEmployeeCardStatus_InvalidTransition_DeaktiviranaToAktivna(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetKarticaZaStatusChange", ctx, "4666661234567890").
		Return(&domain.KarticaZaStatusChange{TrenutniStatus: "DEAKTIVIRANA"}, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.ChangeEmployeeCardStatus(ctx, "4666661234567890", "AKTIVNA")
	assert.Error(t, err)
}

// ─── CreateKarticaZaVlasnika ──────────────────────────────────────────────────

func TestCreateKarticaZaVlasnika_InvalidCardType(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetRacunInfo", ctx, int64(1)).
		Return(&domain.RacunInfo{VrstaRacuna: "LICNI", ValutaOznaka: "RSD", MesecniLimit: 100000}, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.CreateKarticaZaVlasnika(ctx, 1, "INVALID_TYPE")
	assert.ErrorIs(t, err, domain.ErrNepoznatTipKartice)
}

func TestCreateKarticaZaVlasnika_DinaCardNotRSD(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetRacunInfo", ctx, int64(1)).
		Return(&domain.RacunInfo{VrstaRacuna: "DEVIZNI", ValutaOznaka: "EUR", MesecniLimit: 5000}, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.CreateKarticaZaVlasnika(ctx, 1, domain.TipKarticaDinaCard)
	assert.ErrorIs(t, err, domain.ErrDinaCardSamoRSD)
}

func TestCreateKarticaZaVlasnika_LicniLimitExceeded(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetRacunInfo", ctx, int64(1)).
		Return(&domain.RacunInfo{VrstaRacuna: "LICNI", ValutaOznaka: "RSD", MesecniLimit: 100000}, nil)
	repo.On("CountKarticeZaRacun", ctx, int64(1)).Return(int64(2), nil) // at limit (max 2)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.CreateKarticaZaVlasnika(ctx, 1, domain.TipKarticaVisa)
	assert.ErrorIs(t, err, domain.ErrKarticaLimitPremasen)
}

func TestCreateKarticaZaVlasnika_PoslovniVlasnikAlreadyHasCard(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetRacunInfo", ctx, int64(2)).
		Return(&domain.RacunInfo{VrstaRacuna: "POSLOVNI", ValutaOznaka: "RSD", MesecniLimit: 500000}, nil)
	repo.On("HasVlasnikovaKarticaPostoji", ctx, int64(2)).Return(true, nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	_, err := svc.CreateKarticaZaVlasnika(ctx, 2, domain.TipKarticaMastercard)
	assert.ErrorIs(t, err, domain.ErrKarticaLimitPremasen)
}

func TestCreateKarticaZaVlasnika_Success(t *testing.T) {
	repo := &mocks.MockKarticaRepository{}
	ctx := context.Background()
	repo.On("GetRacunInfo", ctx, int64(3)).
		Return(&domain.RacunInfo{VrstaRacuna: "LICNI", ValutaOznaka: "RSD", MesecniLimit: 100000}, nil)
	repo.On("CountKarticeZaRacun", ctx, int64(3)).Return(int64(0), nil)
	repo.On("CreateKartica", ctx, mock.Anything).Return(int64(10), nil)

	svc := newKarticaService(repo, &mocks.MockCardRequestStore{}, &mocks.MockNotificationSender{})
	id, err := svc.CreateKarticaZaVlasnika(ctx, 3, domain.TipKarticaVisa)
	require.NoError(t, err)
	assert.Equal(t, int64(10), id)
}
