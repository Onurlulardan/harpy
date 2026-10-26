package runtime

import (
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestIAMSystemResourcesSnapshotScopeRequiresPairedGeneration(t *testing.T) {
	actorID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	tenantID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	request, err := IAMSystemResourcesSnapshotScopeRequest(
		actorID, tenantID,
		map[string]int64{"iam": 7, "system_resource_layout": 7},
	)
	if err != nil || request.Mode != platformtx.SnapshotQuery {
		t.Fatalf("request=%#v err=%v", request, err)
	}
	if _, err = IAMSystemResourcesSnapshotScopeRequest(
		actorID, tenantID,
		map[string]int64{"iam": 7, "system_resource_layout": 8},
	); err == nil {
		t.Fatal("generation drift was accepted")
	}
}
