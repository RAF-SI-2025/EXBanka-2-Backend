//go:build e2e

package e2e_saga

// SG-03: F1 natural failure — buyer has insufficient funds.
// Expected: FAILED, log=[{RESERVE_FUNDS, FAILED}], no side effects.
//
// Run: go test -tags e2e -run TestSG03 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG03(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	// Contract requires 10 × 300 = 3000 USD; buyer only has 500.
	s := SeedReadyContract(t, db, SeedOpts{
		BuyerBalance:   500,
		SellerQty:      10,
		ContractAmount: 10,
		StrikePrice:    300,
	})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, nil)
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusFailed, exec.Status)
	// F1 failed → compensateFrom(PENDING) → PENDING, no compensators run
	assert.Equal(t, StepPending, exec.CurrentStep)

	// Step log: exactly 1 entry — F1 FAILED
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 1)
	assert.Equal(t, StepReserveFunds, logEntries[0].Step)
	assert.Equal(t, "FAILED", logEntries[0].StepStatus)

	// I3: no reserved funds
	_, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0")

	// I2: seller shares unchanged
	assert.Equal(t, 10, PublicShareQty(t, db, s.SellerUserID, s.ListingID))

	// I6: contract still VALID
	assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID))

	// No transactions created
	assert.Equal(t, 0, CountTransactions(t, db, s.BuyerAccountID))
}
