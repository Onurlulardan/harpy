package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestChangeStateMutatesIAMAndResolvesLayoutInOneCommandScope(t *testing.T) {
	fixture := newTestApplication(t)
	input, _ := publicresource.NewStateInput(
		publicresource.Users, mustID(t, testRecordID), 1,
		publicresource.StateInactive, "state-key",
	)
	result, err := fixture.application.ChangeState(
		context.Background(), fixture.fixture.admission, input,
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("state error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
