package logger

import "go.uber.org/zap"

const (
	envLocal = "local"
	envProd  = "prod"
)

func Setup(env string) *zap.Logger {
	var logger *zap.Logger

	switch env {
	case envLocal:
		logger, _ = zap.NewDevelopment(
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
			zap.AddStacktrace(zap.WarnLevel),
		)
	case envProd:
		logger, _ = zap.NewProduction(
			zap.AddCaller(),
			zap.AddStacktrace(zap.WarnLevel),
		)
	default:
		logger, _ = zap.NewProduction()
	}

	return logger
}
