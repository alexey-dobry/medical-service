package rest

func (s *RESTServer) initRoutes() {
	s.fiberApp.Post("/auth/login", s.handleLogin())
	s.fiberApp.Post("/auth/logout", s.handleLogout())
	s.fiberApp.Post("/auth/refresh", s.handleRefresh())
}
