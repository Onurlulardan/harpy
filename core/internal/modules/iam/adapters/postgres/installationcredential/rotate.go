package installationcredentialpostgres

import (
	"context"
	_ "embed"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
)

//go:embed sql/tenant/rotate.sql
var rotateSQL string

func (storage *store) Rotate(
	ctx context.Context,
	command installationcredential.RotateCommand,
) (installationcredential.RotateReceipt, error) {
	if storage == nil || storage.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return installationcredential.RotateReceipt{}, installationcredential.ErrStorage
	}
	binding := command.Binding
	salt, tag := command.Credential.Salt(), command.Credential.Tag()
	var receipt installationcredential.RotateReceipt
	err := storage.queryRow(ctx, rotateSQL, binding.OperationID, binding.TenantID,
		binding.FirstOwnerUserID, binding.CurrentCredentialID,
		binding.CurrentCredentialVersion, binding.NextCredentialID,
		binding.NextCredentialVersion, salt[:], tag[:],
	).Scan(&receipt.OperationID, &receipt.TenantID, &receipt.UserID,
		&receipt.PreviousCredentialID, &receipt.CredentialID,
		&receipt.CredentialVersion, &receipt.CredentialRevision,
		&receipt.PreviousMarkerRevision, &receipt.MarkerRevision)
	clear(salt[:])
	clear(tag[:])
	if err != nil {
		return installationcredential.RotateReceipt{}, installationcredential.ErrStorage
	}
	return receipt, nil
}
