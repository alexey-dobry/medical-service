package client

import "github.com/google/uuid"

type CreateDoctorRequest struct {
	FirstName      string
	MiddleName     string
	LastName       string
	Phone          string
	Email          string
	Sex            string
	BirthDate      string
	Specialty      string
	WorkExperience string
	Description    string
	Services       []string
	ProfilePicture *ProfilePicture
}

type ProfilePicture struct {
	Filename string
	Content  []byte
	MimeType string
}

type Client interface {
	CreateDoctor(request CreateDoctorRequest) error
	DeleteDoctor(id uuid.UUID) error

	Close() error
}
