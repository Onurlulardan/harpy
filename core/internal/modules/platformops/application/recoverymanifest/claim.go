package recoverymanifest

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	RecoveryManifestClaim(context.Context, platformopsapi.RecoveryArtifactClaim) (platformopsapi.RecoveryArtifact, error)
}

type RecoveryManifestOperations struct{ backend Backend }

func NewRecoveryManifestOperations(backend Backend) *RecoveryManifestOperations {
	return &RecoveryManifestOperations{backend: backend}
}

func (operations *RecoveryManifestOperations) ClaimRecoveryManifest(
	ctx context.Context, claim platformopsapi.RecoveryArtifactClaim,
) (platformopsapi.RecoveryArtifact, error) {
	return operations.backend.RecoveryManifestClaim(ctx, claim)
}
