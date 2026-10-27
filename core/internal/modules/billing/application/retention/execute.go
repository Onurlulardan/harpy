package retention

import (
	"context"
	"reflect"

	"github.com/google/uuid"
)

type Commands struct {
	store  Store
	nextID IDSource
}

func NewCommands(store Store, nextID IDSource) (*Commands, error) {
	if nilStore(store) || nextID == nil {
		return nil, ErrConfiguration
	}
	return &Commands{store: store, nextID: nextID}, nil
}

func (commands *Commands) Execute(ctx context.Context, claim Claim) (Outcome, error) {
	if commands == nil || ctx == nil || ctx.Err() != nil ||
		claim.validate() != nil || nilStore(commands.store) || commands.nextID == nil {
		return Outcome{}, ErrClaimInvalid
	}
	nextID, err := commands.nextID()
	if err != nil || nextID == uuid.Nil || nextID.Version() != uuid.Version(7) {
		return Outcome{}, ErrConfiguration
	}
	outcome, err := commands.store.Execute(ctx, claim, nextID)
	if err != nil {
		return Outcome{}, err
	}
	if err := outcome.Validate(claim.TenantID); err != nil {
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
