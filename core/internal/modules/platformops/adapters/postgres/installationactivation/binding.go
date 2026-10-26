package installationactivation

import (
	"context"
	_ "embed"
	"errors"
)

var ErrUnavailable = errors.New("platform operations installation activation unavailable")

//go:embed sql/platform/initialize.sql
var initializeSQL string

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

func (binding Binding) Initialize(
	ctx context.Context,
	operationID, environmentID, tenantID [16]byte,
	schemaName, mode string,
	generation any,
	trustBundleSHA256 any,
	manifestSHA256 string,
) (bool, error) {
	if binding.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return false, ErrUnavailable
	}
	var replay bool
	err := binding.queryRow(ctx, initializeSQL,
		operationID, environmentID, tenantID, schemaName, mode,
		generation, trustBundleSHA256, manifestSHA256,
	).Scan(&replay)
	return replay, err
}
