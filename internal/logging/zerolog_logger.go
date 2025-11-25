// Package logging provides high-performance structured logging using zerolog.
// It replaces the custom StructuredLogger with zerolog for better performance
// and standardized JSON output, with request ID correlation and context-aware methods.
package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Context keys for extracting request metadata
type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	UserIDKey    ctxKey = "user_id"
	TenantIDKey  ctxKey = "tenant_id"
	OperationKey ctxKey = "operation"
)

// Logger wraps zerolog.Logger with application-specific methods
type Logger struct {
	zl zerolog.Logger
}

// Config holds logger configuration options
type Config struct {
	Level       string // debug, info, warn, error
	Pretty      bool   // Use console writer for development
	ServiceName string
	Version     string
}

// DefaultConfig returns production-ready default configuration
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Pretty:      false,
		ServiceName: "agromart",
		Version:     "1.0.0",
	}
}

// New creates a new Logger instance with the given configuration
func New(cfg Config) *Logger {
	// Set global log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure timestamp format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var output io.Writer = os.Stdout

	// Use console writer for pretty printing in development
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	// Create base logger with service context
	zl := zerolog.New(output).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("version", cfg.Version).
		Logger()

	return &Logger{zl: zl}
}

// NewFromEnv creates a logger configured from environment variables
func NewFromEnv() *Logger {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	pretty := os.Getenv("LOG_PRETTY") == "true" || os.Getenv("ENV") == "development"

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "agromart"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0"
	}

	return New(Config{
		Level:       level,
		Pretty:      pretty,
		ServiceName: serviceName,
		Version:     version,
	})
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a new logger with context fields extracted
func (l *Logger) WithContext(ctx context.Context) *Logger {
	zl := l.zl.With().Logger()

	// Extract request ID
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		zl = zl.With().Str("request_id", reqID).Logger()
	}

	// Extract user ID
	if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok && userID != uuid.Nil {
		zl = zl.With().Str("user_id", userID.String()).Logger()
	}

	// Extract tenant ID
	if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && tenantID != uuid.Nil {
		zl = zl.With().Str("tenant_id", tenantID.String()).Logger()
	}

	// Extract operation
	if op, ok := ctx.Value(OperationKey).(string); ok && op != "" {
		zl = zl.With().Str("operation", op).Logger()
	}

	return &Logger{zl: zl}
}

// With returns a new logger with the specified field added
func (l *Logger) With() *LoggerContext {
	return &LoggerContext{ctx: l.zl.With()}
}

// LoggerContext provides a fluent interface for adding context fields
type LoggerContext struct {
	ctx zerolog.Context
}

// Str adds a string field to the logger context
func (c *LoggerContext) Str(key, val string) *LoggerContext {
	c.ctx = c.ctx.Str(key, val)
	return c
}

// Int adds an int field to the logger context
func (c *LoggerContext) Int(key string, val int) *LoggerContext {
	c.ctx = c.ctx.Int(key, val)
	return c
}

// UUID adds a UUID field to the logger context
func (c *LoggerContext) UUID(key string, val uuid.UUID) *LoggerContext {
	c.ctx = c.ctx.Str(key, val.String())
	return c
}

// Err adds an error field to the logger context
func (c *LoggerContext) Err(err error) *LoggerContext {
	c.ctx = c.ctx.Err(err)
	return c
}

// Logger returns the final Logger with all context fields
func (c *LoggerContext) Logger() *Logger {
	return &Logger{zl: c.ctx.Logger()}
}

// Debug logs a debug message
func (l *Logger) Debug() *Event {
	return &Event{e: l.zl.Debug()}
}

// Info logs an info message
func (l *Logger) Info() *Event {
	return &Event{e: l.zl.Info()}
}

// Warn logs a warning message
func (l *Logger) Warn() *Event {
	return &Event{e: l.zl.Warn()}
}

// Error logs an error message
func (l *Logger) Error() *Event {
	return &Event{e: l.zl.Error()}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal() *Event {
	return &Event{e: l.zl.Fatal()}
}

// IsDebugEnabled returns true if debug level is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.zl.Debug().Enabled()
}

// Event wraps zerolog.Event for fluent logging
type Event struct {
	e *zerolog.Event
}

// Str adds a string field
func (e *Event) Str(key, val string) *Event {
	e.e = e.e.Str(key, val)
	return e
}

// Int adds an int field
func (e *Event) Int(key string, val int) *Event {
	e.e = e.e.Int(key, val)
	return e
}

// Int64 adds an int64 field
func (e *Event) Int64(key string, val int64) *Event {
	e.e = e.e.Int64(key, val)
	return e
}

// Float64 adds a float64 field
func (e *Event) Float64(key string, val float64) *Event {
	e.e = e.e.Float64(key, val)
	return e
}

// Bool adds a bool field
func (e *Event) Bool(key string, val bool) *Event {
	e.e = e.e.Bool(key, val)
	return e
}

// UUID adds a UUID field
func (e *Event) UUID(key string, val uuid.UUID) *Event {
	e.e = e.e.Str(key, val.String())
	return e
}

// Err adds an error field
func (e *Event) Err(err error) *Event {
	e.e = e.e.Err(err)
	return e
}

// Dur adds a duration field
func (e *Event) Dur(key string, d time.Duration) *Event {
	e.e = e.e.Dur(key, d)
	return e
}

// Time adds a time field
func (e *Event) Time(key string, t time.Time) *Event {
	e.e = e.e.Time(key, t)
	return e
}

// Interface adds any field using reflection
func (e *Event) Interface(key string, val interface{}) *Event {
	e.e = e.e.Interface(key, val)
	return e
}

// Msg sends the log message
func (e *Event) Msg(msg string) {
	e.e.Msg(msg)
}

// Msgf sends a formatted log message
func (e *Event) Msgf(format string, args ...interface{}) {
	e.e.Msgf(format, args...)
}

// Send sends the log event without a message
func (e *Event) Send() {
	e.e.Send()
}

// Enabled returns whether the event is enabled (level is high enough)
func (e *Event) Enabled() bool {
	return e.e.Enabled()
}

// Global logger instance
var global *Logger

func init() {
	global = NewFromEnv()
}

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) {
	global = l
}

// Global returns the global logger instance
func Global() *Logger {
	return global
}

// Convenience functions using global logger

// Debug logs a debug message using the global logger
func Debug() *Event {
	return global.Debug()
}

// Info logs an info message using the global logger
func Info() *Event {
	return global.Info()
}

// Warn logs a warning message using the global logger
func Warn() *Event {
	return global.Warn()
}

// Error logs an error message using the global logger
func Error() *Event {
	return global.Error()
}

// Fatal logs a fatal message using the global logger
func Fatal() *Event {
	return global.Fatal()
}

// WithContext returns a logger with context fields using the global logger
func WithContext(ctx context.Context) *Logger {
	return global.WithContext(ctx)
}

// ContextWithRequestID adds a request ID to the context
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// ContextWithUserID adds a user ID to the context
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// ContextWithTenantID adds a tenant ID to the context
func ContextWithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// ContextWithOperation adds an operation name to the context
func ContextWithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, OperationKey, operation)
}
