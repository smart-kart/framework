package logger

import (
	"context"
	"log"
	"os"
)

type contextKey string

const loggerKey contextKey = "logger"

// Logger interface
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// defaultLogger implements Logger interface
type defaultLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// New creates a new logger
func New() Logger {
	return &defaultLogger{
		infoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLog: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	l.infoLog.Printf(msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	l.errorLog.Printf(msg, args...)
}

func (l *defaultLogger) Fatal(msg string, args ...interface{}) {
	l.errorLog.Fatalf(msg, args...)
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	l.debugLog.Printf(msg, args...)
}

// WithContext adds logger to context
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return New()
}

// RestrictedGet return basic logger for framework internal usage
func RestrictedGet() Logger {
	return New()
}