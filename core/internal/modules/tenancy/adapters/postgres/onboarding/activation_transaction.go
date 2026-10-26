package onboarding

import (
	"context"
	"sync"
	"sync/atomic"

	onboardingapp "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/onboarding"
)

type activationTransaction struct {
	store      *postgresStore
	bind       func(onboardingapp.ActivationOwner) (onboardingapp.ActivationOwnerExecutor, error)
	bindSchema func(context.Context, string) error
	open       atomic.Bool
	mutex      sync.Mutex
	schema     string
}

func (transaction *activationTransaction) usable() bool {
	return transaction != nil && transaction.store != nil && transaction.bind != nil &&
		transaction.bindSchema != nil && transaction.open.Load()
}

func (transaction *activationTransaction) close() {
	if transaction != nil {
		transaction.open.Store(false)
	}
}

var _ onboardingapp.ActivationTransaction = (*activationTransaction)(nil)
