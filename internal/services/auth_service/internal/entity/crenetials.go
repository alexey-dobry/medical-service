package entity

import "time"

type Role string

const (
	PatientRole Role = "PATIENT"
	AdminRole   Role = "ADMIN"
	DoctorRole  Role = "DOCTOR"
)

type Credentials struct {
	ID           string
	UserID       string
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r Role) Valid() bool {
	switch r {
	case PatientRole, AdminRole, DoctorRole:
		return true
	default:
		return false
	}
}
