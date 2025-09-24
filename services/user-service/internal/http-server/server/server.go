package server

import (
	"context"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/http-server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	userHandler *handlers.UserHandler
	srv         *http.Server
	router      *chi.Mux
}

func New(cfg *config.Config, userHandler *handlers.UserHandler) *Server {
	router := chi.NewMux()

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Addr,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		Handler:      router,
	}

	return &Server{
		userHandler: userHandler,
		srv:         srv,
		router:      router,
	}
}

func (s *Server) Start() error {
	s.setupRoutes()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Route("/api", func(r chi.Router) {
		r.Route("/users", s.userHandler.RegisterRoutes)
	})
}
