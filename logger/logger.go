package logger

import (
	"encoding/json"
	"fmt"
	"io"
	logGolang "log"
	"os"
	"path"
	"runtime"
	"time"
)

// LogFormat defines the output format of the logger.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

type Logger struct {
	isProduction bool
	writer       io.Writer
	format       LogFormat
}

type LogLevel string

const (
	infoLevel  LogLevel = "INFO"
	warnLevel  LogLevel = "WARN"
	errorLevel LogLevel = "ERROR"
	debugLevel LogLevel = "DEBUG"
)

// NewLogger creates a new Logger instance using text format.
// In production, debug logs are disabled.
func NewLogger(isProduction bool) *Logger {
	return &Logger{isProduction: isProduction, writer: os.Stdout, format: LogFormatText}
}

// NewLoggerWithFormat creates a new Logger with the specified format.
// In production, debug logs are disabled.
func NewLoggerWithFormat(isProduction bool, format LogFormat) *Logger {
	return &Logger{isProduction: isProduction, writer: os.Stdout, format: format}
}

// WithWriter returns a new Logger with a custom writer.
// Useful for testing — inject a bytes.Buffer to capture output.
func (l *Logger) WithWriter(w io.Writer) *Logger {
	return &Logger{isProduction: l.isProduction, writer: w, format: l.format}
}

// caller extracts the file and line of the real caller (2 levels up).
func (l *Logger) caller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "???", 0
	}
	return path.Base(file), line
}

func (l *Logger) log(level LogLevel, fileName string, line int, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	if l.format == LogFormatJSON {
		entry := map[string]any{
			"level": string(level),
			"ts":    time.Now().Format(time.RFC3339),
			"file":  fileName,
			"line":  line,
			"msg":   message,
		}
		b, _ := json.Marshal(entry)
		if level == errorLevel {
			logGolang.Fatalln(string(b))
		}
		fmt.Fprintln(l.writer, string(b))
		return
	}

	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[KEEL] [%s] [%s] [%s:%d] %s", timeStamp, level, fileName, line, message)
	if level == errorLevel {
		logGolang.Fatalln(logLine)
	}
	fmt.Fprintln(l.writer, logLine)
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(infoLevel, f, line, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(warnLevel, f, line, format, args...)
}

// Error logs an error message and exits the application.
func (l *Logger) Error(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(errorLevel, f, line, format, args...)
}

// Debug logs a debug message. Disabled in production.
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.isProduction {
		f, line := l.caller()
		l.log(debugLevel, f, line, format, args...)
	}
}
