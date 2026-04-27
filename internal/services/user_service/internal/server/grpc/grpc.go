package grpc

import (
	userrpc "github.com/alexey-dobry/medical-service/internal/pkg/gen/user"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/grpc/user"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"google.golang.org/grpc"
)

func New(logger logger.Logger, store store.Store, client client.Client) *grpc.Server {
	s := grpc.NewServer()

	userrpc.RegisterUserServer(s, user.New(logger, store, client))

	return s
}
