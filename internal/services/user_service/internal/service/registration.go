package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/client"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/domain/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/store"
	"github.com/google/uuid"
)

var (
	// ErrAlreadyExists — пользователь с таким email уже зарегистрирован.
	ErrAlreadyExists = errors.New("user already exists")
)

// Registration создаёт аккаунты в двух сервисах: учётные данные в auth_service,
// профиль в user_service. Распределённой транзакции нет, поэтому сбой после
// создания учётных данных откатывается компенсирующим DeleteUser (сага).
type Registration struct {
	auth   client.Client
	store  store.Store
	logger logger.Logger
	retry  RetryPolicy
}

func NewRegistration(auth client.Client, store store.Store, logger logger.Logger, retry RetryPolicy) *Registration {
	return &Registration{
		auth:   auth,
		store:  store,
		logger: logger.WithFields("layer", "registration service"),
		retry:  retry,
	}
}

// PatientRegistration — провалидированный запрос регистрации пациента.
type PatientRegistration struct {
	User     model.User
	Password string
	// Photo необязательно.
	Photo *PhotoUpload
}

// RegisterPatient создаёт учётные данные и профиль, возвращает пару токенов от auth_service.
func (r *Registration) RegisterPatient(ctx context.Context, req PatientRegistration) (client.RegisterPatientResponse, error) {
	if req.Photo != nil {
		if err := req.Photo.Validate(); err != nil {
			return client.RegisterPatientResponse{}, err
		}
	}

	tokens, err := r.auth.RegisterPatient(ctx, client.RegisterRequest{
		UserID:   req.User.ID,
		Email:    req.User.Email,
		Role:     string(pkgmodel.PatientRole),
		Password: req.Password,
	})
	if err != nil {
		return client.RegisterPatientResponse{}, err
	}

	if _, err := r.saveProfile(ctx, req.User, req.Photo, func(user model.User) error {
		return r.store.User().AddPatient(ctx, user)
	}); err != nil {
		r.compensate(ctx, req.User.ID)
		return client.RegisterPatientResponse{}, err
	}

	return tokens, nil
}

// DoctorRegistration — провалидированный запрос админа на добавление врача.
type DoctorRegistration struct {
	User     model.User
	Doctor   model.DoctorAdditionalData
	Password string
	Photo    *PhotoUpload
}

// RegisterDoctor создаёт учётные данные врача в auth_service (чтобы он мог
// залогиниться), профиль с фото и запись в поисковом индексе.
func (r *Registration) RegisterDoctor(ctx context.Context, req DoctorRegistration) error {
	if req.Photo == nil {
		return fmt.Errorf("%w: photo is required for a doctor", ErrInvalidPhoto)
	}
	if err := req.Photo.Validate(); err != nil {
		return err
	}

	err := r.auth.RegisterDoctor(ctx, client.RegisterRequest{
		UserID:   req.User.ID,
		Email:    req.User.Email,
		Role:     string(pkgmodel.DoctorRole),
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	photo, err := r.saveProfile(ctx, req.User, req.Photo, func(user model.User) error {
		return r.store.User().AddDoctor(ctx, user, req.Doctor)
	})
	if err != nil {
		r.compensate(ctx, req.User.ID)
		return err
	}

	err = r.store.Search().AddDoctor(ctx, model.DoctorSearchParams{
		ID:         req.User.ID,
		FirstName:  req.User.FirstName,
		MiddleName: req.User.MiddleName,
		LastName:   req.User.LastName,
		Sex:        req.User.Sex,
		Specialty:  req.Doctor.Specialty,
		Services:   req.Doctor.Services,
	})
	if err != nil {
		r.deleteProfile(ctx, req.User.ID)
		if photo != nil {
			r.removePhoto(*photo)
		}
		r.compensate(ctx, req.User.ID)
		return fmt.Errorf("failed to index doctor: %w", err)
	}

	return nil
}

// saveProfile загружает необязательное фото, сохраняет профиль через addUser
// и затем метаданные фото. При сбое откатывает свои шаги; компенсацию
// auth_service выполняет вызывающий.
func (r *Registration) saveProfile(ctx context.Context, user model.User, upload *PhotoUpload, addUser func(model.User) error) (*model.Photo, error) {
	var photo *model.Photo
	if upload != nil {
		uploaded, err := r.uploadPhoto(upload, user.ID)
		if err != nil {
			return nil, err
		}
		photo = &uploaded
		user.PhotoID = uploaded.ID
	}

	if err := addUser(user); err != nil {
		if photo != nil {
			r.removePhoto(*photo)
		}
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	if photo != nil {
		if err := r.store.Meta().Create(ctx, *photo); err != nil {
			r.deleteProfile(ctx, user.ID)
			r.removePhoto(*photo)
			return nil, fmt.Errorf("failed to save photo metadata: %w", err)
		}
	}

	return photo, nil
}

// deleteProfile — best-effort удаление профиля, созданного в рамках
// откатываемой регистрации.
func (r *Registration) deleteProfile(ctx context.Context, userID string) {
	id, err := uuid.Parse(userID)
	if err == nil {
		err = r.store.User().DeleteUser(context.WithoutCancel(ctx), id)
	}
	if err != nil {
		r.logger.Errorf("Failed to roll back profile of user %s: %s", userID, err)
	}
}

// compensate удаляет учётные данные из auth_service. Использует контекст
// запроса (сохраняет значения), но не отменяется вместе с ним: обрыв клиента
// не должен оставлять осиротевшие учётные данные. Вся компенсация ограничена
// retry.Timeout.
func (r *Registration) compensate(ctx context.Context, userID string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.retry.Timeout)
	defer cancel()

	err := retry(ctx, r.retry, func(ctx context.Context) error {
		return r.auth.DeleteUser(ctx, client.DeleteRequest{UserID: userID})
	})
	if err != nil {
		// Аккаунт остался в auth_service без профиля и требует ручной чистки —
		// логируем так, чтобы это нельзя было пропустить.
		r.logger.Errorf("COMPENSATION FAILED: credentials of user %s were not deleted from auth_service: %s", userID, err)
		return
	}

	r.logger.Warnf("Registration of user %s rolled back: credentials deleted", userID)
}
