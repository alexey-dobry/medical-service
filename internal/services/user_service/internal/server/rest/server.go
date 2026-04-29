package rest

import (
	"net"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/gofiber/fiber/v2"
)

type RESTServer struct {
	fiberApp   *fiber.App
	grpcClient client.Client

	middlewareConfig middleware.Config

	logger logger.Logger
	store  store.Store
}

func New(logger logger.Logger, store store.Store, grpcClient client.Client, cfg Config) *RESTServer {
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	s := &RESTServer{
		fiberApp:         fiberApp,
		grpcClient:       grpcClient,
		middlewareConfig: cfg.MiddlewareConfig,
		logger:           logger.WithFields("layer", "rest server api", "server"),
		store:            store,
	}

	s.initRoutes()

	return s
}

func (s *RESTServer) Listener(listener net.Listener) error {
	return s.fiberApp.Listener(listener)
}
