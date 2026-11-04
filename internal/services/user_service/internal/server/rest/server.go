package rest

import (
	"net"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/service"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/gofiber/fiber/v2"
)

type RESTServer struct {
	fiberApp     *fiber.App
	grpcClient   client.Client
	registration *service.Registration

	middlewareConfig middleware.Config
	tokenValidator   middleware.TokenValidator

	logger logger.Logger
	store  store.Store

	secureCookies bool
}

func New(
	logger logger.Logger,
	store store.Store,
	grpcClient client.Client,
	registration *service.Registration,
	tokenValidator middleware.TokenValidator,
	cfg Config,
) *RESTServer {
	fiberApp := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})

	s := &RESTServer{
		fiberApp:         fiberApp,
		grpcClient:       grpcClient,
		registration:     registration,
		middlewareConfig: cfg.MiddlewareConfig,
		tokenValidator:   tokenValidator,
		logger:           logger.WithFields("layer", "rest server"),
		store:            store,
		secureCookies:    !cfg.InsecureCookies,
	}

	s.initRoutes()

	return s
}

func (s *RESTServer) Listener(listener net.Listener) error {
	return s.fiberApp.Listener(listener)
}

// setAuthCookies выставляет те же куки, что auth_service отдаёт при логине.
func (s *RESTServer) setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   s.secureCookies,
		SameSite: "Strict",
		Path:     "/auth",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   s.secureCookies,
		SameSite: "Strict",
		Path:     "/",
	})
}
