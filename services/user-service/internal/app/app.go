package app

import (
	"context"
	"time"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/http-server/handlers"
	"github.com/erknas/ecom/user-service/internal/http-server/middleware"
	"github.com/erknas/ecom/user-service/internal/http-server/server"
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

	userService := service.NewUserService(storage, storage, log)

	authService := service.NewAuthService(storage, jwtManager, jwtManager, log)

	middleware := middleware.New(authService, log)

	userHandler := handlers.New(userService, authService, middleware, log)

	httpServer := server.New(cfg, userHandler)

	return &App{
		HTTPServer: httpServer,
	}
}
