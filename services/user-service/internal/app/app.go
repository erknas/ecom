package app

import (
	"context"
	"time"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/http/handlers"
	mw "github.com/erknas/ecom/user-service/internal/http/middleware"
	"github.com/erknas/ecom/user-service/internal/http/server"
	"github.com/erknas/ecom/user-service/internal/lib/jwt"
	"github.com/erknas/ecom/user-service/internal/service"
	"github.com/erknas/ecom/user-service/internal/storage/postgres"
	"go.uber.org/zap"
)

type App struct {
	HTTPServer *server.Server
}

func New(ctx context.Context, cfg *config.Config, log *zap.Logger) *App {
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	storage, err := postgres.New(ctx, cfg)
	if err != nil {
		panic(err)
	}

	jwtManager := jwt.New(cfg)

	service := service.New(storage, jwtManager, log)

	handlers := handlers.New(service, log)

	middleware := mw.New(jwtManager, log)

	server := server.New(cfg, handlers, middleware)

	return &App{
		HTTPServer: server,
	}
}
