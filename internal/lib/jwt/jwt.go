package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type CustomClaims struct {
	AccountID int64  `json:"uid"`
	Email     string `json:"email"`
	Role      int32  `json:"role"`
	AppID     int32  `json:"app_id"`
	jwt.RegisteredClaims
}

// ParseToken parses and validates a JWT string using the given secret key.
func ParseToken(tokenString string, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token: token is not valid")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

func generateJTI() string {
	return uuid.New().String()
}

const duration = 1 * time.Hour

func CreateMockToken(secret string) string {
	claims := CustomClaims{
		AccountID: 123,
		Email:     "test@example.com",
		Role:      1,
		AppID:     1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			ID:        generateJTI(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}

	return tokenString
}
