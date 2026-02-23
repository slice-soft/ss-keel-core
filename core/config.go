package core

type KConfig struct {
	Port        int
	ServiceName string
	Env         string
	Docs        DocsConfig
}

type DocsConfig struct {
	Path    string
	Title   string
	Version string
}

func applyDefaults(cfg KConfig) KConfig {
	if cfg.Port == 0 {
		cfg.Port = 3000
	}
	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "Keel App"
	}
	if cfg.Docs.Path == "" {
		cfg.Docs.Path = "/docs"
	}
	if cfg.Docs.Title == "" {
		cfg.Docs.Title = cfg.ServiceName
	}
	if cfg.Docs.Version == "" {
		cfg.Docs.Version = "1.0.0"
	}
	return cfg
}

func (c KConfig) isProduction() bool { return c.Env == "production" }
func (c KConfig) docsEnabled() bool  { return !c.isProduction() }
