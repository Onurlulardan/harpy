package providerapp

import (
	"context"
	"errors"
	"testing"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type commandStoreStub struct{ called bool }

func (store *commandStoreStub) CreateCommand(
	context.Context,
	CommandMutation,
) (CommandSnapshot, error) {
	store.called = true
	return CommandSnapshot{}, nil
}

func TestCreateCommandRejectsUnverifiedTenantScopeBeforeStorage(t *testing.T) {
	store := &commandStoreStub{}
	commands, err := NewCommandCommands(store, uuid.NewV7, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = commands.CreateCommand(
		context.Background(), platformtx.TenantScope{}, SubmitCommandRequest{
			SubscriptionID: uuid.New(), Kind: CommandCheckout,
			EffectKey: "checkout:tenant:one", RequestSHA256: [32]byte{1},
		},
	)
	if !errors.Is(err, ErrCommandRequest) || store.called {
		t.Fatalf("CreateCommand() error = %v, store called = %v", err, store.called)
	}
}
