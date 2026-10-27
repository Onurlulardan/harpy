package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) RejectFailure(
	ctx context.Context,
	failure ports.RecoveryExecutionFailure,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validIdentity(failure.RecoveryExecutionIdentity) ||
		!codePattern.MatchString(failure.Code) {
		return ErrInvalid
	}
	changed, err := service.store.RejectRecoveryExecution(ctx, failure)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
