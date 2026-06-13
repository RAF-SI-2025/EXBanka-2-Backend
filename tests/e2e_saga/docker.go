//go:build e2e

package e2e_saga

import (
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

// composeFile returns the path to the infra docker-compose.yml.
// Override with E2E_COMPOSE env var.
func composeFile() string {
	return envOr("E2E_COMPOSE", "../EXBanka-2-Infrastructure/docker-compose.yml")
}

// Pause runs `docker compose pause <svc>`.
func Pause(t *testing.T, svc string) {
	t.Helper()
	runCompose(t, "pause", svc)
}

// Unpause runs `docker compose unpause <svc>`.
func Unpause(t *testing.T, svc string) {
	t.Helper()
	runCompose(t, "unpause", svc)
}

// Kill sends SIGKILL to the container: `docker compose kill -s KILL <svc>`.
func Kill(t *testing.T, svc string) {
	t.Helper()
	cmd := exec.Command("docker", "compose", "-f", composeFile(), "kill", "-s", "KILL", svc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Kill(%s): %v", svc, err)
	}
}

// Up runs `docker compose up -d <svc>`.
func Up(t *testing.T, svc string) {
	t.Helper()
	runCompose(t, "up", "-d", svc)
}

// Docker compose service names as used in docker-compose.yml.
const (
	SvcBankService = "bank-service"
	SvcUserService = "user-service"
	SvcDB          = "db"
)

// WaitForBankServiceReady polls GET /healthz until it returns 200 or timeout elapses.
// Use after docker compose up to confirm bank-service is accepting requests.
func WaitForBankServiceReady(t *testing.T, timeout time.Duration) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(BaseURL() + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("WaitForBankServiceReady: service not ready within %s", timeout)
}

func runCompose(t *testing.T, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"compose", "-f", composeFile()}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker compose %v: %v", args, err)
	}
}
