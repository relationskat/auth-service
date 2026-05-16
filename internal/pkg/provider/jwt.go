package provider

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type claims struct {
	ID       string `json:"id"`
	IsAccess bool   `json:"is_access"`
	jwt.RegisteredClaims
}

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

func (p *Provider) NewJwt(id uuid.UUID, role string, isAccess bool) (string, error) {
	expirationTime := 24 * time.Hour
	if !isAccess {
		expirationTime = 30 * 24 * time.Hour
	}

	claims := &claims{
		ID:       id.String(),
		IsAccess: isAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(p.privateKey)
}

func (p *Provider) parseJwt(tokenStr string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{},
		error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return p.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func (p *Provider) ValidateToken(tokenStr string, mustBeAccess bool) (*claims, error) {
	c, err := p.parseJwt(tokenStr)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		default:
			return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
		}
	}

	exp, err := c.GetExpirationTime()
	if err != nil || exp == nil {
		return nil, ErrTokenInvalid
	}
	if time.Now().After(exp.Time) {
		return nil, ErrTokenExpired
	}

	if c.IsAccess != mustBeAccess {
		return nil, ErrTokenInvalid
	}

	return c, nil
}
