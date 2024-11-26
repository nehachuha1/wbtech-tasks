package log

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

func NewLogger(path string) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{path, "stdout"}

	logger, err := config.Build()

	if err != nil {
		panic(fmt.Sprintf("failed to configure logger: %v", err))
	}

	defer logger.Sync()
	sugaredLogger := logger.Sugar()
	sugaredLogger.Infow("started logger", "source", "logger", "time", time.Now().String())
	return sugaredLogger
}
