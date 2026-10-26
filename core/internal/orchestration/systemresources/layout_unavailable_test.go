package systemresources

import (
	"context"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestQueryReturnsIAMRecordsWhenLayoutIsUnavailable(t *testing.T) {
	fixture := newTestApplication(t)
	fixture.layout.err = layoutapi.ErrUnavailable
	result, err := fixture.application.Query(
		context.Background(), fixture.fixture.admission, testQueryInput(t),
	)
	if err != nil {
		t.Fatalf("query error=%v", err)
	}
	if _, err := result.IAM(); err != nil {
		t.Fatalf("iam query=%v", err)
	}
	if _, err := result.Presentation(); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("presentation error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}

func TestGetReturnsIAMRecordWhenLayoutIsUnavailable(t *testing.T) {
	fixture := newTestApplication(t)
	fixture.layout.err = layoutapi.ErrUnavailable
	input, _ := publicresource.NewGetInput(publicresource.Users, mustID(t, testRecordID))
	result, err := fixture.application.Get(
		context.Background(), fixture.fixture.admission, input,
	)
	if err != nil {
		t.Fatalf("get error=%v", err)
	}
	if _, err := result.IAM(); err != nil {
		t.Fatalf("iam read=%v", err)
	}
	if _, err := result.Presentation(); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("presentation error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
