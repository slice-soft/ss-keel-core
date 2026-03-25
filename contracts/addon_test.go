package contracts

import (
	"testing"
	"time"
)

// --- mocks ---

type addonMock struct{ id string }

func (a addonMock) ID() string { return a.id }

type debuggableMock struct {
	id     string
	label  string
	events chan PanelEvent
}

func (d *debuggableMock) PanelID() string                { return d.id }
func (d *debuggableMock) PanelLabel() string             { return d.label }
func (d *debuggableMock) PanelEvents() <-chan PanelEvent { return d.events }

type panelRegistryMock struct {
	addons []Debuggable
}

func (p *panelRegistryMock) RegisterAddon(d Debuggable) { p.addons = append(p.addons, d) }

type manifestableMock struct{ manifest AddonManifest }

func (m manifestableMock) Manifest() AddonManifest { return m.manifest }

// --- compile-time assertions ---

var (
	_ Addon         = addonMock{}
	_ Debuggable    = (*debuggableMock)(nil)
	_ PanelRegistry = (*panelRegistryMock)(nil)
	_ Manifestable  = manifestableMock{}
)

// --- tests ---

func TestAddonID(t *testing.T) {
	var a Addon = addonMock{id: "redis"}
	if a.ID() != "redis" {
		t.Fatalf("ID() = %q, want %q", a.ID(), "redis")
	}
}

func TestPanelEventFields(t *testing.T) {
	now := time.Now()
	e := PanelEvent{
		Timestamp: now,
		AddonID:   "gorm",
		Label:     "query executed",
		Detail:    map[string]any{"sql": "SELECT 1", "duration_ms": 12},
		Level:     "info",
	}

	if e.AddonID != "gorm" {
		t.Fatalf("AddonID = %q, want %q", e.AddonID, "gorm")
	}
	if e.Level != "info" {
		t.Fatalf("Level = %q, want %q", e.Level, "info")
	}
	if e.Label != "query executed" {
		t.Fatalf("Label = %q, want %q", e.Label, "query executed")
	}
	if e.Detail["sql"] != "SELECT 1" {
		t.Fatalf("Detail[sql] = %v, want %q", e.Detail["sql"], "SELECT 1")
	}
	if !e.Timestamp.Equal(now) {
		t.Fatalf("Timestamp = %v, want %v", e.Timestamp, now)
	}
}

func TestDebuggablePanelEvents(t *testing.T) {
	ch := make(chan PanelEvent, 1)
	d := &debuggableMock{id: "jwt", label: "JWT", events: ch}

	if d.PanelID() != "jwt" {
		t.Fatalf("PanelID() = %q, want %q", d.PanelID(), "jwt")
	}
	if d.PanelLabel() != "JWT" {
		t.Fatalf("PanelLabel() = %q, want %q", d.PanelLabel(), "JWT")
	}

	sent := PanelEvent{AddonID: "jwt", Label: "token issued", Level: "info"}
	ch <- sent

	got := <-d.PanelEvents()
	if got.Label != sent.Label {
		t.Fatalf("event Label = %q, want %q", got.Label, sent.Label)
	}
	if got.AddonID != sent.AddonID {
		t.Fatalf("event AddonID = %q, want %q", got.AddonID, sent.AddonID)
	}
}

func TestPanelRegistryRegisterAddon(t *testing.T) {
	registry := &panelRegistryMock{}
	a := &debuggableMock{id: "redis"}
	b := &debuggableMock{id: "gorm"}

	registry.RegisterAddon(a)
	registry.RegisterAddon(b)

	if len(registry.addons) != 2 {
		t.Fatalf("expected 2 addons, got %d", len(registry.addons))
	}
	if registry.addons[0].PanelID() != "redis" {
		t.Fatalf("first addon ID = %q, want %q", registry.addons[0].PanelID(), "redis")
	}
	if registry.addons[1].PanelID() != "gorm" {
		t.Fatalf("second addon ID = %q, want %q", registry.addons[1].PanelID(), "gorm")
	}
}

func TestManifestableManifest(t *testing.T) {
	m := manifestableMock{
		manifest: AddonManifest{
			ID:           "gorm",
			Version:      "2.0.0",
			Capabilities: []string{"database"},
			Resources:    []string{"postgres"},
			EnvVars: []EnvVar{
				{Key: "DB_DSN", ConfigKey: "database.url", Required: true, Secret: true, Source: "gorm"},
			},
		},
	}

	got := m.Manifest()

	if got.ID != "gorm" {
		t.Fatalf("ID = %q, want %q", got.ID, "gorm")
	}
	if got.Version != "2.0.0" {
		t.Fatalf("Version = %q, want %q", got.Version, "2.0.0")
	}
	if len(got.Capabilities) != 1 || got.Capabilities[0] != "database" {
		t.Fatalf("Capabilities = %v, want [database]", got.Capabilities)
	}
	if len(got.Resources) != 1 || got.Resources[0] != "postgres" {
		t.Fatalf("Resources = %v, want [postgres]", got.Resources)
	}
	if len(got.EnvVars) != 1 {
		t.Fatalf("expected 1 EnvVar, got %d", len(got.EnvVars))
	}

	ev := got.EnvVars[0]
	if ev.Key != "DB_DSN" {
		t.Fatalf("EnvVar.Key = %q, want %q", ev.Key, "DB_DSN")
	}
	if ev.ConfigKey != "database.url" {
		t.Fatalf("EnvVar.ConfigKey = %q, want %q", ev.ConfigKey, "database.url")
	}
	if !ev.Required {
		t.Fatal("EnvVar.Required should be true")
	}
	if !ev.Secret {
		t.Fatal("EnvVar.Secret should be true")
	}
	if ev.Source != "gorm" {
		t.Fatalf("EnvVar.Source = %q, want %q", ev.Source, "gorm")
	}
}
