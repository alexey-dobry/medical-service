package rest

import (
	"net"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store"
	"github.com/gofiber/fiber/v2"
)

type RESTServer struct {
	fiberApp *fiber.App

	logger     logger.Logger
	store      store.Store
	jwtHandler jwt.JWTHandler
}

func New(logger logger.Logger, store store.Store, jwtHandler jwt.JWTHandler, cfg Config) *RESTServer {
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	s := &RESTServer{
		fiberApp: fiberApp,

		logger:     logger.WithFields("layer", "rest server api", "server"),
		store:      store,
		jwtHandler: jwtHandler,
	}

	s.initRoutes()

	return s
}

func (s *RESTServer) Listener(listener net.Listener) error {
	return s.fiberApp.Listener(listener)
}
