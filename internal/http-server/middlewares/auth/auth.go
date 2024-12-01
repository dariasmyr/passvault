package authrest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"passvault/internal/lib/jwt"
	"passvault/internal/lib/logger/sl"
	"time"
)

type UserClaims struct {
	AccountID int64
	Email     string
	Role      int32
	AppID     int32
}

type UserClaimsKeyType struct{}

var UserClaimsKey = UserClaimsKeyType{}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

func New(log *slog.Logger, appSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				// Allow unauthenticated users to proceed
				next.ServeHTTP(w, r)
				return
			}

			// Parse and validate token
			claims, err := jwt.ParseToken(tokenStr, appSecret)
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))

				// Handle invalid token by setting error in context and halting request
				ctx := context.WithValue(r.Context(), ErrInvalidToken, err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				next.ServeHTTP(w, r.WithContext(ctx)) // End request if token is invalid
				return
			}

			if claims.ExpiresAt.Before(time.Now()) {
				log.Warn("token expired", sl.Err(err))

				ctx := context.WithValue(r.Context(), ErrTokenExpired, ErrTokenExpired)
				http.Error(w, "Token expired", http.StatusUnauthorized)
				next.ServeHTTP(w, r.WithContext(ctx)) // End request if token is expired
				return
			}

			log.Info("user authorized", slog.Any("claims", claims))

			userClaims := &UserClaims{
				AccountID: claims.AccountID,
				Email:     claims.Email,
				Role:      claims.Role,
				AppID:     claims.AppID,
			}

			// Inject claims into the request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, userClaims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	if !ok {
		return nil, errors.New("user claims not found in context")
	}
	return claims, nil
}

func GetAuthErrorFromContext(ctx context.Context) error {
	if err := ctx.Value(ErrInvalidToken); err != nil {
		return ErrInvalidToken
	}
	if err := ctx.Value(ErrTokenExpired); err != nil {
		return ErrTokenExpired
	}
	return nil
}
