package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type systemTransferBatchAuthorizer struct {
	runtime *WorkerAuthorizationRuntime
}

func newSystemTransferBatchAuthorizer(
	runtime *WorkerAuthorizationRuntime,
) (*systemTransferBatchAuthorizer, error) {
	if runtime == nil {
		return nil, systemtransfer.ErrDependencies
	}
	return &systemTransferBatchAuthorizer{runtime: runtime}, nil
}

func (authorizer *systemTransferBatchAuthorizer) AuthorizeImportBatch(
	ctx context.Context,
	scope platformtx.TenantScope,
	attempt systemtransfer.Attempt,
) (publicresource.ActorEvidence, error) {
	if authorizer == nil || authorizer.runtime == nil || attempt.Validate() != nil ||
		scope.Mode() != platformtx.Command ||
		scope.OperationID() != SystemImportExecuteOperation {
		return publicresource.ActorEvidence{}, systemtransfer.ErrAuthorizationCheckpoint
	}
	allowed, _, tokenVersion, currentVersion, _, err :=
		authorizer.runtime.AuthorizePinned(ctx, scope,
			"iam."+string(attempt.Resource), "import", attempt.AuthorizationVersion)
	if err != nil || !allowed || tokenVersion < 1 ||
		currentVersion != attempt.AuthorizationVersion {
		return publicresource.ActorEvidence{}, systemtransfer.ErrAuthorizationCheckpoint
	}
	if err = verifySystemTransferImportCatalogPin(ctx, scope, attempt); err != nil {
		return publicresource.ActorEvidence{}, err
	}
	evidence, err := publicresource.NewActorEvidence(
		publicresource.ID(scope.TenantID()), publicresource.ID(scope.ActorID()),
		tokenVersion, currentVersion,
	)
	if err != nil {
		return publicresource.ActorEvidence{}, systemtransfer.ErrAuthorizationCheckpoint
	}
	return evidence, nil
}

var _ systemtransfer.BatchAuthorizer = (*systemTransferBatchAuthorizer)(nil)
