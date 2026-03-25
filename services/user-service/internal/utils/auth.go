package utils

import auth "banka-backend/shared/auth"

// AccessClaims is a type alias for auth.AccessClaims.
// Exported here so service-layer code can reference it without importing shared/auth directly.
type AccessClaims = auth.AccessClaims

// GenerateTokens wraps auth.GenerateTokens for use within the user-service.
func GenerateTokens(userID, email, userType string, permissions []string, accessSecret, refreshSecret string) (string, string, error) {
	return auth.GenerateTokens(userID, email, userType, permissions, accessSecret, refreshSecret)
}
