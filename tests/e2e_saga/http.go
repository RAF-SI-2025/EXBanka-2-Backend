//go:build e2e

package e2e_saga

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

// BaseURL returns the bank-service base URL.
// Override with E2E_BANK_URL env var.
func BaseURL() string {
	return envOr("E2E_BANK_URL", "http://localhost:8083")
}

// ExerciseResponse is the JSON body returned by POST /api/otc/contracts/{id}/execute.
type ExerciseResponse struct {
	Message    string `json:"message"`
	SagaID     int64  `json:"sagaId"`
	ContractID int64  `json:"contractId"`
	Status     string `json:"status"`
}

// ErrorResponse is the JSON body returned on 4xx/5xx responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Exercise sends POST /api/otc/contracts/{id}/execute.
// token is placed in Authorization: Bearer <token>; pass "" to omit.
// headers accepts arbitrary extra headers — pass X-Saga-* fault-injection headers here.
// Returns (httpStatus, decoded ExerciseResponse). Fatals on network error.
func Exercise(t *testing.T, contractID int64, token string, headers map[string]string) (int, ExerciseResponse) {
	t.Helper()
	status, resp, err := ExerciseTry(t, contractID, token, headers, http.DefaultClient)
	if err != nil {
		t.Fatalf("Exercise: %v", err)
	}
	return status, resp
}

// ExerciseTry is like Exercise but returns a non-nil error on network/timeout failures
// instead of fataling. Pass a custom *http.Client to set timeouts (e.g. SG-09a).
func ExerciseTry(t *testing.T, contractID int64, token string, headers map[string]string, client *http.Client) (int, ExerciseResponse, error) {
	t.Helper()

	url := fmt.Sprintf("%s/api/otc/contracts/%d/execute", BaseURL(), contractID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte{}))
	if err != nil {
		return 0, ExerciseResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, ExerciseResponse{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var out ExerciseResponse
	_ = json.Unmarshal(body, &out)
	return resp.StatusCode, out, nil
}
