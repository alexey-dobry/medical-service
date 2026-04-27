package user

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/alexey-dobry/medical-service/internal/pkg/gen/user"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GRPCServer) CreateDoctor(ctx context.Context, req *pb.CreateDoctorRequest) (*emptypb.Empty, error) {
	birthDate, _ := time.Parse("2006-01-02", req.BirthDate)

	user := model.User{
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		Email:      req.Email,
		Sex:        req.Sex,
		BirthDate:  birthDate,
	}

	if err := user.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user field value")
	}

	userID := uuid.New()
	workExperience, _ := strconv.Atoi(req.WorkExperience)

	doctorData := model.DoctorAdditionalData{
		UserID:         userID.String(),
		Specialty:      req.Specialty,
		WorkExperience: workExperience,
		Description:    req.Description,
	}

	err := s.store.User().AddDoctor(user, doctorData)
	if err != nil {
		s.logger.Errorf("Failed to add new user to data: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	searchParams := model.DoctorSearchParams{
		ID:         userID.String(),
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		Sex:        req.Sex,
		Services:   req.Services,
	}

	err = s.store.Search().AddDoctor(searchParams)
	if err != nil {

		s.store.User().DeleteUser(userID)

		s.logger.Errorf("Failed to add new doctor search params: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	photoID := uuid.New()
	storageKey := fmt.Sprintf("files/%s", photoID.String())

	photo := model.Photo{
		ID:         photoID.String(),
		Name:       "profile_photo",
		MimeType:   req.ProfilePicture.MimeType,
		Size:       int64(len(req.ProfilePicture.Content)),
		StorageKey: storageKey,
	}

	s.store.Meta().Create(photo)

	s.store.Photos().Put(
		storageKey,
		bytes.NewReader(req.ProfilePicture.Content),
		int64(len(req.ProfilePicture.Content)),
		req.ProfilePicture.MimeType,
	)

	return &emptypb.Empty{}, nil
}

func (s *GRPCServer) DeleteDoctor(ctx context.Context, req *pb.DeleteDoctorRequest) (*emptypb.Empty, error) {
	err := s.grpcClient.DeleteUser(client.DeleteRequest{
		UserID: req.UserId,
	})
	if err != nil {
		s.logger.Errorf("Failed to delete user credentials: %s", err)
		return nil, status.Error(codes.Internal, "Internal server error")
	}

	id, _ := uuid.Parse(req.UserId)

	for range 5 {
		err = s.store.User().DeleteUser(id)
		if err == nil {
			break
		}
	}

	var photoStorageKey string
	for range 5 {
		photoStorageKey, err = s.store.Meta().Delete(id)
		if err == nil {
			break
		}
	}

	for range 5 {
		err = s.store.Photos().Delete(photoStorageKey)
		if err == nil {
			break
		}
	}

	return &emptypb.Empty{}, nil
}
