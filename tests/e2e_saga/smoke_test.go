//go:build e2e

package e2e_saga

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmoke verifies all 5 helpers work end-to-end against a live stack.
// Fold this into SG-01 in Phase 2, or delete once SG-01 covers the happy path.
//
// Run: go test -tags e2e -run TestSmoke -v ./tests/e2e_saga/
func TestSmoke(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})

	token := TokenFor(t, s.BuyerUserID, "CLIENT")

	httpStatus, resp := Exercise(t, s.ContractID, token, nil)
	require.Equal(t, http.StatusAccepted, httpStatus, "expect 202 Accepted")
	require.NotZero(t, resp.SagaID, "sagaId must be non-zero")

	exec := WaitForTerminal(t, db, resp.SagaID, 30*time.Second)
	assert.Equal(t, StatusCompleted, exec.Status, "happy path must complete")
	assert.Equal(t, StepCompleted, exec.CurrentStep)

	log := StepLog(t, db, resp.SagaID)
	require.Len(t, log, 5, "happy path: exactly 5 step log entries")
	wantSteps := []string{
		StepReserveFunds,
		StepReserveSecurities,
		StepTransferFunds,
		StepTransferOwnership,
		StepCompleted,
	}
	for i, entry := range log {
		assert.Equal(t, wantSteps[i], entry.Step, "log[%d].step", i)
		assert.Equal(t, "COMPLETED", entry.StepStatus, "log[%d].step_status", i)
		assert.Equal(t, 1, entry.Attempt, "log[%d].attempt", i)
	}

	// Invariant I3: no reserved funds remain
	_, buyerReserved := AccountBalances(t, db, s.BuyerAccountID)
	assert.Equal(t, 0.0, buyerReserved, "I3: buyer rezervisana_sredstva must be 0")

	// Invariant I6: contract is EXERCISED
	assert.Equal(t, "EXERCISED", ContractStatus(t, db, s.ContractID), "I6: contract must be EXERCISED")
}
