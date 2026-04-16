package model

// DoctorAdditionalData is a model
// which stores additional data for User model if user has Doctor Role
type DoctorAdditionalData struct {
	UserID         string `validate:"required,uuid"`
	Specialty      string `validate:"required,max=15"`
	WorkExperience uint   `validate:"required,min=1,max=40"`
	Description    string `validate:"required,max=255"`
}
