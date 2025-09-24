package app

import (
	"context"
	"time"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/http-server/handlers"
	"github.com/erknas/ecom/user-service/internal/http-server/server"
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

	userService := service.New(storage, storage)

	userHandler := handlers.New(userService, log)

	httpServer := server.New(cfg, userHandler)

	return &App{
		HTTPServer: httpServer,
	}
}
