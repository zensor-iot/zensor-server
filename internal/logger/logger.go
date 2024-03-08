package logger

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

func NewDefaultLogger() Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	logger, _ := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	return logger.Sugar()
}

var once sync.Once
var defaultLogger Logger

func getDefaultLogger() Logger {
	once.Do(func() {
		defaultLogger = NewDefaultLogger()
	})
	return defaultLogger
}

func Debug(msg string, keysAndValues ...interface{}) {
	getDefaultLogger().Debugw(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	getDefaultLogger().Infow(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	getDefaultLogger().Warnw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	getDefaultLogger().Errorw(msg, keysAndValues...)
}

func String(key string, val string) zap.Field {
	return zap.String(key, val)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func Reflect(key string, val interface{}) zap.Field {
	return zap.Reflect(key, val)
}

func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}
