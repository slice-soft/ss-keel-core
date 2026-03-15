package contracts

import (
	"context"
	"testing"
)

type testApp struct {
	registered bool
}

type testRoute struct {
	path string
}

type moduleMock struct{}

func (moduleMock) Register(app *testApp) {
	app.registered = true
}

type controllerMock struct {
	routes []testRoute
}

func (c controllerMock) Routes() []testRoute {
	return c.routes
}

type healthCheckerMock struct {
	name string
	err  error
}

func (h healthCheckerMock) Name() string                  { return h.name }
func (h healthCheckerMock) Check(_ context.Context) error { return h.err }

type pageQueryMock struct {
	Page  int
	Limit int
}

type entityMock struct {
	ID string
}

type pageMock struct {
	Data []entityMock
}

type repositoryMock struct{}

func (repositoryMock) FindByID(_ context.Context, id string) (*entityMock, error) {
	return &entityMock{ID: id}, nil
}

func (repositoryMock) FindAll(_ context.Context, _ pageQueryMock) (pageMock, error) {
	return pageMock{Data: []entityMock{{ID: "1"}}}, nil
}

func (repositoryMock) Create(_ context.Context, _ *entityMock) error {
	return nil
}

func (repositoryMock) Update(_ context.Context, _ string, _ *entityMock) error {
	return nil
}

func (repositoryMock) Patch(_ context.Context, _ string, _ *entityMock) error {
	return nil
}

func (repositoryMock) Delete(_ context.Context, _ string) error {
	return nil
}

var (
	_ Module[*testApp]                                        = moduleMock{}
	_ Controller[testRoute]                                   = controllerMock{}
	_ Controller[testRoute]                                   = ControllerFunc[testRoute](func() []testRoute { return nil })
	_ HealthChecker                                           = healthCheckerMock{}
	_ Repository[entityMock, string, pageQueryMock, pageMock] = repositoryMock{}
)

func TestModuleRegister(t *testing.T) {
	app := &testApp{}
	var m Module[*testApp] = moduleMock{}

	m.Register(app)

	if !app.registered {
		t.Fatal("module should register into the host app")
	}
}

func TestControllerFuncRoutes(t *testing.T) {
	routes := []testRoute{{path: "/users"}}
	controller := ControllerFunc[testRoute](func() []testRoute {
		return routes
	})

	got := controller.Routes()
	if len(got) != 1 || got[0].path != "/users" {
		t.Fatalf("unexpected routes: %+v", got)
	}
}

func TestRepositoryContractCallable(t *testing.T) {
	ctx := context.Background()
	repo := repositoryMock{}

	entity, err := repo.FindByID(ctx, "42")
	if err != nil {
		t.Fatal(err)
	}
	if entity == nil || entity.ID != "42" {
		t.Fatalf("unexpected entity: %+v", entity)
	}

	page, err := repo.FindAll(ctx, pageQueryMock{Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Data) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
}
