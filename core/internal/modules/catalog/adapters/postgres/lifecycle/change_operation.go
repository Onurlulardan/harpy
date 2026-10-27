package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/update_operation.sql
var updateOperationSQL string

func (storage *store) ChangeOperation(
	ctx context.Context, mutation catalogapp.OperationMutation,
) (catalogdomain.Operation, error) {
	value := mutation.Operation
	if !value.ID.Valid() || !value.Plan.Valid() || mutation.ExpectedRevision < 1 ||
		value.Revision != mutation.ExpectedRevision+1 || mutation.FailureCode != value.FailureCode {
		return catalogdomain.Operation{}, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return catalogdomain.Operation{}, unavailable()
	}
	var destructive any
	if value.DestructiveStarted != nil {
		destructive = *value.DestructiveStarted
	}
	var failure any
	if value.FailureCode != "" {
		failure = value.FailureCode
	}
	operationID, err := uuid.Parse(value.ID.String())
	if err != nil {
		return catalogdomain.Operation{}, catalogapp.ErrInvalid
	}
	stored, _, err := scanOperation(storage.queryRow(ctx, updateOperationSQL,
		operationID, value.State, destructive, failure, value.Revision, mutation.ExpectedRevision))
	if err != nil || stored.ID != value.ID || stored.Plan.SHA256 != value.Plan.SHA256 ||
		stored.Kind != value.Kind || !stored.StartedAt.Equal(value.StartedAt) {
		return catalogdomain.Operation{}, unavailable()
	}
	return stored, nil
}
