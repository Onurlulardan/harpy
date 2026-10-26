package bulkdeletepostgres

import (
	"context"
	_ "embed"
	"errors"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

//go:embed sql/tenant/reserve.sql
var reserveSQL string

func (storage *store) ReserveBulkDelete(
	ctx context.Context,
	scope tenantScope,
	command identitydomain.BulkDeleteReceiptCommand,
) (bool, error) {
	if command.Validate() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != uuid.UUID(command.ActorID) {
		return false, identitydomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return false, err
	}
	var replayed bool
	err = bound.database.QueryRow(
		ctx, reserveSQL, uuid.UUID(command.OperationID), uuid.UUID(command.ActorID),
		command.ResourceKind, command.RequestSHA256[:], command.DeletedCount,
	).Scan(&replayed)
	if err != nil {
		var state interface{ SQLState() string }
		if errors.As(err, &state) && state.SQLState() == "23505" {
			return false, identitydomain.ErrBulkDeleteConflict
		}
		return false, identitydomain.ErrStorage
	}
	return replayed, nil
}
