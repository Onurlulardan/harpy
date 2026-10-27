package recoverymanifest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	RecoveryManifestClaim(context.Context, ports.RecoveryArtifactClaim) (ports.RecoveryArtifactRecord, error)
}

type RecoveryManifestOperations struct{ backend Backend }

func NewRecoveryManifestOperations(backend Backend) *RecoveryManifestOperations {
	return &RecoveryManifestOperations{backend: backend}
}

func (operations *RecoveryManifestOperations) ClaimRecoveryManifest(
	ctx context.Context, claim ports.RecoveryArtifactClaim,
) (ports.RecoveryArtifactRecord, error) {
	return operations.backend.RecoveryManifestClaim(ctx, claim)
}
