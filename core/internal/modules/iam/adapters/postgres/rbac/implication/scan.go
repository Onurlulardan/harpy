package implicationpostgres

import (
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

func scanMutation(row scanner) (rbacdomain.MutationResult, error) {
	var result rbacdomain.MutationResult
	var id uuid.UUID
	if row == nil || row.Scan(&id, &result.Revision, &result.AuthorizationClock, &result.Replayed) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	result.ID = rbacdomain.ID(id)
	if result.Validate(result.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return result, nil
}
func scanImplication(row scanner) (rbacdomain.Implication, error) {
	var result rbacdomain.Implication
	var id, sourceID, impliedID uuid.UUID
	if row == nil || row.Scan(&id, &sourceID, &impliedID, &result.RetirementContractKey, &result.State, &result.HitCount, &result.Revision) != nil {
		return rbacdomain.Implication{}, rbacdomain.ErrStorage
	}
	result.ID, result.SourcePermissionID, result.ImpliedPermissionID = rbacdomain.ID(id), rbacdomain.ID(sourceID), rbacdomain.ID(impliedID)
	return result, nil
}
