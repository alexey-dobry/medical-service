package rest

import "github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"

func (s *RESTServer) initRoutes() {
	s.fiberApp.Post("/users/register", s.handleCreatePatientProfile())
	s.fiberApp.Get("/users/profile/patient", s.handleGetPatientProfile(), middleware.ValidateJWT(s.middlewareConfig))
	s.fiberApp.Get("/users/profile/doctor", s.handleGetDoctorProfile(), middleware.ValidateJWT(s.middlewareConfig))
	s.fiberApp.Patch("/users/profile", s.handleUpdateProfile(), middleware.ValidateJWT(s.middlewareConfig))
	s.fiberApp.Get("/users/doctors/search", s.handleSearchDoctors(), middleware.ValidateJWT(s.middlewareConfig))
	s.fiberApp.Get("/users/doctors/{doctor_id}/details", s.handleGetDoctorDetails(), middleware.ValidateJWT(s.middlewareConfig))
}
