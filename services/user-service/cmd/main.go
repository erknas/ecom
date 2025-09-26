package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erknas/ecom/user-service/internal/app"
	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	cfg := config.MustLoad()

	log := logger.Setup(cfg.Env)

	app := app.New(ctx, cfg, log)

	log.Debug("config loaded", zap.Any("cfg", cfg))

	log.Info("starting HTTP server", zap.String("addr", cfg.HTTPServer.Addr))

	go func() {
		if err := app.HTTPServer.Start(); err != nil {
			log.Error("start server error", zap.Error(err))
			cancel()
		}
	}()

	<-ctx.Done()

	log.Info("shutting down app")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer shutdownCancel()

	if err := app.HTTPServer.Stop(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("app shutdown gracefully")
}
