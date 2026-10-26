package management

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/health"
)

type handlerFixture struct {
	handler *Handler
	monitor *health.Monitor
	errors  []error
}

func newHandlerFixture(t *testing.T) handlerFixture {
	t.Helper()
	now := func() time.Time { return time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC) }
	monitor, err := health.NewMonitor(runtimev1.ServiceCoreAPI, 30*time.Second, now)
	if err != nil {
		t.Fatal(err)
	}
	info, err := buildinfo.Read(runtimev1.ServiceCoreAPI, runtimev1.EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	fixture := handlerFixture{monitor: monitor}
	handler, err := NewHandler(monitor, info, func(err error) {
		fixture.errors = append(fixture.errors, err)
	})
	if err != nil {
		t.Fatal(err)
	}
	fixture.handler = handler
	return fixture
}
