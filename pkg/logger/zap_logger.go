package logger

import (
	"log"

	"go.uber.org/zap"
)

// NewZapLogger creates a new configured Zap logger.
func NewZapLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return logger
}
