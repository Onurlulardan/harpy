package systemresources

import (
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func commandScopeRequest(
	admission Admission,
	operation string,
) (platformtx.ScopeRequest, error) {
	actorID, tenantID, groupGeneration, err := admission.Identity()
	if err != nil || operation == "" {
		return platformtx.ScopeRequest{}, ErrContract
	}
	generations, err := admission.Generations()
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	request, err := platformtx.AdmitCommandScope(
		uuid.UUID(actorID), uuid.UUID(tenantID), operation, AuthorityGroup,
		groupGeneration, generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	return request, nil
}
