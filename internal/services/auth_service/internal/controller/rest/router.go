package rest

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	v1 "github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/controller/rest/V1"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func NewRouter(
	app *fiber.App,
	a usecase.Auth,
	jwtManager *jwt.Handler,
	l logger.Logger,
) {
	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewRoutes(apiV1Group, jwtManager, a, l, validator.New())
	}
}
