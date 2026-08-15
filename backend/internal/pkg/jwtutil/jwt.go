// Package jwtutil issues and parses JWT access tokens. Refresh tokens are
// opaque random strings handled by the auth module (stored hashed).
package jwtutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the JWT payload. school_id and roles[] are the primary
// authorization inputs for tenant scoping and RBAC.
type Claims struct {
	UserID      uuid.UUID `json:"sub"`
	SchoolID    uuid.UUID `json:"school_id"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions,omitempty"`
	DeviceID    string    `json:"device_id,omitempty"`
	jwt.RegisteredClaims
}

// Issuer issues and parses access tokens.
type Issuer struct {
	secret   []byte
	accessTTL time.Duration
}

// New creates an Issuer. In production use a strong secret (>= 32 bytes).
func New(secret string, accessTTL time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), accessTTL: accessTTL}
}

// Issue creates a signed access token for the given claims.
func (i *Issuer) Issue(userID, schoolID uuid.UUID, roles, permissions []string, deviceID string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(i.accessTTL)
	claims := Claims{
		UserID:      userID,
		SchoolID:    schoolID,
		Roles:       roles,
		Permissions: permissions,
		DeviceID:    deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        uuid.NewString(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, exp, nil
}

// Parse validates a token and returns its claims.
func (i *Issuer) Parse(raw string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt: unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}
	return claims, nil
}
