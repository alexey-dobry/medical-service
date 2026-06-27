package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrFailedToCastType        = errors.New("failed to cast type")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

type TokenType string

const (
	AccessToken  TokenType = "ACCESS"
	RefreshToken TokenType = "REFRESH"
)

type Handler struct {
	accessSecret  string
	refreshSecret string

	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(cfg Config) *Handler {
	return &Handler{
		accessSecret:  cfg.AccessSecret,
		refreshSecret: cfg.RefreshSecret,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

func (h *Handler) GenerateToken(userID, role string, jwtType TokenType) (string, error) {
	TTL := h.getTTL(jwtType)
	secret := h.getSecret(jwtType)

	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(TTL)),
		},
		Role: role,
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("jwt - GenerateToken() - failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (h *Handler) ParseToken(tokenString string, jwtType TokenType) (string, string, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}

		return h.getSecret(jwtType), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("jwt - ParseToken() - failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", "", fmt.Errorf("jwt - ParseToken() - %w", ErrFailedToCastType)
	}
	sub, _ := claims.GetSubject()
	role, _ := claims.GetRole()

	return sub, role, nil
}

func (h *Handler) getSecret(jwtType TokenType) []byte {
	switch jwtType {
	case AccessToken:
		return []byte(h.accessSecret)
	case RefreshToken:
		return []byte(h.refreshSecret)
	default:
		return []byte{}
	}
}

func (h *Handler) getTTL(jwtType TokenType) time.Duration {
	switch jwtType {
	case AccessToken:
		return h.accessTTL
	case RefreshToken:
		return h.refreshTTL
	default:
		return time.Duration(0)
	}
}
