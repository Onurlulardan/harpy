package systemresourcequery

import (
	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func SnapshotScopeRequest(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	generations map[string]int64,
) (platformtx.ScopeRequest, error) {
	return queryruntime.IAMSnapshotScopeRequest(actorID, tenantID, generations)
}

func SystemResourceSnapshotScopeRequest(
	actorID uuid.UUID,
	tenantID uuid.UUID,
	generations map[string]int64,
) (platformtx.ScopeRequest, error) {
	return queryruntime.IAMSystemResourcesSnapshotScopeRequest(
		actorID, tenantID, generations,
	)
}
