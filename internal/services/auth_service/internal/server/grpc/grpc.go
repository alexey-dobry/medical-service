package grpc

import (
	authrpc "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/grpc/auth"

	"google.golang.org/grpc"
)

func New(
	logger logger.Logger,
	credentialsRepo repository.CredentialsRepository,
	blacklistRepo repository.BlacklistRepository,
	jwtHandler *jwt.Handler,
) *grpc.Server {
	s := grpc.NewServer()

	authrpc.RegisterAuthServer(s, auth.New(logger, credentialsRepo, blacklistRepo, jwtHandler))

	return s
}
