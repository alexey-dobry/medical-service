package rest

import (
	"net"

	"github.com/alexey-dobry/medical-service/internal/pkg/health"
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/gofiber/fiber/v2"
)

type RESTServer struct {
	fiberApp *fiber.App

	logger          logger.Logger
	credentialsRepo repository.CredentialsRepository
	blacklistRepo   repository.BlacklistRepository
	jwtHandler      *jwt.Handler
	healthChecker   *health.Checker

	secureCookies bool
}

func New(
	logger logger.Logger,
	credentialsRepo repository.CredentialsRepository,
	blacklistRepo repository.BlacklistRepository,
	jwtHandler *jwt.Handler,
	health *health.Checker,
	cfg Config,
) *RESTServer {
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	s := &RESTServer{
		fiberApp: fiberApp,

		logger:          logger.WithFields("layer", "rest server"),
		credentialsRepo: credentialsRepo,
		blacklistRepo:   blacklistRepo,
		jwtHandler:      jwtHandler,
		healthChecker:   health,
		secureCookies:   !cfg.InsecureCookies,
	}

	s.initRoutes()

	return s
}

func (s *RESTServer) Listener(listener net.Listener) error {
	return s.fiberApp.Listener(listener)
}
