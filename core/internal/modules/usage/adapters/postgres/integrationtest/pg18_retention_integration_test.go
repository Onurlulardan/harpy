//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/google/uuid"
)

type retentionIntegrationIDs struct{ value uuid.UUID }

func (ids retentionIntegrationIDs) NewV7() (uuid.UUID, error) { return ids.value, nil }

type retentionIntegrationDatabase struct{ connection *sql.Conn }

func (database retentionIntegrationDatabase) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) usage.PostgresRow {
	return database.connection.QueryRowContext(ctx, statement, arguments...)
}

func TestUsageRetentionAndLegalHoldOnFreshPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startUsageDatabase(t, ctx)
	resetAndApplyUsage(t, ctx, database)
	exerciseRetentionHold(t, ctx, database)
}

func exerciseRetentionHold(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC)
	environment, tenantA, tenantB, actor := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	old := now.Add(-500 * 24 * time.Hour)
	if _, _, err := recordFact(t, ctx, database,
		newFact(environment, tenantA, actor, "held-fact", "/api/forms/{id}", old)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := recordFact(t, ctx, database,
		newFact(environment, tenantB, actor, "deletable-fact", "/api/workflow/{id}", old)); err != nil {
		t.Fatal(err)
	}
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `SET ROLE harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	holdID := uuid.Must(uuid.NewV7())
	var placed bool
	err = connection.QueryRowContext(ctx, `SELECT platform.usage_legal_hold_place_v1(
$1,$2,$3,'legal.case',$4,NULL,$5)`, holdID, environment, tenantA,
		now.Add(-90*24*time.Hour), now).Scan(&placed)
	if err != nil || !placed {
		t.Fatalf("place hold=%v err=%v", placed, err)
	}
	deleteBefore := now.Add(-usage.RetentionAge)
	receiptOne := uuid.Must(uuid.NewV7())
	firstClaim := retentionIntegrationClaim(t, environment, receiptOne, now)
	commands := retentionIntegrationCommands(t, connection)
	first, err := commands.Execute(ctx, firstClaim)
	if err != nil || first.DeletedCount != 1 {
		t.Fatalf("held prune outcome=%+v err=%v", first, err)
	}
	replay, err := commands.Execute(ctx, firstClaim)
	if err != nil || replay.DeletedCount != first.DeletedCount {
		t.Fatalf("held prune replay=%+v first=%+v err=%v", replay, first, err)
	}
	_, err = connection.ExecContext(ctx, `SELECT hold_id FROM platform.usage_legal_holds`)
	assertSQLState(t, err, "42501")
	var released bool
	err = connection.QueryRowContext(ctx, `SELECT platform.usage_legal_hold_release_v1($1,$2)`,
		holdID, now.Add(time.Microsecond)).Scan(&released)
	if err != nil || !released {
		t.Fatalf("release hold=%v err=%v", released, err)
	}
	receiptTwo := uuid.Must(uuid.NewV7())
	secondClaim := retentionIntegrationClaim(t, environment, receiptTwo,
		now.Add(2*time.Microsecond))
	second, err := commands.Execute(ctx, secondClaim)
	if err != nil || second.DeletedCount != 1 {
		t.Fatalf("released prune outcome=%+v err=%v", second, err)
	}
	var deleted int
	err = connection.QueryRowContext(ctx, `SELECT platform.usage_retention_prune_v1($1,$2,$3,100,$4)`,
		receiptTwo, environment, deleteBefore.Add(-time.Hour), now.Add(3*time.Microsecond)).Scan(&deleted)
	assertSQLState(t, err, "P6902")
	if _, err = connection.ExecContext(ctx, `RESET ROLE; SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	err = connection.QueryRowContext(ctx, `SELECT platform.usage_retention_prune_v1($1,$2,$3,1,$4)`,
		uuid.Must(uuid.NewV7()), environment, deleteBefore, now).Scan(&deleted)
	assertSQLState(t, err, "42501")
	if _, err = connection.ExecContext(ctx, `RESET ROLE`); err != nil {
		t.Fatal(err)
	}
	var remaining int
	err = database.QueryRowContext(ctx, `SELECT count(*) FROM platform.usage_request_facts
WHERE environment_id=$1`, environment).Scan(&remaining)
	if err != nil || remaining != 0 {
		t.Fatalf("retention remaining=%d err=%v", remaining, err)
	}
	immutable := newFact(environment, tenantA, actor, "immutable-fact",
		"/api/forms/{id}", now)
	if _, _, err = recordFact(t, ctx, database, immutable); err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx,
		`DELETE FROM platform.usage_request_facts WHERE fact_id=$1`, immutable.ID)
	assertSQLState(t, err, "P6903")
	assertBoundedRetention(t, ctx, database, environment, now)
}

func retentionIntegrationCommands(
	t *testing.T,
	connection *sql.Conn,
) *usage.RetentionCommands {
	t.Helper()
	nextID := uuid.Must(uuid.NewV7())
	binding, err := usage.NewRetentionJobBinding(retentionIntegrationIDs{nextID})
	if err != nil {
		t.Fatal(err)
	}
	commands, err := binding.Bind(retentionIntegrationDatabase{connection})
	if err != nil {
		t.Fatal(err)
	}
	return commands
}

func retentionIntegrationClaim(t *testing.T, environmentID, jobID uuid.UUID,
	cycleAt time.Time,
) usage.RetentionJobClaim {
	t.Helper()
	authority, err := usage.NewRetentionAuthority(usage.RetentionAuthorityInput{
		EnvironmentID: environmentID, ScopeKind: usage.RetentionScopeKind,
		ScopeID: environmentID, GroupID: usage.RetentionAuthorityGroup,
		GroupGeneration: 1, Domain: usage.RetentionAuthorityDomain,
		DomainGeneration: 1, WriterOrigin: usage.RetentionWriterOrigin,
	})
	if err != nil {
		t.Fatal(err)
	}
	reference, digest, err := usage.RetentionReference(environmentID, cycleAt)
	if err != nil {
		t.Fatal(err)
	}
	return usage.RetentionJobClaim{JobID: jobID, AttemptID: uuid.Must(uuid.NewV7()),
		Authority: authority, Kind: usage.RetentionJobKind,
		PayloadRevision: usage.RetentionPayloadRevision, PayloadReference: reference,
		PayloadSHA256: digest, LeaseGeneration: 1, StartedAt: cycleAt}
}

func assertBoundedRetention(t *testing.T, ctx context.Context, database *sql.DB,
	environment uuid.UUID, now time.Time,
) {
	t.Helper()
	connection, _ := database.Conn(ctx)
	defer func() {
		_, _ = connection.ExecContext(ctx, `RESET ROLE`)
		_ = connection.Close()
	}()
	_, _ = connection.ExecContext(ctx, `SET ROLE harpy_adopter`)
	var deleted int
	err := connection.QueryRowContext(ctx, `SELECT platform.usage_retention_prune_v1($1,$2,$3,1,$4)`,
		uuid.Must(uuid.NewV7()), environment, now.Add(-24*time.Hour), now).Scan(&deleted)
	assertSQLState(t, err, "P6901")
}
