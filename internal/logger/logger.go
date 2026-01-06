package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the global logger instance
var Logger *zap.Logger

// Init initializes the logger based on configuration
func Init(level string, format string, output string, filePath string, maxSize int, maxBackups int, maxAge int) error {
	var config zap.Config

	// Set log level
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure based on format
	if format == "json" {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevel)
		config.Encoding = "json"
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevel)
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	if output == "file" && filePath != "" {
		// File output with rotation
		lumberjackLogger := &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize,    // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,     // days
			Compress:   true,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	} else {
		// Stdout output
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Build logger
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		writeSyncer,
		zapLevel,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

