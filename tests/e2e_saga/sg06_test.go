//go:build e2e

package e2e_saga

// SG-06: F4 (TRANSFER_OWNERSHIP) forced failure via X-Saga-Force-Fail: F4.
// Compensation: C3 (return funds) → C2 (restore securities) → C1 (release reserved).
// Expected: FAILED, log=[RF✓,RS✓,TF✓,TO✗, TF_COMP, RS_COMP, RF_COMP].
//
// Run: SAGA_TEST_MODE=true go test -tags e2e -run TestSG06 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG06(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, map[string]string{
		"X-Saga-Force-Fail": "F4",
	})
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusFailed, exec.Status)
	// compensateFrom(TRANSFER_FUNDS) → final step = TRANSFER_FUNDS
	assert.Equal(t, StepTransferFunds, exec.CurrentStep)

	// Log: F1✓ F2✓ F3✓ F4✗ C3 C2 C1
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 7)

	assert.Equal(t, StepReserveFunds, logEntries[0].Step)
	assert.Equal(t, "COMPLETED", logEntries[0].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[1].Step)
	assert.Equal(t, "COMPLETED", logEntries[1].StepStatus)

	assert.Equal(t, StepTransferFunds, logEntries[2].Step)
	assert.Equal(t, "COMPLETED", logEntries[2].StepStatus)

	assert.Equal(t, StepTransferOwnership, logEntries[3].Step)
	assert.Equal(t, "FAILED", logEntries[3].StepStatus)

	assert.Equal(t, StepTransferFunds, logEntries[4].Step)
	assert.Equal(t, "COMPENSATED", logEntries[4].StepStatus)

	assert.Equal(t, StepReserveSecurities, logEntries[5].Step)
	assert.Equal(t, "COMPENSATED", logEntries[5].StepStatus)

	assert.Equal(t, StepReserveFunds, logEntries[6].Step)
	assert.Equal(t, "COMPENSATED", logEntries[6].StepStatus)

	// I1: funds restored — buyer gets money back (C3 ran)
	buyerBal, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 5000.0, buyerBal, 0.01, "I1: buyer balance restored")
	assert.Equal(t, 0.0, buyerRes, "I3: buyer reserved=0")

	// I2: seller shares restored by C2
	assert.Equal(t, 10, PublicShareQty(t, db, s.SellerUserID, s.ListingID))

	// I6: contract still VALID
	assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID))

	// C3 creates 2 rollback transactions (debit seller + credit buyer)
	assert.Equal(t, 2, CountTransactions(t, db, s.BuyerAccountID), "rollback credit + forward debit")
}
