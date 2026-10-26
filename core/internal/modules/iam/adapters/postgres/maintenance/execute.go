package maintenance

import (
	"context"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) Execute(
	ctx context.Context,
	claim maintenanceapp.Claim,
	nextID uuid.UUID,
) (maintenanceapp.Outcome, error) {
	if ctx == nil || ctx.Err() != nil || store == nil ||
		!store.runtime.usable() || claim.Validate() != nil ||
		nextID == uuid.Nil || nextID.Version() != uuid.Version(7) {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	switch claim.Kind {
	case maintenanceapp.KindDirectoryRepair:
		return store.executeDirectoryRepair(ctx, claim, nextID)
	case maintenanceapp.KindDirectoryChecksum:
		return store.executeDirectoryChecksum(ctx, claim, nextID)
	case maintenanceapp.KindCredentialBcryptRetire:
		return store.executeCredentialBcryptRetire(ctx, claim, nextID)
	case maintenanceapp.KindIdentifierRetire:
		return store.executeIdentifierRetire(ctx, claim, nextID)
	case maintenanceapp.KindDirectoryTombstonePurge:
		return store.executeDirectoryTombstonePurge(ctx, claim, nextID)
	case maintenanceapp.KindSecurityArtifactPurge:
		return store.executeSecurityArtifactPurge(ctx, claim, nextID)
	case maintenanceapp.KindProfileBackfill:
		return store.executeProfileBackfill(ctx, claim, nextID)
	default:
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
}
