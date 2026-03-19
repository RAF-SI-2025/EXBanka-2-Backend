// Package worker sadrži pozadinske radnike koji se pokreću kao goroutine
// paralelno sa gRPC serverom.
package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"banka-backend/services/bank-service/internal/domain"
)

// =============================================================================
// Konstante
// =============================================================================

const (
	// Tipovi kreditnih notifikacija.
	eventTypeUspeh      = "CREDIT_RATA_USPEH"
	eventTypeUpozorenje = "CREDIT_RATA_UPOZORENJE"
	eventTypeKazna      = "CREDIT_RATA_KAZNA"
)

// =============================================================================
// InstallmentWorker
// =============================================================================

// InstallmentWorker je pozadinski radnik koji jednom dnevno:
//  1. Naplaćuje sve dospele rate (status NEPLACENO, ocekivani_datum_dospeca ≤ danas).
//  2. Ponovni pokušaj naplate za rate u statusu KASNI čiji je sledeci_pokusaj ≤ sada.
//
// Poslovni tokovi (Issue #5):
//   - Uspeh:        rata → PLACENO; kredit datum_sledece_rate pomeren; UPLATA transakcija; SUCCESS notifikacija.
//   - 1. neuspeh:   rata → KASNI; sledeci_pokusaj = sada + retryAfter; WARNING notifikacija.
//   - 2. neuspeh:   kredit nominalna_stopa += penaltyPct; kredit → U_KASNJENJU; PENALTY notifikacija.
//
// Idempotentnost: ProcessInstallmentPayment koristi SELECT FOR UPDATE i proverava
// status_placanja pre svake naplate — dvostruka naplata je nemoguća čak i ako
// worker bude ponovo podignut tokom iste sesije.
type InstallmentWorker struct {
	kreditRepo domain.KreditRepository // za sve DB operacije vezane za kredit/rate
	publisher  NotificationPublisher   // za slanje email notifikacija

	interval   time.Duration // koliko često se worker pokreće (default 24h)
	retryAfter time.Duration // kada se zakazuje ponovni pokušaj (default 72h)
	penaltyPct float64       // kazneni % koji se dodaje nominalnoj stopi (default 0.05)
}

// NewInstallmentWorker konstruktor — sve zavisnosti se injektuju.
func NewInstallmentWorker(
	kreditRepo domain.KreditRepository,
	publisher NotificationPublisher,
	interval time.Duration,
	retryAfter time.Duration,
	penaltyPct float64,
) *InstallmentWorker {
	return &InstallmentWorker{
		kreditRepo: kreditRepo,
		publisher:  publisher,
		interval:   interval,
		retryAfter: retryAfter,
		penaltyPct: penaltyPct,
	}
}

// Start pokreće worker u tekućoj goroutini (pozivati sa go worker.Start(ctx)).
// Blokira sve dok ctx ne bude otkazan (graceful shutdown).
// Dnevni job se pokreće odmah pri startu, a zatim na svaki ticker.
func (w *InstallmentWorker) Start(ctx context.Context) {
	log.Printf("[worker] InstallmentWorker pokrenut (interval=%s retryAfter=%s penaltyPct=%.4f%%)",
		w.interval, w.retryAfter, w.penaltyPct)

	// Pokreni odmah — ne čekaj prvi tick.
	w.runDailyJob(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.runDailyJob(ctx)

		case <-ctx.Done():
			log.Printf("[worker] InstallmentWorker prima signal za zaustavljanje — kraj")
			return
		}
	}
}

// =============================================================================
// Dnevni posao
// =============================================================================

func (w *InstallmentWorker) runDailyJob(ctx context.Context) {
	now := time.Now().UTC()
	log.Printf("[worker] ═══ Početak dnevne naplate rata: %s ═══", now.Format("2006-01-02 15:04:05"))

	// Faza 1: Prvotna naplata dospelih rata.
	firstCount, firstOk := w.processFirstAttempts(ctx, now)
	// Faza 2: Ponovni pokušaj za rate u kašnjenju.
	retryCount, retryOk := w.processRetryAttempts(ctx, now)

	log.Printf("[worker] ═══ Završetak dnevne naplate rata: faza1=%d/%d faza2=%d/%d ═══",
		firstOk, firstCount, retryOk, retryCount)
}

// processFirstAttempts obrađuje rate sa statusom NEPLACENO čiji je datum dospeća ≤ asOf.
// Vraća (ukupan_broj, uspešnih).
func (w *InstallmentWorker) processFirstAttempts(ctx context.Context, asOf time.Time) (total, ok int) {
	dueList, err := w.kreditRepo.GetDueInstallments(ctx, asOf)
	if err != nil {
		log.Printf("[worker] GREŠKA: dohvat dospelih rata: %v", err)
		return 0, 0
	}

	log.Printf("[worker] Faza 1: pronađeno %d dospelih rata za naplatu", len(dueList))

	for _, due := range dueList {
		if w.attemptPayment(ctx, due, false) {
			ok++
		}
		total++
	}
	return total, ok
}

// processRetryAttempts obrađuje rate sa statusom KASNI čiji je sledeci_pokusaj ≤ asOf.
// Vraća (ukupan_broj, uspešnih).
func (w *InstallmentWorker) processRetryAttempts(ctx context.Context, asOf time.Time) (total, ok int) {
	retryList, err := w.kreditRepo.GetRetryInstallments(ctx, asOf)
	if err != nil {
		log.Printf("[worker] GREŠKA: dohvat rata za ponovni pokušaj: %v", err)
		return 0, 0
	}

	log.Printf("[worker] Faza 2: pronađeno %d rata za ponovni pokušaj naplate", len(retryList))

	for _, due := range retryList {
		if w.attemptPayment(ctx, due, true) {
			ok++
		}
		total++
	}
	return total, ok
}

// =============================================================================
// Osnovna logika naplate jedne rate
// =============================================================================

// attemptPayment pokušava da naplati jednu ratu.
// isRetry=false → prvi pokušaj; isRetry=true → ponovni pokušaj (già in KASNI).
// Vraća true ako je naplata uspela.
func (w *InstallmentWorker) attemptPayment(
	ctx context.Context,
	due domain.DueInstallment,
	isRetry bool,
) bool {
	attemptLabel := "prvi pokušaj"
	if isRetry {
		attemptLabel = fmt.Sprintf("ponovni pokušaj #%d", due.BrojPokusaja+1)
	}

	log.Printf("[worker] → Naplata rate ID=%d (%s): kredit=%d racun=%s iznos=%.2f %s dospece=%s",
		due.RataID, attemptLabel,
		due.KreditID, due.BrojRacuna,
		due.IznosRate, due.Valuta,
		due.OcekivaniDatumDospeca.Format("2006-01-02"))

	input := domain.ProcessInstallmentInput{
		RataID:     due.RataID,
		KreditID:   due.KreditID,
		BrojRacuna: due.BrojRacuna,
		IznosRate:  due.IznosRate,
		Valuta:     due.Valuta,
	}

	err := w.kreditRepo.ProcessInstallmentPayment(ctx, input)

	switch {
	// ── Uspeh ────────────────────────────────────────────────────────────────
	case err == nil:
		log.Printf("[worker] ✓ Rata ID=%d uspešno naplaćena: kredit=%d racun=%s iznos=%.2f %s",
			due.RataID, due.KreditID, due.BrojRacuna, due.IznosRate, due.Valuta)
		w.notifySuccess(due)
		return true

	// ── Idempotentnost: rata je već plaćena ───────────────────────────────────
	case errors.Is(err, domain.ErrRataVecPlacena):
		log.Printf("[worker] ℹ Rata ID=%d je već plaćena — idempotentno preskačemo", due.RataID)
		return true // tretiramo kao uspeh jer je cilj postignut

	// ── Nedovoljno sredstava ──────────────────────────────────────────────────
	case errors.Is(err, domain.ErrInsufficientFunds):
		if isRetry {
			w.handleSecondFailure(ctx, due)
		} else {
			w.handleFirstFailure(ctx, due)
		}
		return false

	// ── Neočekivana greška ────────────────────────────────────────────────────
	default:
		log.Printf("[worker] GREŠKA: naplata rate ID=%d (kredit=%d): %v",
			due.RataID, due.KreditID, err)
		return false
	}
}

// handleFirstFailure: nema sredstava na prvom pokušaju → zakažemo retry za +retryAfter.
func (w *InstallmentWorker) handleFirstFailure(ctx context.Context, due domain.DueInstallment) {
	nextRetry := time.Now().UTC().Add(w.retryAfter)

	log.Printf("[worker] ✗ Rata ID=%d (kredit=%d racun=%s): nema sredstava (%.2f %s). "+
		"Zakazujem ponovni pokušaj za: %s",
		due.RataID, due.KreditID, due.BrojRacuna,
		due.IznosRate, due.Valuta,
		nextRetry.Format("2006-01-02 15:04:05"))

	if markErr := w.kreditRepo.MarkInstallmentFailed(ctx, due.RataID, nextRetry); markErr != nil {
		log.Printf("[worker] GREŠKA: označavanje rate %d kao KASNI: %v", due.RataID, markErr)
	}

	w.notifyWarning(due, nextRetry)
}

// handleSecondFailure: nema sredstava na ponovnom pokušaju → kaznena stopa.
func (w *InstallmentWorker) handleSecondFailure(ctx context.Context, due domain.DueInstallment) {
	log.Printf("[worker] ✗✗ PONOVNI NEUSPEH rate ID=%d (kredit=%d racun=%s): nema sredstava (%.2f %s). "+
		"Primenjujem kaznenu stopu +%.4f%% na kredit",
		due.RataID, due.KreditID, due.BrojRacuna,
		due.IznosRate, due.Valuta, w.penaltyPct)

	if penaltyErr := w.kreditRepo.ApplyLatePaymentPenalty(ctx, due.KreditID, w.penaltyPct); penaltyErr != nil {
		log.Printf("[worker] GREŠKA: primena kazne na kredit %d: %v", due.KreditID, penaltyErr)
	} else {
		log.Printf("[worker] ✓ Kaznena stopa +%.4f%% primenjena na kredit=%d; status → U_KASNJENJU",
			w.penaltyPct, due.KreditID)
	}

	w.notifyPenalty(due)
}

// =============================================================================
// Notifikacije
// =============================================================================

// notifySuccess šalje email obaveštenje o uspešnoj naplati rate.
func (w *InstallmentWorker) notifySuccess(due domain.DueInstallment) {
	event := KreditEmailEvent{
		Type: eventTypeUspeh,
		// TODO: Popuniti Email pozivom user-service.GetClientEmail(due.VlasnikID).
		// Bank-service ne čuva email adrese klijenata (cross-service referenca).
		Email:     "",
		Token:     "",
		VlasnikID: due.VlasnikID,
		KreditID:  due.KreditID,
		IznosRate: due.IznosRate,
		Valuta:    due.Valuta,
		Subject:   "Uspešna naplata rate kredita",
		Body: fmt.Sprintf(
			"Poštovani, obaveštavamo Vas da je rata Vašeg kredita (ID: %d) u iznosu od %.2f %s "+
				"uspešno naplaćena sa računa %s.",
			due.KreditID, due.IznosRate, due.Valuta, due.BrojRacuna,
		),
	}
	w.publishEvent(due.RataID, event)
}

// notifyWarning šalje email upozorenje da naplata nije uspela — zakazan retry.
func (w *InstallmentWorker) notifyWarning(due domain.DueInstallment, nextRetry time.Time) {
	event := KreditEmailEvent{
		Type:  eventTypeUpozorenje,
		Email: "", // TODO: user-service lookup
		Token: "",

		VlasnikID: due.VlasnikID,
		KreditID:  due.KreditID,
		IznosRate: due.IznosRate,
		Valuta:    due.Valuta,
		Subject:   "Neuspešna naplata rate — potrebno je obezbediti sredstva",
		Body: fmt.Sprintf(
			"Poštovani, naplata rate Vašeg kredita (ID: %d) u iznosu od %.2f %s nije uspela "+
				"usled nedovoljnog stanja na računu %s. "+
				"Sledeći pokušaj naplate biće izvršen %s. "+
				"Molimo Vas da obezbedite dovoljna sredstva na računu.",
			due.KreditID, due.IznosRate, due.Valuta, due.BrojRacuna,
			nextRetry.Format("02.01.2006. u 15:04"),
		),
	}
	w.publishEvent(due.RataID, event)
}

// notifyPenalty šalje email obaveštenje da je primenjena kaznena kamata.
func (w *InstallmentWorker) notifyPenalty(due domain.DueInstallment) {
	event := KreditEmailEvent{
		Type:  eventTypeKazna,
		Email: "", // TODO: user-service lookup
		Token: "",

		VlasnikID: due.VlasnikID,
		KreditID:  due.KreditID,
		IznosRate: due.IznosRate,
		Valuta:    due.Valuta,
		Subject:   "Obaveštenje o kašnjenju u otplati — primenjena kaznena kamata",
		Body: fmt.Sprintf(
			"Poštovani, obaveštavamo Vas da je usled neuspešne naplate rate Vašeg kredita "+
				"(ID: %d) u iznosu od %.2f %s na računu %s, "+
				"primenjena kaznena kamatna stopa od +%.4f%%. "+
				"Molimo Vas da što pre regulišete obavezu radi izbegavanja daljih posledica.",
			due.KreditID, due.IznosRate, due.Valuta, due.BrojRacuna, w.penaltyPct,
		),
	}
	w.publishEvent(due.RataID, event)
}

// publishEvent šalje event na notification publisher; greška se samo loguje
// jer neuspeh notifikacije ne sme sprečiti nastavak naplate ostalih rata.
func (w *InstallmentWorker) publishEvent(rataID int64, event KreditEmailEvent) {
	if err := w.publisher.Publish(event); err != nil {
		log.Printf("[worker] WARN: slanje notifikacije tipa %q za ratu %d: %v",
			event.Type, rataID, err)
		return
	}
	log.Printf("[worker] ✉ Notifikacija tipa %q objavljena za ratu %d (vlasnik=%d)",
		event.Type, rataID, event.VlasnikID)
}
