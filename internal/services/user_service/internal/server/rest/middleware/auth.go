package middleware

import (
	"context"
	"errors"

	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrNoIdentity   = errors.New("no identity in request context")
)

type Config struct {
	CookieName string `yaml:"cookie_name" env-default:"accessToken"`
}

type Identity struct {
	UserID uuid.UUID
	Role   model.Role
}

type TokenValidator interface {
	ValidateAccessToken(ctx context.Context, token string) (Identity, error)
}

type identityKey struct{}

func ValidateJWT(cfg Config, validator TokenValidator) fiber.Handler {
	if cfg.CookieName == "" {
		cfg.CookieName = "accessToken"
	}

	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies(cfg.CookieName)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing access token",
			})
		}

		identity, err := validator.ValidateAccessToken(c.UserContext(), tokenString)
		if errors.Is(err, ErrInvalidToken) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to validate token",
			})
		}

		c.Locals(identityKey{}, identity)

		return c.Next()
	}
}

func IdentityFrom(c *fiber.Ctx) (Identity, error) {
	identity, ok := c.Locals(identityKey{}).(Identity)
	if !ok {
		return Identity{}, ErrNoIdentity
	}

	return identity, nil
}

func RequireRole(roles ...model.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identity, err := IdentityFrom(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthenticated",
			})
		}

		for _, role := range roles {
			if identity.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient role",
		})
	}
}
