package systemtransfer

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func workerEvidence(scope tenantScope,
	evidence publicresource.ActorEvidence,
	attempt systemtransferapp.Attempt) (int64, int64, error) {
	tenantID, tenantErr := evidence.TenantID()
	actorID, actorErr := evidence.ActorID()
	token, authorization, versionErr := evidence.Versions()
	if tenantErr != nil || actorErr != nil || versionErr != nil ||
		uuid.UUID(tenantID) != attempt.TenantID ||
		uuid.UUID(actorID) != attempt.ActorID ||
		scope.TenantID() != attempt.TenantID || scope.ActorID() != attempt.ActorID ||
		authorization != attempt.AuthorizationVersion {
		return 0, 0, systemtransferapp.ErrAuthorizationCheckpoint
	}
	return token, authorization, nil
}
