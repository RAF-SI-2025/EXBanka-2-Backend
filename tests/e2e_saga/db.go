//go:build e2e

package e2e_saga

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a *sql.DB to the test database.
// Reads E2E_DB_* env vars; falls back to docker-compose defaults.
func Connect(t *testing.T) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=core_banking",
		envOr("E2E_DB_HOST", "localhost"),
		envOr("E2E_DB_PORT", "5433"),
		envOr("E2E_DB_USER", "bank_admin"),
		envOr("E2E_DB_PASS", "super_secret_password"),
		envOr("E2E_DB_NAME", "bank_db"),
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Connect: sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Connect: ping failed (is docker stack up?): %v", err)
	}
	return db
}

// Truncate wipes transactional test tables in FK-safe order.
// Reference data (exchange, listing, valuta) is preserved — seeded via upsert.
func Truncate(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE
			core_banking.otc_saga_step_log,
			core_banking.otc_saga_executions,
			core_banking.otc_contracts,
			core_banking.otc_offers,
			core_banking.transakcija,
			core_banking.public_shares,
			core_banking.racun
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("Truncate: %v", err)
	}
}

// SeedOpts configures SeedReadyContract. Zero values use SG-01 defaults.
type SeedOpts struct {
	BuyerUserID    int64
	SellerUserID   int64
	BuyerBalance   float64   // default 5000
	SellerQty      int       // default 10
	ContractAmount int32     // default 10
	StrikePrice    float64   // default 300 (USD)
	Premium        float64   // default 0
	SettlementDate time.Time // default tomorrow
	ContractStatus string    // default "VALID"
}

// Seeded holds IDs of all rows inserted by SeedReadyContract.
type Seeded struct {
	ContractID      int64
	OfferID         int64
	BuyerUserID     int64
	SellerUserID    int64
	BuyerAccountID  int64
	SellerAccountID int64
	ListingID       int64
}

// SeedReadyContract inserts a complete ready-to-exercise OTC contract.
// Upserts reference rows (valuta USD, exchange XTEST, listing TEST_E2E_OTC);
// inserts fresh buyer/seller racun, public_shares, otc_offer, otc_contract.
func SeedReadyContract(t *testing.T, db *sql.DB, opts SeedOpts) Seeded {
	t.Helper()

	if opts.BuyerUserID == 0 {
		opts.BuyerUserID = 1001
	}
	if opts.SellerUserID == 0 {
		opts.SellerUserID = 1002
	}
	if opts.BuyerBalance == 0 {
		opts.BuyerBalance = 5000
	}
	if opts.SellerQty == 0 {
		opts.SellerQty = 10
	}
	if opts.ContractAmount == 0 {
		opts.ContractAmount = 10
	}
	if opts.StrikePrice == 0 {
		opts.StrikePrice = 300
	}
	if opts.SettlementDate.IsZero() {
		opts.SettlementDate = time.Now().UTC().AddDate(0, 0, 1)
	}
	if opts.ContractStatus == "" {
		opts.ContractStatus = "VALID"
	}

	// 1. Upsert valuta USD
	var valutaID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.valuta (naziv, oznaka, simbol, zemlja)
		VALUES ('US Dollar', 'USD', '$', 'United States')
		ON CONFLICT (oznaka) DO UPDATE SET naziv = EXCLUDED.naziv
		RETURNING id
	`).Scan(&valutaID); err != nil {
		t.Fatalf("SeedReadyContract: upsert valuta: %v", err)
	}

	// 2. Upsert exchange XTEST (open_time/close_time are NOT NULL after migration 000023)
	var exchangeID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.exchange (name, acronym, mic_code, polity, currency_id, timezone, open_time, close_time)
		VALUES ('Test Exchange', 'TEST', 'XTEST', 'Test', $1, 'UTC', '00:00:00', '23:59:59')
		ON CONFLICT (mic_code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, valutaID).Scan(&exchangeID); err != nil {
		t.Fatalf("SeedReadyContract: upsert exchange: %v", err)
	}

	// 3. Upsert listing TEST_E2E_OTC
	var listingID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.listing (ticker, name, exchange_id, listing_type, price)
		VALUES ('TEST_E2E_OTC', 'Test E2E OTC Stock', $1, 'STOCK', 300)
		ON CONFLICT (ticker) DO UPDATE SET price = EXCLUDED.price
		RETURNING id
	`, exchangeID).Scan(&listingID); err != nil {
		t.Fatalf("SeedReadyContract: upsert listing: %v", err)
	}

	now := time.Now().UTC()
	futureExpiry := now.AddDate(1, 0, 0)

	// 4. Insert buyer racun
	var buyerAccountID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.racun
			(broj_racuna, id_zaposlenog, id_vlasnika, id_valute,
			 kategorija_racuna, vrsta_racuna, naziv_racuna,
			 stanje_racuna, rezervisana_sredstva,
			 datum_kreiranja, datum_isteka, status)
		VALUES
			($1, 9999, $2, $3,
			 'DEVIZNI', 'LICNI', 'Test Buyer Account',
			 $4, 0,
			 $5, $6, 'AKTIVAN')
		RETURNING id
	`,
		fmt.Sprintf("E2E-B-%d", opts.BuyerUserID),
		opts.BuyerUserID, valutaID,
		opts.BuyerBalance, now, futureExpiry,
	).Scan(&buyerAccountID); err != nil {
		t.Fatalf("SeedReadyContract: insert buyer racun: %v", err)
	}

	// 5. Insert seller racun
	var sellerAccountID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.racun
			(broj_racuna, id_zaposlenog, id_vlasnika, id_valute,
			 kategorija_racuna, vrsta_racuna, naziv_racuna,
			 stanje_racuna, rezervisana_sredstva,
			 datum_kreiranja, datum_isteka, status)
		VALUES
			($1, 9999, $2, $3,
			 'DEVIZNI', 'LICNI', 'Test Seller Account',
			 0, 0,
			 $4, $5, 'AKTIVAN')
		RETURNING id
	`,
		fmt.Sprintf("E2E-S-%d", opts.SellerUserID),
		opts.SellerUserID, valutaID,
		now, futureExpiry,
	).Scan(&sellerAccountID); err != nil {
		t.Fatalf("SeedReadyContract: insert seller racun: %v", err)
	}

	// 6. Insert seller public_shares
	if _, err := db.Exec(`
		INSERT INTO core_banking.public_shares (listing_id, user_id, quantity)
		VALUES ($1, $2, $3)
	`, listingID, opts.SellerUserID, opts.SellerQty); err != nil {
		t.Fatalf("SeedReadyContract: insert public_shares: %v", err)
	}

	// 7. Insert otc_offer (ACCEPTED — required by otc_contracts.offer_id FK)
	var offerID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.otc_offers
			(listing_id, seller_id, buyer_id,
			 buyer_account_id, seller_account_id,
			 amount, price_per_stock, premium, settlement_date,
			 status, modified_by)
		VALUES
			($1, $2, $3,
			 $4, $5,
			 $6, $7, $8, $9,
			 'ACCEPTED', $3)
		RETURNING id
	`,
		listingID, opts.SellerUserID, opts.BuyerUserID,
		buyerAccountID, sellerAccountID,
		opts.ContractAmount, opts.StrikePrice, opts.Premium,
		opts.SettlementDate.Format("2006-01-02"),
	).Scan(&offerID); err != nil {
		t.Fatalf("SeedReadyContract: insert otc_offer: %v", err)
	}

	// 8. Insert otc_contract
	var contractID int64
	if err := db.QueryRow(`
		INSERT INTO core_banking.otc_contracts
			(offer_id, listing_id, seller_id, buyer_id,
			 buyer_account_id, seller_account_id,
			 amount, strike_price, premium, settlement_date, status)
		VALUES
			($1, $2, $3, $4,
			 $5, $6,
			 $7, $8, $9, $10, $11)
		RETURNING id
	`,
		offerID, listingID, opts.SellerUserID, opts.BuyerUserID,
		buyerAccountID, sellerAccountID,
		opts.ContractAmount, opts.StrikePrice, opts.Premium,
		opts.SettlementDate.Format("2006-01-02"),
		opts.ContractStatus,
	).Scan(&contractID); err != nil {
		t.Fatalf("SeedReadyContract: insert otc_contract: %v", err)
	}

	return Seeded{
		ContractID:      contractID,
		OfferID:         offerID,
		BuyerUserID:     opts.BuyerUserID,
		SellerUserID:    opts.SellerUserID,
		BuyerAccountID:  buyerAccountID,
		SellerAccountID: sellerAccountID,
		ListingID:       listingID,
	}
}

// AccountBalances returns (stanje_racuna, rezervisana_sredstva) for an account.
func AccountBalances(t *testing.T, db *sql.DB, accountID int64) (balance, reserved float64) {
	t.Helper()
	if err := db.QueryRow(`
		SELECT stanje_racuna, rezervisana_sredstva
		FROM core_banking.racun WHERE id = $1
	`, accountID).Scan(&balance, &reserved); err != nil {
		t.Fatalf("AccountBalances(%d): %v", accountID, err)
	}
	return
}

// PublicShareQty returns total quantity in public_shares for (userID, listingID).
// Returns 0 if no rows (seller held nothing after compensation).
func PublicShareQty(t *testing.T, db *sql.DB, userID, listingID int64) int {
	t.Helper()
	var qty int
	if err := db.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0)
		FROM core_banking.public_shares
		WHERE user_id = $1 AND listing_id = $2
	`, userID, listingID).Scan(&qty); err != nil {
		t.Fatalf("PublicShareQty(%d, %d): %v", userID, listingID, err)
	}
	return qty
}

// CountTransactions returns the number of transakcija rows for an account.
func CountTransactions(t *testing.T, db *sql.DB, accountID int64) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM core_banking.transakcija WHERE racun_id = $1
	`, accountID).Scan(&n); err != nil {
		t.Fatalf("CountTransactions(%d): %v", accountID, err)
	}
	return n
}

// ContractStatus returns the current status string of an OTC contract.
func ContractStatus(t *testing.T, db *sql.DB, contractID int64) string {
	t.Helper()
	var s string
	if err := db.QueryRow(`
		SELECT status FROM core_banking.otc_contracts WHERE id = $1
	`, contractID).Scan(&s); err != nil {
		t.Fatalf("ContractStatus(%d): %v", contractID, err)
	}
	return s
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
