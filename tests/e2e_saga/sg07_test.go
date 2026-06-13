//go:build e2e

package e2e_saga

// SG-07: F5 (COMPLETED/markCompleted) forced failure via X-Saga-Force-Fail: F5.
// Compensation: C4 → C3 → C2 → C1.
// Expected: FAILED, log=[RF✓,RS✓,TF✓,TO✓,COMPLETED✗,TO_COMP,TF_COMP,RS_COMP,RF_COMP].
//
// Run: SAGA_TEST_MODE=true go test -tags e2e -run TestSG07 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG07(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, map[string]string{
		"X-Saga-Force-Fail": "F5",
	})
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusFailed, exec.Status)
	// compensateFrom(TRANSFER_OWNERSHIP) → final step = TRANSFER_OWNERSHIP
	assert.Equal(t, StepTransferOwnership, exec.CurrentStep)

	// Log: F1✓ F2✓ F3✓ F4✓ F5✗ C4 C3 C2 C1 — 9 entries
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 9)

	assert.Equal(t, StepReserveFunds, logEntries[0].Step)
	assert.Equal(t, "COMPLETED", logEntries[0].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[1].Step)
	assert.Equal(t, "COMPLETED", logEntries[1].StepStatus)

	assert.Equal(t, StepTransferFunds, logEntries[2].Step)
	assert.Equal(t, "COMPLETED", logEntries[2].StepStatus)

	assert.Equal(t, StepTransferOwnership, logEntries[3].Step)
	assert.Equal(t, "COMPLETED", logEntries[3].StepStatus)

	assert.Equal(t, StepCompleted, logEntries[4].Step)
	assert.Equal(t, "FAILED", logEntries[4].StepStatus)

	assert.Equal(t, StepTransferOwnership, logEntries[5].Step)
	assert.Equal(t, "COMPENSATED", logEntries[5].StepStatus)

	assert.Equal(t, StepTransferFunds, logEntries[6].Step)
	assert.Equal(t, "COMPENSATED", logEntries[6].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[7].Step)
	assert.Equal(t, "COMPENSATED", logEntries[7].StepStatus)

	assert.Equal(t, StepReserveFunds, logEntries[8].Step)
	assert.Equal(t, "COMPENSATED", logEntries[8].StepStatus)

	// I1: funds restored
	buyerBal, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 5000.0, buyerBal, 0.01, "I1: buyer balance restored")
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0")

	// I2: seller shares restored by C2
	assert.Equal(t, 10, PublicShareQty(t, db, s.SellerUserID, s.ListingID))

	// I6: contract still VALID (F5 failed, never EXERCISED)
	assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID))
}
