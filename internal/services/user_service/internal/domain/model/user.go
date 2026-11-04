package model

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
)

type User struct {
	ID         string    `validate:"required,uuid"`
	FirstName  string    `validate:"required"`
	MiddleName string    `validate:"required"`
	LastName   string    `validate:"required"`
	Phone      string    `validate:"required,e164"`
	Email      string    `validate:"required,email"`
	Sex        string    `validate:"required"`
	BirthDate  time.Time `validate:"required"`
	PhotoID    string    `validate:"omitempty,uuid"`
}

func (u *User) Validate() error {
	return validator.V.Struct(u)
}
