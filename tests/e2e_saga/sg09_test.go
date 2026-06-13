//go:build e2e

package e2e_saga

// SG-09: Infrastructure failure before/during F1.
//
// SG-09a: docker compose pause bank-service before the HTTP request.
//         bank-service's HTTP server is frozen (SIGSTOP); connection times out.
//         Expected: HTTP timeout error, no saga execution created in DB.
//
// SG-09b/c: Toxiproxy (latency > RPC timeout / partition). SKIPPED —
//           Toxiproxy is not present in the current docker-compose setup.
//           Add toxiproxy service to docker-compose.yml to enable these tests.
//
// Run: go test -tags e2e -run TestSG09 -v ./tests/e2e_saga/

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSG09a_ServicePausedBeforeRequest(t *testing.T) {
	db := Connect(t)
	defer db.Close()

	Truncate(t, db)
	s := SeedReadyContract(t, db, SeedOpts{})
	token := TokenFor(t, s.BuyerUserID, "CLIENT")

	paused := true
	Pause(t, SvcBankService)
	t.Cleanup(func() {
		if paused {
			Unpause(t, SvcBankService)
		}
	})

	// Use 3s timeout — bank is frozen, so the request will hang.
	client := &http.Client{Timeout: 3 * time.Second}
	_, _, err := ExerciseTry(t, s.ContractID, token, nil, client)

	// Service is paused — expect timeout or connection error.
	require.Error(t, err, "paused service should cause request error/timeout")

	// Unpause so DB checks below can observe any queued request processing.
	paused = false
	Unpause(t, SvcBankService)

	// Give the service a moment to process any queued connection (if any).
	time.Sleep(300 * time.Millisecond)

	// No saga should have been created (request never reached the handler).
	var sagaID int64
	scanErr := db.QueryRow(`
		SELECT id FROM core_banking.otc_saga_executions WHERE contract_id = $1
	`, s.ContractID).Scan(&sagaID)
	if scanErr == nil {
		// A saga was created (service processed the request after unpause).
		// Wait for it to reach terminal status and verify no side effects.
		t.Logf("SG-09a: service processed queued request after unpause (sagaID=%d), waiting for terminal", sagaID)
		exec := WaitForTerminal(t, db, sagaID, 15*time.Second)
		assert.Equal(t, StatusFailed, exec.Status, "if saga ran, it should fail (F1 should fail due to infra error)")
		_, buyerRes := AccountBalances(t, db, s.BuyerAccountID)
		assert.Equal(t, 0.0, buyerRes, "I3: no stuck reservations")
	} else {
		assert.Equal(t, sql.ErrNoRows, scanErr, "no saga created when service was paused")
	}
}

func TestSG09b_ToxiproxyLatency(t *testing.T) {
	t.Skip("SG-09b: Toxiproxy not in docker-compose.yml — add toxiproxy service to enable")
}

func TestSG09c_ToxiproxyPartition(t *testing.T) {
	t.Skip("SG-09c: Toxiproxy not in docker-compose.yml — add toxiproxy service to enable")
}
