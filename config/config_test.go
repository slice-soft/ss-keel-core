package config

import (
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		wantPanic bool
	}{
		{
			name:      "existing variable returns value",
			envKey:    "TEST_GET_ENV",
			envValue:  "hello",
			wantPanic: false,
		},
		{
			name:      "missing variable panics",
			envKey:    "TEST_GET_ENV_MISSING",
			envValue:  "",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not panic")
					}
				}()
				GetEnv(tt.envKey)
				return
			}

			got := GetEnv(tt.envKey)
			if got != tt.envValue {
				t.Errorf("GetEnv() = %v, want %v", got, tt.envValue)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "existing variable returns value",
			envKey:       "TEST_OR_DEFAULT",
			envValue:     "custom",
			defaultValue: "default",
			want:         "custom",
		},
		{
			name:         "missing variable returns default",
			envKey:       "TEST_OR_DEFAULT_MISSING",
			envValue:     "",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := GetEnvOrDefault(tt.envKey, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		want      int
		wantPanic bool
	}{
		{
			name:     "valid integer",
			envKey:   "TEST_INT",
			envValue: "42",
			want:     42,
		},
		{
			name:      "invalid integer panics",
			envKey:    "TEST_INT_INVALID",
			envValue:  "notanint",
			wantPanic: true,
		},
		{
			name:      "missing variable panics",
			envKey:    "TEST_INT_MISSING",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not panic")
					}
				}()
				GetEnvInt(tt.envKey)
				return
			}

			got := GetEnvInt(tt.envKey)
			if got != tt.want {
				t.Errorf("GetEnvInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvIntOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		want         int
	}{
		{
			name:         "valid integer returns value",
			envKey:       "TEST_INT_OR_DEFAULT",
			envValue:     "8080",
			defaultValue: 3000,
			want:         8080,
		},
		{
			name:         "missing variable returns default",
			envKey:       "TEST_INT_OR_DEFAULT_MISSING",
			defaultValue: 3000,
			want:         3000,
		},
		{
			name:         "invalid integer returns default",
			envKey:       "TEST_INT_OR_DEFAULT_INVALID",
			envValue:     "notanint",
			defaultValue: 3000,
			want:         3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := GetEnvIntOrDefault(tt.envKey, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvIntOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvUint(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		want      uint
		wantPanic bool
	}{
		{
			name:     "valid uint",
			envKey:   "TEST_UINT",
			envValue: "100",
			want:     100,
		},
		{
			name:      "negative value panics",
			envKey:    "TEST_UINT_NEGATIVE",
			envValue:  "-1",
			wantPanic: true,
		},
		{
			name:      "invalid value panics",
			envKey:    "TEST_UINT_INVALID",
			envValue:  "notauint",
			wantPanic: true,
		},
		{
			name:      "missing variable panics",
			envKey:    "TEST_UINT_MISSING",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not panic")
					}
				}()
				GetEnvUint(tt.envKey)
				return
			}

			got := GetEnvUint(tt.envKey)
			if got != tt.want {
				t.Errorf("GetEnvUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		want      bool
		wantPanic bool
	}{
		{
			name:     "true value",
			envKey:   "TEST_BOOL_TRUE",
			envValue: "true",
			want:     true,
		},
		{
			name:     "false value",
			envKey:   "TEST_BOOL_FALSE",
			envValue: "false",
			want:     false,
		},
		{
			name:     "1 is true",
			envKey:   "TEST_BOOL_ONE",
			envValue: "1",
			want:     true,
		},
		{
			name:     "0 is false",
			envKey:   "TEST_BOOL_ZERO",
			envValue: "0",
			want:     false,
		},
		{
			name:      "invalid value panics",
			envKey:    "TEST_BOOL_INVALID",
			envValue:  "notabool",
			wantPanic: true,
		},
		{
			name:      "missing variable panics",
			envKey:    "TEST_BOOL_MISSING",
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic but did not panic")
					}
				}()
				GetEnvBool(tt.envKey)
				return
			}

			got := GetEnvBool(tt.envKey)
			if got != tt.want {
				t.Errorf("GetEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvBoolOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		want         bool
	}{
		{
			name:         "true value returns true",
			envKey:       "TEST_BOOL_OR_DEFAULT_TRUE",
			envValue:     "true",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "missing variable returns default true",
			envKey:       "TEST_BOOL_OR_DEFAULT_MISSING_T",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "missing variable returns default false",
			envKey:       "TEST_BOOL_OR_DEFAULT_MISSING_F",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "invalid value returns default",
			envKey:       "TEST_BOOL_OR_DEFAULT_INVALID",
			envValue:     "notabool",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := GetEnvBoolOrDefault(tt.envKey, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvBoolOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
