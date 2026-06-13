//go:build e2e

package e2e_saga

// SG-04: F2 natural failure — seller has insufficient securities.
// Expected: FAILED, log=[{RF,COMPLETED},{RS,FAILED},{RF,COMPENSATED}].
// Buyer reserved funds are restored by C1.
//
// Run: go test -tags e2e -run TestSG04 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG04(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	// Seller only has 3 shares; contract requires 10.
	s := SeedReadyContract(t, db, SeedOpts{
		BuyerBalance:   5000,
		SellerQty:      3,
		ContractAmount: 10,
		StrikePrice:    300,
	})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, nil)
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusFailed, exec.Status)
	// compensateFrom(RESERVE_FUNDS) → final step = RESERVE_FUNDS
	assert.Equal(t, StepReserveFunds, exec.CurrentStep)

	// Step log: F1 ok, F2 fail, C1 ok
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 3, "expect 3 log entries: F1 ok, F2 fail, C1 compensated")

	assert.Equal(t, StepReserveFunds, logEntries[0].Step)
	assert.Equal(t, "COMPLETED", logEntries[0].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[1].Step)
	assert.Equal(t, "FAILED", logEntries[1].StepStatus)

	assert.Equal(t, StepReserveFunds, logEntries[2].Step)
	assert.Equal(t, "COMPENSATED", logEntries[2].StepStatus)

	// I3: buyer reserved funds released by C1
	_, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0 after C1")

	// Buyer balance unchanged (F3 never ran)
	buyerBal, _ := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 5000.0, buyerBal, 0.01, "I1: buyer balance unchanged")

	// I2: seller shares restored (C2 not needed since F2 failed before removing shares)
	assert.Equal(t, 3, PublicShareQty(t, db, s.SellerUserID, s.ListingID))

	// I6: contract still VALID
	assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID))
}
