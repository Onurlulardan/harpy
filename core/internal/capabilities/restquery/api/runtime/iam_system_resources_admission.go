package runtime

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorcontext"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func IAMSystemResourcesSnapshotScopeRequest(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	generations map[string]int64,
) (platformtx.ScopeRequest, error) {
	actor, err := actorcontext.Issue(actorID, tenantID)
	if err != nil {
		return platformtx.ScopeRequest{}, fmt.Errorf("REST Query actor admission failed")
	}
	operation, err := authorization.NewSnapshotOperationAdmission().
		IssueIAMSystemResources(generations)
	if err != nil {
		return platformtx.ScopeRequest{},
			fmt.Errorf("REST Query IAM system-resource admission failed")
	}
	return platformtx.ScopeRequest{
		Actor: actor, Operation: operation, Mode: platformtx.SnapshotQuery,
	}, nil
}
