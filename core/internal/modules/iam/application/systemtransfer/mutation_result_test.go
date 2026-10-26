package systemtransfer

import (
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestMutationResultIdentityUsesOriginalReceiptRevisionOnReplay(t *testing.T) {
	result, err := fixtureMutationResultVersions(publicresource.Roles,
		publicresource.ID(fixtureRecordID), 6, 4, true)
	if err != nil {
		t.Fatal(err)
	}
	id, revision, err := mutationResultIdentity(result)
	if err != nil || id != fixtureRecordID || revision != 4 {
		t.Fatalf("id=%s revision=%d err=%v", id, revision, err)
	}
}
