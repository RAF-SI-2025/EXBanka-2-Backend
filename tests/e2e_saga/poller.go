//go:build e2e

package e2e_saga

import (
	"database/sql"
	"testing"
	"time"
)

// SAGA status constants — mirrors domain.OTCSagaStatus (duplicated to avoid internal/ import).
const (
	StatusInProgress         = "IN_PROGRESS"
	StatusCompleted          = "COMPLETED"
	StatusFailed             = "FAILED"
	StatusCompensating       = "COMPENSATING"
	StatusCompensationFailed = "COMPENSATION_FAILED"
)

// SAGA step constants — mirrors domain.OTCSagaStep.
const (
	StepPending           = "PENDING"
	StepReserveFunds      = "RESERVE_FUNDS"
	StepReserveSecurities = "RESERVE_SECURITIES"
	StepTransferFunds     = "TRANSFER_FUNDS"
	StepTransferOwnership = "TRANSFER_OWNERSHIP"
	StepCompleted         = "COMPLETED"
)

// Execution mirrors a row from otc_saga_executions.
type Execution struct {
	ID                  int64
	ContractID          int64
	CurrentStep         string
	Status              string
	BuyerReservedAmount float64
	ErrorMessage        sql.NullString
	RetryCount          int
	InitiatedBy         int64
}

// StepLogEntry mirrors a row from otc_saga_step_log.
type StepLogEntry struct {
	ID          int64
	ExecutionID int64
	Step        string
	StepStatus  string
	ErrorMsg    sql.NullString
	Attempt     int
}

// WaitForTerminal polls otc_saga_executions until the SAGA reaches a terminal
// status (COMPLETED, FAILED, COMPENSATION_FAILED) or timeout elapses.
// Fails the test on timeout — enforces invariant I5.
func WaitForTerminal(t *testing.T, db *sql.DB, sagaID int64, timeout time.Duration) Execution {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if exec, ok := queryExecution(t, db, sagaID); ok && isTerminal(exec.Status) {
			return exec
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("WaitForTerminal: saga %d did not reach terminal status within %s", sagaID, timeout)
	return Execution{}
}

// WaitForStatusStep polls until the SAGA has the given status+step pair.
// Used for mid-flight coordination (e.g. pause a service while saga is between steps).
func WaitForStatusStep(t *testing.T, db *sql.DB, sagaID int64, status, step string, timeout time.Duration) Execution {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if exec, ok := queryExecution(t, db, sagaID); ok && exec.Status == status && exec.CurrentStep == step {
			return exec
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("WaitForStatusStep: saga %d did not reach status=%s step=%s within %s", sagaID, status, step, timeout)
	return Execution{}
}

// StepLog returns all step log entries for sagaID ordered by insertion (id ASC).
func StepLog(t *testing.T, db *sql.DB, sagaID int64) []StepLogEntry {
	t.Helper()
	rows, err := db.Query(`
		SELECT id, execution_id, step, step_status, error_msg, attempt
		FROM core_banking.otc_saga_step_log
		WHERE execution_id = $1
		ORDER BY id
	`, sagaID)
	if err != nil {
		t.Fatalf("StepLog(%d): query: %v", sagaID, err)
	}
	defer rows.Close()

	var entries []StepLogEntry
	for rows.Next() {
		var e StepLogEntry
		if err := rows.Scan(&e.ID, &e.ExecutionID, &e.Step, &e.StepStatus, &e.ErrorMsg, &e.Attempt); err != nil {
			t.Fatalf("StepLog(%d): scan: %v", sagaID, err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("StepLog(%d): rows: %v", sagaID, err)
	}
	return entries
}

func queryExecution(t *testing.T, db *sql.DB, sagaID int64) (Execution, bool) {
	t.Helper()
	var e Execution
	err := db.QueryRow(`
		SELECT id, contract_id, current_step, status,
		       buyer_reserved_amount, error_message, retry_count, initiated_by
		FROM core_banking.otc_saga_executions
		WHERE id = $1
	`, sagaID).Scan(
		&e.ID, &e.ContractID, &e.CurrentStep, &e.Status,
		&e.BuyerReservedAmount, &e.ErrorMessage, &e.RetryCount, &e.InitiatedBy,
	)
	if err == sql.ErrNoRows {
		return Execution{}, false
	}
	if err != nil {
		t.Fatalf("queryExecution(%d): %v", sagaID, err)
	}
	return e, true
}

func isTerminal(status string) bool {
	switch status {
	case StatusCompleted, StatusFailed, StatusCompensationFailed:
		return true
	}
	return false
}
