package rest

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/health"
	"github.com/gofiber/fiber/v2"
)

func (s *RESTServer) handleReadinessCheck() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !s.healthChecker.Ready() {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"status": "not ready"},
			)
		}

		ok, results := s.healthChecker.Run(ctx.UserContext())
		code := fiber.StatusOK
		if !ok {
			code = fiber.StatusServiceUnavailable
		}

		return ctx.Status(code).JSON(fiber.Map{"status": health.StatusText(ok, results), "checks": results})
	}
}
