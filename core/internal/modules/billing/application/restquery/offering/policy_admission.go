package offering

import (
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func policyAdmission(
	scope transaction.GlobalScope,
	request restauthorization.GlobalPolicyRequest,
) (bool, string, bool, error) {
	environmentID, environmentErr := request.EnvironmentID()
	actorID, hasActor, actorErr := request.ActorID()
	operation, operationErr := request.OperationID()
	public, publicErr := request.Public()
	owner, ownerErr := request.Owner()
	if environmentErr != nil || actorErr != nil || operationErr != nil ||
		publicErr != nil || ownerErr != nil || environmentID != scope.EnvironmentID() ||
		owner != Owner {
		return false, "", false, ErrContract
	}
	if public {
		if hasActor || actorID != uuid.Nil || scope.ActorID() != uuid.Nil {
			return false, "", false, ErrContract
		}
		return true, operation,
			operation == string(transaction.PublicBillingCatalogList), nil
	}
	if !hasActor || actorID == uuid.Nil || actorID != scope.ActorID() {
		return false, "", false, ErrContract
	}
	return false, operation, operation == platformListOperation, nil
}
