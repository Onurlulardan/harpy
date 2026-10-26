package systemresources

import (
	"context"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestCreateKeepsIAMMutationWhenLayoutIsUnavailable(t *testing.T) {
	fixture := newTestApplication(t)
	fixture.layout.err = layoutapi.ErrUnavailable
	prepared := &testPrepared{resource: publicresource.Users}
	result, err := fixture.application.Create(
		context.Background(), fixture.fixture.admission, prepared,
	)
	if err != nil {
		t.Fatalf("create error=%v", err)
	}
	if _, err := result.IAM(); err != nil {
		t.Fatalf("iam mutation=%v", err)
	}
	if _, err := result.Presentation(); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("presentation error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}

func TestCreateReturnsExactPresentedMutation(t *testing.T) {
	fixture := newTestApplication(t)
	result, err := fixture.application.Create(
		context.Background(), fixture.fixture.admission,
		&testPrepared{resource: publicresource.Users},
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("create error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
