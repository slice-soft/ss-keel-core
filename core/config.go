package core

type KConfig struct {
	DisableHealth bool
	Port          int
	ServiceName   string
	Env           string
	Docs          DocsConfig
}

type DocsConfig struct {
	Path        string
	Title       string
	Version     string
	Description string
	Contact     *DocsContact
	License     *DocsLicense
	Servers     []string // format: "https://api.example.com - Description"
	Tags        []DocsTag
}

type DocsContact struct {
	Name  string
	URL   string
	Email string
}

type DocsLicense struct {
	Name string
	URL  string
}

type DocsTag struct {
	Name        string
	Description string
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
