package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger for structured logging.
type Logger struct {
	*zap.SugaredLogger
}

// Config defines logger configuration options.
type Config struct {
	Level  string // DEBUG, INFO, WARNING, ERROR, CRITICAL
	Format string // json or text
}

// New creates a new structured logger instance.
// It configures zap logger based on the provided configuration.
//
// Parameters:
//   - cfg: Logger configuration (level and format)
//
// Returns:
//   - *Logger: Configured logger instance
//   - error: Configuration or initialization error
func New(cfg Config) (*Logger, error) {
	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// Configure encoder based on format
	var zapConfig zap.Config
	if cfg.Format == "json" {
		// JSON format for production (machine-readable)
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Text format for development (human-readable)
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Build logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}, nil
}

// parseLevel converts string log level to zapcore.Level.
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel, nil
	case "INFO":
		return zapcore.InfoLevel, nil
	case "WARNING", "WARN":
		return zapcore.WarnLevel, nil
	case "ERROR":
		return zapcore.ErrorLevel, nil
	case "CRITICAL", "FATAL":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Sync flushes any buffered log entries.
// Applications should call Sync before exiting.
func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}
