package auth

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
)

type GRPCServer struct {
	pb.UnimplementedAuthServer

	logger          logger.Logger
	credentialsRepo repository.CredentialsRepository
	jwtHandler      jwt.JWTHandler
}

func New(logger logger.Logger, credentialsRepo repository.CredentialsRepository, jwtHandler jwt.JWTHandler) *GRPCServer {
	return &GRPCServer{
		logger:          logger.WithFields("layer", "grpc server api", "server"),
		credentialsRepo: credentialsRepo,
		jwtHandler:      jwtHandler,
	}
}
