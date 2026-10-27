package installationactivation

import (
	"context"
	_ "embed"
	"errors"
)

var ErrUnavailable = errors.New("installation activation event store unavailable")

//go:embed sql/platform/record_event.sql
var recordEventSQL string

//go:embed sql/platform/enqueue_initial_admin_setup.sql
var enqueueInitialAdminSetupSQL string

type Row interface{ Scan(...any) error }

type Database struct {
	QueryRow func(context.Context, string, ...any) Row
}

type Binding struct {
	queryRow func(context.Context, string, ...any) Row
}

func Bind(database Database) (Binding, error) {
	if database.QueryRow == nil {
		return Binding{}, ErrUnavailable
	}
	return Binding{queryRow: database.QueryRow}, nil
}

func (binding Binding) Record(
	ctx context.Context,
	operationID, environmentID, tenantID, actorID [16]byte,
	schemaName, manifestSHA256 string,
) (bool, error) {
	if binding.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return false, ErrUnavailable
	}
	var replay bool
	err := binding.queryRow(ctx, recordEventSQL,
		operationID, environmentID, tenantID, actorID,
		schemaName, manifestSHA256,
	).Scan(&replay)
	return replay, err
}

func (binding Binding) EnqueueInitialAdminSetup(
	ctx context.Context,
	operationID, environmentID, tenantID, artifactID [16]byte,
	artifactReference, schemaName, manifestSHA256 string,
) (bool, error) {
	if binding.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return false, ErrUnavailable
	}
	var replay bool
	err := binding.queryRow(ctx, enqueueInitialAdminSetupSQL,
		operationID, environmentID, tenantID, artifactID, artifactReference,
		schemaName, manifestSHA256,
	).Scan(&replay)
	return replay, err
}
