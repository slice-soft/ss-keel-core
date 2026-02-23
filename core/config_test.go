package core

import "testing"

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    KConfig
		expected KConfig
	}{
		{
			name:  "empty config applies all defaults",
			input: KConfig{},
			expected: KConfig{
				Port:        3000,
				ServiceName: "Keel App",
				Env:         "development",
				Docs: DocsConfig{
					Path:    "/docs",
					Title:   "Keel App",
					Version: "1.0.0",
				},
			},
		},
		{
			name: "partial config applies missing defaults",
			input: KConfig{
				Port:        8080,
				ServiceName: "My API",
			},
			expected: KConfig{
				Port:        8080,
				ServiceName: "My API",
				Env:         "development",
				Docs: DocsConfig{
					Path:    "/docs",
					Title:   "My API",
					Version: "1.0.0",
				},
			},
		},
		{
			name: "full config keeps all values",
			input: KConfig{
				Port:        9000,
				ServiceName: "Custom API",
				Env:         "production",
				Docs: DocsConfig{
					Path:    "/api-docs",
					Title:   "Custom Docs",
					Version: "2.0.0",
				},
			},
			expected: KConfig{
				Port:        9000,
				ServiceName: "Custom API",
				Env:         "production",
				Docs: DocsConfig{
					Path:    "/api-docs",
					Title:   "Custom Docs",
					Version: "2.0.0",
				},
			},
		},
		{
			name: "docs title defaults to service name",
			input: KConfig{
				ServiceName: "Orders API",
			},
			expected: KConfig{
				Port:        3000,
				ServiceName: "Orders API",
				Env:         "development",
				Docs: DocsConfig{
					Path:    "/docs",
					Title:   "Orders API",
					Version: "1.0.0",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyDefaults(tt.input)
			if got.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.expected.Port)
			}
			if got.ServiceName != tt.expected.ServiceName {
				t.Errorf("ServiceName = %v, want %v", got.ServiceName, tt.expected.ServiceName)
			}
			if got.Env != tt.expected.Env {
				t.Errorf("Env = %v, want %v", got.Env, tt.expected.Env)
			}
			if got.Docs.Path != tt.expected.Docs.Path {
				t.Errorf("Docs.Path = %v, want %v", got.Docs.Path, tt.expected.Docs.Path)
			}
			if got.Docs.Title != tt.expected.Docs.Title {
				t.Errorf("Docs.Title = %v, want %v", got.Docs.Title, tt.expected.Docs.Title)
			}
			if got.Docs.Version != tt.expected.Docs.Version {
				t.Errorf("Docs.Version = %v, want %v", got.Docs.Version, tt.expected.Docs.Version)
			}
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name string
		cfg  KConfig
		want bool
	}{
		{
			name: "production env",
			cfg:  KConfig{Env: "production"},
			want: true,
		},
		{
			name: "development env",
			cfg:  KConfig{Env: "development"},
			want: false,
		},
		{
			name: "empty env",
			cfg:  KConfig{Env: ""},
			want: false,
		},
		{
			name: "staging env",
			cfg:  KConfig{Env: "staging"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.isProduction(); got != tt.want {
				t.Errorf("isProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocsEnabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  KConfig
		want bool
	}{
		{
			name: "disabled in production",
			cfg:  KConfig{Env: "production"},
			want: false,
		},
		{
			name: "enabled in development",
			cfg:  KConfig{Env: "development"},
			want: true,
		},
		{
			name: "enabled in staging",
			cfg:  KConfig{Env: "staging"},
			want: true,
		},
		{
			name: "enabled when env is empty",
			cfg:  KConfig{Env: ""},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.docsEnabled(); got != tt.want {
				t.Errorf("docsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
