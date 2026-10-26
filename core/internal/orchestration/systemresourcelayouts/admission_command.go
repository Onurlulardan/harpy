package systemresourcelayouts

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func commandScopeRequest(
	admission publicresource.Admission,
	operation string,
) (platformtx.ScopeRequest, error) {
	actorID, tenantID, generation, err := admission.Identity()
	if err != nil || operation == "" {
		return platformtx.ScopeRequest{}, ErrContract
	}
	generations, err := admission.Generations()
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	request, err := platformtx.AdmitCommandScope(
		uuid.UUID(actorID), uuid.UUID(tenantID), operation,
		publicresource.AuthorityGroup, generation, generations,
	)
	if err != nil {
		return platformtx.ScopeRequest{}, ErrContract
	}
	return request, nil
}
