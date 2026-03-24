package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	auth "banka-backend/shared/auth"
)

const (
	interceptorAccessSecret  = "test-access-secret-32-chars-long!"
	interceptorRefreshSecret = "test-refresh-secret-32-chars-lon!"
)

// publicMethod and protectedMethod are stand-in RPC names.
const (
	publicMethod    = "/user.UserService/Login"
	protectedMethod = "/user.UserService/GetAllEmployees"
)

// runInterceptor invokes AuthInterceptor.Unary() for the given full-method path.
func runInterceptor(ctx context.Context, fullMethod, secret string, publicMethods []string) (interface{}, error) {
	ai := auth.NewAuthInterceptor(secret, publicMethods)
	info := &grpc.UnaryServerInfo{FullMethod: fullMethod}

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	resp, err := ai.Unary()(ctx, nil, info, handler)
	_ = capturedCtx
	return resp, err
}

// makeToken creates a valid access token signed with interceptorAccessSecret.
func makeToken(userID, email, userType string, permissions []string) string {
	tok, _ := auth.GenerateAccessToken(userID, email, userType, permissions, interceptorAccessSecret)
	return tok
}

// bearerCtx creates a gRPC incoming context with Authorization: Bearer <token>.
func bearerCtx(token string) context.Context {
	md := metadata.Pairs("authorization", "Bearer "+token)
	return metadata.NewIncomingContext(context.Background(), md)
}

// ─── Public methods bypass auth ───────────────────────────────────────────────

func TestAuthInterceptor_PublicMethod_Bypass(t *testing.T) {
	tests := []string{
		"/user.UserService/Login",
		"/user.UserService/HealthCheck",
		"/user.UserService/SetPassword",
		"/user.UserService/ActivateAccount",
		"/user.UserService/RefreshToken",
	}

	for _, method := range tests {
		t.Run(method, func(t *testing.T) {
			resp, err := runInterceptor(
				context.Background(), method,
				interceptorAccessSecret,
				[]string{method},
			)
			require.NoError(t, err)
			assert.Equal(t, "ok", resp)
		})
	}
}

// ─── Protected method: valid token ────────────────────────────────────────────

func TestAuthInterceptor_ValidToken_InjectsClaimsInContext(t *testing.T) {
	token := makeToken("42", "user@test.com", "EMPLOYEE", []string{"VIEW_ACCOUNTS"})
	ctx := bearerCtx(token)

	ai := auth.NewAuthInterceptor(interceptorAccessSecret, nil)
	info := &grpc.UnaryServerInfo{FullMethod: protectedMethod}

	var gotClaims *auth.AccessClaims
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		claims, ok := auth.ClaimsFromContext(ctx)
		require.True(t, ok)
		gotClaims = claims
		return "ok", nil
	}

	resp, err := ai.Unary()(ctx, nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
	require.NotNil(t, gotClaims)
	assert.Equal(t, "42", gotClaims.Subject)
	assert.Equal(t, "EMPLOYEE", gotClaims.UserType)
	assert.Equal(t, []string{"VIEW_ACCOUNTS"}, gotClaims.Permissions)
}

// ─── Protected method: error cases ────────────────────────────────────────────

func TestAuthInterceptor_Errors(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		wantCode codes.Code
	}{
		{
			name:     "no metadata",
			ctx:      context.Background(),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "missing authorization header",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs()),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "non-bearer scheme",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic dXNlcjpwYXNz")),
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "malformed token",
			ctx:      metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer not.a.real.token")),
			wantCode: codes.Unauthenticated,
		},
		{
			name: "refresh token used as access token",
			ctx: func() context.Context {
				_, refresh, _ := auth.GenerateTokens("1", "a@b.com", "EMPLOYEE", nil,
					interceptorAccessSecret, interceptorRefreshSecret)
				return bearerCtx(refresh)
			}(),
			wantCode: codes.Unauthenticated,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runInterceptor(tc.ctx, protectedMethod, interceptorAccessSecret, nil)
			require.Error(t, err)
			s, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tc.wantCode, s.Code())
		})
	}
}

// ─── ClaimsFromContext on empty context ───────────────────────────────────────

func TestClaimsFromContext_NoClaims(t *testing.T) {
	claims, ok := auth.ClaimsFromContext(context.Background())
	assert.False(t, ok)
	assert.Nil(t, claims)
}

// ─── NewContextWithClaims ─────────────────────────────────────────────────────

func TestNewContextWithClaims_RoundTrip(t *testing.T) {
	injected := &auth.AccessClaims{
		Email:    "test@example.com",
		UserType: "ADMIN",
	}
	ctx := auth.NewContextWithClaims(context.Background(), injected)
	got, ok := auth.ClaimsFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, injected.Email, got.Email)
	assert.Equal(t, injected.UserType, got.UserType)
}
