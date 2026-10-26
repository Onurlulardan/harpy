package restoreexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) StartDrain(
	ctx context.Context,
	identity ports.RestoreExecutionIdentity,
) error {
	if service == nil || service.store == nil || ctx == nil ||
		ctx.Err() != nil || !identity.Valid() {
		return ErrTransition
	}
	value, err := service.store.StartRestoreDrain(ctx, identity)
	return changed(err, value)
}
