package model

import "time"

type Role string

const (
	PatientRole = "PATIENT"
	AdminRole   = "ADMIN"
	DoctorRole  = "DOCTOR"
)

var RoleValue = map[string]Role{
	"PATIENT": PatientRole,
	"ADMIN":   AdminRole,
	"DOCTOR":  DoctorRole,
}

// User is a model which stores data of user in a system
type User struct {
	FirstName  string    `validate:"required"`
	MiddleName string    `validate:"required"`
	LastName   string    `validate:"required"`
	Phone      string    `validate:"required,e164"`
	Email      string    `validate:"required,email"`
	Sex        string    `validate:"required"`
	Role       Role      `validate:"required"`
	BirthDate  time.Time `validate:"required"`
}
