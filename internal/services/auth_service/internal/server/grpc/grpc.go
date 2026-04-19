package grpc

import (
	authrpc "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/repository"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/server/grpc/auth"

	"google.golang.org/grpc"
)

func NewServer(logger logger.Logger, repository repository.CredentialsRepository, jwtHandler jwt.JWTHandler) *grpc.Server {
	s := grpc.NewServer()

	authrpc.RegisterAuthServer(s, auth.New(logger, repository, jwtHandler))

	return s
}
