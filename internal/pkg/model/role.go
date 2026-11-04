package model

// Role — роль пользователя, общая для всех сервисов.
type Role string

const (
	PatientRole Role = "PATIENT"
	DoctorRole  Role = "DOCTOR"
	AdminRole   Role = "ADMIN"
)

// ParseRole преобразует сырую строку в известную Role.
func ParseRole(raw string) (Role, bool) {
	switch r := Role(raw); r {
	case PatientRole, DoctorRole, AdminRole:
		return r, true
	default:
		return "", false
	}
}
