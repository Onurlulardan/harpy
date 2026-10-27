package systemresources

import (
	restqueryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func snapshotScopeRequest(
	admission Admission,
) (platformtx.ScopeRequest, error) {
	actorID, tenantID, _, err := admission.Identity()
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	generations, err := admission.Generations()
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	request, err := restqueryruntime.IAMSystemResourcesSnapshotScopeRequest(
		uuid.UUID(actorID), uuid.UUID(tenantID), generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	return request, nil
}
