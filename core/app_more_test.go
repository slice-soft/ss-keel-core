package core

import (
	"context"
	"testing"

	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core/httpx"
)

type schedulerSpy struct {
	started bool
	stopped bool
}

func (s *schedulerSpy) Add(_ contracts.Job) error { return nil }
func (s *schedulerSpy) Start()                    { s.started = true }
func (s *schedulerSpy) Stop(_ context.Context)    { s.stopped = true }

type moduleSpy struct {
	registered bool
}

func (m *moduleSpy) Register(_ *App) {
	m.registered = true
}

func TestLoggerGetter(t *testing.T) {
	app := New(KConfig{DisableHealth: true})
	if app.Logger() == nil {
		t.Fatal("Logger() returned nil")
	}
}

func TestRegisterSchedulerAddsShutdownHook(t *testing.T) {
	app := New(KConfig{DisableHealth: true})
	s := &schedulerSpy{}

	app.RegisterScheduler(s)
	if app.scheduler == nil {
		t.Fatal("scheduler should be set")
	}
	if len(app.shutdownHooks) != 1 {
		t.Fatalf("shutdownHooks len = %d, want 1", len(app.shutdownHooks))
	}

	if err := app.shutdownHooks[0](context.Background()); err != nil {
		t.Fatal(err)
	}
	if !s.stopped {
		t.Fatal("scheduler Stop() should be called by shutdown hook")
	}
}

func TestGroupUseCallsModuleRegister(t *testing.T) {
	app := New(KConfig{DisableHealth: true})
	g := app.Group("/v1")

	m := &moduleSpy{}
	g.Use(m)
	if !m.registered {
		t.Fatal("group.Use() should call module Register()")
	}
}

func TestNewPageAlias(t *testing.T) {
	p := httpx.NewPage([]int{1, 2}, 5, 1, 2)
	if p.Total != 5 || p.TotalPages != 3 {
		t.Fatalf("unexpected page: %+v", p)
	}
}

func TestNoopSpanMethods(t *testing.T) {
	s := noopSpan{}
	s.SetAttribute("k", "v")
	s.RecordError(nil)
	s.End()
}
