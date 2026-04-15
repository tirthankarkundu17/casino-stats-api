package logger

import (
	"go.uber.org/zap"
)

// New creates and returns a new production SugaredLogger.
func New() *zap.SugaredLogger {
	zapLogger, _ := zap.NewProduction()
	return zapLogger.Sugar()
}
