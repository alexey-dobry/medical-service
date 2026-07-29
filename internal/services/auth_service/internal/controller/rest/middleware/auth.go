package middleware

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

type errorResponse struct {
	Error string `json:"error"`
}

func AuthPair(jwtManager jwt.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		refreshToken := ctx.Cookies("refreshToken")
		accessToken := ctx.Cookies("accessToken")

		if refreshToken == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Error: "missing refresh token cookie"})
		}

		if accessToken == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Error: "missing access token cookie"})
		}

		refreshClaims, err := jwtManager.ParseToken(refreshToken, jwt.RefreshToken)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired refresh token"})
		}

		accessClaims, err := jwtManager.ParseToken(accessToken, jwt.AccessToken)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorResponse{Error: "invalid or expired access token"})
		}

		ctx.Locals("accessClaims", accessClaims)
		ctx.Locals("refreshClaims", refreshClaims)

		return ctx.Next()
	}
}
