package v1

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/controller/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func NewRoutes(routesV1Group fiber.Router, jwtManager *jwt.Handler, a usecase.Auth, l logger.Logger, v *validator.Validate) *V1 {
	r := &V1{
		a: a,
		l: l,
		v: v,
	}

	authGroup := routesV1Group.Group("/auth")
	{
		authGroup.Post("/login", r.login)
		authGroup.Post("/logout", r.logout)
	}

	protected := authGroup.Group("", middleware.AuthPair(*jwtManager))
	{
		protected.Post("/refresh", r.refresh)
	}

	return r
}

// func (s *V1) initRoutes() {
// 	s.fiberApp.Post("/auth/login", s.handleLogin())
// 	s.fiberApp.Post("/auth/logout", s.handleLogout())
// 	s.fiberApp.Post("/auth/refresh", s.handleRefresh())
// }
