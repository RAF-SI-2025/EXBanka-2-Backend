//go:build e2e

package e2e_saga

// SG-02: Pre-saga validation — 4 subtests, each checks a different precondition failure.
// None should create a saga execution; all return HTTP 4xx.
//
// Run: go test -tags e2e -run TestSG02 -v ./tests/e2e_saga/

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG02(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	noSagaCreated := func(t *testing.T, contractID int64) {
		t.Helper()
		var id int64
		err := db.QueryRow(`
			SELECT id FROM core_banking.otc_saga_executions WHERE contract_id = $1
		`, contractID).Scan(&id)
		assert.Equal(t, sql.ErrNoRows, err, "no saga execution should be created on pre-saga failure")
	}

	t.Run("a_wrong_caller", func(t *testing.T) {
		Truncate(t, db)
		s := SeedReadyContract(t, db, SeedOpts{})
		// Use seller's token instead of buyer's
		wrongToken := TokenFor(t, s.SellerUserID, "CLIENT")
		status, _ := Exercise(t, s.ContractID, wrongToken, nil)
		assert.Equal(t, http.StatusForbidden, status, "non-buyer should get 403")
		noSagaCreated(t, s.ContractID)
	})

	t.Run("b_nonexistent_contract", func(t *testing.T) {
		Truncate(t, db)
		s := SeedReadyContract(t, db, SeedOpts{})
		token := TokenFor(t, s.BuyerUserID, "CLIENT")
		status, _ := Exercise(t, 999999999, token, nil)
		assert.Equal(t, http.StatusNotFound, status, "nonexistent contract should get 404")
	})

	t.Run("c_already_exercised", func(t *testing.T) {
		Truncate(t, db)
		s := SeedReadyContract(t, db, SeedOpts{ContractStatus: "EXERCISED"})
		token := TokenFor(t, s.BuyerUserID, "CLIENT")
		status, _ := Exercise(t, s.ContractID, token, nil)
		assert.Equal(t, http.StatusConflict, status, "already exercised contract should get 409")
		noSagaCreated(t, s.ContractID)
	})

	t.Run("d_past_settlement_date", func(t *testing.T) {
		Truncate(t, db)
		past := time.Now().UTC().AddDate(0, 0, -1)
		s := SeedReadyContract(t, db, SeedOpts{SettlementDate: past})
		token := TokenFor(t, s.BuyerUserID, "CLIENT")
		status, _ := Exercise(t, s.ContractID, token, nil)
		assert.Equal(t, http.StatusConflict, status, "expired settlement date should get 409")
		noSagaCreated(t, s.ContractID)
	})
}

func TestSG02c_ExpiredStatus(t *testing.T) {
	db := Connect(t)
	defer db.Close()
	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{ContractStatus: "EXPIRED"})
	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	status, _ := Exercise(t, s.ContractID, token, nil)
	require.Equal(t, http.StatusConflict, status, "EXPIRED contract should get 409")
}
