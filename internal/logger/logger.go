package logger

import (
	"sync"

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
