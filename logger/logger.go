package logger

import (
	"fmt"
	logGolang "log"
	"os"
	"path"
	"runtime"
	"time"
)

type Logger struct {
	isProduction bool
}

type LogLevel string

const (
	infoLevel  LogLevel = "INFO"
	warnLevel  LogLevel = "WARN"
	errorLevel LogLevel = "ERROR"
	debugLevel LogLevel = "DEBUG"
)

func NewLogger(isProduction bool) *Logger {
	return &Logger{isProduction: isProduction}
}

// caller extracts the file and line of the real caller (2 levels up)
func (l *Logger) caller() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "???", 0
	}
	return path.Base(file), line
}

func (l *Logger) log(level LogLevel, fileName string, line int, format string, args ...interface{}) {
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s", timeStamp, level, fileName, line, message)
	if level == errorLevel {
		logGolang.Fatalln(logLine)
	}
	fmt.Fprintln(os.Stdout, logLine)
}

func (l *Logger) Info(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(infoLevel, f, line, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(warnLevel, f, line, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	f, line := l.caller()
	l.log(errorLevel, f, line, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.isProduction {
		f, line := l.caller()
		l.log(debugLevel, f, line, format, args...)
	}
}
