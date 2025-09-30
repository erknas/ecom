package server

import (
	"context"
	"net/http"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/http/handlers"
	"github.com/erknas/ecom/user-service/internal/lib/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthMiddleware interface {
	WithJWTAuth() func(next http.Handler) http.Handler
}

type Server struct {
	mw       AuthMiddleware
	handlers *handlers.UserHandlers
	srv      *http.Server
	router   *chi.Mux
}

func New(cfg *config.Config, handlers *handlers.UserHandlers, mw AuthMiddleware) *Server {
	router := chi.NewMux()

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Addr,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		Handler:      router,
	}

	return &Server{
		mw:       mw,
		handlers: handlers,
		srv:      srv,
		router:   router,
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
		r.Post("/register", api.MakeHTTPFunc(s.handlers.HandleRegister))
		r.Post("/login", api.MakeHTTPFunc(s.handlers.HandleLogin))

		r.Group(func(r chi.Router) {
			r.Use(s.mw.WithJWTAuth())
			r.Get("/me", api.MakeHTTPFunc(s.handlers.HandleGetUser))
			r.Put("/me/update", api.MakeHTTPFunc(s.handlers.HandleUpdateUser))
		})
	})
}
