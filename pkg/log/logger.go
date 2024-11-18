package log

import (
	"go.uber.org/zap"
	"time"
)

func NewLogger() *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"../../logs/logs.log", "stdout"}

	logger, err := config.Build()

	if err != nil {
		panic("failed to configure logger")
	}

	defer logger.Sync()
	sugaredLogger := logger.Sugar()
	sugaredLogger.Infow("started logger", "source", "logger", "time", time.Now().String())
	return sugaredLogger
}
