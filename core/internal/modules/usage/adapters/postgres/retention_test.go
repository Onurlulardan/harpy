package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type retentionRowStub struct {
	deleted int
	err     error
}

func (row retentionRowStub) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	*destinations[0].(*int) = row.deleted
	return nil
}

type retentionDatabaseStub struct {
	statement string
	arguments []any
	row       Row
}

func (database *retentionDatabaseStub) QueryRow(
	_ context.Context,
	statement string,
	arguments ...any,
) Row {
	database.statement = statement
	database.arguments = append([]any(nil), arguments...)
	return database.row
}

type retentionStateError string

func (value retentionStateError) Error() string    { return string(value) }
func (value retentionStateError) SQLState() string { return string(value) }

func TestRetentionStoreUsesOnlySealedRoutineAndCanonicalArguments(t *testing.T) {
	request := retentionRequest(t)
	database := &retentionDatabaseStub{row: retentionRowStub{deleted: 17}}
	store, err := NewRetentionStore(database)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := store.Prune(context.Background(), request)
	if err != nil || deleted != 17 {
		t.Fatalf("retention prune deleted=%d err=%v", deleted, err)
	}
	if database.statement != retentionPruneSQL ||
		!strings.Contains(database.statement, "platform.usage_retention_prune_v1(") ||
		strings.Contains(database.statement, "usage_request_facts") {
		t.Fatalf("retention statement bypassed sealed routine: %q", database.statement)
	}
	if len(database.arguments) != 5 || database.arguments[0] != request.ReceiptID ||
		database.arguments[1] != request.EnvironmentID ||
		database.arguments[2] != request.DeleteBefore ||
		database.arguments[3] != request.BatchLimit ||
		database.arguments[4] != request.ObservedAt {
		t.Fatalf("retention arguments drifted: %#v", database.arguments)
	}
}

func TestRetentionStoreFailsClosedBeforeSQLAndNormalizesConflict(t *testing.T) {
	request := retentionRequest(t)
	database := &retentionDatabaseStub{row: retentionRowStub{
		err: retentionStateError("P6902"),
	}}
	store, _ := NewRetentionStore(database)
	if _, err := store.Prune(context.Background(), request); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("retention conflict error=%v", err)
	}
	database.statement = ""
	request.BatchLimit--
	if _, err := store.Prune(context.Background(), request); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("invalid retention request error=%v", err)
	}
	if database.statement != "" {
		t.Fatal("invalid retention request reached PostgreSQL")
	}
}

func retentionRequest(t *testing.T) application.RetentionPrune {
	t.Helper()
	environmentID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	observedAt := time.Date(2026, 8, 2, 2, 0, 0, 0, time.UTC)
	authority, err := application.NewRetentionAuthority(
		application.RetentionAuthorityInput{EnvironmentID: environmentID,
			ScopeKind: application.RetentionScopeKind, ScopeID: environmentID,
			GroupID: application.RetentionAuthorityGroup, GroupGeneration: 3,
			Domain: application.RetentionAuthorityDomain, DomainGeneration: 5,
			WriterOrigin: application.RetentionWriterOrigin},
	)
	if err != nil {
		t.Fatal(err)
	}
	return application.RetentionPrune{
		ReceiptID:     uuid.MustParse("019c1000-0000-7000-8000-000000000001"),
		EnvironmentID: environmentID,
		DeleteBefore:  observedAt.Add(-application.RetentionAge),
		BatchLimit:    application.RetentionBatchLimit,
		ObservedAt:    observedAt, Authority: authority,
	}
}
