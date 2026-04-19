package model

import (
	"time"

	"github.com/alexey-dobry/medical-service/internal/pkg/validator"
)

// Appointment is a model to store appointment data
type Appointment struct {
	PatientID string    `validate:"required,uuid"`
	DoctorID  string    `validate:"required,uuid"`
	StartTime time.Time `validate:"required"`
	EndTime   time.Time `validate:"required"`
}

func (a *Appointment) Validate() error {
	return validator.V.Struct(a)
}
