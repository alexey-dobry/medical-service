package v1

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/logger"
	"github.com/alexey-dobry/medical-service/internal/services/auth_service/internal/usecase"
	"github.com/go-playground/validator/v10"
)

type V1 struct {
	a usecase.Auth
	l logger.Logger
	v *validator.Validate
}
