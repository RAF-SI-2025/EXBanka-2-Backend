// White-box testovi za InstallmentWorker.
// Paket je isti kao produkcijski kod radi pristupa neeksportovanim metodama.
package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"banka-backend/services/bank-service/internal/domain"
	"banka-backend/services/bank-service/mocks"
)

// ─── mockPublisher ────────────────────────────────────────────────────────────

// mockPublisher je lokalni testify mock za NotificationPublisher interfejs.
type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(event KreditEmailEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

// ─── Pomoćne funkcije ─────────────────────────────────────────────────────────

// newTestWorker kreira InstallmentWorker sa ubrizganim zavisnostima za testove.
func newTestWorker(repo domain.KreditRepository, pub NotificationPublisher) *InstallmentWorker {
	return NewInstallmentWorker(
		repo,
		pub,
		time.Hour,    // interval — ne koristi se u unit testovima
		72*time.Hour, // retryAfter
		0.05,         // penaltyPct
	)
}

// testDue kreira minimalni DueInstallment za testove.
func testDue() domain.DueInstallment {
	return domain.DueInstallment{
		RataID:                1,
		KreditID:              10,
		IznosRate:             5_000,
		Valuta:                "RSD",
		OcekivaniDatumDospeca: time.Now().UTC().AddDate(0, -1, 0),
		BrojPokusaja:          0,
		BrojRacuna:            "111-222-333",
		VlasnikID:             42,
	}
}

// expectedInput kreira ProcessInstallmentInput koji odgovara testDue().
func expectedInput(due domain.DueInstallment) domain.ProcessInstallmentInput {
	return domain.ProcessInstallmentInput{
		RataID:     due.RataID,
		KreditID:   due.KreditID,
		BrojRacuna: due.BrojRacuna,
		IznosRate:  due.IznosRate,
		Valuta:     due.Valuta,
	}
}

// ─── attemptPayment: tok uspeha ──────────────────────────────────────────────

func TestAttemptPayment_Success(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()

	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).Return(nil)
	pub.On("Publish", mock.MatchedBy(func(e KreditEmailEvent) bool {
		return e.Type == eventTypeUspeh && e.KreditID == due.KreditID
	})).Return(nil)

	ok := w.attemptPayment(context.Background(), due, false)
	assert.True(t, ok, "uspešna naplata mora vraćati true")
}

// ─── attemptPayment: idempotentnost ─────────────────────────────────────────

func TestAttemptPayment_AlreadyPaid(t *testing.T) {
	// Rata je već naplaćena — tretiramo kao uspeh; notifikacija se NE šalje.
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()

	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).Return(domain.ErrRataVecPlacena)

	ok := w.attemptPayment(context.Background(), due, false)
	assert.True(t, ok, "već plaćena rata se tretira kao uspeh (idempotentnost)")
	pub.AssertNotCalled(t, "Publish")
}

// ─── attemptPayment: prvi neuspeh ─────────────────────────────────────────────

func TestAttemptPayment_FirstFailure_MarksRetry(t *testing.T) {
	// Nema sredstava na prvom pokušaju →
	//   - MarkInstallmentFailed(rataID, now+72h)
	//   - Publish upozorenje (CREDIT_RATA_UPOZORENJE)
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()
	before := time.Now().UTC()

	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).
		Return(domain.ErrInsufficientFunds)

	repo.On("MarkInstallmentFailed", mock.Anything, due.RataID,
		mock.MatchedBy(func(nextRetry time.Time) bool {
			// nextRetry treba da bude između now+71h i now+73h
			min := before.Add(71 * time.Hour)
			max := before.Add(73 * time.Hour)
			return nextRetry.After(min) && nextRetry.Before(max)
		}),
	).Return(nil)

	pub.On("Publish", mock.MatchedBy(func(e KreditEmailEvent) bool {
		return e.Type == eventTypeUpozorenje && e.KreditID == due.KreditID
	})).Return(nil)

	ok := w.attemptPayment(context.Background(), due, false)
	assert.False(t, ok, "neuspela naplata mora vraćati false")
}

// ─── attemptPayment: drugi neuspeh ────────────────────────────────────────────

func TestAttemptPayment_SecondFailure_AppliesPenalty(t *testing.T) {
	// Nema sredstava na ponovnom pokušaju (isRetry=true) →
	//   - ApplyLatePaymentPenalty(kreditID, 0.05)
	//   - Publish kazna (CREDIT_RATA_KAZNA)
	//   - MarkInstallmentFailed se NE poziva (penalty zamenjuje retry logiku)
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()
	due.BrojPokusaja = 1 // već je bilo jedno neuspelo

	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).
		Return(domain.ErrInsufficientFunds)

	repo.On("ApplyLatePaymentPenalty", mock.Anything, due.KreditID, 0.05).Return(nil)

	pub.On("Publish", mock.MatchedBy(func(e KreditEmailEvent) bool {
		return e.Type == eventTypeKazna && e.KreditID == due.KreditID
	})).Return(nil)

	ok := w.attemptPayment(context.Background(), due, true)
	assert.False(t, ok, "neuspela naplata mora vraćati false")
}

// ─── attemptPayment: neočekivana greška ──────────────────────────────────────

func TestAttemptPayment_UnexpectedError(t *testing.T) {
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()

	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).
		Return(assert.AnError)

	ok := w.attemptPayment(context.Background(), due, false)
	assert.False(t, ok, "greška mora vraćati false")
	pub.AssertNotCalled(t, "Publish")
}

// ─── runDailyJob: integracija ─────────────────────────────────────────────────

func TestRunDailyJob_NoInstallments(t *testing.T) {
	// Nema dospelih rata — dnevni job mora proći bez grešaka i bez notifikacija.
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)

	repo.On("GetDueInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{}, nil)
	repo.On("GetRetryInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{}, nil)

	w.runDailyJob(context.Background())
	pub.AssertNotCalled(t, "Publish")
}

func TestRunDailyJob_SuccessfulInstallment(t *testing.T) {
	// Jedna dospela rata; naplata uspeva → tačno jedna SUCCESS notifikacija.
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()

	repo.On("GetDueInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{due}, nil)
	repo.On("GetRetryInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{}, nil)
	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).Return(nil)
	pub.On("Publish", mock.MatchedBy(func(e KreditEmailEvent) bool {
		return e.Type == eventTypeUspeh
	})).Return(nil)

	w.runDailyJob(context.Background())
}

func TestRunDailyJob_RetryInstallmentWithPenalty(t *testing.T) {
	// Jedna rata u kašnjenju (retry faza); naplata ne uspeva → kazna + KAZNA notifikacija.
	repo := mocks.NewMockKreditRepository(t)
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(repo, pub)
	due := testDue()

	repo.On("GetDueInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{}, nil)
	repo.On("GetRetryInstallments", mock.Anything, mock.Anything).
		Return([]domain.DueInstallment{due}, nil)
	repo.On("ProcessInstallmentPayment", mock.Anything, expectedInput(due)).
		Return(domain.ErrInsufficientFunds)
	repo.On("ApplyLatePaymentPenalty", mock.Anything, due.KreditID, 0.05).Return(nil)
	pub.On("Publish", mock.MatchedBy(func(e KreditEmailEvent) bool {
		return e.Type == eventTypeKazna
	})).Return(nil)

	w.runDailyJob(context.Background())
}

// ─── publishEvent: otpornost na greške notifikacije ─────────────────────────

func TestPublishEvent_NotificationFailureDoesNotPanic(t *testing.T) {
	// Greška pri slanju notifikacije NE sme sprečiti nastavak rada workera.
	pub := &mockPublisher{}
	defer pub.AssertExpectations(t)

	w := newTestWorker(mocks.NewMockKreditRepository(t), pub)

	pub.On("Publish", mock.Anything).Return(assert.AnError)

	// Ne sme da panikuje
	assert.NotPanics(t, func() {
		w.publishEvent(1, KreditEmailEvent{Type: eventTypeUspeh, KreditID: 10, VlasnikID: 42})
	})
}
