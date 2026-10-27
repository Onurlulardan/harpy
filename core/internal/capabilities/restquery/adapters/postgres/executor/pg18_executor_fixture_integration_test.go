//go:build integration

package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func pg18NewExecutor(
	t *testing.T,
	codec cursor.Codec,
	policy planner.Policy,
	ports ...*pg18OwnerPort,
) Executor {
	t.Helper()
	registrations := make([]execution.Registration, len(ports))
	for index, port := range ports {
		registration, err := execution.NewRegistration(port.owner, port)
		if err != nil {
			t.Fatal(err)
		}
		registrations[index] = registration
	}
	registry, err := execution.NewRegistry(registrations)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := New(registry, codec, policy)
	if err != nil {
		t.Fatal(err)
	}
	return executor
}
