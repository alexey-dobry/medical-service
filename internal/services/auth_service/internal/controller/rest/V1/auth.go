package v1

import (
	"errors"

	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/controller/rest/V1/request"
	"github.com/gofiber/fiber/v2"
)

func (r *V1) login(ctx *fiber.Ctx) error {
	var body request.Register

	err := ctx.BodyParser(&body)
	if err != nil {
		r.l.Errorf("Failed to parse request body: %w", err)

		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	if err = r.v.Struct(body); err != nil {
		r.l.Errorf("Validation failed for request %s : %w", ctx.Locals("requestid"), err)

		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	refreshToken, accessToken, err := r.a.Login(ctx.Context(), body.Email, body.Password)

	if errors.Is(err, jwt.ErrUnexpectedSigningMethod) {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	} else if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/auth/refresh",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})

	return ctx.SendStatus(fiber.StatusOK)
}

func (r *V1) logout(ctx *fiber.Ctx) error {
	refreshClaims := ctx.Locals("refreshClaims").(*jwt.Claims)
	accessClaims := ctx.Locals("accessClaims").(*jwt.Claims)

	if err := r.a.Logout(ctx.Context(), refreshClaims, accessClaims); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:   "accessToken",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	ctx.Cookie(&fiber.Cookie{
		Name:   "refreshToken",
		Value:  "",
		MaxAge: -1,
		Path:   "/auth/refresh",
	})

	return ctx.SendStatus(fiber.StatusOK)
}

func (r *V1) refresh(ctx *fiber.Ctx) error {
	refreshClaims := ctx.Locals("refreshClaims").(*jwt.Claims)
	accessClaims := ctx.Locals("accessClaims").(*jwt.Claims)

	accessToken, err := r.a.Refresh(ctx.Context(), refreshClaims, accessClaims)
	if errors.Is(err, jwt.ErrUnexpectedSigningMethod) {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	} else if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
	})

	return ctx.SendStatus(fiber.StatusOK)
}
