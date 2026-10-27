//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type factFixture struct {
	ID, Environment, Tenant, Actor uuid.UUID
	Source                         [32]byte
	Method, Route                  string
	Status                         int
	Duration                       int64
	Occurred                       time.Time
}

func newFact(environment, tenant, actor uuid.UUID, source, route string,
	occurred time.Time,
) factFixture {
	return factFixture{ID: uuid.Must(uuid.NewV7()), Environment: environment,
		Tenant: tenant, Actor: actor, Source: sha256.Sum256([]byte(source)),
		Method: "GET", Route: route, Status: 200, Duration: 2500, Occurred: occurred}
}

func (fact factFixture) digest(t *testing.T) [32]byte {
	t.Helper()
	observation, err := domain.NormalizeObservation(domain.Observation{
		ID: domain.ID(fact.ID), EnvironmentID: domain.ID(fact.Environment),
		TenantID: domain.ID(fact.Tenant), ActorID: domain.ID(fact.Actor),
		SourceIdentity: fact.Source, Method: fact.Method,
		RoutePattern: fact.Route, StatusCode: fact.Status,
		DurationMicros: fact.Duration, OccurredAt: fact.Occurred})
	if err != nil {
		t.Fatal(err)
	}
	return observation.Digest()
}

func recordFact(t *testing.T, ctx context.Context, database *sql.DB,
	fact factFixture,
) (uuid.UUID, bool, error) {
	t.Helper()
	digest := fact.digest(t)
	var id uuid.UUID
	var replayed bool
	err := database.QueryRowContext(ctx, `SELECT fact_id,replayed FROM platform.usage_record_fact_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, fact.ID, fact.Environment, fact.Tenant,
		fact.Actor, fact.Source[:], digest[:], fact.Method, fact.Route, fact.Status,
		fact.Duration, fact.Occurred).Scan(&id, &replayed)
	return id, replayed, err
}

func assertSQLState(t *testing.T, err error, expected string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != expected {
		t.Fatalf("SQLSTATE=%v, want %s", err, expected)
	}
}
