//go:build e2e

package e2e_saga

// SG-11: Coordinator (bank-service) killed mid-saga, restart triggers recovery.
//
// Architecture note: bank-service is both the orchestrator and executor.
// Killing it with SIGKILL terminates the SAGA goroutine mid-execution.
// On restart, RecoverInProgressSagas() finds the IN_PROGRESS execution and
// resumes it from the last committed step (exec.CurrentStep in DB).
//
// Test flow:
//   - Inject 5s delay before F3 to create a kill window
//   - Start exercise
//   - Wait for current_step=RESERVE_SECURITIES (F2 done, F3 delay started)
//   - Kill bank-service (SIGKILL)
//   - Restart bank-service (docker compose up -d)
//   - Wait for bank-service to be ready
//   - Recovery runs: resumes from F3, runs F3..F5
//   - Expected: COMPLETED, all invariants satisfied
//
// Run: SAGA_TEST_MODE=true go test -tags e2e -run TestSG11 -v -timeout 120s ./tests/e2e_saga/

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG11(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")
	httpStatus, resp := Exercise(t, s.ContractID, token, map[string]string{
		"X-Saga-Inject-Delay": "F3:5000", // 5s delay opens kill window
	})
	require.Equal(t, http.StatusAccepted, httpStatus)
	require.NotZero(t, resp.SagaID)

	// Wait for F2 to complete and F3 delay to start.
	WaitForStatusStep(t, db, resp.SagaID, StatusInProgress, StepReserveSecurities, 15*time.Second)

	// Kill bank-service while SAGA goroutine is sleeping inside F3 delay.
	t.Logf("SG-11: killing bank-service (sagaID=%d, step=RESERVE_SECURITIES)", resp.SagaID)
	Kill(t, SvcBankService)

	// Restart bank-service; RecoverInProgressSagas runs on startup.
	Up(t, SvcBankService)
	WaitForBankServiceReady(t, 60*time.Second)
	t.Logf("SG-11: bank-service restarted, waiting for SAGA recovery")

	// Recovery resumes F3 (without delay — no fault headers), then F4 and F5.
	// Expect COMPLETED since kill happened before F3's actual DB work.
	exec := WaitForTerminal(t, db, resp.SagaID, 60*time.Second)

	// Both COMPLETED and FAILED are valid per spec (depends on kill timing).
	// In our setup (kill during 5s delay, before F3 DB work) → deterministically COMPLETED.
	assert.Equal(t, StatusCompleted, exec.Status, "recovery should complete the saga")

	// I3: no stuck reservations
	_, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
	assert.Equal(t, 0.0, buyerRes, "I3: no stuck reserved funds after recovery")

	if exec.Status == StatusCompleted {
		assert.Equal(t, StepCompleted, exec.CurrentStep)
		buyerBal, _ := AccountBalances(t, db, s.BuyerAccountID)
		assert.InDelta(t, 2000.0, buyerBal, 0.01, "I1: buyer paid 3000")
		assert.Equal(t, 0, PublicShareQty(t, db, s.SellerUserID, s.ListingID), "I2: shares transferred")
		assert.Equal(t, "EXERCISED", ContractStatus(t, db, s.ContractID), "I6")
	} else {
		// FAILED path: all compensators should have run, no stuck state
		assert.Equal(t, "VALID", ContractStatus(t, db, s.ContractID), "I6: VALID if not completed")
		assert.Equal(t, 10, PublicShareQty(t, db, s.SellerUserID, s.ListingID), "I2: shares restored")
	}
}
