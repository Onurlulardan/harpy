package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestGetReadsIAMAndLayoutInOneSnapshotScope(t *testing.T) {
	fixture := newTestApplication(t)
	input, _ := publicresource.NewGetInput(publicresource.Users, mustID(t, testRecordID))
	result, err := fixture.application.Get(
		context.Background(), fixture.fixture.admission, input,
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("get error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
