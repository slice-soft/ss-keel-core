package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// newTestLogger creates a Logger with a buffer for capturing output.
func newTestLogger(isProduction bool) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := NewLogger(isProduction).WithWriter(buf)
	return log, buf
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name         string
		isProduction bool
	}{
		{
			name:         "development logger",
			isProduction: false,
		},
		{
			name:         "production logger",
			isProduction: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewLogger(tt.isProduction)
			if log == nil {
				t.Fatal("NewLogger() returned nil")
			}
			if log.isProduction != tt.isProduction {
				t.Errorf("isProduction = %v, want %v", log.isProduction, tt.isProduction)
			}
			if log.writer == nil {
				t.Error("writer should not be nil")
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantMsg string
		wantLvl string
	}{
		{
			name:    "simple message",
			format:  "server started",
			wantMsg: "server started",
			wantLvl: "INFO",
		},
		{
			name:    "formatted message",
			format:  "listening on port %d",
			args:    []interface{}{3000},
			wantMsg: "listening on port 3000",
			wantLvl: "INFO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newTestLogger(false)
			log.Info(tt.format, tt.args...)

			output := buf.String()
			if !strings.Contains(output, tt.wantLvl) {
				t.Errorf("output missing level %v, got: %v", tt.wantLvl, output)
			}
			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("output missing message %v, got: %v", tt.wantMsg, output)
			}
		})
	}
}

func TestWarn(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		args    []interface{}
		wantMsg string
		wantLvl string
	}{
		{
			name:    "simple warning",
			format:  "something looks off",
			wantMsg: "something looks off",
			wantLvl: "WARN",
		},
		{
			name:    "formatted warning",
			format:  "HTTP Error [%d]: %s",
			args:    []interface{}{404, "not found"},
			wantMsg: "HTTP Error [404]: not found",
			wantLvl: "WARN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newTestLogger(false)
			log.Warn(tt.format, tt.args...)

			output := buf.String()
			if !strings.Contains(output, tt.wantLvl) {
				t.Errorf("output missing level %v, got: %v", tt.wantLvl, output)
			}
			if !strings.Contains(output, tt.wantMsg) {
				t.Errorf("output missing message %v, got: %v", tt.wantMsg, output)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name         string
		isProduction bool
		format       string
		wantOutput   bool
	}{
		{
			name:         "debug visible in development",
			isProduction: false,
			format:       "route registered",
			wantOutput:   true,
		},
		{
			name:         "debug hidden in production",
			isProduction: true,
			format:       "route registered",
			wantOutput:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newTestLogger(tt.isProduction)
			log.Debug(tt.format)

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.wantOutput {
				t.Errorf("wantOutput = %v, got output: %q", tt.wantOutput, output)
			}
			if tt.wantOutput && !strings.Contains(output, "DEBUG") {
				t.Errorf("output missing DEBUG level, got: %v", output)
			}
		})
	}
}

func TestLogFormat(t *testing.T) {
	tests := []struct {
		name         string
		logFunc      func(l *Logger)
		wantLevel    string
		wantContains []string
	}{
		{
			name:      "info format contains timestamp level file and message",
			logFunc:   func(l *Logger) { l.Info("test message") },
			wantLevel: "INFO",
			wantContains: []string{
				"INFO",
				"logger_test.go",
				"test message",
			},
		},
		{
			name:      "warn format contains timestamp level file and message",
			logFunc:   func(l *Logger) { l.Warn("warn message") },
			wantLevel: "WARN",
			wantContains: []string{
				"WARN",
				"logger_test.go",
				"warn message",
			},
		},
		{
			name:      "debug format contains timestamp level file and message",
			logFunc:   func(l *Logger) { l.Debug("debug message") },
			wantLevel: "DEBUG",
			wantContains: []string{
				"DEBUG",
				"logger_test.go",
				"debug message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, buf := newTestLogger(false)
			tt.logFunc(log)

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q, got: %v", want, output)
				}
			}
		})
	}
}

func TestWithWriter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "WithWriter returns new logger with custom writer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := NewLogger(false)
			buf := &bytes.Buffer{}
			custom := original.WithWriter(buf)

			if custom == original {
				t.Error("WithWriter() should return a new Logger instance")
			}
			if custom.writer != buf {
				t.Error("WithWriter() writer should be the provided buffer")
			}
			if custom.isProduction != original.isProduction {
				t.Error("WithWriter() should preserve isProduction")
			}
			if custom.format != original.format {
				t.Error("WithWriter() should preserve format")
			}
		})
	}
}

func TestNewLoggerWithFormat(t *testing.T) {
	t.Run("text format by default", func(t *testing.T) {
		log := NewLogger(false)
		if log.format != LogFormatText {
			t.Errorf("format = %v, want text", log.format)
		}
	})

	t.Run("JSON format constructor", func(t *testing.T) {
		log := NewLoggerWithFormat(false, LogFormatJSON)
		if log.format != LogFormatJSON {
			t.Errorf("format = %v, want json", log.format)
		}
	})
}

func TestJSONLogFormat(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(l *Logger)
		wantLevel string
		wantMsg   string
	}{
		{
			name:      "info produces valid JSON",
			logFunc:   func(l *Logger) { l.Info("server started on port %d", 3000) },
			wantLevel: "INFO",
			wantMsg:   "server started on port 3000",
		},
		{
			name:      "warn produces valid JSON",
			logFunc:   func(l *Logger) { l.Warn("something failed") },
			wantLevel: "WARN",
			wantMsg:   "something failed",
		},
		{
			name:      "debug produces valid JSON in dev",
			logFunc:   func(l *Logger) { l.Debug("debug event") },
			wantLevel: "DEBUG",
			wantMsg:   "debug event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log := NewLoggerWithFormat(false, LogFormatJSON).WithWriter(buf)
			tt.logFunc(log)

			line := strings.TrimSpace(buf.String())
			if line == "" {
				t.Fatal("no output produced")
			}

			var entry map[string]any
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("output is not valid JSON: %v — got: %q", err, line)
			}

			if entry["level"] != tt.wantLevel {
				t.Errorf("level = %v, want %v", entry["level"], tt.wantLevel)
			}
			if entry["msg"] != tt.wantMsg {
				t.Errorf("msg = %v, want %v", entry["msg"], tt.wantMsg)
			}
			if entry["ts"] == "" || entry["ts"] == nil {
				t.Error("ts field should be present")
			}
			if entry["file"] == "" || entry["file"] == nil {
				t.Error("file field should be present")
			}
			if entry["line"] == nil {
				t.Error("line field should be present")
			}
		})
	}
}
