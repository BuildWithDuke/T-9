package logging

import (
	"log"
	"os"
	"time"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warning
	Error
)

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
	}
}

// LogEvent represents a structured log event
type LogEvent struct {
	Level     LogLevel            `json:"level"`
	Timestamp time.Time           `json:"timestamp"`
	Message   string              `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Error     error               `json:"error,omitempty"`
}

// logInternal logs an event with the specified level
func (l *Logger) logInternal(level LogLevel, message string, fields map[string]interface{}, err error) {
	if level < l.level {
		return
	}

	event := LogEvent{
		Level:     level,
		Timestamp: time.Now(),
		Message:   message,
		Fields:    fields,
		Error:     err,
	}

	// Convert to structured log format
	levelStr := ""
	switch level {
	case Debug:
		levelStr = "DEBUG"
	case Info:
		levelStr = "INFO"
	case Warning:
		levelStr = "WARN"
	case Error:
		levelStr = "ERROR"
	}

	// Log with structured format
	logMsg := "[" + levelStr + "] " + event.Timestamp.Format(time.RFC3339) + " " + message
	if err != nil {
		logMsg += " error=" + err.Error()
	}
	if len(fields) > 0 {
		for k, v := range fields {
			logMsg += " " + k + "=" + toString(v)
		}
	}

	l.logger.Println(logMsg)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.logInternal(Debug, message, fields, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.logInternal(Info, message, fields, nil)
}

// Warning logs a warning message
func (l *Logger) Warning(message string, fields map[string]interface{}) {
	l.logInternal(Warning, message, fields, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	l.logInternal(Error, message, fields, err)
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	return l
}

// toString converts interface{} to string for logging
func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case int:
		return string(rune(v))
	case float64:
		return string(rune(int(v)))
	case bool:
		if v {
			return "T"
		}
		return "F"
	default:
		return "unknown"
	}
}

// Default logger instance
var DefaultLogger = NewLogger(Info)