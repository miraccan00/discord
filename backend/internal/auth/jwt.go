package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when a token fails verification.
var ErrInvalidToken = errors.New("invalid token")

// Issuer signs and verifies HS256 JWTs for authenticated users.
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

// NewIssuer builds an Issuer from a signing secret and token lifetime.
func NewIssuer(secret string, ttlMinutes int) *Issuer {
	return &Issuer{
		secret: []byte(secret),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

// Issue returns a signed token whose subject is the given username.
func (i *Issuer) Issue(username string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

// Verify parses and validates a token, returning its subject (username).
func (i *Issuer) Verify(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return i.secret, nil
	})
	if err != nil || claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
