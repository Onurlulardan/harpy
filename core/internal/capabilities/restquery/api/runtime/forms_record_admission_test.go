package runtime

import (
	"testing"

	"github.com/google/uuid"
)

func TestFormsRecordGetSnapshotScopeRequestRequiresExactEvidence(t *testing.T) {
	actorID, tenantID := uuid.New(), uuid.New()
	request, err := FormsRecordGetSnapshotScopeRequest(actorID, tenantID, 7)
	if err != nil || request.Mode != "snapshot_query" {
		t.Fatalf("request=%+v error=%v", request, err)
	}
	for _, invalid := range []struct {
		actor, tenant uuid.UUID
		generation    int64
	}{
		{uuid.Nil, tenantID, 7},
		{actorID, uuid.Nil, 7},
		{actorID, tenantID, 0},
	} {
		if _, err := FormsRecordGetSnapshotScopeRequest(
			invalid.actor, invalid.tenant, invalid.generation,
		); err == nil {
			t.Fatal("invalid Forms record snapshot evidence was accepted")
		}
	}
}
