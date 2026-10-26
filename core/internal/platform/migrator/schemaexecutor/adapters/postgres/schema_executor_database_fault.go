package postgres

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgconn"
)

func schemaExecutorClassifyError(err error) error {
	if err == nil {
		return nil
	}
	kind, _ := schemaexecutor.FaultDetails(err)
	if kind != schemaexecutor.FaultTransient {
		return err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return schemaexecutor.NewFault(
			schemaexecutor.FaultTransient, "schema_operation_interrupted", err,
		)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return schemaexecutor.NewFault(
			schemaexecutor.FaultTransient, "schema_database_unavailable", err,
		)
	}
	if len(postgresError.Code) < 2 {
		return schemaexecutor.NewFault(
			schemaexecutor.FaultPermanent, "schema_database_permanent", err,
		)
	}
	class := postgresError.Code[:2]
	switch class {
	case "08", "40", "53", "57", "58":
		return schemaexecutor.NewFault(
			schemaexecutor.FaultTransient, "schema_database_transient", err,
		)
	case "22", "23", "42":
		return schemaexecutor.NewFault(
			schemaexecutor.FaultBlocked, "schema_precondition_failed", err,
		)
	default:
		return schemaexecutor.NewFault(
			schemaexecutor.FaultPermanent, "schema_database_permanent", err,
		)
	}
}
