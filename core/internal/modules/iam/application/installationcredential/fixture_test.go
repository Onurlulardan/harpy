package installationcredential

import (
	"context"
	"testing"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

type memoryStore struct {
	create func(ActivationBinding, CreateCommand) (Receipt, error)
	rotate func(RotateCommand) (RotateReceipt, error)
}

func (storage *memoryStore) Create(
	_ context.Context,
	binding ActivationBinding,
	command CreateCommand,
) (Receipt, error) {
	return storage.create(binding, command)
}

func (storage *memoryStore) Rotate(
	_ context.Context,
	command RotateCommand,
) (RotateReceipt, error) {
	return storage.rotate(command)
}

func credentialFixture(t *testing.T) Credential {
	t.Helper()
	credential, err := NewCredential([16]byte{1}, [32]byte{2})
	if err != nil {
		t.Fatal(err)
	}
	return credential
}

func activationBindingFixture() ActivationBinding {
	return ActivationBinding{
		OperationID: uuid.New(), EnvironmentID: uuid.New(),
		TenantID: uuid.New(), FirstUserID: uuid.New(), Kind: "environment_initialize",
	}
}

func rotationBindingFixture(t *testing.T) RotationBinding {
	t.Helper()
	operationID, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return RotationBinding{
		OperationID: operationID, TenantID: uuid.New(), FirstOwnerUserID: uuid.New(),
		CurrentCredentialID: uuid.New(), CurrentCredentialVersion: 1,
		NextCredentialID: uuid.New(), NextCredentialVersion: 2,
		CurrentAttemptID: uuid.New(), NextAttemptID: uuid.New(),
		Cause:          credentialdeliveryv1.CauseOutcomeUnknown,
		ManifestSHA256: credentialdeliveryv1.ManifestSHA256(),
	}
}
