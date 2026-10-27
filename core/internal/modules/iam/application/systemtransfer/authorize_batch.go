package systemtransfer

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func (worker *Worker) AuthorizeBatch(
	ctx context.Context,
	scope BatchScope,
	attempt Attempt,
) (publicresource.ActorEvidence, error) {
	if scope.TenantID() != attempt.TenantID || scope.ActorID() != attempt.ActorID {
		return publicresource.ActorEvidence{}, ErrAuthorizationCheckpoint
	}
	evidence, err := worker.authorizer.AuthorizeImportBatch(
		ctx, scope.IAMScope(), attempt,
	)
	if errors.Is(err, ErrImportCatalogPinStale) ||
		errors.Is(err, ErrWorkerUnavailable) {
		return publicresource.ActorEvidence{}, err
	}
	if err != nil || evidence.Validate() != nil {
		return publicresource.ActorEvidence{}, ErrAuthorizationCheckpoint
	}
	tenantID, tenantErr := evidence.TenantID()
	actorID, actorErr := evidence.ActorID()
	_, authorizationVersion, versionErr := evidence.Versions()
	if tenantErr != nil || actorErr != nil || versionErr != nil ||
		tenantID != publicresource.ID(attempt.TenantID) ||
		actorID != publicresource.ID(attempt.ActorID) ||
		authorizationVersion != attempt.AuthorizationVersion {
		return publicresource.ActorEvidence{}, ErrAuthorizationCheckpoint
	}
	return evidence, nil
}
