package logger

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

var logger log.Logger

func InitLogger() {
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.Caller(4))
}

func GetLogger() log.Logger {
	return logger
}

func LogError(keyvals ...any) {
	_ = level.Error(logger).Log(keyvals...)
}

func LogInfo(keyvals ...any) {
	_ = level.Info(logger).Log(keyvals...)
}

func LogDebug(keyvals ...any) {
	_ = level.Debug(logger).Log(keyvals...)
}
