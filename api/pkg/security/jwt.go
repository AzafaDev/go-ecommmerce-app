package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claims struct {
	ID   uuid.UUID `json:"id"`
	Role string    `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(secretKey string, expiry time.Duration, userID uuid.UUID, userRole string) (string, error) {
	claims := Claims{
		ID:   userID,
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}

func ParseToken(secretKey string, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("method is invalid")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error in parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("error in converting to claim: %w", err)
	}

	return claims, nil
}
