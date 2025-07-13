package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

// GinStyleFormatter formats logs to match Gin's style
type GinStyleFormatter struct{}

// Format implements logrus.Formatter interface
func (f *GinStyleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006/01/02 - 15:04:05")
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message
	
	logLine := fmt.Sprintf("[APP] %s | %s | %s\n", timestamp, level, message)
	return []byte(logLine), nil
}

// InitLogger initializes the global logger with specified level
func InitLogger(levelStr string) {
	logger = logrus.New()
	
	// Set log level
	level := parseLogLevel(levelStr)
	logger.SetLevel(level)
	
	// Set Gin-style formatter
	logger.SetFormatter(&GinStyleFormatter{})
}

// InitFileLogging sets up file logging in addition to console logging
func InitFileLogging(logDir string) error {
	if logger == nil {
		return nil
	}

	// Ensure log directory exists
	if err := EnsureDir(logDir); err != nil {
		return err
	}

	// Create system log file
	logFile := filepath.Join(logDir, "system.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Create multi-writer for both console and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(multiWriter)

	return nil
}

// parseLogLevel converts string to logrus.Level
func parseLogLevel(levelStr string) logrus.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// SetAsDefaultLogger sets our logger as the default for the standard log package
func SetAsDefaultLogger() {
	if logger != nil {
		// Create a writer that uses logrus
		writer := logger.Writer()
		defer writer.Close()
		
		// Redirect standard log to logrus
		logrus.SetOutput(logger.Out)
		logrus.SetLevel(logger.Level)
		logrus.SetFormatter(logger.Formatter)
	}
}

// Debug logs debug level messages
func Debug(format string, v ...interface{}) {
	if logger != nil {
		logger.Debugf(format, v...)
	}
}

// Info logs info level messages
func Info(format string, v ...interface{}) {
	if logger != nil {
		logger.Infof(format, v...)
	}
}

// Warn logs warning level messages
func Warn(format string, v ...interface{}) {
	if logger != nil {
		logger.Warnf(format, v...)
	}
}

// Error logs error level messages
func Error(format string, v ...interface{}) {
	if logger != nil {
		logger.Errorf(format, v...)
	}
}

// Fatal logs error level messages and exits
func Fatal(format string, v ...interface{}) {
	if logger != nil {
		logger.Fatalf(format, v...)
	} else {
		os.Exit(1)
	}
}

// GetLogLevel returns the current log level as string
func GetLogLevel() string {
	if logger != nil {
		return logger.Level.String()
	}
	return "info"
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return logger != nil && logger.Level >= logrus.DebugLevel
}

// IsInfoEnabled returns true if info logging is enabled
func IsInfoEnabled() bool {
	return logger != nil && logger.Level >= logrus.InfoLevel
}

// GetLogger returns the underlying logrus logger for advanced usage
func GetLogger() *logrus.Logger {
	return logger
}