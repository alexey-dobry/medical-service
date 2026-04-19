package model

import "github.com/alexey-dobry/medical-service/internal/pkg/validator"

// Credentials is a model to store users data used for authorization
type Credentials struct {
	UserID       string `validate:"required,uuid"`
	Email        string `validate:"required,email"`
	Role         Role   `validate:"required"`
	PasswordHash string `validate:"required"`
}

func (c *Credentials) Validate() error {
	return validator.V.Struct(c)
}
