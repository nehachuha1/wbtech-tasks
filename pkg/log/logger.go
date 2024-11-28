package log

import (
	"fmt"
	"go.uber.org/zap"
)

// Инициализация логгера в зависимости от того, какой файл с логами нужен
func NewLogger(path string) *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{path, "stdout"}

	logger, err := config.Build()

	if err != nil {
		panic(fmt.Sprintf("failed to configure logger: %v", err))
	}

	defer logger.Sync()
	sugaredLogger := logger.Sugar()
	sugaredLogger.Info("started logger")
	return sugaredLogger
}
