package client

type RegisterRequest struct {
	UserID   string
	Email    string
	Role     string
	Password string
}

type RegisterPatientResponse struct {
	AccessToken  string
	RefreshToken string
}

type DeleteRequest struct {
	UserID string
}

type Client interface {
	RegisterPatient(request RegisterRequest) (RegisterPatientResponse, error)
	RegisterDoctor(request RegisterRequest) error
	DeleteUser(DeleteRequest) error
}
