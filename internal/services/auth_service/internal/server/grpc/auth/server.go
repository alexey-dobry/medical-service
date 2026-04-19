package auth

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
)

type ServerAPI struct {
	pb.UnimplementedAuthServer

	logger     logger.Logger
	repository repository.CredentialsRepository
	jwtHandler jwt.JWTHandler
}

func New(logger logger.Logger, repository repository.CredentialsRepository, jwtHandler jwt.JWTHandler) *ServerAPI {
	return &ServerAPI{
		repository: repository,
		logger:     logger.WithFields("layer", "grpc server api", "server"),
		jwtHandler: jwtHandler,
	}
}
