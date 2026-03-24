// exchange_repository.go — atomic exchange transfer execution.
//
// Implements domain.ExchangeTransferRepository. Reuses racunModel and
// transakcijaModel from account_repository.go (same package).
package repository

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"banka-backend/services/bank-service/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// trezorVlasnikID je ID korisnika trezor@exbanka.rs u user-service-u,
// seeded u 000010_seed_banka_firma.up.sql (admin=1, trezor=2).
const trezorVlasnikID = 2

type exchangeTransferRepository struct {
	db *gorm.DB
}

// NewExchangeTransferRepository creates a new ExchangeTransferRepository.
func NewExchangeTransferRepository(db *gorm.DB) domain.ExchangeTransferRepository {
	return &exchangeTransferRepository{db: db}
}

// exchangeAccountInfo is the pre-flight read projection (no lock).
type exchangeAccountInfo struct {
	ID                  int64   `gorm:"column:id"`
	BrojRacuna          string  `gorm:"column:broj_racuna"`
	IDVlasnika          int64   `gorm:"column:id_vlasnika"`
	StanjeRacuna        float64 `gorm:"column:stanje_racuna"`
	RezervovanaSredstva float64 `gorm:"column:rezervisana_sredstva"`
	Status              string  `gorm:"column:status"`
	ValutaOznaka        string  `gorm:"column:valuta_oznaka"`
}

// fetchAccountInfo queries a single account with its currency oznaka.
func (r *exchangeTransferRepository) fetchAccountInfo(ctx context.Context, db *gorm.DB, accountID int64) (*exchangeAccountInfo, error) {
	var row exchangeAccountInfo
	err := db.WithContext(ctx).Raw(`
		SELECT
			ra.id,
			ra.broj_racuna,
			ra.id_vlasnika,
			ra.stanje_racuna,
			ra.rezervisana_sredstva,
			ra.status,
			v.oznaka AS valuta_oznaka
		FROM core_banking.racun ra
		JOIN core_banking.valuta v ON v.id = ra.id_valute
		WHERE ra.id = ?
	`, accountID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, domain.ErrAccountNotFound
	}
	return &row, nil
}

// fetchTreasuryID vraća ID trezorskog računa banke za datu valutu.
// Trezorski računi su vlasništvo korisnika trezorVlasnikID (trezor@exbanka.rs).
func (r *exchangeTransferRepository) fetchTreasuryID(ctx context.Context, db *gorm.DB, currencyOznaka string) (int64, error) {
	var id int64
	err := db.WithContext(ctx).Raw(`
		SELECT ra.id
		FROM core_banking.racun ra
		JOIN core_banking.valuta v ON v.id = ra.id_valute
		WHERE ra.id_vlasnika = ?
		  AND v.oznaka = ?
		  AND ra.status = 'AKTIVAN'
		LIMIT 1
	`, trezorVlasnikID, currencyOznaka).Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("trezorski račun za valutu %s nije pronađen", currencyOznaka)
	}
	return id, nil
}

// ExecuteTransfer validates both accounts and atomically executes a 4-way journal entry
// (četvorostruko knjiženje) within a single DB transaction (ACID):
//
//  1. Klijentski izvorišni račun  → zadužuje se za input.Amount        (MENJACNICA)
//  2. Trezor banke (from valuta)  → odobrava se za input.Amount        (MENJACNICA)
//  3. Trezor banke (to valuta)    → zadužuje se za conversion.Bruto    (MENJACNICA)
//  4. Klijentski odredišni račun  → odobrava se za conversion.Result   (MENJACNICA)
//
// Provizija (= Bruto - Result) ostaje u trezoru banke (to valuta).
// Svi računi se zaključavaju u determinističkom redosledu (id ASC) radi sprečavanja deadlock-a.
func (r *exchangeTransferRepository) ExecuteTransfer(
	ctx context.Context,
	input domain.ExchangeTransferInput,
	conversion domain.ExchangeConversionResult,
) (*domain.ExchangeTransferResult, error) {

	// ── Pre-flight validacija (bez locka) ────────────────────────────────────
	src, err := r.fetchAccountInfo(ctx, r.db, input.SourceAccountID)
	if err != nil {
		return nil, err
	}
	tgt, err := r.fetchAccountInfo(ctx, r.db, input.TargetAccountID)
	if err != nil {
		return nil, err
	}

	if src.IDVlasnika != input.VlasnikID {
		return nil, domain.ErrExchangeAccountNotOwned
	}
	if tgt.IDVlasnika != input.VlasnikID {
		return nil, domain.ErrExchangeAccountNotOwned
	}
	if src.Status != "AKTIVAN" {
		return nil, domain.ErrExchangeAccountInactive
	}
	if tgt.Status != "AKTIVAN" {
		return nil, domain.ErrExchangeAccountInactive
	}
	if src.ValutaOznaka != input.FromOznaka {
		return nil, domain.ErrExchangeWrongCurrency
	}
	if tgt.ValutaOznaka != input.ToOznaka {
		return nil, domain.ErrExchangeWrongCurrency
	}
	if src.StanjeRacuna-src.RezervovanaSredstva < input.Amount {
		return nil, domain.ErrExchangeInsufficientFunds
	}

	// Pre-flight: pronađi trezorske račune (unlocked read).
	treasuryFromID, err := r.fetchTreasuryID(ctx, r.db, input.FromOznaka)
	if err != nil {
		return nil, err
	}
	treasuryToID, err := r.fetchTreasuryID(ctx, r.db, input.ToOznaka)
	if err != nil {
		return nil, err
	}

	// ── Atomična egzekucija ───────────────────────────────────────────────────
	var result *domain.ExchangeTransferResult

	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Prikupi sve ID-jeve i deduplikuj (sigurnost za edge case-ove).
		rawIDs := []int64{input.SourceAccountID, input.TargetAccountID, treasuryFromID, treasuryToID}
		seen := make(map[int64]struct{}, len(rawIDs))
		uniqueIDs := make([]int64, 0, len(rawIDs))
		for _, id := range rawIDs {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				uniqueIDs = append(uniqueIDs, id)
			}
		}

		// Zaključaj sve račune u id ASC redosledu — sprečava deadlock.
		var locked []racunModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id IN ?", uniqueIDs).
			Order("id ASC").
			Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != len(uniqueIDs) {
			return domain.ErrAccountNotFound
		}

		// Napravi mapu iz zaključanih redova.
		lockedByID := make(map[int64]*racunModel, len(locked))
		for i := range locked {
			lockedByID[locked[i].ID] = &locked[i]
		}

		srcLocked := lockedByID[input.SourceAccountID]
		if srcLocked == nil || lockedByID[input.TargetAccountID] == nil ||
			lockedByID[treasuryFromID] == nil || lockedByID[treasuryToID] == nil {
			return domain.ErrAccountNotFound
		}

		// Ponovna validacija sredstava nakon locka.
		availableNow := srcLocked.StanjeRacuna - srcLocked.RezervovanaSredstva
		if availableNow < input.Amount {
			return domain.ErrExchangeInsufficientFunds
		}

		now := time.Now().UTC()
		opis := fmt.Sprintf(
			"Menjačnica: %.4g %s → %.4g %s (provizija: %.4g %s)",
			input.Amount, input.FromOznaka,
			conversion.Result, input.ToOznaka,
			conversion.Provizija, input.ToOznaka,
		)

		// 1. Zaduži klijentski izvorišni račun.
		if err := tx.Model(&racunModel{}).
			Where("id = ?", input.SourceAccountID).
			Update("stanje_racuna", gorm.Expr("stanje_racuna - ?", input.Amount)).Error; err != nil {
			return fmt.Errorf("zaduži klijentski izvor: %w", err)
		}

		// 2. Odobri trezor banke u from valuti.
		if err := tx.Model(&racunModel{}).
			Where("id = ?", treasuryFromID).
			Update("stanje_racuna", gorm.Expr("stanje_racuna + ?", input.Amount)).Error; err != nil {
			return fmt.Errorf("odobri trezor from: %w", err)
		}

		// 3. Zaduži trezor banke u to valuti za bruto iznos.
		//    Provizija (bruto - neto) ostaje u trezoru.
		if err := tx.Model(&racunModel{}).
			Where("id = ?", treasuryToID).
			Update("stanje_racuna", gorm.Expr("stanje_racuna - ?", conversion.Bruto)).Error; err != nil {
			return fmt.Errorf("zaduži trezor to: %w", err)
		}

		// 4. Odobri klijentski odredišni račun neto iznosom.
		if err := tx.Model(&racunModel{}).
			Where("id = ?", input.TargetAccountID).
			Update("stanje_racuna", gorm.Expr("stanje_racuna + ?", conversion.Result)).Error; err != nil {
			return fmt.Errorf("odobri klijentski odredišni: %w", err)
		}

		// Audit trail: 4 transakcije, sve tipa MENJACNICA.
		entries := []transakcijaModel{
			{
				RacunID:          input.SourceAccountID,
				TipTransakcije:   "MENJACNICA",
				Iznos:            input.Amount,
				Opis:             opis,
				VremeIzvrsavanja: now,
				Status:           "IZVRSEN",
			},
			{
				RacunID:          treasuryFromID,
				TipTransakcije:   "MENJACNICA",
				Iznos:            input.Amount,
				Opis:             opis,
				VremeIzvrsavanja: now,
				Status:           "IZVRSEN",
			},
			{
				RacunID:          treasuryToID,
				TipTransakcije:   "MENJACNICA",
				Iznos:            conversion.Bruto,
				Opis:             opis,
				VremeIzvrsavanja: now,
				Status:           "IZVRSEN",
			},
			{
				RacunID:          input.TargetAccountID,
				TipTransakcije:   "MENJACNICA",
				Iznos:            conversion.Result,
				Opis:             opis,
				VremeIzvrsavanja: now,
				Status:           "IZVRSEN",
			},
		}
		if err := tx.Create(&entries).Error; err != nil {
			return fmt.Errorf("upiši transakcije: %w", err)
		}

		referenceID := fmt.Sprintf("KNV-%s-%06d",
			now.Format("20060102150405"),
			100000+rand.Intn(900000),
		)

		result = &domain.ExchangeTransferResult{
			ReferenceID:     referenceID,
			SourceAccountID: input.SourceAccountID,
			TargetAccountID: input.TargetAccountID,
			FromOznaka:      input.FromOznaka,
			ToOznaka:        input.ToOznaka,
			OriginalAmount:  input.Amount,
			GrossAmount:     conversion.Bruto,
			Provizija:       conversion.Provizija,
			NetAmount:       conversion.Result,
			ViaRSD:          conversion.ViaRSD,
			RateNote:        conversion.RateNote,
		}
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
