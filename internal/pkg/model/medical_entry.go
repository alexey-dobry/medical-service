package model

type EntryType string

const (
	AppointmentConclusion = "APPOINTMENT_CONCLUSION"
	TestResult            = "TEST_RESULT"
)

// MedicalEntry is a model which stores data of medical entry created either after
// an appointment or after medical test
type MedicalEntry struct {
	UserID      string    `validate:"required,uuid"`
	Type        EntryType `validate:"required"`
	Description string    `validate:"required,max=500"`
}
