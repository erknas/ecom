package main

import (
	"context"
	"fmt"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/logger"
	"github.com/erknas/ecom/user-service/internal/storage/postgres"
	"go.uber.org/zap"
)

func main() {
	var (
		ctx = context.Background()
		cfg = config.MustLoad()
		log = logger.Setup(cfg.Env)
	)

	log.Debug("config loaded",
		zap.Any("cfg", cfg),
	)

	storage, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Error("failed to init storage",
			zap.Error(err),
		)
	}

	fmt.Println(storage)
}
