package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type JWTHandler interface {
	ExpiresAt() TTL

	GenerateJWTPair(claims Claims) (refreshToken, accessToken string, err error)

	ValidateJWT(token string, jwtType TokenType) (Claims, error)
}

type jwtHandler struct {
	access_secret  []byte
	refresh_secret []byte
	ttl            TTL
}

func NewHandler(cfg Config) (JWTHandler, error) {
	handler := jwtHandler{
		access_secret:  []byte(cfg.AccessSecret),
		refresh_secret: []byte(cfg.RefreshSecret),
		ttl: TTL{
			AccessTTL:  cfg.TTL.AccessTTL,
			RefreshTTL: cfg.TTL.RefreshTTL,
		},
	}

	return &handler, nil
}

func (h *jwtHandler) GenerateJWT(claims Claims, jwtType TokenType) (string, error) {
	var secret []byte
	var TTL time.Duration
	if jwtType == AccessToken {
		secret = h.access_secret
		TTL, _ = time.ParseDuration(h.ttl.AccessTTL)
	} else {
		secret = h.refresh_secret
		TTL, _ = time.ParseDuration(h.ttl.RefreshTTL)
	}

	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(TTL))

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", ErrFailedToGenerateJWT
	}

	return token, nil
}

func (h *jwtHandler) ExpiresAt() TTL {
	return h.ttl
}

func (h *jwtHandler) GenerateJWTPair(claims Claims) (refreshToken, accessToken string, err error) {
	refreshToken, err = h.GenerateJWT(Claims{ID: claims.ID}, RefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate refresh token: %s", err)
	}

	accessToken, err = h.GenerateJWT(claims, AccessToken)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate access token: %s", err)
	}

	return refreshToken, accessToken, nil
}

func (h *jwtHandler) ValidateJWT(token string, jwtType TokenType) (Claims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		var secret []byte
		if jwtType == AccessToken {
			secret = h.access_secret
		} else {
			secret = h.refresh_secret
		}
		return secret, nil
	})
	if errors.Is(err, jwt.ErrTokenExpired) {
		return Claims{}, ErrJWTTokenExpired
	} else if err != nil {
		return Claims{}, fmt.Errorf("error parsing jwt with claims: %w", err)
	}

	if claims, ok := jwtToken.Claims.(*Claims); ok && jwtToken.Valid {
		return *claims, nil
	}

	return Claims{}, ErrSignatureInvalid
}
