package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging with context support
type Logger struct {
	prefix string
	logger *log.Logger
}

// LogLevel represents the severity of a log message
type LogLevel string

const (
	// LevelDebug for detailed debugging information
	LevelDebug LogLevel = "DEBUG"
	// LevelInfo for general informational messages
	LevelInfo LogLevel = "INFO"
	// LevelWarn for warning messages
	LevelWarn LogLevel = "WARN"
	// LevelError for error messages
	LevelError LogLevel = "ERROR"
	// LevelFatal for fatal error messages
	LevelFatal LogLevel = "FATAL"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
)

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger("")
}

// NewLogger creates a new logger with an optional prefix
func NewLogger(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
		logger: log.New(os.Stdout, "", 0),
	}
}

// formatMessage creates a structured log message
func (l *Logger) formatMessage(level LogLevel, ctx context.Context, msg string, fields map[string]interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)

	logMsg := fmt.Sprintf("[%s] %s", timestamp, level)

	if l.prefix != "" {
		logMsg += fmt.Sprintf(" [%s]", l.prefix)
	}

	// Add context fields
	if ctx != nil {
		if reqID := ctx.Value(RequestIDKey); reqID != nil {
			logMsg += fmt.Sprintf(" [req_id=%v]", reqID)
		}
		if userID := ctx.Value(UserIDKey); userID != nil {
			logMsg += fmt.Sprintf(" [user_id=%v]", userID)
		}
	}

	logMsg += fmt.Sprintf(" %s", msg)

	// Add additional fields
	if len(fields) > 0 {
		logMsg += " |"
		for key, value := range fields {
			logMsg += fmt.Sprintf(" %s=%v", key, value)
		}
	}

	return logMsg
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.DebugContext(nil, msg, nil)
}

// DebugContext logs a debug message with context and fields
func (l *Logger) DebugContext(ctx context.Context, msg string, fields map[string]interface{}) {
	l.logger.Println(l.formatMessage(LevelDebug, ctx, msg, fields))
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.InfoContext(nil, msg, nil)
}

// InfoContext logs an info message with context and fields
func (l *Logger) InfoContext(ctx context.Context, msg string, fields map[string]interface{}) {
	l.logger.Println(l.formatMessage(LevelInfo, ctx, msg, fields))
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.WarnContext(nil, msg, nil)
}

// WarnContext logs a warning message with context and fields
func (l *Logger) WarnContext(ctx context.Context, msg string, fields map[string]interface{}) {
	l.logger.Println(l.formatMessage(LevelWarn, ctx, msg, fields))
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.ErrorContext(nil, msg, nil)
}

// ErrorContext logs an error message with context and fields
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields map[string]interface{}) {
	l.logger.Println(l.formatMessage(LevelError, ctx, msg, fields))
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(msg string) {
	l.FatalContext(nil, msg, nil)
}

// FatalContext logs a fatal error with context and fields, then exits
func (l *Logger) FatalContext(ctx context.Context, msg string, fields map[string]interface{}) {
	l.logger.Println(l.formatMessage(LevelFatal, ctx, msg, fields))
	os.Exit(1)
}

// Default logger functions

// Debug logs a debug message using the default logger
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

// DebugContext logs a debug message with context using the default logger
func DebugContext(ctx context.Context, msg string, fields map[string]interface{}) {
	defaultLogger.DebugContext(ctx, msg, fields)
}

// Info logs an info message using the default logger
func Info(msg string) {
	defaultLogger.Info(msg)
}

// InfoContext logs an info message with context using the default logger
func InfoContext(ctx context.Context, msg string, fields map[string]interface{}) {
	defaultLogger.InfoContext(ctx, msg, fields)
}

// Warn logs a warning message using the default logger
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// WarnContext logs a warning message with context using the default logger
func WarnContext(ctx context.Context, msg string, fields map[string]interface{}) {
	defaultLogger.WarnContext(ctx, msg, fields)
}

// Error logs an error message using the default logger
func Error(msg string) {
	defaultLogger.Error(msg)
}

// ErrorContext logs an error message with context using the default logger
func ErrorContext(ctx context.Context, msg string, fields map[string]interface{}) {
	defaultLogger.ErrorContext(ctx, msg, fields)
}

// Fatal logs a fatal error using the default logger and exits
func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

// FatalContext logs a fatal error with context using the default logger and exits
func FatalContext(ctx context.Context, msg string, fields map[string]interface{}) {
	defaultLogger.FatalContext(ctx, msg, fields)
}
