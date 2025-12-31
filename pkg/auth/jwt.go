package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	jwt.RegisteredClaims
}

type JwtIssuer struct{}

func verifyJWT(tokenStr string, key []byte) (*claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return key, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (JwtIssuer) Issue(playerID string, key []byte) (string, error) {
	claims := &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   playerID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}
