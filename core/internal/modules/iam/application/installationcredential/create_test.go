package installationcredential

import (
	"context"
	"testing"
)

func TestCreateLocksCredentialToActivationIdentity(t *testing.T) {
	binding := activationBindingFixture()
	credentialID, err := InitialCredentialID(binding.OperationID, binding.FirstUserID)
	if err != nil {
		t.Fatal(err)
	}
	command := CreateCommand{
		OperationID: binding.OperationID, TenantID: binding.TenantID,
		UserID: binding.FirstUserID, CredentialID: credentialID,
		CredentialVersion: 1, Credential: credentialFixture(t),
	}
	store := &memoryStore{create: func(
		got ActivationBinding, gotCommand CreateCommand,
	) (Receipt, error) {
		if got != binding || gotCommand != command {
			t.Fatal("activation binding drifted")
		}
		return Receipt{
			OperationID: command.OperationID, TenantID: command.TenantID,
			UserID: command.UserID, CredentialID: command.CredentialID,
			CredentialVersion: 1, CredentialRevision: 1, MarkerRevision: 1,
		}, nil
	}}
	if _, err = Create(context.Background(), binding, store, command); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRejectsWrongFirstUser(t *testing.T) {
	binding := activationBindingFixture()
	command := CreateCommand{
		OperationID: binding.OperationID, TenantID: binding.TenantID,
		UserID: binding.TenantID, CredentialID: [16]byte{9},
		CredentialVersion: 1, Credential: credentialFixture(t),
	}
	called := false
	store := &memoryStore{create: func(
		ActivationBinding, CreateCommand,
	) (Receipt, error) {
		called = true
		return Receipt{}, nil
	}}
	if _, err := Create(context.Background(), binding, store, command); err != ErrInvalid ||
		called {
		t.Fatalf("wrong user error=%v called=%t", err, called)
	}
}
