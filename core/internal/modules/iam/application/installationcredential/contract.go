package installationcredential

import (
	"context"
	"reflect"

	"github.com/google/uuid"
)

const Purpose = "installation_initial_v1"

type ActivationBinding struct {
	OperationID, EnvironmentID, TenantID, FirstUserID uuid.UUID
	Kind                                              string
}

type CreateCommand struct {
	OperationID, TenantID, UserID, CredentialID uuid.UUID
	CredentialVersion                           int64
	Credential                                  Credential
}

type Receipt struct {
	OperationID, TenantID, UserID, CredentialID uuid.UUID
	CredentialVersion, CredentialRevision       int64
	MarkerRevision                              int64
}

func (receipt Receipt) valid(command CreateCommand) bool {
	return receipt.OperationID == command.OperationID && receipt.TenantID == command.TenantID &&
		receipt.UserID == command.UserID && receipt.CredentialID == command.CredentialID &&
		receipt.CredentialVersion == command.CredentialVersion &&
		receipt.CredentialRevision == 1 && receipt.MarkerRevision == 1
}

type Store interface {
	Create(context.Context, ActivationBinding, CreateCommand) (Receipt, error)
	Rotate(context.Context, RotateCommand) (RotateReceipt, error)
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
