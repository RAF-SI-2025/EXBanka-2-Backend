//go:build e2e

package e2e_saga

// SG-10: Service paused mid-saga (during F3 delay).
//
// Architecture note: bank-service is both the orchestrator and executor.
// Pausing it freezes the SAGA goroutine mid-execution (SIGSTOP). On SIGCONT
// the goroutine resumes exactly where it left off — no state is lost.
//
// This test verifies goroutine resilience to SIGSTOP/SIGCONT:
//   - Inject 5s delay before F3 to create a pause window
//   - Start exercise
//   - Wait for F2 to complete (current_step=RESERVE_SECURITIES → delay starts)
//   - Pause bank-service (goroutine freezes inside the sleep)
//   - Verify DB state doesn't change while paused
//   - Unpause bank-service
//   - Verify SAGA completes successfully (F3..F5 run after resume)
//
// Run: SAGA_TEST_MODE=true go test -tags e2e -run TestSG10 -v ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG10(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, map[string]string{
		"X-Saga-Inject-Delay": "F3:5000", // 5s delay before TRANSFER_FUNDS
	})
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	// Wait for F2 to finish — current_step=RESERVE_SECURITIES means the 5s delay has started.
	WaitForStatusStep(t, db, resp.SagaID, StatusInProgress, StepReserveSecurities, 15*time.Second)

	// Pause bank-service: goroutine freezes inside time.Sleep(5s).
	paused := true
	Pause(t, SvcBankService)
	t.Cleanup(func() {
		if paused {
			Unpause(t, SvcBankService)
		}
	})

	// Verify saga is frozen (status/step shouldn't change for 1s).
	time.Sleep(1 * time.Second)
	execMid, _ := queryExecution(t, db, resp.SagaID)
	assert.Equal(t, StepReserveSecurities, execMid.CurrentStep, "saga should be frozen while paused")
	assert.Equal(t, StatusInProgress, execMid.Status, "saga should be frozen while paused")

	// Unpause — goroutine resumes the sleep then F3 runs normally.
	paused = false
	Unpause(t, SvcBankService)

	// SAGA should complete successfully after resume.
	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusCompleted, exec.Status, "SAGA should complete after unpause")
	assert.Equal(t, StepCompleted, exec.CurrentStep)

	// All 5 steps completed
	logEntries := StepLog(t, db, resp.SagaID)
	require.Len(t, logEntries, 5)
	for i, e := range logEntries {
		assert.Equal(t, "COMPLETED", e.StepStatus, "log[%d].step_status", i)
	}

	// All invariants
	buyerBal, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.InDelta(t, 2000.0, buyerBal, 0.01, "I1: buyer paid 3000")
	assert.Equal(t, 0.0, buyerRes, "I3: no stuck reservations")
	assert.Equal(t, 0, PublicShareQty(t, db, s.SellerUserID, s.ListingID), "I2: seller shares transferred")
	assert.Equal(t, "EXERCISED", ContractStatus(t, db, s.ContractID), "I6")
}
