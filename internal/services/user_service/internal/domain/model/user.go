package model

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
)

// User is a model which stores data of user in a system
type User struct {
	FirstName  string    `validate:"required"`
	MiddleName string    `validate:"required"`
	LastName   string    `validate:"required"`
	Phone      string    `validate:"required,e164"`
	Email      string    `validate:"required,email"`
	Sex        string    `validate:"required"`
	BirthDate  time.Time `validate:"required"`
}

func (u *User) Validate() error {
	return validator.V.Struct(u)
}
