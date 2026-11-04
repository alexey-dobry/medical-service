package model

import (
	pkgmodel "github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
)

type Credentials struct {
	UserID       string        `validate:"required,uuid"`
	Email        string        `validate:"required,email"`
	Role         pkgmodel.Role `validate:"required"`
	PasswordHash string        `validate:"required"`
}

func (c *Credentials) Validate() error {
	return validator.V.Struct(c)
}
