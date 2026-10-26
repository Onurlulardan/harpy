package installationdelivery

import (
	"context"
	"sync"
	"sync/atomic"

	onboarding "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type recoveryTransaction struct {
	store      *store
	bindSchema func(context.Context, string) error
	bindIAM    func() (onboarding.RecoveryOwnerExecutor, error)
	open       atomic.Bool
	mutex      sync.Mutex
	schema     string
}

func (transaction *recoveryTransaction) usable() bool {
	return transaction != nil && transaction.store != nil && transaction.bindSchema != nil &&
		transaction.bindIAM != nil && transaction.open.Load()
}

func (transaction *recoveryTransaction) close() {
	if transaction != nil {
		transaction.open.Store(false)
	}
}

func (transaction *recoveryTransaction) BindTenantSchema(
	ctx context.Context, schema string,
) error {
	if !transaction.usable() || ctx == nil || ctx.Err() != nil {
		return errRecoveryUnavailable
	}
	transaction.mutex.Lock()
	defer transaction.mutex.Unlock()
	if transaction.schema != "" {
		if transaction.schema == schema {
			return nil
		}
		return errRecoveryConflict
	}
	if err := transaction.bindSchema(ctx, schema); err != nil {
		return err
	}
	transaction.schema = schema
	return nil
}

func (transaction *recoveryTransaction) BindIAM() (
	onboarding.RecoveryOwnerExecutor, error,
) {
	transaction.mutex.Lock()
	prepared := transaction.schema != ""
	transaction.mutex.Unlock()
	if !transaction.usable() || !prepared {
		return onboarding.RecoveryOwnerExecutor{}, errRecoveryUnavailable
	}
	return transaction.bindIAM()
}

var _ onboarding.RecoveryTransaction = (*recoveryTransaction)(nil)
