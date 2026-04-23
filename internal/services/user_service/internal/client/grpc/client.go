package grpc

import (
	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type grpcClient struct {
	Auth pb.AuthClient

	conn *grpc.ClientConn
}

func New(logger logger.Logger, cfg Config) client.Client {
	logger.WithFields("layer", "grpc client", "client")

	creds, err := credentials.NewClientTLSFromFile(cfg.CertFilePath, cfg.ServerNameOverride)
	if err != nil {
		logger.Fatalf("failed to load CA cert: %v", err)
	}

	conn, err := grpc.NewClient(
		cfg.Host+":"+cfg.Port,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		logger.Fatalf("failed to connect: %v", err)
	}

	authClient := pb.NewAuthClient(conn)

	return &grpcClient{
		Auth: authClient,
	}
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
