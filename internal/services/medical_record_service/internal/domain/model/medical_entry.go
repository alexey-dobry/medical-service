package model

import "github.com/alexey-dobry/medical-service/internal/pkg/validator"

type RecordType string

const (
	AppointmentConclusion = "APPOINTMENT_CONCLUSION"
	TestResult            = "TEST_RESULT"
)

// MedicalEntry is a model which stores data of medical entry created either after
// an appointment or after medical test
type MedicalRecord struct {
	ID              string
	PatientID       string     `validate:"required,uuid"`
	DoctorID        string     `validate:"required,uuid"`
	Type            RecordType `validate:"required"`
	Conclusion      string     `validate:"required,max=100"`
	Description     string     `validate:"required,max=500"`
	Recommendations string     `validate:"required,max=500"`
	Date            string     `validate:"required"`
}

func (m *MedicalRecord) Validate() error {
	return validator.V.Struct(m)
}
