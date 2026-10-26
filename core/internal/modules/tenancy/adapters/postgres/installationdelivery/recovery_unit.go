package installationdelivery

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type recoveryBoundary struct {
	transaction *recoveryTransaction
	commit      func(context.Context) error
	rollback    func() error
}

type recoveryUnit struct {
	begin func(context.Context) (recoveryBoundary, error)
}

func (unit *recoveryUnit) Run(
	ctx context.Context,
	callback func(context.Context, onboarding.RecoveryTransaction) error,
) error {
	if unit == nil || unit.begin == nil || ctx == nil || ctx.Err() != nil || callback == nil {
		return errRecoveryUnavailable
	}
	boundary, err := unit.begin(ctx)
	if err != nil {
		return err
	}
	finalized := atomic.Bool{}
	defer func() {
		boundary.transaction.close()
		if !finalized.Load() {
			_ = boundary.rollback()
		}
	}()
	if err = callback(ctx, boundary.transaction); err != nil {
		rollbackErr := boundary.rollback()
		finalized.Store(true)
		return errors.Join(err, rollbackErr)
	}
	boundary.transaction.close()
	finalized.Store(true)
	if err = boundary.commit(ctx); err != nil {
		return fmt.Errorf("commit installation recovery: %w", err)
	}
	return nil
}

var _ onboarding.RecoveryUnitOfWork = (*recoveryUnit)(nil)
