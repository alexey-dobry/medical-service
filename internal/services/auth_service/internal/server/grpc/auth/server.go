package auth

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
)

type GRPCServer struct {
	pb.UnimplementedAuthServer

	logger          logger.Logger
	credentialsRepo repository.CredentialsRepository
	blacklistRepo   repository.BlacklistRepository
	jwtHandler      *jwt.Handler
}

func New(
	logger logger.Logger,
	credentialsRepo repository.CredentialsRepository,
	blacklistRepo repository.BlacklistRepository,
	jwtHandler *jwt.Handler,
) *GRPCServer {
	return &GRPCServer{
		logger:          logger.WithFields("layer", "grpc server"),
		credentialsRepo: credentialsRepo,
		blacklistRepo:   blacklistRepo,
		jwtHandler:      jwtHandler,
	}
}
