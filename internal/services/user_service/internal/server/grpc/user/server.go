package user

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/user"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
)

type GRPCServer struct {
	pb.UnimplementedUserServer

	grpcClient client.Client
	logger     logger.Logger
	store      store.Store
}

func New(logger logger.Logger, store store.Store, client client.Client) *GRPCServer {
	return &GRPCServer{
		grpcClient: client,
		logger:     logger.WithFields("layer", "grpc server api", "server"),
		store:      store,
	}
}
