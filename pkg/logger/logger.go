package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger for application-wide logging
type Logger struct {
	zerolog.Logger
}

// Global logger instance
var globalLogger *Logger

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Pretty     bool   // Enable console pretty printing
	TimeFormat string // Time format (default: RFC3339)
}

// New creates a new configured logger and sets it as global
func New(cfg Config) *Logger {
	// Parse level
	level := zerolog.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global level
	zerolog.SetGlobalLevel(level)

	// Create logger
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	globalLogger = &Logger{Logger: logger}
	return globalLogger
}

// Get returns the global logger instance
func Get() *Logger {
	if globalLogger == nil {
		return Default()
	}
	return globalLogger
}

// Default returns the global logger
func Default() *Logger {
	return &Logger{Logger: log.Logger}
}

// WithGameID adds game ID to logger context
func (l *Logger) WithGameID(gameID string) *Logger {
	return &Logger{Logger: l.With().Str("game_id", gameID).Logger()}
}

// WithPlayerID adds player ID to logger context
func (l *Logger) WithPlayerID(playerID string) *Logger {
	return &Logger{Logger: l.With().Str("player_id", playerID).Logger()}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.With().Str("request_id", requestID).Logger()}
}

// WithError adds error to logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With().Err(err).Logger()}
}
