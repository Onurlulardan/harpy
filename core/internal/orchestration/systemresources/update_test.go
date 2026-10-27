package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestUpdateMutatesIAMAndResolvesLayoutInOneCommandScope(t *testing.T) {
	fixture := newTestApplication(t)
	input, _ := publicresource.NewUpdateInput(
		publicresource.Users, mustID(t, testRecordID), 1,
		fixture.fixture.values, "update-key",
	)
	result, err := fixture.application.Update(
		context.Background(), fixture.fixture.admission, input,
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("update error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
