//go:build e2e

package e2e_saga

// SG-01: Happy path — full SAGA completes successfully.
//
// Run: go test -tags e2e -run TestSG01 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG01(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{
		BuyerBalance: 5000,
		SellerQty:    10,
		ContractAmount: 10,
		StrikePrice:  300,
	})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, nil)
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)

	// Status
	assert.Equal(t, StatusCompleted, exec.Status)
	assert.Equal(t, StepCompleted, exec.CurrentStep)

	// Step log: exactly 5 entries, all COMPLETED
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 5)
	wantSteps := []string{
		StepReserveFunds,
		StepReserveSecurities,
		StepTransferFunds,
		StepTransferOwnership,
		StepCompleted,
	}
	for i, e := range logEntries {
		assert.Equal(t, wantSteps[i], e.Step, "log[%d].step", i)
		assert.Equal(t, "COMPLETED", e.StepStatus, "log[%d].step_status", i)
		assert.Equal(t, 1, e.Attempt, "log[%d].attempt", i)
	}

	// I1: money conserved — buyer paid 3000, seller received 3000
	buyerBal, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 2000.0, buyerBal, 0.01, "I1: buyer balance 5000-3000=2000")
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0 after COMPLETED")

	_, sellerRes := AccountBalances(t, db, s.SellerAccountID)
	assert.Equal(t, 0.0, sellerRes, "I3: seller reserved=0")

	// I2: seller has 0 public_shares for this listing (deducted in F2, not returned)
	sellerShares := PublicShareQty(t, db, s.SellerUserID, s.ListingID)
	assert.Equal(t, 0, sellerShares, "I2: seller public_shares=0 after exercise")

	// I6: contract is EXERCISED
	assert.Equal(t, "EXERCISED", ContractStatus(t, db, s.ContractID), "I6")

	// I4: 2 transactions created (buyer debit + seller credit)
	assert.Equal(t, 1, CountTransactions(t, db, s.BuyerAccountID), "I4: buyer has 1 transaction")
	assert.Equal(t, 1, CountTransactions(t, db, s.SellerAccountID), "I4: seller has 1 transaction")
}
