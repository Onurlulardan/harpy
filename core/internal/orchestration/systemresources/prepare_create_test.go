package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestPrepareCreateCompletesPasswordWorkBeforeTenantScope(t *testing.T) {
	fixture := newTestApplication(t)
	input, err := publicresource.NewCreateInput(
		publicresource.Users, fixture.fixture.values, []byte("strong-password"), "create-key",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Destroy()
	prepared, err := fixture.application.PrepareCreate(context.Background(), input)
	if err != nil || prepared == nil || fixture.iam.preparedCalls != 1 ||
		fixture.coordinator.calls != 0 {
		t.Fatalf(
			"prepared=%v error=%v calls=%d scopes=%d",
			prepared != nil, err, fixture.iam.preparedCalls, fixture.coordinator.calls,
		)
	}
	prepared.Destroy()
}
