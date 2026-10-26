package retention

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Commands struct {
	store  Store
	nextID IDSource
	clock  Clock
}

func NewCommands(store Store, nextID IDSource, clock Clock) (*Commands, error) {
	if nilStore(store) || nextID == nil || clock == nil {
		return nil, ErrConfiguration
	}
	return &Commands{store: store, nextID: nextID, clock: clock}, nil
}

func (commands *Commands) Purge(ctx context.Context, claim Claim) (Outcome, error) {
	if commands == nil || ctx == nil || ctx.Err() != nil || claim.Check() != nil ||
		nilStore(commands.store) || commands.nextID == nil || commands.clock == nil {
		return Outcome{}, ErrClaimInvalid
	}
	now := commands.clock().UTC().Truncate(time.Microsecond)
	nextID, err := commands.nextID()
	if err != nil || now.IsZero() || nextID == uuid.Nil || nextID.Version() != uuid.Version(7) {
		return Outcome{}, ErrConfiguration
	}
	outcome, err := commands.store.ApplyCycle(ctx, Cycle{
		Claim: claim, CompletedAt: now, SuccessorJobID: nextID,
	})
	if err != nil {
		return Outcome{}, err
	}
	if err = outcome.Validate(claim.TenantID); err != nil {
		return Outcome{}, err
	}
	return outcome, nil
}

func nilStore(store Store) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
