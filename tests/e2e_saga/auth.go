//go:build e2e

package e2e_saga

import (
	"fmt"
	"strconv"
	"testing"

	sharedauth "banka-backend/shared/auth"
)

// Secret returns the JWT access secret for signing test tokens.
// Reads JWT_ACCESS_SECRET env var; falls back to the docker-compose default.
func Secret() string {
	return envOr("JWT_ACCESS_SECRET", "super_secret_jwt_access_key")
}

// TokenFor generates a valid access JWT for the given userID and userType.
// For the exercise endpoint, userType should be "CLIENT".
// Use a different userID than the contract buyer to get a "not the buyer" token (SG-02a).
func TokenFor(t *testing.T, userID int64, userType string) string {
	t.Helper()
	tok, err := sharedauth.GenerateAccessToken(
		strconv.FormatInt(userID, 10),
		fmt.Sprintf("e2e-user-%d@test.local", userID),
		userType,
		nil,
		Secret(),
	)
	if err != nil {
		t.Fatalf("TokenFor(%d): %v", userID, err)
	}
	return tok
}
