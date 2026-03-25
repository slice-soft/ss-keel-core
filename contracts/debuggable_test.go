package contracts

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/a-h/templ"
)

// --- mocks ---

type debuggableMock struct {
	id     string
	label  string
	events chan PanelEvent
}

func (d *debuggableMock) PanelID() string              { return d.id }
func (d *debuggableMock) PanelLabel() string           { return d.label }
func (d *debuggableMock) PanelEvents() <-chan PanelEvent { return d.events }

type debuggableWithViewMock struct {
	debuggableMock
}

func (d *debuggableWithViewMock) PanelView() templ.Component {
	return templ.ComponentFunc(func(_ context.Context, _ io.Writer) error { return nil })
}

type panelRegistryMock struct {
	addons []Debuggable
}

func (p *panelRegistryMock) RegisterAddon(d Debuggable) {
	p.addons = append(p.addons, d)
}

type manifestableMock struct{}

func (m manifestableMock) Manifest() AddonManifest {
	return AddonManifest{
		ID:           "mock",
		Version:      "1.0.0",
		Capabilities: []string{"cache"},
		Resources:    []string{"redis"},
		EnvVars: []EnvVar{
			{Key: "REDIS_URL", Required: true, Secret: false, Source: "mock"},
		},
	}
}

type addonMock struct{}

func (a addonMock) ID() string { return "mock" }

// --- compile-time assertions ---

var (
	_ Debuggable         = (*debuggableMock)(nil)
	_ DebuggableWithView = (*debuggableWithViewMock)(nil)
	_ PanelRegistry      = (*panelRegistryMock)(nil)
	_ Manifestable       = manifestableMock{}
	_ Addon              = addonMock{}
)

// --- tests ---

func TestPanelEventFields(t *testing.T) {
	now := time.Now()
	e := PanelEvent{
		Timestamp: now,
		AddonID:   "gorm",
		Category:  CategoryQuery,
		Label:     "SELECT",
		Detail:    map[string]any{"rows": 10},
		Level:     "info",
	}
	if e.AddonID != "gorm" {
		t.Fatalf("AddonID = %q, want %q", e.AddonID, "gorm")
	}
	if e.Category != CategoryQuery {
		t.Fatalf("Category = %q, want %q", e.Category, CategoryQuery)
	}
	if e.Label != "SELECT" {
		t.Fatalf("Label = %q, want %q", e.Label, "SELECT")
	}
	if e.Level != "info" {
		t.Fatalf("Level = %q, want %q", e.Level, "info")
	}
	if e.Detail["rows"] != 10 {
		t.Fatalf("Detail[rows] = %v", e.Detail["rows"])
	}
	if e.Timestamp != now {
		t.Fatalf("Timestamp not preserved")
	}
}

func TestPanelCategories(t *testing.T) {
	cases := []struct{ got, want string }{
		{CategoryQuery, "query"},
		{CategoryAuth, "auth"},
		{CategoryCache, "cache"},
		{CategoryRequest, "request"},
		{CategoryGeneric, "generic"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("category = %q, want %q", c.got, c.want)
		}
	}
}

func TestDebuggableContract(t *testing.T) {
	ch := make(chan PanelEvent, 1)
	d := &debuggableMock{id: "test-addon", label: "Test Addon", events: ch}
	if d.PanelID() != "test-addon" {
		t.Fatalf("PanelID = %q", d.PanelID())
	}
	if d.PanelLabel() != "Test Addon" {
		t.Fatalf("PanelLabel = %q", d.PanelLabel())
	}
	if d.PanelEvents() == nil {
		t.Fatal("PanelEvents() returned nil channel")
	}
}

func TestDebuggableWithViewContract(t *testing.T) {
	ch := make(chan PanelEvent, 1)
	d := &debuggableWithViewMock{debuggableMock{id: "view-addon", label: "View Addon", events: ch}}
	comp := d.PanelView()
	if comp == nil {
		t.Fatal("PanelView() returned nil")
	}
}

func TestPanelRegistryContract(t *testing.T) {
	reg := &panelRegistryMock{}
	ch := make(chan PanelEvent, 1)
	d := &debuggableMock{id: "a", label: "A", events: ch}
	reg.RegisterAddon(d)
	if len(reg.addons) != 1 {
		t.Fatalf("expected 1 addon, got %d", len(reg.addons))
	}
	if reg.addons[0].PanelID() != "a" {
		t.Fatalf("registered addon ID = %q", reg.addons[0].PanelID())
	}
}

func TestManifestableContract(t *testing.T) {
	m := manifestableMock{}
	manifest := m.Manifest()
	if manifest.ID != "mock" {
		t.Fatalf("ID = %q", manifest.ID)
	}
	if len(manifest.Capabilities) != 1 || manifest.Capabilities[0] != "cache" {
		t.Fatalf("Capabilities = %v", manifest.Capabilities)
	}
	if len(manifest.EnvVars) != 1 || manifest.EnvVars[0].Key != "REDIS_URL" {
		t.Fatalf("EnvVars = %v", manifest.EnvVars)
	}
}

func TestAddonContract(t *testing.T) {
	a := addonMock{}
	if a.ID() != "mock" {
		t.Fatalf("ID = %q, want %q", a.ID(), "mock")
	}
}

func TestEnvVarFields(t *testing.T) {
	ev := EnvVar{
		Key:         "DB_DSN",
		Description: "Database connection string",
		Required:    true,
		Secret:      true,
		Default:     "localhost:5432",
		Source:      "gorm",
	}
	if ev.Key != "DB_DSN" {
		t.Fatalf("Key = %q", ev.Key)
	}
	if ev.Description != "Database connection string" {
		t.Fatalf("Description = %q", ev.Description)
	}
	if !ev.Required {
		t.Fatal("Required should be true")
	}
	if !ev.Secret {
		t.Fatal("Secret should be true")
	}
	if ev.Default != "localhost:5432" {
		t.Fatalf("Default = %q", ev.Default)
	}
	if ev.Source != "gorm" {
		t.Fatalf("Source = %q, want %q", ev.Source, "gorm")
	}
}

func TestAddonManifestFields(t *testing.T) {
	m := AddonManifest{
		ID:           "gorm",
		Version:      "2.0.0",
		Capabilities: []string{"database"},
		Resources:    []string{"postgres"},
		EnvVars:      []EnvVar{{Key: "DB_DSN", Required: true, Secret: true, Source: "gorm"}},
	}
	if m.ID != "gorm" {
		t.Fatalf("ID = %q", m.ID)
	}
	if m.Version != "2.0.0" {
		t.Fatalf("Version = %q", m.Version)
	}
	if m.Capabilities[0] != "database" {
		t.Fatalf("Capabilities[0] = %q", m.Capabilities[0])
	}
	if m.Resources[0] != "postgres" {
		t.Fatalf("Resources[0] = %q", m.Resources[0])
	}
	if m.EnvVars[0].Key != "DB_DSN" {
		t.Fatalf("EnvVars[0].Key = %q", m.EnvVars[0].Key)
	}
}
