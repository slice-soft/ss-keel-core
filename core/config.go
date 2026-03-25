package core

type KConfig struct {
	DisableHealth bool
	Port          int    `keel:"server.port,required"`
	ServiceName   string `keel:"app.name,required"`
	Env           string `keel:"app.env,required"`
	Docs          DocsConfig
}

type DocsConfig struct {
	Path        string `keel:"docs.path,required"`
	Title       string `keel:"docs.title,required"`
	Version     string `keel:"docs.version,required"`
	Description string `keel:"docs.description"`
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

// applyDefaults fills in default values for any missing configuration options.
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

// isProduction returns true if the environment is production.
func (c KConfig) isProduction() bool { return c.Env == "production" }

// docsEnabled returns true if API documentation should be generated.
func (c KConfig) docsEnabled() bool { return !c.isProduction() }
