package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) UpdateLease(
	ctx context.Context,
	identity ports.RecoveryExecutionIdentity,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil || !validIdentity(identity) {
		return ErrInvalid
	}
	changed, err := service.store.UpdateRecoveryLease(ctx, identity)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
