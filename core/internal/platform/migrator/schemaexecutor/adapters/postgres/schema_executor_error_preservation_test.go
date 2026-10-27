package postgres

import (
	"context"
	"errors"
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type schemaExecutorErrorRow struct{ err error }

func (row schemaExecutorErrorRow) Scan(...any) error { return row.err }

type schemaExecutorErrorQuerier struct{ err error }

func (querier schemaExecutorErrorQuerier) QueryRow(
	context.Context, string, ...any,
) pgx.Row {
	return schemaExecutorErrorRow{err: querier.err}
}

func TestCatalogOperationalErrorsRemainTransient(t *testing.T) {
	const tenantID = "00000000-0000-4000-8000-000000000001"
	cause := context.DeadlineExceeded
	querier := schemaExecutorErrorQuerier{err: cause}
	if _, err := schemaExecutorResolveTenant(
		context.Background(), querier, tenantID,
	); !errors.Is(err, cause) {
		t.Fatalf("resolve tenant error = %v", err)
	}
	err := schemaExecutorVerifyIndexes(
		context.Background(), querier, "tenant_example",
		schemav1.IdentifierMapping{Indexes: []schemav1.Identifier{{
			LogicalID: "x", PhysicalName: "idx_example",
		}}}, []schemav1.VerifyIndexContract{{IndexID: "x"}},
	)
	if !errors.Is(err, cause) {
		t.Fatalf("verify index error = %v", err)
	}
	classified := schemaExecutorClassifyError(err)
	kind, _ := schemaexecutor.FaultDetails(classified)
	if kind != schemaexecutor.FaultTransient {
		t.Fatalf("classified kind = %s", kind)
	}
}

func TestClassifyShortSQLStateDoesNotPanic(t *testing.T) {
	err := schemaExecutorClassifyError(&pgconn.PgError{Code: "X"})
	kind, _ := schemaexecutor.FaultDetails(err)
	if kind != schemaexecutor.FaultPermanent {
		t.Fatalf("classified kind = %s", kind)
	}
}
