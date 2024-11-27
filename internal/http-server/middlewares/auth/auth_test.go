package authrest

import (
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"passvault/internal/lib/jwt"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test_secret"
	token := jwt.CreateMockToken(secret)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	middleware := New(log, secret)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserClaimsKey).(*UserClaims)
		require.True(t, ok)
		require.Equal(t, int64(123), claims.AccountID)
		w.WriteHeader(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
