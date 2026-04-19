package auth

import (
	"context"
	"strings"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/auth"
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/jwt"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/domain/utils"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerAPI) RegisterPatient(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorf("Failed to hash password: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	credentials := model.Credentials{
		UserID:       req.UserId,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err = credentials.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user field value")
	}

	err = s.repository.Add(credentials)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil, status.Error(codes.AlreadyExists, "Account with specified email already exists")
	} else if err != nil {
		s.logger.Errorf("Failed to add new user to data: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	refreshToken, accessToken, err := s.jwtHandler.GenerateJWTPair(jwt.Claims{
		ID:   req.UserId,
		Role: model.RoleValue[req.Role],
	})

	if err != nil {
		s.logger.Errorf("Failed to generate token pair: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	response := pb.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return &response, nil
}

func (s *ServerAPI) RegisterDoctor(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorf("Failed to hash password: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	credentials := model.Credentials{
		UserID:       req.UserId,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err = credentials.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user field value")
	}

	err = s.repository.Add(credentials)
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil, status.Error(codes.AlreadyExists, "Account with specified email already exists")
	} else if err != nil {
		s.logger.Errorf("Failed to add new user to data: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	return &emptypb.Empty{}, nil
}
