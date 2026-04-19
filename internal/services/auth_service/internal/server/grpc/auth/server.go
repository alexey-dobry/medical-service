package auth

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/store"
)

type GRPCServer struct {
	pb.UnimplementedAuthServer

	logger     logger.Logger
	repository store.CredentialsRepository
	jwtHandler jwt.JWTHandler
}

func New(logger logger.Logger, repository store.CredentialsRepository, jwtHandler jwt.JWTHandler) *GRPCServer {
	return &GRPCServer{
		logger:     logger.WithFields("layer", "grpc server api", "server"),
		repository: repository,
		jwtHandler: jwtHandler,
	}
}
