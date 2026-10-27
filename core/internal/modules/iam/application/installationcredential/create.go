package installationcredential

import (
	"context"

	"github.com/google/uuid"
)

func Create(
	ctx context.Context,
	binding ActivationBinding,
	store Store,
	command CreateCommand,
) (Receipt, error) {
	if ctx == nil || ctx.Err() != nil || nilInterface(store) ||
		!validCreate(binding, command) {
		return Receipt{}, ErrInvalid
	}
	receipt, err := store.Create(ctx, binding, command)
	if err != nil {
		return Receipt{}, err
	}
	if !receipt.valid(command) {
		return Receipt{}, ErrStorage
	}
	return receipt, nil
}

func validCreate(binding ActivationBinding, command CreateCommand) bool {
	credentialID, err := InitialCredentialID(command.OperationID, command.UserID)
	return err == nil && command.OperationID != uuid.Nil && command.TenantID != uuid.Nil &&
		command.UserID != uuid.Nil && command.CredentialVersion == 1 && command.Credential.Valid() &&
		command.OperationID == binding.OperationID && command.TenantID == binding.TenantID &&
		command.UserID == binding.FirstUserID && command.CredentialID == credentialID &&
		binding.EnvironmentID != uuid.Nil && binding.Kind == "environment_initialize"
}
