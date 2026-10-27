package systemhealth

import (
	"context"
	"testing"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ called bool }

func (backend *backendFixture) SystemHealthRead(
	context.Context,
) (platformopsapi.SystemHealth, error) {
	backend.called = true
	return platformopsapi.SystemHealth{Status: platformopsapi.HealthHealthy}, nil
}

func TestReadForwardsSystemHealthRequest(t *testing.T) {
	backend := &backendFixture{}
	result, err := NewSystemHealthOperations(backend).ReadSystemHealth(t.Context())
	if err != nil || !backend.called || result.Status != platformopsapi.HealthHealthy {
		t.Fatalf("ReadSystemHealth() = %#v, %v", result, err)
	}
}
