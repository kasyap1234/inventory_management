package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Duration  *time.Duration         `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() *StructuredLogger {
	return &StructuredLogger{
		logger: log.New(os.Stdout, "", 0),
		level:  LogLevelInfo,
	}
}

// SetLevel sets the minimum log level
func (l *StructuredLogger) SetLevel(level LogLevel) {
	l.level = level
}

// shouldLog determines if a message should be logged based on level
func (l *StructuredLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levels[level] >= levels[l.level]
}

// log writes a structured log entry
func (l *StructuredLogger) log(entry LogEntry) {
	if !l.shouldLog(entry.Level) {
		return
	}

	entry.Timestamp = time.Now()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.logger.Printf("LOG_ERROR: Failed to marshal log entry: %v", err)
		l.logger.Printf("%s [%s] %s", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
		return
	}

	l.logger.Println(string(jsonData))
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string) {
	l.log(LogEntry{
		Level:   LogLevelDebug,
		Message: message,
	})
}

// DebugWithContext logs a debug message with context
func (l *StructuredLogger) DebugWithContext(ctx context.Context, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:   LogLevelDebug,
		Message: message,
		Fields:  fields,
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string) {
	l.log(LogEntry{
		Level:   LogLevelInfo,
		Message: message,
	})
}

// InfoWithContext logs an info message with context
func (l *StructuredLogger) InfoWithContext(ctx context.Context, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: message,
		Fields:  fields,
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string) {
	l.log(LogEntry{
		Level:   LogLevelWarn,
		Message: message,
	})
}

// WarnWithContext logs a warning message with context
func (l *StructuredLogger) WarnWithContext(ctx context.Context, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Level:   LogLevelWarn,
		Message: message,
		Fields:  fields,
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, err error) {
	entry := LogEntry{
		Level:   LogLevelError,
		Message: message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.log(entry)
}

// ErrorWithContext logs an error message with context
func (l *StructuredLogger) ErrorWithContext(ctx context.Context, message string, err error, fields map[string]interface{}) {
	entry := LogEntry{
		Level:   LogLevelError,
		Message: message,
		Fields:  fields,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(message string, err error) {
	entry := LogEntry{
		Level:   LogLevelFatal,
		Message: message,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	l.log(entry)
	os.Exit(1)
}

// LogError implements ErrorLogger interface for EnhancedErrorHandler
func (l *StructuredLogger) LogError(ctx context.Context, err *EnhancedError) {
	entry := LogEntry{
		Level:     LogLevelError,
		Message:   fmt.Sprintf("Error in %s: %s", err.Context.Operation, err.Message),
		RequestID: err.Context.RequestID,
		UserID:    err.Context.UserID,
		TenantID:  err.Context.TenantID,
		Operation: err.Context.Operation,
		Resource:  err.Context.Resource,
		Fields: map[string]interface{}{
			"error_code":     err.Code,
			"error_severity": err.Severity,
			"method":         err.Context.Method,
			"path":           err.Context.Path,
			"user_agent":     err.Context.UserAgent,
			"ip_address":     err.Context.IPAddress,
		},
	}

	if err.Cause != nil {
		entry.Error = err.Cause.Error()
	}

	if err.Details != nil {
		entry.Fields["error_details"] = err.Details
	}

	if err.Context.StackTrace != "" {
		entry.Fields["stack_trace"] = err.Context.StackTrace
	}

	l.log(entry)
}

// LogCriticalError implements ErrorLogger interface for critical errors
func (l *StructuredLogger) LogCriticalError(ctx context.Context, err *EnhancedError) {
	// Log as error first
	l.LogError(ctx, err)

	// Send additional alert for critical errors
	l.sendCriticalAlert(err)
}

// sendCriticalAlert sends alerts for critical errors
func (l *StructuredLogger) sendCriticalAlert(err *EnhancedError) {
	// This would integrate with alerting systems like PagerDuty, Slack, etc.
	// For now, we'll log a special alert entry
	alertEntry := LogEntry{
		Level:   LogLevelError,
		Message: "CRITICAL ALERT: System requires immediate attention",
		Fields: map[string]interface{}{
			"alert_type":    "critical_error",
			"error_code":    err.Code,
			"error_message": err.Message,
			"operation":     err.Context.Operation,
			"request_id":    err.Context.RequestID,
			"tenant_id":     err.Context.TenantID,
			"user_id":       err.Context.UserID,
			"timestamp":     err.Context.Timestamp,
		},
	}

	l.log(alertEntry)
}

// LogOperation logs the start and completion of an operation
func (l *StructuredLogger) LogOperation(ctx context.Context, operation string, resource string, fn func() error) error {
	start := time.Now()

	// Log operation start
	l.InfoWithContext(ctx, fmt.Sprintf("Starting %s", operation), map[string]interface{}{
		"operation": operation,
		"resource":  resource,
	})

	// Execute operation
	err := fn()
	duration := time.Since(start)

	// Log operation completion
	if err != nil {
		l.ErrorWithContext(ctx, fmt.Sprintf("Failed %s", operation), err, map[string]interface{}{
			"operation": operation,
			"resource":  resource,
			"duration":  duration.String(),
		})
	} else {
		l.InfoWithContext(ctx, fmt.Sprintf("Completed %s", operation), map[string]interface{}{
			"operation": operation,
			"resource":  resource,
			"duration":  duration.String(),
		})
	}

	return err
}

// LogPerformance logs performance metrics
func (l *StructuredLogger) LogPerformance(ctx context.Context, operation string, duration time.Duration, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"operation":       operation,
		"duration_ms":     duration.Milliseconds(),
		"performance_log": true,
	}

	// Merge metadata
	for k, v := range metadata {
		fields[k] = v
	}

	level := LogLevelInfo
	message := fmt.Sprintf("Performance: %s completed in %s", operation, duration)

	// Flag slow operations
	if duration > time.Second*5 {
		level = LogLevelWarn
		message = fmt.Sprintf("SLOW OPERATION: %s took %s", operation, duration)
		fields["slow_operation"] = true
	}

	entry := LogEntry{
		Level:   level,
		Message: message,
		Fields:  fields,
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// LogAudit logs audit events
func (l *StructuredLogger) LogAudit(ctx context.Context, action string, resource string, resourceID string, changes map[string]interface{}) {
	entry := LogEntry{
		Level:   LogLevelInfo,
		Message: fmt.Sprintf("Audit: %s %s", action, resource),
		Fields: map[string]interface{}{
			"audit_log":   true,
			"action":      action,
			"resource":    resource,
			"resource_id": resourceID,
			"changes":     changes,
		},
	}
	l.addContextToEntry(ctx, &entry)
	l.log(entry)
}

// addContextToEntry adds context information to a log entry
func (l *StructuredLogger) addContextToEntry(ctx context.Context, entry *LogEntry) {
	if userID, ok := GetUserIDFromContext(ctx); ok {
		entry.UserID = &userID
	}
	if tenantID, ok := GetTenantIDFromContext(ctx); ok {
		entry.TenantID = &tenantID
	}

	// Try to get request ID from context
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			entry.RequestID = id
		}
	}
}

// Global logger instance
var globalLogger = NewStructuredLogger()

// SetGlobalLogLevel sets the global log level
func SetGlobalLogLevel(level LogLevel) {
	globalLogger.SetLevel(level)
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *StructuredLogger {
	return globalLogger
}
