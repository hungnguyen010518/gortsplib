package utils

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	LogDebug(v ...interface{})
	LogInfo(v ...interface{})
	LogWarn(v ...interface{})
	LogError(v ...interface{})
	LogFatal(v ...interface{})
}

type ZapLogger struct {
	logger *zap.Logger
}

func CreateZapLogger() ZapLogger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	// Custom time encoder to remove 'Z' and milliseconds
	encoderCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05"))
	}
	encoderCfg.LevelKey = "level"                        // Change the key to "level"
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder // Capitalize log levels

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     false, // Disable caller information
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stdout",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
	}

	return ZapLogger{logger: zap.Must(config.Build(zap.AddCallerSkip(1)))}
}

func (zapLogger ZapLogger) LogDebug(v ...interface{}) {
	zapLogger.logger.Debug(fmt.Sprintln(v...))
}

func (zapLogger ZapLogger) LogInfo(v ...interface{}) {
	zapLogger.logger.Info(fmt.Sprintln(v...))
}

func (zapLogger ZapLogger) LogWarn(v ...interface{}) {
	zapLogger.logger.Warn(fmt.Sprintln(v...))
}

func (zapLogger ZapLogger) LogError(v ...interface{}) {
	zapLogger.logger.Error(fmt.Sprintln(v...))
}

func (zapLogger ZapLogger) LogFatal(v ...interface{}) {
	zapLogger.logger.Fatal(fmt.Sprintln(v...))
}
