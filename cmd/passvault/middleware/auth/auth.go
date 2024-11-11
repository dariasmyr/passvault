package logger

import (
	"net/http"
	"strings"
	"errors"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrFailedIsAdminCheck = errors.New("failed to check if user is admin")
)

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}

func New(
	log *slog.Logger,
	appSecret string
) func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr == "" {
			// It's ok, if user is not authorized
			next.ServeHTTP(w, r)
			return
		}

		claims, err := jwt.Parse(tokenStr, appSecret)

		if err != nil {
			log.Warn("failed to parse token", sl.Err(err))

			// But if token is invalid, we shouldn't handle request
			ctx := context.WithValue(r.Context(), errorKey, ErrInvalidToken)
			next.ServeHTTP(w, r.WithContext(ctx))

			return
		}

		log.Info("user authorized", slog.Any("claims", claims))

        // TODO: Add parsed user data from context
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}