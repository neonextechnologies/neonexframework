package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Color returns ANSI color code for log level
func (l LogLevel) Color() string {
	switch l {
	case DebugLevel:
		return "\033[36m" // Cyan
	case InfoLevel:
		return "\033[32m" // Green
	case WarnLevel:
		return "\033[33m" // Yellow
	case ErrorLevel:
		return "\033[31m" // Red
	case FatalLevel:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

// Fields represents structured log fields
type Fields map[string]interface{}

// Logger interface defines logging methods
type Logger interface {
	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)
	Fatal(msg string, fields ...Fields)

	With(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	SetLevel(level LogLevel)
	SetFormatter(formatter Formatter)
	AddWriter(writer io.Writer)
}

// Entry represents a log entry
type Entry struct {
	Time    time.Time
	Level   LogLevel
	Message string
	Fields  Fields
	File    string
	Line    int
	Context context.Context
}

// Formatter interface for log formatting
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// StandardLogger is the default logger implementation
type StandardLogger struct {
	mu        sync.RWMutex
	level     LogLevel
	formatter Formatter
	writers   []io.Writer
	fields    Fields
	ctx       context.Context
	caller    bool
	colorize  bool
}

// NewLogger creates a new logger instance
func NewLogger() *StandardLogger {
	return &StandardLogger{
		level:     InfoLevel,
		formatter: NewTextFormatter(),
		writers:   []io.Writer{os.Stdout},
		fields:    make(Fields),
		caller:    true,
		colorize:  true,
	}
}

// SetLevel sets the minimum log level
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetFormatter sets the log formatter
func (l *StandardLogger) SetFormatter(formatter Formatter) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.formatter = formatter
}

// AddWriter adds a writer to the logger
func (l *StandardLogger) AddWriter(writer io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, writer)
}

// EnableCaller enables/disables caller information
func (l *StandardLogger) EnableCaller(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.caller = enabled
}

// EnableColor enables/disables colored output
func (l *StandardLogger) EnableColor(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colorize = enabled
}

// With returns a new logger with additional fields
func (l *StandardLogger) With(fields Fields) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &StandardLogger{
		level:     l.level,
		formatter: l.formatter,
		writers:   l.writers,
		fields:    newFields,
		ctx:       l.ctx,
		caller:    l.caller,
		colorize:  l.colorize,
	}
}

// WithContext returns a new logger with context
func (l *StandardLogger) WithContext(ctx context.Context) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &StandardLogger{
		level:     l.level,
		formatter: l.formatter,
		writers:   l.writers,
		fields:    l.fields,
		ctx:       ctx,
		caller:    l.caller,
		colorize:  l.colorize,
	}
}

// Debug logs a debug message
func (l *StandardLogger) Debug(msg string, fields ...Fields) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *StandardLogger) Info(msg string, fields ...Fields) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *StandardLogger) Warn(msg string, fields ...Fields) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *StandardLogger) Error(msg string, fields ...Fields) {
	l.log(ErrorLevel, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *StandardLogger) Fatal(msg string, fields ...Fields) {
	l.log(FatalLevel, msg, fields...)
	os.Exit(1)
}

// log handles the actual logging
func (l *StandardLogger) log(level LogLevel, msg string, fieldsSlice ...Fields) {
	l.mu.RLock()

	// Check if we should log this level
	if level < l.level {
		l.mu.RUnlock()
		return
	}

	// Merge fields
	mergedFields := make(Fields)
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for _, fields := range fieldsSlice {
		for k, v := range fields {
			mergedFields[k] = v
		}
	}

	// Get caller info
	file := ""
	line := 0
	if l.caller {
		_, f, ln, ok := runtime.Caller(2)
		if ok {
			file = f
			line = ln
			// Shorten file path
			if idx := strings.LastIndex(file, "/"); idx >= 0 {
				file = file[idx+1:]
			}
		}
	}

	entry := &Entry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  mergedFields,
		File:    file,
		Line:    line,
		Context: l.ctx,
	}

	formatter := l.formatter
	writers := l.writers
	l.mu.RUnlock()

	// Format the entry
	formatted, err := formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format log entry: %v\n", err)
		return
	}

	// Write to all writers
	for _, writer := range writers {
		writer.Write(formatted)
	}
}

// Global logger instance
var defaultLogger = NewLogger()

// SetGlobalLevel sets the global logger level
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetGlobalFormatter sets the global logger formatter
func SetGlobalFormatter(formatter Formatter) {
	defaultLogger.SetFormatter(formatter)
}

// AddGlobalWriter adds a writer to the global logger
func AddGlobalWriter(writer io.Writer) {
	defaultLogger.AddWriter(writer)
}

// EnableGlobalCaller enables caller info for global logger
func EnableGlobalCaller(enabled bool) {
	defaultLogger.EnableCaller(enabled)
}

// EnableGlobalColor enables colored output for global logger
func EnableGlobalColor(enabled bool) {
	defaultLogger.EnableColor(enabled)
}

// Global logging functions
func Debug(msg string, fields ...Fields) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...Fields) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...Fields) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...Fields) {
	defaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...Fields) {
	defaultLogger.Fatal(msg, fields...)
}

func With(fields Fields) Logger {
	return defaultLogger.With(fields)
}

func WithContext(ctx context.Context) Logger {
	return defaultLogger.WithContext(ctx)
}
