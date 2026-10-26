package maintenance

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
		return nil, ErrCommandInvalid
	}
	return &Commands{store: store, nextID: nextID}, nil
}

func (commands *Commands) Execute(ctx context.Context, claim Claim) (Outcome, error) {
	if commands == nil || ctx == nil || ctx.Err() != nil ||
		claim.Validate() != nil || nilStore(commands.store) || commands.nextID == nil {
		return Outcome{}, ErrCommandInvalid
	}
	nextID, err := commands.nextID()
	if err != nil || !stableID(nextID) || nextID.Version() != uuid.Version(7) {
		return Outcome{}, ErrCommandInvalid
	}
	outcome, err := commands.store.Execute(ctx, claim, nextID)
	if err != nil {
		return Outcome{}, err
	}
	if err := outcome.Validate(claim.Kind); err != nil {
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
