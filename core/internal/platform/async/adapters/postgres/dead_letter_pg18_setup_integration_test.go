//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deadLetterPGFixture struct {
	owner      *sql.DB
	runtime    *pgxpool.Pool
	repository *deadLetterRepository
	consumers  eventsv1.ConsumerCatalog
}

func newDeadLetterPGFixture(
	t *testing.T,
	ctx context.Context,
	origin string,
) deadLetterPGFixture {
	t.Helper()
	owner := startAdoptionPG18(t, ctx, "dead_letter_"+origin)
	applyAsyncACLFixture(t, ctx, owner, origin)
	addDeadLetterAuthorities(t, ctx, owner)
	grantDeadLetterRuntime(t, ctx, owner)
	runtime := openPublisherRuntimePool(t, ctx, owner)
	consumers := deadLetterConsumerCatalog(t)
	return deadLetterPGFixture{
		owner: owner, runtime: runtime, consumers: consumers,
		repository: &deadLetterRepository{
			database: runtime, cleanup: 5 * time.Second, consumers: consumers,
		},
	}
}

func addDeadLetterAuthorities(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	addPublisherSecondTenant(t, ctx, database)
	const statement = `
INSERT INTO platform.cutover_groups VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'billing_forms','v2','v2',1),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'billing_forms','v2','v2',1);
INSERT INTO platform.domain_cutovers VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'forms','v2','v2',1,NULL,pg_catalog.clock_timestamp()),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'forms','v2','v2',1,NULL,pg_catalog.clock_timestamp());
INSERT INTO platform.cutover_group_members VALUES
    ('tenant','11111111-1111-4111-8111-111111111111',
     'billing_forms','forms'),
    ('tenant','22222222-2222-4222-8222-222222222222',
     'billing_forms','forms');`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("add dead-letter authorities: %v", err)
	}
}

func grantDeadLetterRuntime(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
GRANT INSERT (
    lease_owner_id,lease_heartbeat_at,lease_until,created_at,updated_at
) ON platform.jobs TO harpy_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("grant dead-letter runtime: %v", err)
	}
}

func deadLetterConsumerCatalog(t *testing.T) eventsv1.ConsumerCatalog {
	t.Helper()
	policy, err := eventsv1.NewConsumerPolicy(
		"workflow.projector", 1, "workflow", "forms.entry.created", 1,
		"workflow.projector.v1", eventsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := eventsv1.NewConsumerCatalog([]eventsv1.ConsumerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

type deadLetterPermissionFixture struct {
	allowed  map[uuid.UUID]struct{}
	platform bool
}

func (fixture deadLetterPermissionFixture) Verify(
	_ context.Context,
	check deadletter.PermissionCheck,
) error {
	if check.Platform != fixture.platform ||
		(check.Permission != deadletter.PermissionRead &&
			check.Permission != deadletter.PermissionRetry) {
		return deadletter.ErrUnauthorized
	}
	for _, tenantID := range check.TenantIDs {
		if _, exists := fixture.allowed[tenantID]; !exists {
			return deadletter.ErrUnauthorized
		}
	}
	return nil
}

type deadLetterFixedID struct{ value uuid.UUID }

func (source deadLetterFixedID) NewUUID() uuid.UUID { return source.value }
