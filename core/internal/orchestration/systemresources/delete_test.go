package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestDeleteMutatesIAMAndResolvesLayoutInOneCommandScope(t *testing.T) {
	fixture := newTestApplication(t)
	input, _ := publicresource.NewDeleteInput(
		publicresource.Users, mustID(t, testRecordID), 1, "delete-key",
	)
	result, err := fixture.application.Delete(
		context.Background(), fixture.fixture.admission, input,
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("delete error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
