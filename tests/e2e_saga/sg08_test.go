//go:build e2e

package e2e_saga

// SG-08: Compensator fails N times then succeeds.
// Setup: force F3 fail + force C2 to fail 1 time before succeeding.
// Expected: FAILED, log includes 2 C2 entries (1st COMPENSATION_FAILED, 2nd COMPENSATED).
//
// Run: SAGA_TEST_MODE=true go test -tags e2e -run TestSG08 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG08(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, map[string]string{
		"X-Saga-Force-Fail":           "F3",
		"X-Saga-Compensate-Fail":      "C2",
		"X-Saga-Compensate-Fail-Times": "1",
	})
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusFailed, exec.Status)
	assert.Equal(t, StepReserveSecurities, exec.CurrentStep)

	// Log: F1✓ F2✓ F3✗ C2-fail C2-ok C1 — 6 entries
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 6)

	assert.Equal(t, StepReserveFunds, logEntries[0].Step)
	assert.Equal(t, "COMPLETED", logEntries[0].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[1].Step)
	assert.Equal(t, "COMPLETED", logEntries[1].StepStatus)

	assert.Equal(t, StepTransferFunds, logEntries[2].Step)
	assert.Equal(t, "FAILED", logEntries[2].StepStatus)

	// C2 attempt 1: fails
	assert.Equal(t, StepReserveSecurities, logEntries[3].Step)
	assert.Equal(t, "COMPENSATION_FAILED", logEntries[3].StepStatus)
	assert.Equal(t, 1, logEntries[3].Attempt)

	// C2 attempt 2: succeeds
	assert.Equal(t, StepReserveSecurities, logEntries[4].Step)
	assert.Equal(t, "COMPENSATED", logEntries[4].StepStatus)
	assert.Equal(t, 2, logEntries[4].Attempt)

	assert.Equal(t, StepReserveFunds, logEntries[5].Step)
	assert.Equal(t, "COMPENSATED", logEntries[5].StepStatus)

	// I1, I2, I3 — all state restored
	buyerBal, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 5000.0, buyerBal, 0.01, "I1: buyer balance unchanged")
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0")
	assert.Equal(t, 10, PublicShareQty(t, db, s.SellerUserID, s.ListingID), "I2: seller shares restored")
	assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID), "I6: contract still VALID")
}
