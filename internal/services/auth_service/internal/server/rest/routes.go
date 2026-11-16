package rest

import "github.com/gofiber/fiber/v2"

func (s *RESTServer) initRoutes() {
	s.fiberApp.Post("/auth/login", s.handleLogin())
	s.fiberApp.Post("/auth/logout", s.handleLogout())
	s.fiberApp.Post("/auth/refresh", s.handleRefresh())
	s.fiberApp.Get("/healthz", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"status": "ok"})
	})
	s.fiberApp.Get("/readyz", s.handleReadinessCheck())
}
