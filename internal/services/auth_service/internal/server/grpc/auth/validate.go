package auth

import (
	"context"
	"errors"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateAccessToken проверяет токен для других сервисов: подпись, срок действия и чёрный список.
func (s *GRPCServer) ValidateAccessToken(ctx context.Context, req *pb.ValidateAccessTokenRequest) (*pb.ValidateAccessTokenResponse, error) {
	claims, err := s.jwtHandler.ParseToken(req.Token, jwt.AccessToken)
	if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenInvalid) {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	} else if err != nil {
		s.logger.Errorf("Failed to parse access token: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	revoked, err := s.blacklistRepo.IsAccessTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		s.logger.Errorf("Failed to check token blacklist: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	if revoked {
		return nil, status.Error(codes.Unauthenticated, "token revoked")
	}

	return &pb.ValidateAccessTokenResponse{
		UserId: claims.UserID(),
		Role:   string(claims.Role),
	}, nil
}
