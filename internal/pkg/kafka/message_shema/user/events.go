package user

type EventType string

const (
	PatientCreated EventType = "patient.record.created"
)

type PatientCreatedPayload struct {
}
