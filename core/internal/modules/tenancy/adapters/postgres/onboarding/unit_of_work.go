package onboarding

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type activationBoundary struct {
	transaction *activationTransaction
	commit      func(context.Context) error
	rollback    func() error
}

type activationBegin func(context.Context) (activationBoundary, error)

type activationUnitOfWork struct{ begin activationBegin }

func (unit *activationUnitOfWork) Run(
	ctx context.Context,
	callback func(context.Context, onboardingapp.ActivationTransaction) error,
) error {
	if unit == nil || unit.begin == nil || ctx == nil || ctx.Err() != nil || callback == nil {
		return errUnavailable
	}
	boundary, err := unit.begin(ctx)
	if err != nil {
		return err
	}
	transaction := boundary.transaction
	if transaction == nil || boundary.commit == nil || boundary.rollback == nil {
		return errUnavailable
	}
	finalized := atomic.Bool{}
	defer func() {
		transaction.close()
		if !finalized.Load() {
			_ = boundary.rollback()
		}
	}()
	if err := callback(ctx, transaction); err != nil {
		transaction.close()
		rollbackErr := boundary.rollback()
		finalized.Store(true)
		return errors.Join(err, rollbackErr)
	}
	transaction.close()
	finalized.Store(true)
	if err := boundary.commit(ctx); err != nil {
		return fmt.Errorf("commit tenancy activation transaction: %w", err)
	}
	return nil
}

var _ onboardingapp.ActivationUnitOfWork = (*activationUnitOfWork)(nil)
