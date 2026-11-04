package grpc

import (
	"fmt"
	"net"
	"time"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	Auth pb.AuthClient

	conn    *grpc.ClientConn
	logger  logger.Logger
	timeout time.Duration
}

func New(logger logger.Logger, cfg Config) (client.Client, error) {
	logger = logger.WithFields("layer", "grpc client")

	creds := insecure.NewCredentials()
	if cfg.CertFilePath != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(cfg.CertFilePath, cfg.ServerNameOverride)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA cert: %w", err)
		}
		creds = tlsCreds
	}

	conn, err := grpc.NewClient(
		net.JoinHostPort(cfg.Host, cfg.Port),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	return &grpcClient{
		Auth:    pb.NewAuthClient(conn),
		conn:    conn,
		logger:  logger,
		timeout: cfg.Timeout,
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
