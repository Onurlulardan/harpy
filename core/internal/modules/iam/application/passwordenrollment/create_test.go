package passwordenrollment

import (
	"context"
	"testing"
)

type nilStore struct{}

func (*nilStore) Create(context.Context, TenantScope, Command) (MutationResult, error) {
	return MutationResult{}, nil
}

func TestNewCommandsRejectsTypedNilStore(t *testing.T) {
	var store *nilStore
	if _, err := NewCommands(store); err != ErrStorage {
		t.Fatalf("NewCommands() error=%v", err)
	}
}

func TestCredentialOwnsCopies(t *testing.T) {
	var salt [16]byte
	var tag [32]byte
	salt[0], tag[0] = 1, 2
	credential, err := NewCredential(salt, tag)
	if err != nil {
		t.Fatal(err)
	}
	salt[0], tag[0] = 9, 9
	if credential.Salt()[0] != 1 || credential.Tag()[0] != 2 {
		t.Fatal("credential did not retain owned values")
	}
}

func TestCreateRejectsInvalidCommand(t *testing.T) {
	commands, err := NewCommands(&nilStore{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := commands.Create(
		context.Background(), TenantScope{}, Command{},
	); err != ErrInvalidCommand {
		t.Fatalf("Create() error=%v", err)
	}
}
