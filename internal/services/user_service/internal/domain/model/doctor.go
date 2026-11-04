package model

import "github.com/alexey-dobry/medical-service/internal/pkg/validator"

// DoctorAdditionalData хранит дополнительные данные пользователя с ролью врача.
type DoctorAdditionalData struct {
	UserID         string   `validate:"required,uuid"`
	Specialty      string   `validate:"required,max=15"`
	WorkExperience int      `validate:"required,min=1,max=40"`
	Description    string   `validate:"required,max=255"`
	Services       []string `validate:"required"`
}

func (d *DoctorAdditionalData) Validate() error {
	return validator.V.Struct(d)
}

type DoctorSearchParams struct {
	ID         string
	FirstName  string
	MiddleName string
	LastName   string
	Sex        string
	Specialty  string
	Services   []string
}
