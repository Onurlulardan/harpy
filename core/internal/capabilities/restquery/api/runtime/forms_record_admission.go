package runtime

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const formsRecordGetOperation = "forms.record_kernel.get"

func FormsRecordGetSnapshotScopeRequest(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	formsGeneration int64,
) (platformtx.ScopeRequest, error) {
	actor, err := actorcontext.Issue(actorID, tenantID)
	if err != nil {
		return platformtx.ScopeRequest{}, fmt.Errorf("Forms record actor admission failed")
	}
	operation, err := platformtx.NewSnapshotAuthorityIssuer().Issue(
		formsRecordGetOperation, map[string]int64{"forms": formsGeneration},
	)
	if err != nil {
		return platformtx.ScopeRequest{}, fmt.Errorf("Forms record snapshot admission failed")
	}
	return platformtx.ScopeRequest{
		Actor: actor, Operation: operation, Mode: platformtx.SnapshotQuery,
	}, nil
}
