package installationcredentialpostgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	installationcredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/installationcredential"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed sql/tenant/create.sql
var createSQL string

func (storage *store) Create(
	ctx context.Context,
	binding installationcredential.ActivationBinding,
	command installationcredential.CreateCommand,
) (installationcredential.Receipt, error) {
	if storage == nil || storage.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return installationcredential.Receipt{}, installationcredential.ErrStorage
	}
	salt, tag := command.Credential.Salt(), command.Credential.Tag()
	var receipt installationcredential.Receipt
	err := storage.queryRow(ctx, createSQL,
		command.OperationID, binding.EnvironmentID, command.TenantID,
		command.UserID, command.CredentialID, command.CredentialVersion,
		salt[:], tag[:],
	).Scan(&receipt.OperationID, &receipt.TenantID, &receipt.UserID,
		&receipt.CredentialID, &receipt.CredentialVersion,
		&receipt.CredentialRevision, &receipt.MarkerRevision)
	clear(salt[:])
	clear(tag[:])
	if err != nil {
		return installationcredential.Receipt{}, fmt.Errorf("%w: %s", installationcredential.ErrStorage, sqlFailure(err))
	}
	return receipt, nil
}

func sqlFailure(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err.Error()
	}
	detail := pgErr.Message
	if pgErr.Detail != "" {
		detail += ": " + pgErr.Detail
	}
	return "SQLSTATE " + pgErr.Code + ": " + detail
}
