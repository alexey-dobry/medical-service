package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenExpired  = errors.New("token expired")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrInvalidConfig = errors.New("invalid jwt config")
)

type TokenType string

const (
	AccessToken  TokenType = "ACCESS"
	RefreshToken TokenType = "REFRESH"
)

var signingMethod = jwtlib.SigningMethodHS256

type Pair struct {
	AccessToken   string
	RefreshToken  string
	AccessClaims  Claims
	RefreshClaims Claims
}

type Handler struct {
	accessSecret  []byte
	refreshSecret []byte

	accessTTL  time.Duration
	refreshTTL time.Duration

	now func() time.Time
}

type Option func(*Handler)

func WithClock(now func() time.Time) Option {
	return func(h *Handler) { h.now = now }
}

func New(cfg Config, opts ...Option) (*Handler, error) {
	if cfg.AccessSecret == "" || cfg.RefreshSecret == "" {
		return nil, fmt.Errorf("%w: secrets must not be empty", ErrInvalidConfig)
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, fmt.Errorf("%w: ttl must be positive", ErrInvalidConfig)
	}

	h := &Handler{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		now:           time.Now,
	}
	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

func (h *Handler) GenerateToken(userID string, role model.Role, jwtType TokenType) (string, Claims, error) {
	secret, ttl, err := h.params(jwtType)
	if err != nil {
		return "", Claims{}, err
	}

	now := h.now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			ID:        uuid.NewString(),
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
		Role: role,
		Type: jwtType,
	}

	tokenString, err := jwtlib.NewWithClaims(signingMethod, claims).SignedString(secret)
	if err != nil {
		return "", Claims{}, fmt.Errorf("jwt - GenerateToken() - failed to sign token: %w", err)
	}

	return tokenString, claims, nil
}

func (h *Handler) GeneratePair(userID string, role model.Role) (Pair, error) {
	access, accessClaims, err := h.GenerateToken(userID, role, AccessToken)
	if err != nil {
		return Pair{}, err
	}

	refresh, refreshClaims, err := h.GenerateToken(userID, role, RefreshToken)
	if err != nil {
		return Pair{}, err
	}

	return Pair{
		AccessToken:   access,
		RefreshToken:  refresh,
		AccessClaims:  accessClaims,
		RefreshClaims: refreshClaims,
	}, nil
}

func (h *Handler) ParseToken(tokenString string, jwtType TokenType) (Claims, error) {
	secret, _, err := h.params(jwtType)
	if err != nil {
		return Claims{}, err
	}

	token, err := jwtlib.ParseWithClaims(
		tokenString,
		&Claims{},
		func(*jwtlib.Token) (any, error) { return secret, nil },
		jwtlib.WithValidMethods([]string{signingMethod.Alg()}),
		jwtlib.WithExpirationRequired(),
		jwtlib.WithTimeFunc(h.now),
	)
	if errors.Is(err, jwtlib.ErrTokenExpired) {
		return Claims{}, ErrTokenExpired
	} else if err != nil {
		return Claims{}, fmt.Errorf("jwt - ParseToken() - %w: %w", ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, fmt.Errorf("jwt - ParseToken() - %w", ErrTokenInvalid)
	}

	if claims.Type != jwtType {
		return Claims{}, fmt.Errorf("jwt - ParseToken() - %w: expected %s token, got %q", ErrTokenInvalid, jwtType, claims.Type)
	}

	if claims.Subject == "" || claims.ID == "" {
		return Claims{}, fmt.Errorf("jwt - ParseToken() - %w: missing sub or jti", ErrTokenInvalid)
	}

	return *claims, nil
}

func (h *Handler) params(jwtType TokenType) ([]byte, time.Duration, error) {
	switch jwtType {
	case AccessToken:
		return h.accessSecret, h.accessTTL, nil
	case RefreshToken:
		return h.refreshSecret, h.refreshTTL, nil
	default:
		return nil, 0, fmt.Errorf("jwt - unknown token type %q", jwtType)
	}
}
