package environmentoperator

import (
	"context"
	"io"
	"strings"
	"time"
)

const (
	maximumMigrationWait = 5 * time.Minute
	maximumLease         = 5 * time.Minute
)

type Coordinator struct {
	policy     DeploymentPolicy
	onboarding Onboarding
	migration  TenantMigration
	offering   Offering
	authority  ProtectedAuthority
	finalizer  Finalizer
	entropy    io.Reader
	sleep      func(context.Context, time.Duration) error
	leaseOwner string
	poll       time.Duration
}

func New(dependencies Dependencies) (*Coordinator, error) {
	if dependencies.Policy == nil || dependencies.Onboarding == nil ||
		dependencies.Migration == nil || dependencies.Offering == nil ||
		dependencies.Finalizer == nil || dependencies.Entropy == nil ||
		dependencies.Sleep == nil || len(dependencies.LeaseOwner) < 1 ||
		len(dependencies.LeaseOwner) > 128 ||
		strings.TrimSpace(dependencies.LeaseOwner) != dependencies.LeaseOwner ||
		dependencies.PollInterval < time.Millisecond ||
		dependencies.PollInterval > time.Second {
		return nil, ErrInvalid
	}
	return &Coordinator{policy: dependencies.Policy,
		onboarding: dependencies.Onboarding, migration: dependencies.Migration,
		offering: dependencies.Offering, authority: dependencies.Authority,
		finalizer: dependencies.Finalizer, entropy: dependencies.Entropy,
		sleep: dependencies.Sleep, leaseOwner: dependencies.LeaseOwner,
		poll: dependencies.PollInterval}, nil
}
