package rest

func (s *RESTServer) initRoutes() {
	s.fiberApp.Get("/medical-records/{user_id}")
	s.fiberApp.Get("/medical-records/{record_id}")
	s.fiberApp.Get("/medical-records/files/{file_id}")
	s.fiberApp.Post("/medical-records")
}
