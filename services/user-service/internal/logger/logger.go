package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func New(env string) *zap.Logger {
	var logger *zap.Logger

	switch env {
	case envLocal:
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ = cfg.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		)
	case envProd:
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logger, _ = cfg.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zap.WarnLevel),
		)
	default:
		logger, _ = zap.NewProduction()
	}

	zap.ReplaceGlobals(logger)

	return logger
}
