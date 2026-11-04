package rest

import (
	"github.com/alexey-dobry/medical-service/internal/pkg/model"
	"github.com/alexey-dobry/medical-service/internal/services/user_service/internal/server/rest/middleware"
)

// Fiber вызывает обработчики в порядке регистрации, поэтому middleware
// аутентификации должен идти перед защищаемым обработчиком.
func (s *RESTServer) initRoutes() {
	auth := middleware.ValidateJWT(s.middlewareConfig, s.tokenValidator)
	adminOnly := middleware.RequireRole(model.AdminRole)

	s.fiberApp.Post("/users/register", s.handleCreatePatientProfile())
	s.fiberApp.Get("/users/profile/patient", auth, s.handleGetPatientProfile())
	s.fiberApp.Get("/users/profile/doctor", auth, s.handleGetDoctorProfile())
	s.fiberApp.Patch("/users/profile", auth, s.handleUpdateProfile())
	s.fiberApp.Post("/users/doctors", auth, adminOnly, s.handleCreateDoctor())
	s.fiberApp.Delete("/users/doctors/:doctor_id", auth, adminOnly, s.handleDeleteDoctor())
	s.fiberApp.Get("/users/doctors/search", auth, s.handleSearchDoctors())
	s.fiberApp.Get("/users/doctors/:doctor_id/details", auth, s.handleGetDoctorDetails())
}
