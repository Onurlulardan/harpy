package reports

import (
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func policyAdmission(
	scope transaction.GlobalScope,
	request restauthorization.GlobalPolicyRequest,
	spec resourceSpec,
) error {
	environmentID, environmentErr := request.EnvironmentID()
	actorID, hasActor, actorErr := request.ActorID()
	operation, operationErr := request.OperationID()
	public, publicErr := request.Public()
	owner, ownerErr := request.Owner()
	if environmentErr != nil || actorErr != nil || operationErr != nil ||
		publicErr != nil || ownerErr != nil || environmentID != scope.EnvironmentID() ||
		owner != Owner || public || !hasActor || actorID != scope.ActorID() ||
		actorID == [16]byte{} || operation != scope.OperationID() || spec.Operation == "" {
		return ErrContract
	}
	return nil
}
