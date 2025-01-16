package utils

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	authrest "passvault/internal/http-server/middlewares/auth"
	"passvault/internal/lib/jwt"
)

func TestMiddleware(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	secret := "test_secret"
	mockToken := jwt.CreateMockToken(secret)

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mockToken))

	authMiddleware := authrest.New(slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	), secret)

	handler = authMiddleware(handler)

	handler.ServeHTTP(w, r)
}
